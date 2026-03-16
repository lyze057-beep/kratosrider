package model

import (
	"time"
)

// OrderRating 订单评价模型
type OrderRating struct {
	ID           int64     `gorm:"primaryKey" json:"id"`
	OrderID      int64     `gorm:"index;not null" json:"order_id"`      // 订单ID
	RiderID      int64     `gorm:"index;not null" json:"rider_id"`      // 骑手ID
	UserID       int64     `gorm:"index;not null" json:"user_id"`       // 用户ID
	Rating       int32     `gorm:"not null" json:"rating"`              // 评分(1-5)
	Tags         string    `gorm:"size:500" json:"tags"`                // 评价标签(JSON数组)
	Comment      string    `gorm:"size:1000" json:"comment"`            // 评价内容
	Images       string    `gorm:"size:2000" json:"images"`             // 评价图片(JSON数组)
	IsAnonymous  bool      `gorm:"default:false" json:"is_anonymous"`   // 是否匿名
	Reply        string    `gorm:"size:1000" json:"reply"`              // 骑手回复
	RepliedAt    *time.Time `json:"replied_at"`                           // 回复时间
	Status       int32     `gorm:"default:1" json:"status"`             // 状态(1=正常,2=隐藏)
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TableName 指定表名
func (OrderRating) TableName() string {
	return "rider_order_rating"
}

// RatingTag 评价标签模型
type RatingTag struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	TagName   string    `gorm:"size:50;not null;unique" json:"tag_name"` // 标签名称
	TagType   int32     `gorm:"default:1" json:"tag_type"`               // 标签类型(1=好评,2=中评,3=差评)
	SortOrder int32     `gorm:"default:0" json:"sort_order"`             // 排序
	Status    int32     `gorm:"default:1" json:"status"`                 // 状态
	CreatedAt time.Time `json:"created_at"`
}

// TableName 指定表名
func (RatingTag) TableName() string {
	return "rider_rating_tag"
}

// Complaint 投诉模型
type Complaint struct {
	ID            int64     `gorm:"primaryKey" json:"id"`
	TicketNo      string    `gorm:"size:50;uniqueIndex;not null" json:"ticket_no"` // 工单号
	RiderID       int64     `gorm:"index;not null" json:"rider_id"`                // 骑手ID
	OrderID       int64     `gorm:"index" json:"order_id"`                         // 订单ID
	ComplaintType string    `gorm:"size:50;not null" json:"complaint_type"`        // 投诉类型
	Content       string    `gorm:"size:2000;not null" json:"content"`             // 投诉内容
	Images        string    `gorm:"size:2000" json:"images"`                       // 投诉图片(JSON数组)
	ContactPhone  string    `gorm:"size:20" json:"contact_phone"`                  // 联系电话
	Status        int32     `gorm:"default:1" json:"status"`                       // 状态(1=待处理,2=处理中,3=已解决,4=已关闭)
	Priority      int32     `gorm:"default:2" json:"priority"`                     // 优先级(1=高,2=中,3=低)
	HandlerID     int64     `json:"handler_id"`                                    // 处理人ID
	HandlerName   string    `gorm:"size:50" json:"handler_name"`                   // 处理人姓名
	Reply         string    `gorm:"size:2000" json:"reply"`                        // 处理回复
	ResolvedAt    *time.Time `json:"resolved_at"`                                    // 解决时间
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// TableName 指定表名
func (Complaint) TableName() string {
	return "rider_complaint"
}

// ComplaintType 投诉类型模型
type ComplaintType struct {
	ID          int64     `gorm:"primaryKey" json:"id"`
	TypeCode    string    `gorm:"size:50;not null;unique" json:"type_code"`    // 类型编码
	TypeName    string    `gorm:"size:100;not null" json:"type_name"`          // 类型名称
	Description string    `gorm:"size:500" json:"description"`                 // 类型描述
	SortOrder   int32     `gorm:"default:0" json:"sort_order"`                 // 排序
	Status      int32     `gorm:"default:1" json:"status"`                     // 状态
	CreatedAt   time.Time `json:"created_at"`
}

// TableName 指定表名
func (ComplaintType) TableName() string {
	return "rider_complaint_type"
}

// RiderRatingStats 骑手评分统计模型
type RiderRatingStats struct {
	ID            int64     `gorm:"primaryKey" json:"id"`
	RiderID       int64     `gorm:"uniqueIndex;not null" json:"rider_id"`        // 骑手ID
	TotalRatings  int32     `gorm:"default:0" json:"total_ratings"`              // 总评价数
	AvgRating     float64   `gorm:"type:decimal(2,1);default:5.0" json:"avg_rating"` // 平均评分
	FiveStarCount int32     `gorm:"default:0" json:"five_star_count"`            // 5星数量
	FourStarCount int32     `gorm:"default:0" json:"four_star_count"`            // 4星数量
	ThreeStarCount int32    `gorm:"default:0" json:"three_star_count"`           // 3星数量
	TwoStarCount  int32     `gorm:"default:0" json:"two_star_count"`             // 2星数量
	OneStarCount  int32     `gorm:"default:0" json:"one_star_count"`             // 1星数量
	PraiseRate    float64   `gorm:"type:decimal(5,2);default:100.0" json:"praise_rate"` // 好评率
	StatDate      string    `gorm:"size:10;index" json:"stat_date"`               // 统计日期
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// TableName 指定表名
func (RiderRatingStats) TableName() string {
	return "rider_rating_stats"
}
