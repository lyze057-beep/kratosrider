package biz

import (
	"context"
	"hellokratos/internal/data"
	"hellokratos/internal/data/model"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"
)

// MessageUsecase 消息相关的业务逻辑接口
type MessageUsecase interface {
	// SendMessage 发送消息
	SendMessage(ctx context.Context, fromID int64, toID int64, content string, messageType int32) error
	// GetMessages 获取用户的消息列表
	GetMessages(ctx context.Context, userID int64, limit int) ([]*model.Message, error)
	// GetUnreadMessageCount 获取未读消息数量
	GetUnreadMessageCount(ctx context.Context, userID int64) (int64, error)
	// MarkMessageAsRead 标记消息为已读
	MarkMessageAsRead(ctx context.Context, messageID int64) error
	// GetLatestMessages 获取用户最新的消息（用于WebSocket推送）
	GetLatestMessages(ctx context.Context, userID int64, lastID int64, limit int) ([]*model.Message, error)
}

// messageUsecase 消息相关的业务逻辑实现
type messageUsecase struct {
	messageRepo data.MessageRepo
	rdb         *redis.Client
	log         *log.Helper
}

// NewMessageUsecase 创建消息业务逻辑实例
func NewMessageUsecase(messageRepo data.MessageRepo, rdb *redis.Client, logger log.Logger) MessageUsecase {
	return &messageUsecase{
		messageRepo: messageRepo,
		rdb:         rdb,
		log:         log.NewHelper(logger),
	}
}

// SendMessage 发送消息
func (uc *messageUsecase) SendMessage(ctx context.Context, fromID int64, toID int64, content string, messageType int32) error {
	// 创建消息
	message := &model.Message{
		FromID:  fromID,
		ToID:    toID,
		Content: content,
		Type:    messageType,
		Status:  0, // 未读
	}

	// 保存消息
	err := uc.messageRepo.CreateMessage(ctx, message)
	if err != nil {
		uc.log.Error("failed to send message", "err", err)
		return err
	}

	// TODO: 通过WebSocket推送消息给接收者
	uc.log.Info("message sent", "from_id", fromID, "to_id", toID, "content", content)
	return nil
}

// GetMessages 获取用户的消息列表
func (uc *messageUsecase) GetMessages(ctx context.Context, userID int64, limit int) ([]*model.Message, error) {
	return uc.messageRepo.GetMessagesByUserID(ctx, userID, limit)
}

// GetUnreadMessageCount 获取未读消息数量
func (uc *messageUsecase) GetUnreadMessageCount(ctx context.Context, userID int64) (int64, error) {
	return uc.messageRepo.GetUnreadMessageCount(ctx, userID)
}

// MarkMessageAsRead 标记消息为已读
func (uc *messageUsecase) MarkMessageAsRead(ctx context.Context, messageID int64) error {
	return uc.messageRepo.MarkMessageAsRead(ctx, messageID)
}

// GetLatestMessages 获取用户最新的消息（用于WebSocket推送）
func (uc *messageUsecase) GetLatestMessages(ctx context.Context, userID int64, lastID int64, limit int) ([]*model.Message, error) {
	return uc.messageRepo.GetLatestMessages(ctx, userID, lastID, limit)
}
