package model

import (
	"time"
)

// Message 消息模型
type Message struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	FromID    int64     `json:"from_id"`    // 发送者ID
	ToID      int64     `json:"to_id"`      // 接收者ID
	Content   string    `gorm:"size:1000" json:"content"` // 消息内容
	Type      int32     `gorm:"default:0" json:"type"`    // 消息类型：0-文本 1-图片 2-语音
	Status    int32     `gorm:"default:0" json:"status"`   // 状态：0-未读 1-已读
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 指定表名
func (Message) TableName() string {
	return "rider_message"
}
