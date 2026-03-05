package model

import (
	"time"
)

// AIAgentMessage AI客服消息表
type AIAgentMessage struct {
	ID          int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	RiderID     int64     `gorm:"index;not null" json:"rider_id"`
	Content     string    `gorm:"type:text;not null" json:"content"`
	MessageType int32     `gorm:"default:0" json:"message_type"` // 0-用户消息 1-AI回复
	ContentType int32     `gorm:"default:0" json:"content_type"` // 0-文本 1-图片 2-语音
	SessionID   int64     `gorm:"index" json:"session_id"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// TableName 表名
func (AIAgentMessage) TableName() string {
	return "rider_ai_agent_message"
}

// AIAgentFAQ AI客服常见问题表
type AIAgentFAQ struct {
	ID         int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Question   string    `gorm:"type:varchar(500);not null" json:"question"`
	Answer     string    `gorm:"type:text;not null" json:"answer"`
	Category   string    `gorm:"type:varchar(50);index" json:"category"` // order-订单 income-收入 rule-规则 other-其他
	ViewCount  int32     `gorm:"default:0" json:"view_count"`
	IsActive   bool      `gorm:"default:true" json:"is_active"`
	SortOrder  int32     `gorm:"default:0" json:"sort_order"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 表名
func (AIAgentFAQ) TableName() string {
	return "rider_ai_agent_faq"
}

// AIAgentSession AI客服会话表
type AIAgentSession struct {
	ID         int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	RiderID    int64     `gorm:"index;not null" json:"rider_id"`
	StartTime  time.Time `gorm:"autoCreateTime" json:"start_time"`
	EndTime    *time.Time `json:"end_time"`
	Status     int32     `gorm:"default:0" json:"status"` // 0-进行中 1-已结束
	Rating     int32     `json:"rating"` // 评分：1-5星
	Feedback   string    `gorm:"type:text" json:"feedback"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 表名
func (AIAgentSession) TableName() string {
	return "rider_ai_agent_session"
}
