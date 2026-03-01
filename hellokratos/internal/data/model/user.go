package model

import (
	"time"
)

// User 骑手用户模型
type User struct {
	ID                 int64     `gorm:"primaryKey" json:"id"`
	Phone              string    `gorm:"unique;size:20;not null" json:"phone"`
	Password           string    `gorm:"size:100" json:"-"`
	Nickname           string    `gorm:"size:50" json:"nickname"`
	Avatar             string    `gorm:"size:255" json:"avatar"`
	Status             int32     `gorm:"default:0" json:"status"` // 0-正常 1-禁用
	ThirdPartyPlatform string    `gorm:"size:20" json:"third_party_platform"`
	ThirdPartyID       string    `gorm:"size:100" json:"third_party_id"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// TableName 指定表名
func (User) TableName() string {
	return "rider_user"
}
