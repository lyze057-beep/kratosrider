package model

import (
	"time"
)

// Group 群组模型
type Group struct {
	ID          int64     `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:100;not null" json:"name"`        // 群组名称
	Description string    `gorm:"size:500" json:"description"`          // 群组描述
	CreatorID   int64     `gorm:"not null" json:"creator_id"`           // 创建者ID
	MemberCount int32     `gorm:"default:0" json:"member_count"`        // 成员数量
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName 指定表名
func (Group) TableName() string {
	return "rider_group"
}

// GroupMember 群组成员模型
type GroupMember struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	GroupID   int64     `gorm:"not null;index" json:"group_id"`       // 群组ID
	UserID    int64     `gorm:"not null;index" json:"user_id"`        // 用户ID
	Role      int32     `gorm:"default:0" json:"role"`                // 角色：0-成员 1-管理员 2-群主
	JoinedAt  time.Time `json:"joined_at"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 指定表名
func (GroupMember) TableName() string {
	return "rider_group_member"
}

// GroupMessage 群聊消息模型
type GroupMessage struct {
	ID       int64  `gorm:"primaryKey" json:"id"`
	GroupID  int64  `gorm:"not null;index" json:"group_id"`       // 群组ID
	FromID   int64  `gorm:"not null" json:"from_id"`              // 发送者ID
	Content  string `gorm:"size:1000" json:"content"`             // 消息内容
	Type     int32  `gorm:"default:0" json:"type"`                // 消息类型：0-文本 1-图片 2-语音
	Nickname string `gorm:"size:50" json:"nickname"`              // 发送者昵称（冗余存储）
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 指定表名
func (GroupMessage) TableName() string {
	return "rider_group_message"
}
