package model

import (
	"time"
)

// Income 收入模型
type Income struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	RiderID   int64     `json:"rider_id"`   // 骑手ID
	OrderID   int64     `json:"order_id"`   // 订单ID
	Amount    float64   `gorm:"type:decimal(10,2)" json:"amount"` // 金额
	Type      int32     `gorm:"default:0" json:"type"`    // 类型：0-订单收入 1-其他收入
	Status    int32     `gorm:"default:0" json:"status"`   // 状态：0-待结算 1-已结算 2-已提现
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Withdrawal 提现模型
type Withdrawal struct {
	ID           int64     `gorm:"primaryKey" json:"id"`
	RiderID      int64     `json:"rider_id"`      // 骑手ID
	Amount       float64   `gorm:"type:decimal(10,2)" json:"amount"`       // 提现金额
	Platform     string    `gorm:"size:50" json:"platform"`     // 提现平台：alipay, wechat, bank
	Account      string    `gorm:"size:100" json:"account"`      // 提现账号
	Status       int32     `gorm:"default:0" json:"status"`       // 状态：0-待处理 1-处理中 2-成功 3-失败
	TransactionID string    `gorm:"size:100" json:"transaction_id"` // 交易ID
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TableName 指定表名
func (Income) TableName() string {
	return "rider_income"
}

// TableName 指定表名
func (Withdrawal) TableName() string {
	return "rider_withdrawal"
}
