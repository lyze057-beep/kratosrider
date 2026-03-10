package data

import (
	"context"
	"fmt"
	"hellokratos/internal/data/model"
	"time"

	"gorm.io/gorm"
)

// AIAgentRepo AI客服数据访问接口
type AIAgentRepo interface {
	// 消息相关
	CreateMessage(ctx context.Context, message *model.AIAgentMessage) error
	GetMessagesByRiderID(ctx context.Context, riderID int64, lastID int64, limit int) ([]*model.AIAgentMessage, error)
	GetLastMessageBySessionID(ctx context.Context, sessionID int64) (*model.AIAgentMessage, error)

	// FAQ相关
	GetFAQList(ctx context.Context, category string, limit int) ([]*model.AIAgentFAQ, error)
	GetFAQByID(ctx context.Context, id int64) (*model.AIAgentFAQ, error)
	IncrementFAQViewCount(ctx context.Context, id int64) error

	// 会话相关
	CreateSession(ctx context.Context, session *model.AIAgentSession) error
	GetSessionByID(ctx context.Context, id int64) (*model.AIAgentSession, error)
	GetActiveSessionByRiderID(ctx context.Context, riderID int64) (*model.AIAgentSession, error)
	EndSession(ctx context.Context, sessionID int64) error
	RateSession(ctx context.Context, sessionID int64, rating int32, feedback string) error

	// 限流相关
	CheckRateLimit(ctx context.Context, riderID int64) (bool, error)
	UpdateRateLimit(ctx context.Context, riderID int64) error
}

// aiAgentRepo AI客服数据访问实现
type aiAgentRepo struct {
	data *Data
}

// NewAIAgentRepo 创建AI客服数据访问实例
func NewAIAgentRepo(data *Data) AIAgentRepo {
	return &aiAgentRepo{
		data: data,
	}
}

// CreateMessage 创建消息
func (r *aiAgentRepo) CreateMessage(ctx context.Context, message *model.AIAgentMessage) error {
	return r.data.db.WithContext(ctx).Create(message).Error
}

// GetMessagesByRiderID 获取骑手的消息列表
func (r *aiAgentRepo) GetMessagesByRiderID(ctx context.Context, riderID int64, lastID int64, limit int) ([]*model.AIAgentMessage, error) {
	var messages []*model.AIAgentMessage
	query := r.data.db.WithContext(ctx).Where("rider_id = ?", riderID)

	if lastID > 0 {
		query = query.Where("id < ?", lastID)
	}

	err := query.Order("id DESC").Limit(limit).Find(&messages).Error
	return messages, err
}

// GetFAQList 获取FAQ列表
func (r *aiAgentRepo) GetFAQList(ctx context.Context, category string, limit int) ([]*model.AIAgentFAQ, error) {
	var faqs []*model.AIAgentFAQ
	query := r.data.db.WithContext(ctx).Where("is_active = ?", true)

	if category != "" {
		query = query.Where("category = ?", category)
	}

	err := query.Order("sort_order DESC, view_count DESC").Limit(limit).Find(&faqs).Error
	return faqs, err
}

// GetFAQByID 根据ID获取FAQ
func (r *aiAgentRepo) GetFAQByID(ctx context.Context, id int64) (*model.AIAgentFAQ, error) {
	var faq model.AIAgentFAQ
	err := r.data.db.WithContext(ctx).First(&faq, id).Error
	if err != nil {
		return nil, err
	}
	return &faq, nil
}

// IncrementFAQViewCount 增加FAQ查看次数
func (r *aiAgentRepo) IncrementFAQViewCount(ctx context.Context, id int64) error {
	return r.data.db.WithContext(ctx).Model(&model.AIAgentFAQ{}).
		Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}

// CreateSession 创建会话
func (r *aiAgentRepo) CreateSession(ctx context.Context, session *model.AIAgentSession) error {
	return r.data.db.WithContext(ctx).Create(session).Error
}

// GetSessionByID 根据ID获取会话
func (r *aiAgentRepo) GetSessionByID(ctx context.Context, id int64) (*model.AIAgentSession, error) {
	var session model.AIAgentSession
	err := r.data.db.WithContext(ctx).First(&session, id).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// GetActiveSessionByRiderID 获取骑手的活跃会话
func (r *aiAgentRepo) GetActiveSessionByRiderID(ctx context.Context, riderID int64) (*model.AIAgentSession, error) {
	var session model.AIAgentSession
	err := r.data.db.WithContext(ctx).
		Where("rider_id = ? AND status = ?", riderID, 0).
		Order("id DESC").
		First(&session).Error

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	return &session, err
}

// EndSession 结束会话
func (r *aiAgentRepo) EndSession(ctx context.Context, sessionID int64) error {
	return r.data.db.WithContext(ctx).Model(&model.AIAgentSession{}).
		Where("id = ?", sessionID).
		Updates(map[string]interface{}{
			"status":   1,
			"end_time": gorm.Expr("NOW()"),
		}).Error
}

// RateSession 评价会话
func (r *aiAgentRepo) RateSession(ctx context.Context, sessionID int64, rating int32, feedback string) error {
	return r.data.db.WithContext(ctx).Model(&model.AIAgentSession{}).Where("id = ?", sessionID).Updates(map[string]interface{}{
		"rating":   rating,
		"feedback": feedback,
	}).Error
}

// GetLastMessageBySessionID 获取会话的最后一条消息
func (r *aiAgentRepo) GetLastMessageBySessionID(ctx context.Context, sessionID int64) (*model.AIAgentMessage, error) {
	var message model.AIAgentMessage
	err := r.data.db.WithContext(ctx).Where("session_id = ?", sessionID).Order("created_at DESC").First(&message).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &message, nil
}

// CheckRateLimit 检查限流
func (r *aiAgentRepo) CheckRateLimit(ctx context.Context, riderID int64) (bool, error) {
	// 限流键：每分钟和每小时
	minKey := fmt.Sprintf("ai_agent:rate_limit:min:%d", riderID)
	hourKey := fmt.Sprintf("ai_agent:rate_limit:hour:%d", riderID)

	// 检查每分钟限制（10条）
	minCount, err := r.data.rdb.Get(ctx, minKey).Int()
	if err == nil && minCount >= 10 {
		return false, nil
	}

	// 检查每小时限制（100条）
	hourCount, err := r.data.rdb.Get(ctx, hourKey).Int()
	if err == nil && hourCount >= 100 {
		return false, nil
	}

	return true, nil
}

// UpdateRateLimit 更新限流计数
func (r *aiAgentRepo) UpdateRateLimit(ctx context.Context, riderID int64) error {
	// 限流键：每分钟和每小时
	minKey := fmt.Sprintf("ai_agent:rate_limit:min:%d", riderID)
	hourKey := fmt.Sprintf("ai_agent:rate_limit:hour:%d", riderID)

	// 增加计数并设置过期时间
	pipe := r.data.rdb.Pipeline()
	pipe.Incr(ctx, minKey)
	pipe.Expire(ctx, minKey, time.Minute)
	pipe.Incr(ctx, hourKey)
	pipe.Expire(ctx, hourKey, time.Hour)
	_, err := pipe.Exec(ctx)

	return err
}
