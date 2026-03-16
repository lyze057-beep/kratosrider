package model

import (
	"time"
)

// EmergencyHelp 紧急求助模型
type EmergencyHelp struct {
	ID              int64     `gorm:"primaryKey" json:"id"`
	TicketNo        string    `gorm:"size:50;uniqueIndex;not null" json:"ticket_no"`   // 工单号
	RiderID         int64     `gorm:"index;not null" json:"rider_id"`                  // 骑手ID
	OrderID         int64     `gorm:"index" json:"order_id"`                           // 订单ID
	HelpType        string    `gorm:"size:50;not null" json:"help_type"`               // 求助类型
	Description     string    `gorm:"size:2000" json:"description"`                    // 详细描述
	Latitude        float64   `json:"latitude"`                                        // 纬度
	Longitude       float64   `json:"longitude"`                                       // 经度
	Address         string    `gorm:"size:500" json:"address"`                         // 地址
	Images          string    `gorm:"size:2000" json:"images"`                         // 图片(JSON数组)
	Status          int32     `gorm:"default:1" json:"status"`                         // 状态(1=处理中,2=已响应,3=已处理,4=已取消)
	HandlerID       int64     `json:"handler_id"`                                      // 处理人ID
	HandlerName     string    `gorm:"size:50" json:"handler_name"`                     // 处理人姓名
	Result          string    `gorm:"size:1000" json:"result"`                         // 处理结果
	CancelReason    string    `gorm:"size:500" json:"cancel_reason"`                   // 取消原因
	ResolvedAt      *time.Time `json:"resolved_at"`                                    // 解决时间
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// TableName 指定表名
func (EmergencyHelp) TableName() string {
	return "rider_emergency_helps"
}

// InsuranceClaim 保险理赔模型
type InsuranceClaim struct {
	ID              int64     `gorm:"primaryKey" json:"id"`
	ClaimNo         string    `gorm:"size:50;uniqueIndex;not null" json:"claim_no"`    // 理赔编号
	RiderID         int64     `gorm:"index;not null" json:"rider_id"`                  // 骑手ID
	OrderID         int64     `gorm:"index" json:"order_id"`                           // 订单ID
	IncidentType    string    `gorm:"size:50;not null" json:"incident_type"`           // 事故类型
	IncidentTime    time.Time `json:"incident_time"`                                   // 事故发生时间
	IncidentLocation string   `gorm:"size:500" json:"incident_location"`               // 事故地点
	Latitude        float64   `json:"latitude"`                                        // 纬度
	Longitude       float64   `json:"longitude"`                                       // 经度
	IncidentDesc    string    `gorm:"size:2000" json:"incident_desc"`                  // 事故描述
	InjurySituation string    `gorm:"size:1000" json:"injury_situation"`               // 伤情描述
	MedicalExpense  string    `gorm:"size:50" json:"medical_expense"`                  // 医疗费用
	EvidenceImages  string    `gorm:"size:2000" json:"evidence_images"`                // 证据图片(JSON数组)
	MedicalRecords  string    `gorm:"size:2000" json:"medical_records"`                // 医疗记录(JSON数组)
	ContactName     string    `gorm:"size:50" json:"contact_name"`                     // 联系人姓名
	ContactPhone    string    `gorm:"size:20" json:"contact_phone"`                    // 联系人电话
	BankName        string    `gorm:"size:100" json:"bank_name"`                       // 银行名称
	BankAccount     string    `gorm:"size:50" json:"bank_account"`                     // 银行账号
	AccountName     string    `gorm:"size:50" json:"account_name"`                     // 账户名
	Status          int32     `gorm:"default:1" json:"status"`                         // 状态(1=待审核,2=审核中,3=已通过,4=已拒绝,5=已赔付)
	ClaimAmount     string    `gorm:"size:50" json:"claim_amount"`                     // 申请金额
	PaidAmount      string    `gorm:"size:50" json:"paid_amount"`                      // 赔付金额
	RejectReason    string    `gorm:"size:1000" json:"reject_reason"`                  // 拒绝原因
	HandlerID       int64     `json:"handler_id"`                                      // 处理人ID
	HandlerName     string    `gorm:"size:50" json:"handler_name"`                     // 处理人姓名
	Remark          string    `gorm:"size:1000" json:"remark"`                         // 备注
	PaidAt          *time.Time `json:"paid_at"`                                        // 赔付时间
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// TableName 指定表名
func (InsuranceClaim) TableName() string {
	return "rider_insurance_claims"
}

// RiderInsurance 骑手保险信息模型
type RiderInsurance struct {
	ID              int64     `gorm:"primaryKey" json:"id"`
	RiderID         int64     `gorm:"uniqueIndex;not null" json:"rider_id"`            // 骑手ID
	InsuranceNo     string    `gorm:"size:100;not null" json:"insurance_no"`           // 保险单号
	InsuranceType   string    `gorm:"size:50;not null" json:"insurance_type"`          // 保险类型
	CoverageDesc    string    `gorm:"size:1000" json:"coverage_desc"`                  // 保障范围描述
	CoverageAmount  string    `gorm:"size:50" json:"coverage_amount"`                  // 保额
	EffectiveDate   time.Time `json:"effective_date"`                                  // 生效日期
	ExpireDate      time.Time `json:"expire_date"`                                     // 到期日期
	Status          int32     `gorm:"default:1" json:"status"`                         // 状态(1=有效,2=已过期,3=已终止)
	ClaimCount      int32     `gorm:"default:0" json:"claim_count"`                    // 理赔次数
	TotalClaimAmount string   `gorm:"size:50;default:0" json:"total_claim_amount"`     // 累计理赔金额
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// TableName 指定表名
func (RiderInsurance) TableName() string {
	return "rider_insurances"
}

// SafetyEvent 安全事件模型
type SafetyEvent struct {
	ID              int64     `gorm:"primaryKey" json:"id"`
	RiderID         int64     `gorm:"index;not null" json:"rider_id"`                  // 骑手ID
	OrderID         int64     `gorm:"index" json:"order_id"`                           // 订单ID
	EventType       string    `gorm:"size:50;not null" json:"event_type"`              // 事件类型
	Description     string    `gorm:"size:2000" json:"description"`                    // 事件描述
	Latitude        float64   `json:"latitude"`                                        // 纬度
	Longitude       float64   `json:"longitude"`                                       // 经度
	Address         string    `gorm:"size:500" json:"address"`                         // 地址
	Images          string    `gorm:"size:2000" json:"images"`                         // 图片(JSON数组)
	NeedHelp        bool      `gorm:"default:false" json:"need_help"`                  // 是否需要帮助
	Status          int32     `gorm:"default:1" json:"status"`                         // 状态(1=待处理,2=已处理,3=已忽略)
	HandlerID       int64     `json:"handler_id"`                                      // 处理人ID
	HandlerName     string    `gorm:"size:50" json:"handler_name"`                     // 处理人姓名
	Reply           string    `gorm:"size:1000" json:"reply"`                          // 处理回复
	HandledAt       *time.Time `json:"handled_at"`                                     // 处理时间
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// TableName 指定表名
func (SafetyEvent) TableName() string {
	return "rider_safety_events"
}

// SafetyTip 安全提示模型
type SafetyTip struct {
	ID              int64     `gorm:"primaryKey" json:"id"`
	Title           string    `gorm:"size:200;not null" json:"title"`                  // 标题
	Content         string    `gorm:"size:5000;not null" json:"content"`               // 内容
	TipType         string    `gorm:"size:50;not null" json:"tip_type"`                // 提示类型
	Priority        int32     `gorm:"default:0" json:"priority"`                       // 优先级
	IsActive        bool      `gorm:"default:true" json:"is_active"`                   // 是否启用
	PublishTime     time.Time `json:"publish_time"`                                    // 发布时间
	ExpireTime      *time.Time `json:"expire_time"`                                    // 过期时间
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// TableName 指定表名
func (SafetyTip) TableName() string {
	return "rider_safety_tips"
}

// EmergencyContact 紧急联系人模型
type EmergencyContact struct {
	ID              int64     `gorm:"primaryKey" json:"id"`
	RiderID         int64     `gorm:"index;not null" json:"rider_id"`                  // 骑手ID
	Name            string    `gorm:"size:50;not null" json:"name"`                    // 姓名
	Phone           string    `gorm:"size:20;not null" json:"phone"`                   // 电话
	Relation        string    `gorm:"size:50" json:"relation"`                         // 关系
	IsPrimary       bool      `gorm:"default:false" json:"is_primary"`                 // 是否主要联系人
	SortOrder       int32     `gorm:"default:0" json:"sort_order"`                     // 排序
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// TableName 指定表名
func (EmergencyContact) TableName() string {
	return "rider_emergency_contacts"
}
