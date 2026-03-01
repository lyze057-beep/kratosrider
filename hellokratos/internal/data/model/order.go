package model

import (
	"time"
)

// Order 订单模型
type Order struct {
	ID           int64     `gorm:"primaryKey" json:"id"`
	OrderNo      string    `gorm:"unique;size:50;not null" json:"order_no"`
	Status       int32     `gorm:"default:0" json:"status"` // 0-待接单 1-已接单 2-配送中 3-已完成 4-已取消
	Origin       string    `gorm:"size:255;not null" json:"origin"`        // 起点
	Destination  string    `gorm:"size:255;not null" json:"destination"`   // 终点
	OriginLat    float64   `gorm:"type:decimal(10,6)" json:"origin_lat"`   // 起点纬度
	OriginLng    float64   `gorm:"type:decimal(10,6)" json:"origin_lng"`   // 起点经度
	DestLat      float64   `gorm:"type:decimal(10,6)" json:"dest_lat"`     // 终点纬度
	DestLng      float64   `gorm:"type:decimal(10,6)" json:"dest_lng"`     // 终点经度
	Distance     float64   `gorm:"type:decimal(10,2)" json:"distance"`     // 距离（公里）
	Amount       float64   `gorm:"type:decimal(10,2)" json:"amount"`       // 金额
	RiderID      int64     `json:"rider_id"`                                // 骑手ID
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TableName 指定表名
func (Order) TableName() string {
	return "rider_order"
}
