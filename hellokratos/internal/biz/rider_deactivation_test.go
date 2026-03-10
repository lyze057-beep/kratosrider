package biz

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	"hellokratos/internal/data"
	"hellokratos/internal/data/model"
)

// mockDeactivationRepo 模拟注销仓储
type mockDeactivationRepo struct {
	mu              sync.RWMutex
	deactivations   map[int64]*model.RiderDeactivation
	checklists      map[int64]*model.DeactivationChecklist
	cooldowns       map[int64]*model.DeactivationCooldown
	histories       []*model.DeactivationHistory
	riderStatus     map[int64]int32
	pendingOrders   map[int64]bool
	pendingIncome   map[int64]int64
	pendingComplaint map[int64]bool
	equipmentDebt   map[int64]bool
	violationRecord map[int64]bool
	balanceClear    map[int64]bool
}

func newMockDeactivationRepo() *mockDeactivationRepo {
	return &mockDeactivationRepo{
		deactivations:    make(map[int64]*model.RiderDeactivation),
		checklists:       make(map[int64]*model.DeactivationChecklist),
		cooldowns:        make(map[int64]*model.DeactivationCooldown),
		histories:        make([]*model.DeactivationHistory, 0),
		riderStatus:      make(map[int64]int32),
		pendingOrders:    make(map[int64]bool),
		pendingIncome:    make(map[int64]int64),
		pendingComplaint: make(map[int64]bool),
		equipmentDebt:    make(map[int64]bool),
		violationRecord:  make(map[int64]bool),
		balanceClear:     make(map[int64]bool),
	}
}

func (r *mockDeactivationRepo) CreateDeactivation(ctx context.Context, d *model.RiderDeactivation) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if d.ID == 0 {
		d.ID = int64(len(r.deactivations) + 1)
	}
	r.deactivations[d.ID] = d
	return nil
}

func (r *mockDeactivationRepo) GetDeactivationByID(ctx context.Context, id int64) (*model.RiderDeactivation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if d, ok := r.deactivations[id]; ok {
		return d, nil
	}
	return nil, nil
}

func (r *mockDeactivationRepo) GetPendingDeactivation(ctx context.Context, riderID int64) (*model.RiderDeactivation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, d := range r.deactivations {
		if d.RiderID == riderID && d.Status == model.DeactivationStatusPending {
			return d, nil
		}
	}
	return nil, nil
}

func (r *mockDeactivationRepo) GetLatestDeactivation(ctx context.Context, riderID int64) (*model.RiderDeactivation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var latest *model.RiderDeactivation
	for _, d := range r.deactivations {
		if d.RiderID == riderID {
			if latest == nil || d.CreatedAt.After(latest.CreatedAt) {
				latest = d
			}
		}
	}
	return latest, nil
}

func (r *mockDeactivationRepo) UpdateDeactivation(ctx context.Context, d *model.RiderDeactivation) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.deactivations[d.ID] = d
	return nil
}

func (r *mockDeactivationRepo) UpdateDeactivationStatus(ctx context.Context, id int64, status model.DeactivationStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if d, ok := r.deactivations[id]; ok {
		d.Status = status
	}
	return nil
}

func (r *mockDeactivationRepo) GetDeactivationList(ctx context.Context, status model.DeactivationStatus, page, pageSize int) ([]*model.RiderDeactivation, int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*model.RiderDeactivation
	for _, d := range r.deactivations {
		if status == 0 || d.Status == status {
			result = append(result, d)
		}
	}
	return result, int64(len(result)), nil
}

func (r *mockDeactivationRepo) CreateChecklist(ctx context.Context, c *model.DeactivationChecklist) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if c.ID == 0 {
		c.ID = int64(len(r.checklists) + 1)
	}
	r.checklists[c.ID] = c
	return nil
}

func (r *mockDeactivationRepo) GetChecklistByDeactivationID(ctx context.Context, deactivationID int64) (*model.DeactivationChecklist, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, c := range r.checklists {
		if c.DeactivationID == deactivationID {
			return c, nil
		}
	}
	return nil, nil
}

func (r *mockDeactivationRepo) GetCooldown(ctx context.Context, riderID int64) (*model.DeactivationCooldown, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if c, ok := r.cooldowns[riderID]; ok {
		return c, nil
	}
	return nil, nil
}

func (r *mockDeactivationRepo) SetCooldown(ctx context.Context, riderID int64, deactivatedAt, cooldownEndAt *time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cooldowns[riderID] = &model.DeactivationCooldown{
		RiderID:           riderID,
		LastDeactivatedAt: deactivatedAt,
		CooldownEndAt:     cooldownEndAt,
		CanReactivate:     cooldownEndAt == nil || cooldownEndAt.Before(time.Now()),
	}
	return nil
}

func (r *mockDeactivationRepo) CreateHistory(ctx context.Context, h *model.DeactivationHistory) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if h.ID == 0 {
		h.ID = int64(len(r.histories) + 1)
	}
	r.histories = append(r.histories, h)
	return nil
}

func (r *mockDeactivationRepo) GetHistoryByRiderID(ctx context.Context, riderID int64) ([]*model.DeactivationHistory, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*model.DeactivationHistory
	for _, h := range r.histories {
		if h.RiderID == riderID {
			result = append(result, h)
		}
	}
	return result, nil
}

func (r *mockDeactivationRepo) HasPendingOrders(ctx context.Context, riderID int64) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.pendingOrders[riderID], nil
}

func (r *mockDeactivationRepo) HasPendingIncome(ctx context.Context, riderID int64) (bool, int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.pendingIncome[riderID] > 0, r.pendingIncome[riderID], nil
}

func (r *mockDeactivationRepo) HasPendingComplaint(ctx context.Context, riderID int64) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.pendingComplaint[riderID], nil
}

func (r *mockDeactivationRepo) HasEquipmentDebt(ctx context.Context, riderID int64) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.equipmentDebt[riderID], nil
}

func (r *mockDeactivationRepo) HasViolationRecord(ctx context.Context, riderID int64) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.violationRecord[riderID], nil
}

func (r *mockDeactivationRepo) IsBalanceClear(ctx context.Context, riderID int64) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.balanceClear[riderID], nil
}

func (r *mockDeactivationRepo) CalculatePendingIncome(ctx context.Context, riderID int64) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.pendingIncome[riderID], nil
}

func (r *mockDeactivationRepo) CalculatePendingPenalty(ctx context.Context, riderID int64) (int64, error) {
	return 0, nil
}

func (r *mockDeactivationRepo) DeactivateRider(ctx context.Context, riderID int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.riderStatus[riderID] = 2 // 已注销
	return nil
}

func (r *mockDeactivationRepo) GetDeactivationStatistics(ctx context.Context, startDate, endDate string) (*data.DeactivationStatisticsResult, error) {
	return &data.DeactivationStatisticsResult{}, nil
}

// ==================== 测试用例 ====================

// TestApplyDeactivation_Success 测试正常申请注销
func TestApplyDeactivation_Success(t *testing.T) {
	repo := newMockDeactivationRepo()
	uc := NewDeactivationUsecase(repo, log.DefaultLogger)
	ctx := context.Background()

	// 设置骑手状态为正常（无阻碍条件）
	repo.balanceClear[100] = true

	req := &ApplyDeactivationRequest{
		RiderID:      100,
		ReasonType:   model.DeactivationReasonPersonal,
		ReasonDetail: "找到新工作",
		Phone:        "13800138000",
		IP:           "192.168.1.1",
		DeviceID:     "device_001",
	}

	resp, err := uc.ApplyDeactivation(ctx, req)
	if err != nil {
		t.Fatalf("ApplyDeactivation failed: %v", err)
	}

	if resp.DeactivationID == 0 {
		t.Error("DeactivationID should not be 0")
	}

	if resp.Status != model.DeactivationStatusPending {
		t.Errorf("Status should be PENDING, got %d", resp.Status)
	}

	if !resp.Checklist.CanDeactivate {
		t.Error("Should be able to deactivate")
	}

	t.Logf("✅ ApplyDeactivation success: id=%d", resp.DeactivationID)
}

// TestApplyDeactivation_HasPendingOrders 测试有待处理订单
func TestApplyDeactivation_HasPendingOrders(t *testing.T) {
	repo := newMockDeactivationRepo()
	uc := NewDeactivationUsecase(repo, log.DefaultLogger)
	ctx := context.Background()

	// 设置有 pending orders
	repo.pendingOrders[100] = true
	repo.balanceClear[100] = true

	req := &ApplyDeactivationRequest{
		RiderID:    100,
		ReasonType: model.DeactivationReasonPersonal,
		Phone:      "13800138000",
	}

	resp, err := uc.ApplyDeactivation(ctx, req)
	if err != nil {
		t.Fatalf("ApplyDeactivation failed: %v", err)
	}

	if resp.Checklist.CanDeactivate {
		t.Error("Should not be able to deactivate with pending orders")
	}

	if !resp.Checklist.HasPendingOrders {
		t.Error("HasPendingOrders should be true")
	}

	t.Logf("✅ Pending orders check passed")
}

// TestApplyDeactivation_HasPendingIncome 测试有待结算收入
func TestApplyDeactivation_HasPendingIncome(t *testing.T) {
	repo := newMockDeactivationRepo()
	uc := NewDeactivationUsecase(repo, log.DefaultLogger)
	ctx := context.Background()

	// 设置有 pending income (超过1000元)
	repo.pendingIncome[100] = 150000 // 1500元
	repo.balanceClear[100] = true

	req := &ApplyDeactivationRequest{
		RiderID:    100,
		ReasonType: model.DeactivationReasonPersonal,
		Phone:      "13800138000",
	}

	resp, err := uc.ApplyDeactivation(ctx, req)
	if err != nil {
		t.Fatalf("ApplyDeactivation failed: %v", err)
	}

	if resp.Checklist.CanDeactivate {
		t.Error("Should not be able to deactivate with high pending income")
	}

	if !resp.Checklist.HasPendingIncome {
		t.Error("HasPendingIncome should be true")
	}

	t.Logf("✅ Pending income check passed")
}

// TestApplyDeactivation_HasEquipmentDebt 测试有装备欠款
func TestApplyDeactivation_HasEquipmentDebt(t *testing.T) {
	repo := newMockDeactivationRepo()
	uc := NewDeactivationUsecase(repo, log.DefaultLogger)
	ctx := context.Background()

	// 设置有装备欠款
	repo.equipmentDebt[100] = true
	repo.balanceClear[100] = true

	req := &ApplyDeactivationRequest{
		RiderID:    100,
		ReasonType: model.DeactivationReasonPersonal,
		Phone:      "13800138000",
	}

	resp, err := uc.ApplyDeactivation(ctx, req)
	if err != nil {
		t.Fatalf("ApplyDeactivation failed: %v", err)
	}

	if resp.Checklist.CanDeactivate {
		t.Error("Should not be able to deactivate with equipment debt")
	}

	if !resp.Checklist.HasEquipmentDebt {
		t.Error("HasEquipmentDebt should be true")
	}

	t.Logf("✅ Equipment debt check passed")
}

// TestApplyDeactivation_DuplicateApplication 测试重复申请
func TestApplyDeactivation_DuplicateApplication(t *testing.T) {
	repo := newMockDeactivationRepo()
	uc := NewDeactivationUsecase(repo, log.DefaultLogger)
	ctx := context.Background()

	repo.balanceClear[100] = true

	// 第一次申请
	req := &ApplyDeactivationRequest{
		RiderID:    100,
		ReasonType: model.DeactivationReasonPersonal,
		Phone:      "13800138000",
	}
	_, err := uc.ApplyDeactivation(ctx, req)
	if err != nil {
		t.Fatalf("First application failed: %v", err)
	}

	// 第二次申请应该失败
	_, err = uc.ApplyDeactivation(ctx, req)
	if err == nil {
		t.Error("Second application should fail")
	}

	t.Logf("✅ Duplicate application prevention passed")
}

// TestApplyDeactivation_InCooldown 测试冷却期内
func TestApplyDeactivation_InCooldown(t *testing.T) {
	repo := newMockDeactivationRepo()
	uc := NewDeactivationUsecase(repo, log.DefaultLogger)
	ctx := context.Background()

	// 设置冷却期
	cooldownEnd := time.Now().AddDate(0, 0, 15) // 15天后结束
	repo.SetCooldown(ctx, 100, nil, &cooldownEnd)

	req := &ApplyDeactivationRequest{
		RiderID:    100,
		ReasonType: model.DeactivationReasonPersonal,
		Phone:      "13800138000",
	}

	_, err := uc.ApplyDeactivation(ctx, req)
	if err == nil {
		t.Error("Should not allow application during cooldown")
	}

	t.Logf("✅ Cooldown check passed")
}

// TestCancelDeactivation 测试取消注销申请
func TestCancelDeactivation(t *testing.T) {
	repo := newMockDeactivationRepo()
	uc := NewDeactivationUsecase(repo, log.DefaultLogger)
	ctx := context.Background()

	repo.balanceClear[100] = true

	// 先申请
	req := &ApplyDeactivationRequest{
		RiderID:    100,
		ReasonType: model.DeactivationReasonPersonal,
		Phone:      "13800138000",
	}
	resp, _ := uc.ApplyDeactivation(ctx, req)

	// 再取消
	err := uc.CancelDeactivation(ctx, 100, resp.DeactivationID)
	if err != nil {
		t.Fatalf("CancelDeactivation failed: %v", err)
	}

	// 验证状态
	deactivation, _, _ := uc.GetDeactivationStatus(ctx, 100)
	if deactivation.Status != model.DeactivationStatusRejected {
		t.Error("Status should be REJECTED after cancel")
	}

	t.Logf("✅ CancelDeactivation passed")
}

// TestReviewDeactivation_Approve 测试审核通过
func TestReviewDeactivation_Approve(t *testing.T) {
	repo := newMockDeactivationRepo()
	uc := NewDeactivationUsecase(repo, log.DefaultLogger)
	ctx := context.Background()

	repo.balanceClear[100] = true

	// 申请
	req := &ApplyDeactivationRequest{
		RiderID:    100,
		ReasonType: model.DeactivationReasonPersonal,
		Phone:      "13800138000",
	}
	resp, _ := uc.ApplyDeactivation(ctx, req)

	// 审核通过
	reviewReq := &ReviewDeactivationRequest{
		DeactivationID: resp.DeactivationID,
		ReviewerID:     1,
		Action:         "APPROVE",
		Remark:         "同意注销",
	}
	err := uc.ReviewDeactivation(ctx, reviewReq)
	if err != nil {
		t.Fatalf("ReviewDeactivation failed: %v", err)
	}

	// 验证状态
	deactivation, _, _ := uc.GetDeactivationStatus(ctx, 100)
	if deactivation.Status != model.DeactivationStatusApproved {
		t.Errorf("Status should be APPROVED, got %d", deactivation.Status)
	}

	t.Logf("✅ ReviewDeactivation approve passed")
}

// TestReviewDeactivation_Reject 测试审核拒绝
func TestReviewDeactivation_Reject(t *testing.T) {
	repo := newMockDeactivationRepo()
	uc := NewDeactivationUsecase(repo, log.DefaultLogger)
	ctx := context.Background()

	repo.balanceClear[100] = true

	// 申请
	req := &ApplyDeactivationRequest{
		RiderID:    100,
		ReasonType: model.DeactivationReasonPersonal,
		Phone:      "13800138000",
	}
	resp, _ := uc.ApplyDeactivation(ctx, req)

	// 审核拒绝
	reviewReq := &ReviewDeactivationRequest{
		DeactivationID: resp.DeactivationID,
		ReviewerID:     1,
		Action:         "REJECT",
		Remark:         "有未处理订单",
	}
	err := uc.ReviewDeactivation(ctx, reviewReq)
	if err != nil {
		t.Fatalf("ReviewDeactivation failed: %v", err)
	}

	// 验证状态
	deactivation, _, _ := uc.GetDeactivationStatus(ctx, 100)
	if deactivation.Status != model.DeactivationStatusRejected {
		t.Errorf("Status should be REJECTED, got %d", deactivation.Status)
	}

	t.Logf("✅ ReviewDeactivation reject passed")
}

// TestGetDeactivationStatus 测试获取注销状态
func TestGetDeactivationStatus(t *testing.T) {
	repo := newMockDeactivationRepo()
	uc := NewDeactivationUsecase(repo, log.DefaultLogger)
	ctx := context.Background()

	repo.balanceClear[100] = true

	// 申请
	req := &ApplyDeactivationRequest{
		RiderID:    100,
		ReasonType: model.DeactivationReasonPersonal,
		Phone:      "13800138000",
	}
	_, _ = uc.ApplyDeactivation(ctx, req)

	// 获取状态
	deactivation, checklist, err := uc.GetDeactivationStatus(ctx, 100)
	if err != nil {
		t.Fatalf("GetDeactivationStatus failed: %v", err)
	}

	if deactivation == nil {
		t.Fatal("Deactivation should not be nil")
	}

	if deactivation.RiderID != 100 {
		t.Errorf("RiderID should be 100, got %d", deactivation.RiderID)
	}

	if checklist == nil {
		t.Error("Checklist should not be nil")
	}

	t.Logf("✅ GetDeactivationStatus passed")
}

// TestCompleteDeactivationFlow 测试完整的注销流程
func TestCompleteDeactivationFlow(t *testing.T) {
	repo := newMockDeactivationRepo()
	uc := NewDeactivationUsecase(repo, log.DefaultLogger)
	ctx := context.Background()

	// 设置骑手状态正常
	repo.balanceClear[100] = true
	repo.pendingIncome[100] = 5000 // 50元待结算

	// 步骤1: 申请注销
	t.Log("步骤1: 申请注销")
	req := &ApplyDeactivationRequest{
		RiderID:      100,
		ReasonType:   model.DeactivationReasonOtherJob,
		ReasonDetail: "找到全职工作",
		Phone:        "13800138000",
		IP:           "192.168.1.1",
		DeviceID:     "device_001",
	}
	resp, err := uc.ApplyDeactivation(ctx, req)
	if err != nil {
		t.Fatalf("ApplyDeactivation failed: %v", err)
	}
	t.Logf("✅ 申请成功，ID=%d", resp.DeactivationID)

	// 步骤2: 获取状态
	t.Log("步骤2: 获取注销状态")
	deactivation, _, _ := uc.GetDeactivationStatus(ctx, 100)
	if deactivation.Status != model.DeactivationStatusPending {
		t.Fatal("Status should be PENDING")
	}
	t.Logf("✅ 状态为待审核")

	// 步骤3: 审核通过
	t.Log("步骤3: 审核通过")
	reviewReq := &ReviewDeactivationRequest{
		DeactivationID: resp.DeactivationID,
		ReviewerID:     1,
		Action:         "APPROVE",
		Remark:         "审核通过",
	}
	err = uc.ReviewDeactivation(ctx, reviewReq)
	if err != nil {
		t.Fatalf("ReviewDeactivation failed: %v", err)
	}
	t.Logf("✅ 审核通过")

	t.Logf("✅ Complete deactivation flow test passed")
}
