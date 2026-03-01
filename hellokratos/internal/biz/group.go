package biz

import (
	"context"
	"errors"
	"hellokratos/internal/data"
	"hellokratos/internal/data/model"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"
)

// GroupUsecase 群组相关的业务逻辑接口
type GroupUsecase interface {
	// CreateGroup 创建群组
	CreateGroup(ctx context.Context, creatorID int64, name string, description string, memberIDs []int64) (*model.Group, error)
	// JoinGroup 加入群组
	JoinGroup(ctx context.Context, userID int64, groupID int64) error
	// LeaveGroup 退出群组
	LeaveGroup(ctx context.Context, userID int64, groupID int64) error
	// GetGroupList 获取用户加入的群组列表
	GetGroupList(ctx context.Context, userID int64, limit int) ([]*model.Group, error)
	// GetGroupMembers 获取群组成员列表
	GetGroupMembers(ctx context.Context, groupID int64, limit int) ([]*model.GroupMember, error)
	// GetGroupByID 根据ID获取群组
	GetGroupByID(ctx context.Context, groupID int64) (*model.Group, error)

	// SendGroupMessage 发送群聊消息
	SendGroupMessage(ctx context.Context, fromID int64, groupID int64, content string, messageType int32) (*model.GroupMessage, error)
	// GetGroupMessages 获取群聊消息列表
	GetGroupMessages(ctx context.Context, groupID int64, lastID int64, limit int) ([]*model.GroupMessage, error)
	// GetLatestGroupMessages 获取最新的群聊消息（用于WebSocket推送）
	GetLatestGroupMessages(ctx context.Context, groupID int64, lastID int64, limit int) ([]*model.GroupMessage, error)
	// IsGroupMember 检查用户是否是群组成员
	IsGroupMember(ctx context.Context, userID int64, groupID int64) (bool, error)
	// GetUserGroupIDs 获取用户加入的所有群组ID
	GetUserGroupIDs(ctx context.Context, userID int64) ([]int64, error)
}

// groupUsecase 群组相关的业务逻辑实现
type groupUsecase struct {
	groupRepo data.GroupRepo
	rdb       *redis.Client
	log       *log.Helper
}

// NewGroupUsecase 创建群组业务逻辑实例
func NewGroupUsecase(groupRepo data.GroupRepo, rdb *redis.Client, logger log.Logger) GroupUsecase {
	return &groupUsecase{
		groupRepo: groupRepo,
		rdb:       rdb,
		log:       log.NewHelper(logger),
	}
}

// CreateGroup 创建群组
func (uc *groupUsecase) CreateGroup(ctx context.Context, creatorID int64, name string, description string, memberIDs []int64) (*model.Group, error) {
	// 创建群组
	group := &model.Group{
		Name:        name,
		Description: description,
		CreatorID:   creatorID,
		MemberCount: 1, // 创建者默认加入
	}

	err := uc.groupRepo.CreateGroup(ctx, group)
	if err != nil {
		uc.log.Error("failed to create group", "err", err)
		return nil, err
	}

	// 添加创建者为群主
	creatorMember := &model.GroupMember{
		GroupID:  group.ID,
		UserID:   creatorID,
		Role:     2, // 群主
		JoinedAt: time.Now(),
	}
	err = uc.groupRepo.AddGroupMember(ctx, creatorMember)
	if err != nil {
		uc.log.Error("failed to add creator to group", "err", err)
		return nil, err
	}

	// 添加其他初始成员
	for _, memberID := range memberIDs {
		if memberID == creatorID {
			continue // 跳过创建者
		}
		member := &model.GroupMember{
			GroupID:  group.ID,
			UserID:   memberID,
			Role:     0, // 普通成员
			JoinedAt: time.Now(),
		}
		err = uc.groupRepo.AddGroupMember(ctx, member)
		if err != nil {
			uc.log.Error("failed to add member to group", "err", err, "member_id", memberID)
			// 继续添加其他成员
			continue
		}
		group.MemberCount++
	}

	// 更新成员数量
	err = uc.groupRepo.UpdateMemberCount(ctx, group.ID, group.MemberCount)
	if err != nil {
		uc.log.Error("failed to update member count", "err", err)
	}

	return group, nil
}

// JoinGroup 加入群组
func (uc *groupUsecase) JoinGroup(ctx context.Context, userID int64, groupID int64) error {
	// 检查是否已经是成员
	_, err := uc.groupRepo.GetGroupMember(ctx, groupID, userID)
	if err == nil {
		return errors.New("已经是群组成员")
	}

	// 添加成员
	member := &model.GroupMember{
		GroupID:  groupID,
		UserID:   userID,
		Role:     0, // 普通成员
		JoinedAt: time.Now(),
	}
	err = uc.groupRepo.AddGroupMember(ctx, member)
	if err != nil {
		uc.log.Error("failed to join group", "err", err)
		return err
	}

	// 更新成员数量
	group, err := uc.groupRepo.GetGroupByID(ctx, groupID)
	if err != nil {
		return err
	}
	err = uc.groupRepo.UpdateMemberCount(ctx, groupID, group.MemberCount+1)
	if err != nil {
		uc.log.Error("failed to update member count", "err", err)
	}

	return nil
}

// LeaveGroup 退出群组
func (uc *groupUsecase) LeaveGroup(ctx context.Context, userID int64, groupID int64) error {
	// 检查是否是成员
	member, err := uc.groupRepo.GetGroupMember(ctx, groupID, userID)
	if err != nil {
		return errors.New("不是群组成员")
	}

	// 如果是群主，不能退出（需要先转让群主）
	if member.Role == 2 {
		return errors.New("群主不能退出群组，请先转让群主")
	}

	// 移除成员
	err = uc.groupRepo.RemoveGroupMember(ctx, groupID, userID)
	if err != nil {
		uc.log.Error("failed to leave group", "err", err)
		return err
	}

	// 更新成员数量
	group, err := uc.groupRepo.GetGroupByID(ctx, groupID)
	if err != nil {
		return err
	}
	if group.MemberCount > 0 {
		err = uc.groupRepo.UpdateMemberCount(ctx, groupID, group.MemberCount-1)
		if err != nil {
			uc.log.Error("failed to update member count", "err", err)
		}
	}

	return nil
}

// GetGroupList 获取用户加入的群组列表
func (uc *groupUsecase) GetGroupList(ctx context.Context, userID int64, limit int) ([]*model.Group, error) {
	return uc.groupRepo.GetGroupsByUserID(ctx, userID, limit)
}

// GetGroupMembers 获取群组成员列表
func (uc *groupUsecase) GetGroupMembers(ctx context.Context, groupID int64, limit int) ([]*model.GroupMember, error) {
	return uc.groupRepo.GetGroupMembers(ctx, groupID, limit)
}

// GetGroupByID 根据ID获取群组
func (uc *groupUsecase) GetGroupByID(ctx context.Context, groupID int64) (*model.Group, error) {
	return uc.groupRepo.GetGroupByID(ctx, groupID)
}

// SendGroupMessage 发送群聊消息
func (uc *groupUsecase) SendGroupMessage(ctx context.Context, fromID int64, groupID int64, content string, messageType int32) (*model.GroupMessage, error) {
	// 检查是否是群组成员
	isMember, err := uc.IsGroupMember(ctx, fromID, groupID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, errors.New("不是群组成员，无法发送消息")
	}

	// 创建群聊消息
	message := &model.GroupMessage{
		GroupID:  groupID,
		FromID:   fromID,
		Content:  content,
		Type:     messageType,
		Nickname: "", // 可以从用户信息中获取昵称
	}

	err = uc.groupRepo.CreateGroupMessage(ctx, message)
	if err != nil {
		uc.log.Error("failed to send group message", "err", err)
		return nil, err
	}

	uc.log.Info("group message sent", "group_id", groupID, "from_id", fromID)
	return message, nil
}

// GetGroupMessages 获取群聊消息列表
func (uc *groupUsecase) GetGroupMessages(ctx context.Context, groupID int64, lastID int64, limit int) ([]*model.GroupMessage, error) {
	return uc.groupRepo.GetGroupMessages(ctx, groupID, lastID, limit)
}

// GetLatestGroupMessages 获取最新的群聊消息（用于WebSocket推送）
func (uc *groupUsecase) GetLatestGroupMessages(ctx context.Context, groupID int64, lastID int64, limit int) ([]*model.GroupMessage, error) {
	return uc.groupRepo.GetLatestGroupMessages(ctx, groupID, lastID, limit)
}

// IsGroupMember 检查用户是否是群组成员
func (uc *groupUsecase) IsGroupMember(ctx context.Context, userID int64, groupID int64) (bool, error) {
	_, err := uc.groupRepo.GetGroupMember(ctx, groupID, userID)
	if err != nil {
		return false, nil
	}
	return true, nil
}

// GetUserGroupIDs 获取用户加入的所有群组ID
func (uc *groupUsecase) GetUserGroupIDs(ctx context.Context, userID int64) ([]int64, error) {
	groups, err := uc.groupRepo.GetGroupsByUserID(ctx, userID, 1000)
	if err != nil {
		return nil, err
	}

	groupIDs := make([]int64, 0, len(groups))
	for _, group := range groups {
		groupIDs = append(groupIDs, group.ID)
	}
	return groupIDs, nil
}
