package model

import (
	"time"
)

// DeactivationStatus 注销状态
type DeactivationStatus int32

const (
	DeactivationStatusPending   DeactivationStatus = 1 // 待审核
	DeactivationStatusApproved  DeactivationStatus = 2 // 已批准
	DeactivationStatusRejected  DeactivationStatus = 3 // 已拒绝
	DeactivationStatusCompleted DeactivationStatus = 4 // 已完成
)

// DeactivationReason 注销原因类型
type DeactivationReason int32

const (
	DeactivationReasonPersonal      DeactivationReason = 1 // 个人原因
	DeactivationReasonHealth        DeactivationReason = 2 // 健康原因
	DeactivationReasonOtherJob      DeactivationReason = 3 // 找到其他工作
	DeactivationReasonIncome        DeactivationReason = 4 // 收入不满意
	DeactivationReasonTime          DeactivationReason = 5 // 时间不合适
	DeactivationReasonOther         DeactivationReason = 6 // 其他原因
)

// RiderDeactivation 骑手注销申请
// 存储骑手的注销申请记录
// 关联表: rider_user
// 索引: idx_rider_id, idx_status, idx_created_at
type RiderDeactivation struct {
	ID                int64              `gorm:"primaryKey" json:"id"`
	RiderID           int64              `gorm:"not null;index:idx_rider_id" json:"rider_id"`
	ReasonType        DeactivationReason `gorm:"not null" json:"reason_type"`
	ReasonDetail      string             `gorm:"size:500" json:"reason_detail"`
	Status            DeactivationStatus `gorm:"default:1;index:idx_status" json:"status"`
	ApplyTime         time.Time          `json:"apply_time"`
	ReviewTime        *time.Time         `json:"review_time"`
	ReviewerID        int64              `json:"reviewer_id"`
	ReviewRemark      string             `gorm:"size:500" json:"review_remark"`
	CompleteTime      *time.Time         `json:"complete_time"`
	ClearanceStatus   int32              `gorm:"default:0" json:"clearance_status"` // 0-未清算 1-清算中 2-已清算
	PendingIncome     int64              `gorm:"default:0" json:"pending_income"`  // 待结算收入(分)
	PendingPenalty    int64              `gorm:"default:0" json:"pending_penalty"` // 待扣罚款(分)
	FinalSettlement   int64              `gorm:"default:0" json:"final_settlement"` // 最终结算金额(分)
	CreatedAt         time.Time          `json:"created_at"`
	UpdatedAt         time.Time          `json:"updated_at"`
}

// TableName 指定表名
func (RiderDeactivation) TableName() string {
	return "rider_deactivation"
}

// DeactivationChecklist 注销检查项
// 存储骑手注销前的各项检查状态
// 关联表: rider_deactivation
type DeactivationChecklist struct {
	ID                  int64     `gorm:"primaryKey" json:"id"`
	DeactivationID      int64     `gorm:"not null;index" json:"deactivation_id"`
	RiderID             int64     `gorm:"not null" json:"rider_id"`
	HasPendingOrders    bool      `gorm:"default:false" json:"has_pending_orders"`    // 有待处理订单
	HasPendingIncome    bool      `gorm:"default:false" json:"has_pending_income"`    // 有待结算收入
	HasPendingComplaint bool      `gorm:"default:false" json:"has_pending_complaint"` // 有待处理投诉
	HasEquipmentDebt    bool      `gorm:"default:false" json:"has_equipment_debt"`    // 有装备欠款
	HasViolationRecord  bool      `gorm:"default:false" json:"has_violation_record"`  // 有违规记录
	IsBalanceClear      bool      `gorm:"default:false" json:"is_balance_clear"`      // 余额已清零
	AllChecksPassed     bool      `gorm:"default:false" json:"all_checks_passed"`     // 所有检查通过
	CheckDetails        string    `gorm:"size:1000" json:"check_details"`             // 检查详情JSON
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// TableName 指定表名
func (DeactivationChecklist) TableName() string {
	return "rider_deactivation_checklist"
}

// DeactivationCooldown 注销冷却期
// 记录骑手的注销冷却期信息
// 关联表: rider_user
type DeactivationCooldown struct {
	ID             int64     `gorm:"primaryKey" json:"id"`
	RiderID        int64     `gorm:"unique;not null" json:"rider_id"`
	LastDeactivatedAt *time.Time `json:"last_deactivated_at"` // 上次注销时间
	CooldownEndAt  *time.Time `json:"cooldown_end_at"`      // 冷却期结束时间
	DeactivationCount int32   `gorm:"default:0" json:"deactivation_count"` // 注销次数
	CanReactivate  bool      `gorm:"default:true" json:"can_reactivate"`  // 是否可以重新激活
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// TableName 指定表名
func (DeactivationCooldown) TableName() string {
	return "rider_deactivation_cooldown"
}

// DeactivationHistory 注销历史记录
// 存储骑手的注销历史
// 关联表: rider_user
type DeactivationHistory struct {
	ID           int64              `gorm:"primaryKey" json:"id"`
	RiderID      int64              `gorm:"not null;index" json:"rider_id"`
	DeactivationID int64            `json:"deactivation_id"`
	Action       string             `gorm:"size:20;not null" json:"action"` // APPLY-申请 REVIEW-审核 COMPLETE-完成 REACTIVATE-重新激活
	ActionTime   time.Time          `json:"action_time"`
	OperatorID   int64              `json:"operator_id"`
	OperatorType int32              `json:"operator_type"` // 1-骑手 2-系统 3-管理员
	Remark       string             `gorm:"size:500" json:"remark"`
	IP           string             `gorm:"size:50" json:"ip"`
	DeviceID     string             `gorm:"size:100" json:"device_id"`
	CreatedAt    time.Time          `json:"created_at"`
}

// TableName 指定表名
func (DeactivationHistory) TableName() string {
	return "rider_deactivation_history"
}
