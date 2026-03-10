package biz

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	"hellokratos/internal/data"
	"hellokratos/internal/data/model"
)

// DeactivationUsecase 骑手注销业务用例
type DeactivationUsecase struct {
	repo data.DeactivationRepo
	log  *log.Helper
}

// NewDeactivationUsecase 创建注销业务用例
func NewDeactivationUsecase(repo data.DeactivationRepo, logger log.Logger) *DeactivationUsecase {
	return &DeactivationUsecase{
		repo: repo,
		log:  log.NewHelper(logger),
	}
}

// ApplyDeactivationRequest 注销申请请求
type ApplyDeactivationRequest struct {
	RiderID      int64
	ReasonType   model.DeactivationReason
	ReasonDetail string
	Phone        string
	IP           string
	DeviceID     string
}

// ApplyDeactivationResponse 注销申请响应
type ApplyDeactivationResponse struct {
	DeactivationID int64
	Status         model.DeactivationStatus
	Checklist      *DeactivationCheckResult
	Message        string
}

// DeactivationCheckResult 注销检查结果
type DeactivationCheckResult struct {
	CanDeactivate       bool     `json:"can_deactivate"`
	HasPendingOrders    bool     `json:"has_pending_orders"`
	HasPendingIncome    bool     `json:"has_pending_income"`
	HasPendingComplaint bool     `json:"has_pending_complaint"`
	HasEquipmentDebt    bool     `json:"has_equipment_debt"`
	HasViolationRecord  bool     `json:"has_violation_record"`
	IsBalanceClear      bool     `json:"is_balance_clear"`
	BlockingItems       []string `json:"blocking_items"`
}

// ApplyDeactivation 申请注销
func (uc *DeactivationUsecase) ApplyDeactivation(ctx context.Context, req *ApplyDeactivationRequest) (*ApplyDeactivationResponse, error) {
	// 1. 检查骑手是否已有待处理的注销申请
	existing, err := uc.repo.GetPendingDeactivation(ctx, req.RiderID)
	if err != nil {
		return nil, fmt.Errorf("查询待处理申请失败: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("您已有一个待处理的注销申请")
	}

	// 2. 检查冷却期
	cooldown, err := uc.repo.GetCooldown(ctx, req.RiderID)
	if err != nil {
		return nil, fmt.Errorf("查询冷却期失败: %w", err)
	}
	if cooldown != nil && cooldown.CooldownEndAt != nil && cooldown.CooldownEndAt.After(time.Now()) {
		return nil, fmt.Errorf("您处于注销冷却期内，%s后可重新申请", cooldown.CooldownEndAt.Format("2006-01-02"))
	}

	// 3. 执行注销检查
	checkResult, err := uc.performDeactivationCheck(ctx, req.RiderID)
	if err != nil {
		return nil, fmt.Errorf("注销检查失败: %w", err)
	}

	// 4. 创建注销申请
	deactivation := &model.RiderDeactivation{
		RiderID:      req.RiderID,
		ReasonType:   req.ReasonType,
		ReasonDetail: req.ReasonDetail,
		Status:       model.DeactivationStatusPending,
		ApplyTime:    time.Now(),
	}

	if err := uc.repo.CreateDeactivation(ctx, deactivation); err != nil {
		return nil, fmt.Errorf("创建注销申请失败: %w", err)
	}

	// 5. 保存检查清单
	checklist := &model.DeactivationChecklist{
		DeactivationID:      deactivation.ID,
		RiderID:             req.RiderID,
		HasPendingOrders:    checkResult.HasPendingOrders,
		HasPendingIncome:    checkResult.HasPendingIncome,
		HasPendingComplaint: checkResult.HasPendingComplaint,
		HasEquipmentDebt:    checkResult.HasEquipmentDebt,
		HasViolationRecord:  checkResult.HasViolationRecord,
		IsBalanceClear:      checkResult.IsBalanceClear,
		AllChecksPassed:     checkResult.CanDeactivate,
	}
	checkDetailsJSON, _ := json.Marshal(checkResult)
	checklist.CheckDetails = string(checkDetailsJSON)

	if err := uc.repo.CreateChecklist(ctx, checklist); err != nil {
		uc.log.Warnf("保存检查清单失败: %v", err)
	}

	// 6. 记录历史
	history := &model.DeactivationHistory{
		RiderID:      req.RiderID,
		DeactivationID: deactivation.ID,
		Action:       "APPLY",
		ActionTime:   time.Now(),
		OperatorID:   req.RiderID,
		OperatorType: 1,
		Remark:       fmt.Sprintf("申请注销，原因: %s", string(req.ReasonType)),
		IP:           req.IP,
		DeviceID:     req.DeviceID,
	}
	if err := uc.repo.CreateHistory(ctx, history); err != nil {
		uc.log.Warnf("记录历史失败: %v", err)
	}

	uc.log.Infof("骑手申请注销: rider_id=%d, deactivation_id=%d", req.RiderID, deactivation.ID)

	return &ApplyDeactivationResponse{
		DeactivationID: deactivation.ID,
		Status:         deactivation.Status,
		Checklist:      checkResult,
		Message:        uc.getCheckResultMessage(checkResult),
	}, nil
}

// performDeactivationCheck 执行注销检查
func (uc *DeactivationUsecase) performDeactivationCheck(ctx context.Context, riderID int64) (*DeactivationCheckResult, error) {
	result := &DeactivationCheckResult{
		CanDeactivate: true,
		BlockingItems: []string{},
	}

	// 检查1: 是否有待处理订单
	hasPendingOrders, err := uc.repo.HasPendingOrders(ctx, riderID)
	if err != nil {
		return nil, err
	}
	result.HasPendingOrders = hasPendingOrders
	if hasPendingOrders {
		result.CanDeactivate = false
		result.BlockingItems = append(result.BlockingItems, "有待处理订单")
	}

	// 检查2: 是否有待结算收入
	hasPendingIncome, pendingAmount, err := uc.repo.HasPendingIncome(ctx, riderID)
	if err != nil {
		return nil, err
	}
	result.HasPendingIncome = hasPendingIncome
	if hasPendingIncome && pendingAmount > 100000 { // 超过1000元需要处理
		result.CanDeactivate = false
		result.BlockingItems = append(result.BlockingItems, "有待结算收入")
	}

	// 检查3: 是否有待处理投诉
	hasPendingComplaint, err := uc.repo.HasPendingComplaint(ctx, riderID)
	if err != nil {
		return nil, err
	}
	result.HasPendingComplaint = hasPendingComplaint
	if hasPendingComplaint {
		result.CanDeactivate = false
		result.BlockingItems = append(result.BlockingItems, "有待处理投诉")
	}

	// 检查4: 是否有装备欠款
	hasEquipmentDebt, err := uc.repo.HasEquipmentDebt(ctx, riderID)
	if err != nil {
		return nil, err
	}
	result.HasEquipmentDebt = hasEquipmentDebt
	if hasEquipmentDebt {
		result.CanDeactivate = false
		result.BlockingItems = append(result.BlockingItems, "有装备欠款")
	}

	// 检查5: 是否有严重违规记录
	hasViolation, err := uc.repo.HasViolationRecord(ctx, riderID)
	if err != nil {
		return nil, err
	}
	result.HasViolationRecord = hasViolation
	if hasViolation {
		result.BlockingItems = append(result.BlockingItems, "有违规记录")
	}

	// 检查6: 余额是否清零
	isBalanceClear, err := uc.repo.IsBalanceClear(ctx, riderID)
	if err != nil {
		return nil, err
	}
	result.IsBalanceClear = isBalanceClear
	if !isBalanceClear {
		result.BlockingItems = append(result.BlockingItems, "账户余额未清零")
	}

	return result, nil
}

// getCheckResultMessage 获取检查结果消息
func (uc *DeactivationUsecase) getCheckResultMessage(result *DeactivationCheckResult) string {
	if result.CanDeactivate {
		return "检查通过，您的注销申请已提交，请等待审核"
	}
	return fmt.Sprintf("检查未通过，请处理以下事项: %v", result.BlockingItems)
}

// GetDeactivationStatus 获取注销状态
func (uc *DeactivationUsecase) GetDeactivationStatus(ctx context.Context, riderID int64) (*model.RiderDeactivation, *model.DeactivationChecklist, error) {
	deactivation, err := uc.repo.GetLatestDeactivation(ctx, riderID)
	if err != nil {
		return nil, nil, err
	}

	if deactivation == nil {
		return nil, nil, nil
	}

	checklist, err := uc.repo.GetChecklistByDeactivationID(ctx, deactivation.ID)
	if err != nil {
		uc.log.Warnf("获取检查清单失败: %v", err)
	}

	return deactivation, checklist, nil
}

// CancelDeactivation 取消注销申请
func (uc *DeactivationUsecase) CancelDeactivation(ctx context.Context, riderID, deactivationID int64) error {
	deactivation, err := uc.repo.GetDeactivationByID(ctx, deactivationID)
	if err != nil {
		return fmt.Errorf("查询注销申请失败: %w", err)
	}

	if deactivation.RiderID != riderID {
		return fmt.Errorf("无权操作此申请")
	}

	if deactivation.Status != model.DeactivationStatusPending {
		return fmt.Errorf("只能取消待审核的申请")
	}

	// 更新状态为已拒绝（主动取消）
	if err := uc.repo.UpdateDeactivationStatus(ctx, deactivationID, model.DeactivationStatusRejected); err != nil {
		return fmt.Errorf("取消申请失败: %w", err)
	}

	// 记录历史
	history := &model.DeactivationHistory{
		RiderID:      riderID,
		DeactivationID: deactivationID,
		Action:       "CANCEL",
		ActionTime:   time.Now(),
		OperatorID:   riderID,
		OperatorType: 1,
		Remark:       "骑手主动取消注销申请",
	}
	if err := uc.repo.CreateHistory(ctx, history); err != nil {
		uc.log.Warnf("记录历史失败: %v", err)
	}

	uc.log.Infof("骑手取消注销申请: rider_id=%d, deactivation_id=%d", riderID, deactivationID)
	return nil
}

// ==================== 管理员接口 ====================

// ReviewDeactivationRequest 审核注销请求
type ReviewDeactivationRequest struct {
	DeactivationID int64
	ReviewerID     int64
	Action         string // APPROVE-批准 REJECT-拒绝
	Remark         string
}

// ReviewDeactivation 审核注销申请
func (uc *DeactivationUsecase) ReviewDeactivation(ctx context.Context, req *ReviewDeactivationRequest) error {
	deactivation, err := uc.repo.GetDeactivationByID(ctx, req.DeactivationID)
	if err != nil {
		return fmt.Errorf("查询注销申请失败: %w", err)
	}

	if deactivation.Status != model.DeactivationStatusPending {
		return fmt.Errorf("只能审核待处理的申请")
	}

	now := time.Now()

	if req.Action == "APPROVE" {
		deactivation.Status = model.DeactivationStatusApproved
		deactivation.ReviewTime = &now
		deactivation.ReviewerID = req.ReviewerID
		deactivation.ReviewRemark = req.Remark

		// 开始清算流程
		deactivation.ClearanceStatus = 1

		if err := uc.repo.UpdateDeactivation(ctx, deactivation); err != nil {
			return fmt.Errorf("更新申请状态失败: %w", err)
		}

		// 异步执行清算
		go uc.performClearance(context.Background(), deactivation)

	} else {
		deactivation.Status = model.DeactivationStatusRejected
		deactivation.ReviewTime = &now
		deactivation.ReviewerID = req.ReviewerID
		deactivation.ReviewRemark = req.Remark

		if err := uc.repo.UpdateDeactivation(ctx, deactivation); err != nil {
			return fmt.Errorf("更新申请状态失败: %w", err)
		}
	}

	// 记录历史
	history := &model.DeactivationHistory{
		RiderID:      deactivation.RiderID,
		DeactivationID: req.DeactivationID,
		Action:       "REVIEW",
		ActionTime:   now,
		OperatorID:   req.ReviewerID,
		OperatorType: 3,
		Remark:       fmt.Sprintf("%s: %s", req.Action, req.Remark),
	}
	if err := uc.repo.CreateHistory(ctx, history); err != nil {
		uc.log.Warnf("记录历史失败: %v", err)
	}

	uc.log.Infof("审核注销申请: deactivation_id=%d, action=%s, reviewer=%d", req.DeactivationID, req.Action, req.ReviewerID)
	return nil
}

// performClearance 执行清算
func (uc *DeactivationUsecase) performClearance(ctx context.Context, deactivation *model.RiderDeactivation) {
	uc.log.Infof("开始清算: deactivation_id=%d, rider_id=%d", deactivation.ID, deactivation.RiderID)

	// 1. 计算待结算收入
	pendingIncome, err := uc.repo.CalculatePendingIncome(ctx, deactivation.RiderID)
	if err != nil {
		uc.log.Errorf("计算待结算收入失败: %v", err)
		return
	}
	deactivation.PendingIncome = pendingIncome

	// 2. 计算待扣罚款
	pendingPenalty, err := uc.repo.CalculatePendingPenalty(ctx, deactivation.RiderID)
	if err != nil {
		uc.log.Errorf("计算待扣罚款失败: %v", err)
		return
	}
	deactivation.PendingPenalty = pendingPenalty

	// 3. 计算最终结算金额
	finalSettlement := pendingIncome - pendingPenalty
	if finalSettlement < 0 {
		finalSettlement = 0
	}
	deactivation.FinalSettlement = finalSettlement

	// 4. 更新清算状态
	deactivation.ClearanceStatus = 2

	// 5. 完成注销
	now := time.Now()
	deactivation.Status = model.DeactivationStatusCompleted
	deactivation.CompleteTime = &now

	if err := uc.repo.UpdateDeactivation(ctx, deactivation); err != nil {
		uc.log.Errorf("更新清算结果失败: %v", err)
		return
	}

	// 6. 更新骑手状态为已注销
	if err := uc.repo.DeactivateRider(ctx, deactivation.RiderID); err != nil {
		uc.log.Errorf("注销骑手失败: %v", err)
		return
	}

	// 7. 设置冷却期（30天）
	cooldownEnd := now.AddDate(0, 0, 30)
	if err := uc.repo.SetCooldown(ctx, deactivation.RiderID, &now, &cooldownEnd); err != nil {
		uc.log.Errorf("设置冷却期失败: %v", err)
	}

	// 8. 记录历史
	history := &model.DeactivationHistory{
		RiderID:      deactivation.RiderID,
		DeactivationID: deactivation.ID,
		Action:       "COMPLETE",
		ActionTime:   now,
		OperatorID:   0,
		OperatorType: 2,
		Remark:       fmt.Sprintf("清算完成，最终结算: %.2f元", float64(finalSettlement)/100),
	}
	if err := uc.repo.CreateHistory(ctx, history); err != nil {
		uc.log.Warnf("记录历史失败: %v", err)
	}

	uc.log.Infof("清算完成: deactivation_id=%d, final_settlement=%d", deactivation.ID, finalSettlement)
}

// GetDeactivationList 获取注销申请列表（管理员）
func (uc *DeactivationUsecase) GetDeactivationList(ctx context.Context, status model.DeactivationStatus, page, pageSize int) ([]*model.RiderDeactivation, int64, error) {
	return uc.repo.GetDeactivationList(ctx, status, page, pageSize)
}

// GetDeactivationStatistics 获取注销统计（管理员）
func (uc *DeactivationUsecase) GetDeactivationStatistics(ctx context.Context, startDate, endDate string) (*data.DeactivationStatisticsResult, error) {
	return uc.repo.GetDeactivationStatistics(ctx, startDate, endDate)
}


