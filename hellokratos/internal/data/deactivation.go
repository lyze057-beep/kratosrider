package data

import (
	"context"
	"time"

	"hellokratos/internal/data/model"
)

// DeactivationRepo 骑手注销仓储接口
type DeactivationRepo interface {
	// 注销申请相关
	CreateDeactivation(ctx context.Context, deactivation *model.RiderDeactivation) error
	GetDeactivationByID(ctx context.Context, id int64) (*model.RiderDeactivation, error)
	GetPendingDeactivation(ctx context.Context, riderID int64) (*model.RiderDeactivation, error)
	GetLatestDeactivation(ctx context.Context, riderID int64) (*model.RiderDeactivation, error)
	UpdateDeactivation(ctx context.Context, deactivation *model.RiderDeactivation) error
	UpdateDeactivationStatus(ctx context.Context, id int64, status model.DeactivationStatus) error
	GetDeactivationList(ctx context.Context, status model.DeactivationStatus, page, pageSize int) ([]*model.RiderDeactivation, int64, error)

	// 检查清单相关
	CreateChecklist(ctx context.Context, checklist *model.DeactivationChecklist) error
	GetChecklistByDeactivationID(ctx context.Context, deactivationID int64) (*model.DeactivationChecklist, error)

	// 冷却期相关
	GetCooldown(ctx context.Context, riderID int64) (*model.DeactivationCooldown, error)
	SetCooldown(ctx context.Context, riderID int64, deactivatedAt, cooldownEndAt *time.Time) error

	// 历史记录相关
	CreateHistory(ctx context.Context, history *model.DeactivationHistory) error
	GetHistoryByRiderID(ctx context.Context, riderID int64) ([]*model.DeactivationHistory, error)

	// 检查项相关
	HasPendingOrders(ctx context.Context, riderID int64) (bool, error)
	HasPendingIncome(ctx context.Context, riderID int64) (bool, int64, error)
	HasPendingComplaint(ctx context.Context, riderID int64) (bool, error)
	HasEquipmentDebt(ctx context.Context, riderID int64) (bool, error)
	HasViolationRecord(ctx context.Context, riderID int64) (bool, error)
	IsBalanceClear(ctx context.Context, riderID int64) (bool, error)

	// 清算相关
	CalculatePendingIncome(ctx context.Context, riderID int64) (int64, error)
	CalculatePendingPenalty(ctx context.Context, riderID int64) (int64, error)

	// 骑手状态相关
	DeactivateRider(ctx context.Context, riderID int64) error

	// 统计相关
	GetDeactivationStatistics(ctx context.Context, startDate, endDate string) (*DeactivationStatisticsResult, error)
}

// DeactivationStatisticsResult 注销统计结果
type DeactivationStatisticsResult struct {
	TotalApply      int64
	PendingCount    int64
	ApprovedCount   int64
	RejectedCount   int64
	CompletedCount  int64
	AvgProcessTime  float64
	TotalSettlement int64
}
