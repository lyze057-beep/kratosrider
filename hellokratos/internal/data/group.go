package data

import (
	"context"
	"hellokratos/internal/data/model"

	"gorm.io/gorm"
)

// GroupRepo 群组相关的数据访问接口
type GroupRepo interface {
	// CreateGroup 创建群组
	CreateGroup(ctx context.Context, group *model.Group) error
	// GetGroupByID 根据ID获取群组
	GetGroupByID(ctx context.Context, id int64) (*model.Group, error)
	// GetGroupsByUserID 获取用户加入的群组列表
	GetGroupsByUserID(ctx context.Context, userID int64, limit int) ([]*model.Group, error)
	// UpdateGroup 更新群组信息
	UpdateGroup(ctx context.Context, group *model.Group) error
	// DeleteGroup 删除群组
	DeleteGroup(ctx context.Context, id int64) error

	// AddGroupMember 添加群组成员
	AddGroupMember(ctx context.Context, member *model.GroupMember) error
	// GetGroupMembers 获取群组成员列表
	GetGroupMembers(ctx context.Context, groupID int64, limit int) ([]*model.GroupMember, error)
	// GetGroupMember 获取特定成员信息
	GetGroupMember(ctx context.Context, groupID int64, userID int64) (*model.GroupMember, error)
	// RemoveGroupMember 移除群组成员
	RemoveGroupMember(ctx context.Context, groupID int64, userID int64) error
	// UpdateMemberCount 更新群组成员数量
	UpdateMemberCount(ctx context.Context, groupID int64, count int32) error

	// CreateGroupMessage 创建群聊消息
	CreateGroupMessage(ctx context.Context, message *model.GroupMessage) error
	// GetGroupMessages 获取群聊消息列表
	GetGroupMessages(ctx context.Context, groupID int64, lastID int64, limit int) ([]*model.GroupMessage, error)
	// GetLatestGroupMessages 获取最新的群聊消息（用于WebSocket推送）
	GetLatestGroupMessages(ctx context.Context, groupID int64, lastID int64, limit int) ([]*model.GroupMessage, error)
}

// groupRepo 群组相关的数据访问实现
type groupRepo struct {
	db *gorm.DB
}

// NewGroupRepo 创建群组数据访问实例
func NewGroupRepo(data *Data) GroupRepo {
	return &groupRepo{db: data.db}
}

// CreateGroup 创建群组
func (r *groupRepo) CreateGroup(ctx context.Context, group *model.Group) error {
	return r.db.WithContext(ctx).Create(group).Error
}

// GetGroupByID 根据ID获取群组
func (r *groupRepo) GetGroupByID(ctx context.Context, id int64) (*model.Group, error) {
	var group model.Group
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&group).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

// GetGroupsByUserID 获取用户加入的群组列表
func (r *groupRepo) GetGroupsByUserID(ctx context.Context, userID int64, limit int) ([]*model.Group, error) {
	var groups []*model.Group
	err := r.db.WithContext(ctx).
		Table("rider_group").
		Select("rider_group.*").
		Joins("JOIN rider_group_member ON rider_group.id = rider_group_member.group_id").
		Where("rider_group_member.user_id = ?", userID).
		Order("rider_group.created_at desc").
		Limit(limit).
		Find(&groups).Error
	if err != nil {
		return nil, err
	}
	return groups, nil
}

// UpdateGroup 更新群组信息
func (r *groupRepo) UpdateGroup(ctx context.Context, group *model.Group) error {
	return r.db.WithContext(ctx).Save(group).Error
}

// DeleteGroup 删除群组
func (r *groupRepo) DeleteGroup(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&model.Group{}, id).Error
}

// AddGroupMember 添加群组成员
func (r *groupRepo) AddGroupMember(ctx context.Context, member *model.GroupMember) error {
	return r.db.WithContext(ctx).Create(member).Error
}

// GetGroupMembers 获取群组成员列表
func (r *groupRepo) GetGroupMembers(ctx context.Context, groupID int64, limit int) ([]*model.GroupMember, error) {
	var members []*model.GroupMember
	err := r.db.WithContext(ctx).
		Where("group_id = ?", groupID).
		Order("role desc, joined_at asc").
		Limit(limit).
		Find(&members).Error
	if err != nil {
		return nil, err
	}
	return members, nil
}

// GetGroupMember 获取特定成员信息
func (r *groupRepo) GetGroupMember(ctx context.Context, groupID int64, userID int64) (*model.GroupMember, error) {
	var member model.GroupMember
	err := r.db.WithContext(ctx).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		First(&member).Error
	if err != nil {
		return nil, err
	}
	return &member, nil
}

// RemoveGroupMember 移除群组成员
func (r *groupRepo) RemoveGroupMember(ctx context.Context, groupID int64, userID int64) error {
	return r.db.WithContext(ctx).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		Delete(&model.GroupMember{}).Error
}

// UpdateMemberCount 更新群组成员数量
func (r *groupRepo) UpdateMemberCount(ctx context.Context, groupID int64, count int32) error {
	return r.db.WithContext(ctx).
		Model(&model.Group{}).
		Where("id = ?", groupID).
		Update("member_count", count).Error
}

// CreateGroupMessage 创建群聊消息
func (r *groupRepo) CreateGroupMessage(ctx context.Context, message *model.GroupMessage) error {
	return r.db.WithContext(ctx).Create(message).Error
}

// GetGroupMessages 获取群聊消息列表
func (r *groupRepo) GetGroupMessages(ctx context.Context, groupID int64, lastID int64, limit int) ([]*model.GroupMessage, error) {
	var messages []*model.GroupMessage
	query := r.db.WithContext(ctx).Where("group_id = ?", groupID)
	if lastID > 0 {
		query = query.Where("id < ?", lastID)
	}
	err := query.Order("created_at desc").Limit(limit).Find(&messages).Error
	if err != nil {
		return nil, err
	}
	return messages, nil
}

// GetLatestGroupMessages 获取最新的群聊消息（用于WebSocket推送）
func (r *groupRepo) GetLatestGroupMessages(ctx context.Context, groupID int64, lastID int64, limit int) ([]*model.GroupMessage, error) {
	var messages []*model.GroupMessage
	query := r.db.WithContext(ctx).Where("group_id = ?", groupID)
	if lastID > 0 {
		query = query.Where("id > ?", lastID)
	}
	err := query.Order("created_at desc").Limit(limit).Find(&messages).Error
	if err != nil {
		return nil, err
	}
	return messages, nil
}
