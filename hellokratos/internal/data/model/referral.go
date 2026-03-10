package model

import (
	"time"

	"gorm.io/gorm"
)

// ReferralInviteCode 邀请码表
type ReferralInviteCode struct {
	ID           int64     `gorm:"primaryKey;column:id"`
	RiderID      int64     `gorm:"column:rider_id;index:idx_rider_id"`
	InviteCode   string    `gorm:"column:invite_code;uniqueIndex:idx_invite_code;size:20"`
	InviteLink   string    `gorm:"column:invite_link;size:255"`
	Status       int32     `gorm:"column:status;default:1"` // 1-有效 2-过期 3-禁用
	ExpireAt     time.Time `gorm:"column:expire_at"`
	TotalInvited int32     `gorm:"column:total_invited;default:0"`
	ValidInvited int32     `gorm:"column:valid_invited;default:0"`
	TotalRewards int32     `gorm:"column:total_rewards;default:0"`
	CreatedAt    time.Time `gorm:"column:created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at"`
}

func (ReferralInviteCode) TableName() string {
	return "referral_invite_codes"
}

// ReferralRelation 邀请关系表
type ReferralRelation struct {
	ID              int64     `gorm:"primaryKey;column:id"`
	InviterID       int64     `gorm:"column:inviter_id;index:idx_inviter_id"`
	InvitedID       int64     `gorm:"column:invited_id;uniqueIndex:idx_invited_id"`
	InviteCode      string    `gorm:"column:invite_code;size:20"`
	Status          int32     `gorm:"column:status;default:1"` // 1-已注册 2-已完成首单 3-任务达标 4-已奖励 5-无效
	RegisterAt      time.Time `gorm:"column:register_at"`
	FirstOrderAt    *time.Time `gorm:"column:first_order_at"`
	TaskCompletedAt *time.Time `gorm:"column:task_completed_at"`
	RewardAmount    int32     `gorm:"column:reward_amount;default:0"`
	RewardStatus    int32     `gorm:"column:reward_status;default:0"` // 0-无奖励 1-待发放 2-已发放
	IsCheating      bool      `gorm:"column:is_cheating;default:false"`
	CheatingReason  string    `gorm:"column:cheating_reason;size:255"`
	CreatedAt       time.Time `gorm:"column:created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (ReferralRelation) TableName() string {
	return "referral_relations"
}

// ReferralTask 拉新任务表
type ReferralTask struct {
	ID              int64     `gorm:"primaryKey;column:id"`
	TaskName        string    `gorm:"column:task_name;size:100"`
	TaskDesc        string    `gorm:"column:task_desc;size:500"`
	TaskType        int32     `gorm:"column:task_type"` // 1-首单任务 2-在线时长任务 3-完成单量任务
	TargetValue     int32     `gorm:"column:target_value"`
	RewardAmount    int32     `gorm:"column:reward_amount"`
	RewardType      int32     `gorm:"column:reward_type;default:1"` // 1-现金 2-积分
	TimeLimitDays   int32     `gorm:"column:time_limit_days"`
	StartTime       time.Time `gorm:"column:start_time"`
	EndTime         time.Time `gorm:"column:end_time"`
	Status          int32     `gorm:"column:status;default:1"` // 1-有效 2-无效
	CreatedAt       time.Time `gorm:"column:created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at"`
}

func (ReferralTask) TableName() string {
	return "referral_tasks"
}

// ReferralTaskProgress 任务进度表
type ReferralTaskProgress struct {
	ID             int64     `gorm:"primaryKey;column:id"`
	RelationID     int64     `gorm:"column:relation_id;index:idx_relation_task"`
	TaskID         int64     `gorm:"column:task_id;index:idx_relation_task"`
	InviterID      int64     `gorm:"column:inviter_id;index:idx_inviter"`
	InvitedID      int64     `gorm:"column:invited_id"`
	CurrentValue   int32     `gorm:"column:current_value;default:0"`
	TargetValue    int32     `gorm:"column:target_value"`
	Status         int32     `gorm:"column:status;default:0"` // 0-未开始 1-进行中 2-已完成 3-已过期
	StartTime      time.Time `gorm:"column:start_time"`
	Deadline       time.Time `gorm:"column:deadline"`
	CompletedAt    *time.Time `gorm:"column:completed_at"`
	IsClaimed      bool      `gorm:"column:is_claimed;default:false"`
	ClaimedAt      *time.Time `gorm:"column:claimed_at"`
	CreatedAt      time.Time `gorm:"column:created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at"`
}

func (ReferralTaskProgress) TableName() string {
	return "referral_task_progress"
}

// ReferralRewardRecord 奖励记录表
type ReferralRewardRecord struct {
	ID           int64     `gorm:"primaryKey;column:id"`
	RelationID   int64     `gorm:"column:relation_id;index:idx_relation"`
	TaskID       int64     `gorm:"column:task_id"`
	InviterID    int64     `gorm:"column:inviter_id;index:idx_inviter"`
	InvitedID    int64     `gorm:"column:invited_id"`
	RewardAmount int32     `gorm:"column:reward_amount"`
	RewardType   int32     `gorm:"column:reward_type"` // 1-现金 2-积分
	Status       int32     `gorm:"column:status;default:1"` // 1-待发放 2-已发放 3-发放失败
	FailReason   string    `gorm:"column:fail_reason;size:255"`
	CreatedAt    time.Time `gorm:"column:created_at"`
	IssuedAt     *time.Time `gorm:"column:issued_at"`
}

func (ReferralRewardRecord) TableName() string {
	return "referral_reward_records"
}

// ReferralRiskLog 风控日志表
type ReferralRiskLog struct {
	ID             int64     `gorm:"primaryKey;column:id"`
	RelationID     int64     `gorm:"column:relation_id;index:idx_relation"`
	InviterID      int64     `gorm:"column:inviter_id"`
	InvitedID      int64     `gorm:"column:invited_id"`
	RiskType       int32     `gorm:"column:risk_type"` // 1-设备重复 2-IP异常 3-行为异常 4-虚假注册
	RiskLevel      int32     `gorm:"column:risk_level"` // 1-低风险 2-中风险 3-高风险
	RiskDesc       string    `gorm:"column:risk_desc;size:500"`
	DeviceID       string    `gorm:"column:device_id;size:100"`
	IPAddress      string    `gorm:"column:ip_address;size:50"`
	IsConfirmed    bool      `gorm:"column:is_confirmed;default:false"`
	ConfirmedAt    *time.Time `gorm:"column:confirmed_at"`
	CreatedAt      time.Time `gorm:"column:created_at"`
}

func (ReferralRiskLog) TableName() string {
	return "referral_risk_logs"
}

// ReferralStatistics 拉新统计表（每日汇总）
type ReferralStatistics struct {
	ID             int64     `gorm:"primaryKey;column:id"`
	StatDate       string    `gorm:"column:stat_date;uniqueIndex:idx_stat_date;size:10"`
	NewInvited     int32     `gorm:"column:new_invited;default:0"`
	ValidInvited   int32     `gorm:"column:valid_invited;default:0"`
	RewardsAmount  int32     `gorm:"column:rewards_amount;default:0"`
	CheatingCount  int32     `gorm:"column:cheating_count;default:0"`
	CreatedAt      time.Time `gorm:"column:created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at"`
}

func (ReferralStatistics) TableName() string {
	return "referral_statistics"
}
