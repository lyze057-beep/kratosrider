package service

import (
	"context"
	v1 "hellokratos/api/rider/v1"
	"hellokratos/internal/biz"
	"hellokratos/internal/data/model"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

// MessageService 消息服务
type MessageService struct {
	v1.UnimplementedMessageServer
	messageUsecase biz.MessageUsecase
	groupUsecase   biz.GroupUsecase
	log            *log.Helper
}

// NewMessageService 创建消息服务实例
func NewMessageService(messageUsecase biz.MessageUsecase, groupUsecase biz.GroupUsecase, logger log.Logger) *MessageService {
	return &MessageService{
		messageUsecase: messageUsecase,
		groupUsecase:   groupUsecase,
		log:            log.NewHelper(logger),
	}
}

// SendMessage 发送消息
func (s *MessageService) SendMessage(ctx context.Context, req *v1.SendMessageRequest) (*v1.SendMessageReply, error) {
	err := s.messageUsecase.SendMessage(ctx, req.FromId, req.ToId, req.Content, req.Type)
	if err != nil {
		s.log.Error("send message failed", "err", err)
		return &v1.SendMessageReply{
			Success: false,
			Message: err.Error(),
		}, nil
	}
	return &v1.SendMessageReply{
		Success: true,
		Message: "发送成功",
	}, nil
}

// GetMessageList 获取消息列表
func (s *MessageService) GetMessageList(ctx context.Context, req *v1.GetMessageListRequest) (*v1.GetMessageListReply, error) {
	messages, err := s.messageUsecase.GetMessages(ctx, req.UserId, int(req.Limit))
	if err != nil {
		s.log.Error("get message list failed", "err", err)
		return nil, err
	}

	messageInfos := make([]*v1.MessageInfo, 0, len(messages))
	for _, msg := range messages {
		messageInfos = append(messageInfos, s.convertMessageToInfo(msg))
	}

	return &v1.GetMessageListReply{
		Messages: messageInfos,
		Total:    int32(len(messageInfos)),
	}, nil
}

// GetUnreadMessageCount 获取未读消息数量
func (s *MessageService) GetUnreadMessageCount(ctx context.Context, req *v1.GetUnreadMessageCountRequest) (*v1.GetUnreadMessageCountReply, error) {
	count, err := s.messageUsecase.GetUnreadMessageCount(ctx, req.UserId)
	if err != nil {
		s.log.Error("get unread message count failed", "err", err)
		return nil, err
	}
	return &v1.GetUnreadMessageCountReply{
		Count: count,
	}, nil
}

// MarkMessageAsRead 标记消息为已读
func (s *MessageService) MarkMessageAsRead(ctx context.Context, req *v1.MarkMessageAsReadRequest) (*v1.MarkMessageAsReadReply, error) {
	err := s.messageUsecase.MarkMessageAsRead(ctx, req.MessageId)
	if err != nil {
		s.log.Error("mark message as read failed", "err", err)
		return &v1.MarkMessageAsReadReply{
			Success: false,
			Message: err.Error(),
		}, nil
	}
	return &v1.MarkMessageAsReadReply{
		Success: true,
		Message: "标记成功",
	}, nil
}

// SubscribeMessages 实时消息推送（WebSocket）
func (s *MessageService) SubscribeMessages(req *v1.SubscribeMessagesRequest, stream v1.Message_SubscribeMessagesServer) error {
	s.log.Info("user subscribed messages", "user_id", req.UserId)

	// 记录最后一条消息的ID
	var lastMessageID int64

	// 持续监听新消息
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// 检查上下文是否已取消
			select {
			case <-stream.Context().Done():
				s.log.Info("context canceled, stopping message subscription", "user_id", req.UserId)
				return stream.Context().Err()
			default:
			}

			// 获取用户最新的消息
			messages, err := s.messageUsecase.GetLatestMessages(stream.Context(), req.UserId, lastMessageID, 10)
			if err != nil {
				s.log.Error("get messages failed", "err", err)
				continue
			}

			// 推送新消息
			for _, msg := range messages {
				err := stream.Send(&v1.SubscribeMessagesReply{
					Message: s.convertMessageToInfo(msg),
				})
				if err != nil {
					s.log.Error("send message failed", "err", err)
					return err
				}
				// 更新最后一条消息的ID
				if msg.ID > lastMessageID {
					lastMessageID = msg.ID
				}
			}
		case <-stream.Context().Done():
			s.log.Info("user unsubscribed messages", "user_id", req.UserId)
			return nil
		}
	}
}

// ========== 群聊接口实现 ==========

// CreateGroup 创建群组
func (s *MessageService) CreateGroup(ctx context.Context, req *v1.CreateGroupRequest) (*v1.CreateGroupReply, error) {
	group, err := s.groupUsecase.CreateGroup(ctx, req.CreatorId, req.Name, req.Description, req.MemberIds)
	if err != nil {
		s.log.Error("create group failed", "err", err)
		return &v1.CreateGroupReply{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &v1.CreateGroupReply{
		Success: true,
		Message: "群组创建成功",
		Group:   s.convertGroupToInfo(group),
	}, nil
}

// JoinGroup 加入群组
func (s *MessageService) JoinGroup(ctx context.Context, req *v1.JoinGroupRequest) (*v1.JoinGroupReply, error) {
	err := s.groupUsecase.JoinGroup(ctx, req.UserId, req.GroupId)
	if err != nil {
		s.log.Error("join group failed", "err", err)
		return &v1.JoinGroupReply{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &v1.JoinGroupReply{
		Success: true,
		Message: "加入群组成功",
	}, nil
}

// LeaveGroup 退出群组
func (s *MessageService) LeaveGroup(ctx context.Context, req *v1.LeaveGroupRequest) (*v1.LeaveGroupReply, error) {
	err := s.groupUsecase.LeaveGroup(ctx, req.UserId, req.GroupId)
	if err != nil {
		s.log.Error("leave group failed", "err", err)
		return &v1.LeaveGroupReply{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &v1.LeaveGroupReply{
		Success: true,
		Message: "退出群组成功",
	}, nil
}

// GetGroupList 获取群组列表
func (s *MessageService) GetGroupList(ctx context.Context, req *v1.GetGroupListRequest) (*v1.GetGroupListReply, error) {
	groups, err := s.groupUsecase.GetGroupList(ctx, req.UserId, int(req.Limit))
	if err != nil {
		s.log.Error("get group list failed", "err", err)
		return nil, err
	}

	groupInfos := make([]*v1.GroupInfo, 0, len(groups))
	for _, group := range groups {
		groupInfos = append(groupInfos, s.convertGroupToInfo(group))
	}

	return &v1.GetGroupListReply{
		Groups: groupInfos,
		Total:  int32(len(groupInfos)),
	}, nil
}

// GetGroupMembers 获取群组成员列表
func (s *MessageService) GetGroupMembers(ctx context.Context, req *v1.GetGroupMembersRequest) (*v1.GetGroupMembersReply, error) {
	members, err := s.groupUsecase.GetGroupMembers(ctx, req.GroupId, int(req.Limit))
	if err != nil {
		s.log.Error("get group members failed", "err", err)
		return nil, err
	}

	memberInfos := make([]*v1.GroupMemberInfo, 0, len(members))
	for _, member := range members {
		memberInfos = append(memberInfos, s.convertGroupMemberToInfo(member))
	}

	return &v1.GetGroupMembersReply{
		Members: memberInfos,
		Total:   int32(len(memberInfos)),
	}, nil
}

// SendGroupMessage 发送群聊消息
func (s *MessageService) SendGroupMessage(ctx context.Context, req *v1.SendGroupMessageRequest) (*v1.SendGroupMessageReply, error) {
	message, err := s.groupUsecase.SendGroupMessage(ctx, req.FromId, req.GroupId, req.Content, req.Type)
	if err != nil {
		s.log.Error("send group message failed", "err", err)
		return &v1.SendGroupMessageReply{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &v1.SendGroupMessageReply{
		Success:      true,
		Message:      "发送成功",
		GroupMessage: s.convertGroupMessageToInfo(message),
	}, nil
}

// GetGroupMessageList 获取群聊消息列表
func (s *MessageService) GetGroupMessageList(ctx context.Context, req *v1.GetGroupMessageListRequest) (*v1.GetGroupMessageListReply, error) {
	messages, err := s.groupUsecase.GetGroupMessages(ctx, req.GroupId, req.LastId, int(req.Limit))
	if err != nil {
		s.log.Error("get group message list failed", "err", err)
		return nil, err
	}

	messageInfos := make([]*v1.GroupMessageInfo, 0, len(messages))
	for _, msg := range messages {
		messageInfos = append(messageInfos, s.convertGroupMessageToInfo(msg))
	}

	return &v1.GetGroupMessageListReply{
		Messages: messageInfos,
		Total:    int32(len(messageInfos)),
	}, nil
}

// SubscribeGroupMessages 实时群聊消息推送（WebSocket）
func (s *MessageService) SubscribeGroupMessages(req *v1.SubscribeGroupMessagesRequest, stream v1.Message_SubscribeGroupMessagesServer) error {
	s.log.Info("user subscribed group messages", "user_id", req.UserId, "group_id", req.GroupId)

	// 记录每个群组最后一条消息的ID
	lastMessageIDs := make(map[int64]int64)

	// 获取用户加入的群组ID列表
	var groupIDs []int64
	var err error
	if req.GroupId == 0 {
		// 订阅所有加入的群组
		groupIDs, err = s.groupUsecase.GetUserGroupIDs(stream.Context(), req.UserId)
		if err != nil {
			s.log.Error("get user group ids failed", "err", err)
			return err
		}
	} else {
		// 订阅指定群组
		groupIDs = []int64{req.GroupId}
	}

	if len(groupIDs) == 0 {
		s.log.Info("user has no groups to subscribe", "user_id", req.UserId)
		<-stream.Context().Done()
		return nil
	}

	// 持续监听新消息
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// 检查上下文是否已取消
			select {
			case <-stream.Context().Done():
				s.log.Info("context canceled, stopping group message subscription", "user_id", req.UserId)
				return stream.Context().Err()
			default:
			}

			// 遍历所有订阅的群组，获取新消息
			for _, groupID := range groupIDs {
				lastID := lastMessageIDs[groupID]
				messages, err := s.groupUsecase.GetLatestGroupMessages(stream.Context(), groupID, lastID, 10)
				if err != nil {
					s.log.Error("get latest group messages failed", "err", err, "group_id", groupID)
					continue
				}

				// 推送新消息
				for _, msg := range messages {
					err := stream.Send(&v1.SubscribeGroupMessagesReply{
						Message: s.convertGroupMessageToInfo(msg),
					})
					if err != nil {
						s.log.Error("send group message failed", "err", err)
						return err
					}
					// 更新最后一条消息的ID
					if msg.ID > lastMessageIDs[groupID] {
						lastMessageIDs[groupID] = msg.ID
					}
				}
			}
		case <-stream.Context().Done():
			s.log.Info("user unsubscribed group messages", "user_id", req.UserId)
			return nil
		}
	}
}

// ========== 转换方法 ==========

// convertMessageToInfo 将消息模型转换为消息信息
func (s *MessageService) convertMessageToInfo(message *model.Message) *v1.MessageInfo {
	return &v1.MessageInfo{
		Id:        message.ID,
		FromId:    message.FromID,
		ToId:      message.ToID,
		Content:   message.Content,
		Type:      int32(message.Type),
		Status:    int32(message.Status),
		CreatedAt: message.CreatedAt.Format(time.RFC3339),
	}
}

// convertGroupToInfo 将群组模型转换为群组信息
func (s *MessageService) convertGroupToInfo(group *model.Group) *v1.GroupInfo {
	return &v1.GroupInfo{
		Id:          group.ID,
		Name:        group.Name,
		Description: group.Description,
		CreatorId:   group.CreatorID,
		MemberCount: group.MemberCount,
		CreatedAt:   group.CreatedAt.Format(time.RFC3339),
	}
}

// convertGroupMemberToInfo 将群组成员模型转换为群组成员信息
func (s *MessageService) convertGroupMemberToInfo(member *model.GroupMember) *v1.GroupMemberInfo {
	return &v1.GroupMemberInfo{
		UserId:   member.UserID,
		Nickname: "", // 可以从用户信息中获取
		Role:     int32(member.Role),
		JoinedAt: member.JoinedAt.Format(time.RFC3339),
	}
}

// convertGroupMessageToInfo 将群聊消息模型转换为群聊消息信息
func (s *MessageService) convertGroupMessageToInfo(message *model.GroupMessage) *v1.GroupMessageInfo {
	return &v1.GroupMessageInfo{
		Id:           message.ID,
		GroupId:      message.GroupID,
		FromId:       message.FromID,
		FromNickname: message.Nickname,
		Content:      message.Content,
		Type:         int32(message.Type),
		CreatedAt:    message.CreatedAt.Format(time.RFC3339),
	}
}
