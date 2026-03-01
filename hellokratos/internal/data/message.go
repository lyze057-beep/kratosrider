package data

import (
	"context"
	"hellokratos/internal/data/model"

	"gorm.io/gorm"
)

// MessageRepo 消息相关的数据访问接口
type MessageRepo interface {
	// CreateMessage 创建消息
	CreateMessage(ctx context.Context, message *model.Message) error
	// GetMessageByID 根据ID获取消息
	GetMessageByID(ctx context.Context, id int64) (*model.Message, error)
	// GetMessagesByUserID 获取用户的消息列表
	GetMessagesByUserID(ctx context.Context, userID int64, limit int) ([]*model.Message, error)
	// GetUnreadMessageCount 获取未读消息数量
	GetUnreadMessageCount(ctx context.Context, userID int64) (int64, error)
	// MarkMessageAsRead 标记消息为已读
	MarkMessageAsRead(ctx context.Context, messageID int64) error
	// GetLatestMessages 获取用户最新的消息（用于WebSocket推送）
	GetLatestMessages(ctx context.Context, userID int64, lastID int64, limit int) ([]*model.Message, error)
}

// messageRepo 消息相关的数据访问实现
type messageRepo struct {
	db *gorm.DB
}

// NewMessageRepo 创建消息数据访问实例
func NewMessageRepo(data *Data) MessageRepo {
	return &messageRepo{db: data.db}
}

// CreateMessage 创建消息
func (r *messageRepo) CreateMessage(ctx context.Context, message *model.Message) error {
	return r.db.WithContext(ctx).Create(message).Error
}

// GetMessageByID 根据ID获取消息
func (r *messageRepo) GetMessageByID(ctx context.Context, id int64) (*model.Message, error) {
	var message model.Message
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&message).Error
	if err != nil {
		return nil, err
	}
	return &message, nil
}

// GetMessagesByUserID 获取用户的消息列表
func (r *messageRepo) GetMessagesByUserID(ctx context.Context, userID int64, limit int) ([]*model.Message, error) {
	var messages []*model.Message
	err := r.db.WithContext(ctx).Where("to_id = ?", userID).Order("created_at desc").Limit(limit).Find(&messages).Error
	if err != nil {
		return nil, err
	}
	return messages, nil
}

// GetUnreadMessageCount 获取未读消息数量
func (r *messageRepo) GetUnreadMessageCount(ctx context.Context, userID int64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Message{}).Where("to_id = ? AND status = 0", userID).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

// MarkMessageAsRead 标记消息为已读
func (r *messageRepo) MarkMessageAsRead(ctx context.Context, messageID int64) error {
	return r.db.WithContext(ctx).Model(&model.Message{}).Where("id = ?", messageID).Update("status", 1).Error
}

// GetLatestMessages 获取用户最新的消息（用于WebSocket推送）
func (r *messageRepo) GetLatestMessages(ctx context.Context, userID int64, lastID int64, limit int) ([]*model.Message, error) {
	var messages []*model.Message
	query := r.db.WithContext(ctx).Where("to_id = ?", userID)
	if lastID > 0 {
		query = query.Where("id > ?", lastID)
	}
	err := query.Order("created_at desc").Limit(limit).Find(&messages).Error
	if err != nil {
		return nil, err
	}
	return messages, nil
}
