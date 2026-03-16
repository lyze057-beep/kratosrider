package model

import (
	"time"
)

// RiderLocation 骑手实时位置模型
type RiderLocation struct {
	ID           int64     `gorm:"primaryKey" json:"id"`
	RiderID      int64     `gorm:"uniqueIndex;not null" json:"rider_id"` // 骑手ID
	Latitude     float64   `gorm:"type:decimal(10,6);not null" json:"latitude"` // 纬度
	Longitude    float64   `gorm:"type:decimal(10,6);not null" json:"longitude"` // 经度
	Accuracy     float64   `gorm:"type:decimal(5,2);default:0" json:"accuracy"` // 定位精度(米)
	Speed        float64   `gorm:"type:decimal(5,2);default:0" json:"speed"` // 速度(km/h)
	Direction    int32     `gorm:"default:0" json:"direction"` // 方向(0-360度)
	Address      string    `gorm:"size:255" json:"address"` // 详细地址
	City         string    `gorm:"size:50" json:"city"` // 城市
	Province     string    `gorm:"size:50" json:"province"` // 省份
	Country      string    `gorm:"size:50;default:中国" json:"country"` // 国家
	LocationType string    `gorm:"size:20;default:gps" json:"location_type"` // 定位类型(gps/wifi/基站)
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TableName 指定表名
func (RiderLocation) TableName() string {
	return "rider_location"
}

// RiderLocationHistory 骑手位置历史记录模型
type RiderLocationHistory struct {
	ID           int64     `gorm:"primaryKey" json:"id"`
	RiderID      int64     `gorm:"index;not null" json:"rider_id"` // 骑手ID
	Latitude     float64   `gorm:"type:decimal(10,6);not null" json:"latitude"` // 纬度
	Longitude    float64   `gorm:"type:decimal(10,6);not null" json:"longitude"` // 经度
	Accuracy     float64   `gorm:"type:decimal(5,2);default:0" json:"accuracy"` // 定位精度(米)
	Speed        float64   `gorm:"type:decimal(5,2);default:0" json:"speed"` // 速度(km/h)
	Direction    int32     `gorm:"default:0" json:"direction"` // 方向(0-360度)
	Address      string    `gorm:"size:255" json:"address"` // 详细地址
	City         string    `gorm:"size:50" json:"city"` // 城市
	Province     string    `gorm:"size:50" json:"province"` // 省份
	Country      string    `gorm:"size:50;default:中国" json:"country"` // 国家
	LocationType string    `gorm:"size:20;default:gps" json:"location_type"` // 定位类型(gps/wifi/基站)
	CreatedAt    time.Time `json:"created_at"`
}

// TableName 指定表名
func (RiderLocationHistory) TableName() string {
	return "rider_location_history"
}
