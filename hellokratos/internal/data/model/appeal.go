package model

import (
	"time"
)

// Appeal 申诉模型
type Appeal struct {
	ID              int64     `gorm:"primaryKey" json:"id"`
	TicketNo        string    `gorm:"size:50;uniqueIndex;not null" json:"ticket_no"` // 工单号
	RiderID         int64     `gorm:"index;not null" json:"rider_id"`                // 骑手ID
	AppealType      int32     `gorm:"not null" json:"appeal_type"`                   // 申诉类型(1=订单申诉,2=处罚申诉)
	OrderID         int64     `gorm:"index" json:"order_id"`                         // 订单ID
	PenaltyType     string    `gorm:"size:50" json:"penalty_type"`                   // 处罚类型
	PenaltyID       string    `gorm:"size:50" json:"penalty_id"`                     // 处罚ID
	Reason          string    `gorm:"size:100;not null" json:"reason"`               // 申诉原因
	Description     string    `gorm:"size:2000" json:"description"`                  // 详细描述
	EvidenceImages  string    `gorm:"size:2000" json:"evidence_images"`              // 证据图片(JSON数组)
	ContactPhone    string    `gorm:"size:20" json:"contact_phone"`                  // 联系电话
	Status          int32     `gorm:"default:1" json:"status"`                       // 状态(1=待处理,2=处理中,3=已通过,4=已驳回,5=已取消)
	Result          string    `gorm:"size:500" json:"result"`                        // 处理结果
	Reply           string    `gorm:"size:2000" json:"reply"`                        // 处理回复
	HandlerID       int64     `json:"handler_id"`                                     // 处理人ID
	HandlerName     string    `gorm:"size:50" json:"handler_name"`                   // 处理人姓名
	CancelReason    string    `gorm:"size:500" json:"cancel_reason"`                 // 取消原因
	ResolvedAt      *time.Time `json:"resolved_at"`                                   // 解决时间
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// TableName 指定表名
func (Appeal) TableName() string {
	return "rider_appeals"
}

// AppealType 申诉类型模型
type AppealType struct {
	ID              int64     `gorm:"primaryKey" json:"id"`
	TypeCode        string    `gorm:"size:50;uniqueIndex;not null" json:"type_code"`   // 类型编码
	TypeName        string    `gorm:"size:100;not null" json:"type_name"`              // 类型名称
	Description     string    `gorm:"size:500" json:"description"`                     // 类型描述
	AppealCategory  int32     `gorm:"not null" json:"appeal_category"`                 // 申诉类别(1=订单申诉,2=处罚申诉)
	SortOrder       int32     `gorm:"default:0" json:"sort_order"`                     // 排序
	IsActive        bool      `gorm:"default:true" json:"is_active"`                   // 是否启用
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// TableName 指定表名
func (AppealType) TableName() string {
	return "rider_appeal_types"
}

// ExceptionReport 异常报备模型
type ExceptionReport struct {
	ID              int64     `gorm:"primaryKey" json:"id"`
	RiderID         int64     `gorm:"index;not null" json:"rider_id"`                  // 骑手ID
	OrderID         int64     `gorm:"index" json:"order_id"`                           // 订单ID
	ExceptionType   string    `gorm:"size:50;not null" json:"exception_type"`          // 异常类型
	Description     string    `gorm:"size:2000" json:"description"`                    // 详细描述
	Images          string    `gorm:"size:2000" json:"images"`                         // 图片(JSON数组)
	Location        string    `gorm:"size:200" json:"location"`                        // 位置描述
	Latitude        float64   `json:"latitude"`                                        // 纬度
	Longitude       float64   `json:"longitude"`                                       // 经度
	Status          int32     `gorm:"default:1" json:"status"`                         // 状态(1=待处理,2=已处理,3=已忽略)
	HandlerID       int64     `json:"handler_id"`                                      // 处理人ID
	HandlerName     string    `gorm:"size:50" json:"handler_name"`                     // 处理人姓名
	Reply           string    `gorm:"size:2000" json:"reply"`                          // 处理回复
	HandledAt       *time.Time `json:"handled_at"`                                     // 处理时间
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// TableName 指定表名
func (ExceptionReport) TableName() string {
	return "rider_exception_reports"
}

// ExceptionOrder 异常订单模型(用于记录系统检测到的异常订单)
type ExceptionOrder struct {
	ID              int64     `gorm:"primaryKey" json:"id"`
	OrderID         int64     `gorm:"uniqueIndex;not null" json:"order_id"`            // 订单ID
	RiderID         int64     `gorm:"index;not null" json:"rider_id"`                  // 骑手ID
	ExceptionType   string    `gorm:"size:50;not null" json:"exception_type"`          // 异常类型
	ExceptionDesc   string    `gorm:"size:500" json:"exception_desc"`                  // 异常描述
	OccurredAt      time.Time `json:"occurred_at"`                                     // 发生时间
	Status          int32     `gorm:"default:1" json:"status"`                         // 状态(1=未处理,2=已申诉,3=已处理)
	AppealID        int64     `json:"appeal_id"`                                       // 关联申诉ID
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// TableName 指定表名
func (ExceptionOrder) TableName() string {
	return "rider_exception_orders"
}
