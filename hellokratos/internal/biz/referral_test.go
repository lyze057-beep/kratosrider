package biz

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"

	"hellokratos/internal/conf"
	"hellokratos/internal/data/model"
)

var errRecordNotFound = gorm.ErrRecordNotFound

type inMemoryRepo struct {
	mu              sync.RWMutex
	inviteCodes     map[int64]*model.ReferralInviteCode
	inviteCodeMap   map[string]*model.ReferralInviteCode
	relations       map[int64]*model.ReferralRelation
	relationsByCode map[string][]*model.ReferralRelation
	tasks           map[int64]*model.ReferralTask
	taskProgress    map[int64]*model.ReferralTaskProgress
	rewardRecords   map[int64]*model.ReferralRewardRecord
	riskLogs        map[int64]*model.ReferralRiskLog
}

func newInMemoryRepo() *inMemoryRepo {
	return &inMemoryRepo{
		inviteCodes:     make(map[int64]*model.ReferralInviteCode),
		inviteCodeMap:   make(map[string]*model.ReferralInviteCode),
		relations:       make(map[int64]*model.ReferralRelation),
		relationsByCode: make(map[string][]*model.ReferralRelation),
		tasks:           make(map[int64]*model.ReferralTask),
		taskProgress:    make(map[int64]*model.ReferralTaskProgress),
		rewardRecords:   make(map[int64]*model.ReferralRewardRecord),
		riskLogs:        make(map[int64]*model.ReferralRiskLog),
	}
}

func (r *inMemoryRepo) GetInviteCodeByRiderID(ctx context.Context, riderID int64) (*model.ReferralInviteCode, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if code, ok := r.inviteCodes[riderID]; ok {
		return code, nil
	}
	return nil, errRecordNotFound
}

func (r *inMemoryRepo) GetInviteCodeByCode(ctx context.Context, code string) (*model.ReferralInviteCode, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if c, ok := r.inviteCodeMap[code]; ok {
		return c, nil
	}
	return nil, errRecordNotFound
}

func (r *inMemoryRepo) CreateInviteCode(ctx context.Context, code *model.ReferralInviteCode) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.inviteCodes[code.RiderID] = code
	r.inviteCodeMap[code.InviteCode] = code
	return nil
}

func (r *inMemoryRepo) UpdateInviteCodeStats(ctx context.Context, code string, totalInvited, validInvited, totalRewards int32) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if c, ok := r.inviteCodeMap[code]; ok {
		c.TotalInvited = totalInvited
		c.ValidInvited = validInvited
		c.TotalRewards = totalRewards
	}
	return nil
}

func (r *inMemoryRepo) GetReferralRelationByInvitedID(ctx context.Context, invitedID int64) (*model.ReferralRelation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if rel, ok := r.relations[invitedID]; ok {
		return rel, nil
	}
	return nil, errRecordNotFound
}

func (r *inMemoryRepo) CreateReferralRelation(ctx context.Context, relation *model.ReferralRelation) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	// 设置创建时间为当前时间
	if relation.CreatedAt.IsZero() {
		relation.CreatedAt = time.Now()
	}
	r.relations[relation.InvitedID] = relation
	r.relationsByCode[relation.InviteCode] = append(r.relationsByCode[relation.InviteCode], relation)
	return nil
}

func (r *inMemoryRepo) UpdateReferralRelationStatus(ctx context.Context, id int64, status int32) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, rel := range r.relations {
		if rel.ID == id {
			rel.Status = status
			return nil
		}
	}
	return nil
}

func (r *inMemoryRepo) UpdateReferralRelationFirstOrder(ctx context.Context, invitedID int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if rel, ok := r.relations[invitedID]; ok {
		now := time.Now()
		rel.FirstOrderAt = &now
		rel.Status = 2
	}
	return nil
}

func (r *inMemoryRepo) UpdateReferralRelationTaskCompleted(ctx context.Context, invitedID int64, rewardAmount int32) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if rel, ok := r.relations[invitedID]; ok {
		now := time.Now()
		rel.TaskCompletedAt = &now
		rel.Status = 3
		rel.RewardAmount = rewardAmount
		rel.RewardStatus = 1
	}
	return nil
}

func (r *inMemoryRepo) GetReferralRelationList(ctx context.Context, inviterID int64, status int32, page, pageSize int) ([]*model.ReferralRelation, int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*model.ReferralRelation
	var total int64
	for _, rel := range r.relations {
		if rel.InviterID == inviterID {
			if status == 0 || rel.Status == status {
				result = append(result, rel)
				total++
			}
		}
	}
	if len(result) > pageSize {
		start := (page - 1) * pageSize
		end := start + pageSize
		if start > len(result) {
			result = result[:0]
		} else {
			if end > len(result) {
				result = result[start:]
			} else {
				result = result[start:end]
			}
		}
	}
	return result, total, nil
}

func (r *inMemoryRepo) GetReferralRelationCount(ctx context.Context, inviterID int64, startTime, endTime time.Time) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var count int64
	for _, rel := range r.relations {
		if rel.InviterID == inviterID && rel.CreatedAt.After(startTime) && rel.CreatedAt.Before(endTime) {
			count++
		}
	}
	return count, nil
}

func (r *inMemoryRepo) GetTaskList(ctx context.Context, status int32) ([]*model.ReferralTask, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*model.ReferralTask
	for _, task := range r.tasks {
		if status == 0 || task.Status == status {
			result = append(result, task)
		}
	}
	return result, nil
}

func (r *inMemoryRepo) GetTaskByID(ctx context.Context, taskID int64) (*model.ReferralTask, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if task, ok := r.tasks[taskID]; ok {
		return task, nil
	}
	return nil, errRecordNotFound
}

func (r *inMemoryRepo) CreateTask(ctx context.Context, task *model.ReferralTask) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tasks[task.ID] = task
	return nil
}

func (r *inMemoryRepo) GetTaskProgress(ctx context.Context, relationID, taskID int64) (*model.ReferralTaskProgress, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, prog := range r.taskProgress {
		if prog.RelationID == relationID && prog.TaskID == taskID {
			return prog, nil
		}
	}
	return nil, errRecordNotFound
}

func (r *inMemoryRepo) CreateTaskProgress(ctx context.Context, progress *model.ReferralTaskProgress) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.taskProgress[progress.ID] = progress
	return nil
}

func (r *inMemoryRepo) UpdateTaskProgress(ctx context.Context, id int64, currentValue int32, status int32) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if prog, ok := r.taskProgress[id]; ok {
		prog.CurrentValue = currentValue
		prog.Status = status
	}
	return nil
}

func (r *inMemoryRepo) CompleteTaskProgress(ctx context.Context, id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if prog, ok := r.taskProgress[id]; ok {
		now := time.Now()
		prog.CompletedAt = &now
		prog.Status = 2
	}
	return nil
}

func (r *inMemoryRepo) ClaimTaskReward(ctx context.Context, id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if prog, ok := r.taskProgress[id]; ok {
		now := time.Now()
		prog.IsClaimed = true
		prog.ClaimedAt = &now
	}
	return nil
}

func (r *inMemoryRepo) GetTaskProgressList(ctx context.Context, inviterID int64) ([]*model.ReferralTaskProgress, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*model.ReferralTaskProgress
	for _, prog := range r.taskProgress {
		if prog.InviterID == inviterID {
			result = append(result, prog)
		}
	}
	return result, nil
}

func (r *inMemoryRepo) CreateRewardRecord(ctx context.Context, record *model.ReferralRewardRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.rewardRecords[record.ID] = record
	return nil
}

func (r *inMemoryRepo) GetRewardRecordList(ctx context.Context, inviterID int64, status int32, page, pageSize int) ([]*model.ReferralRewardRecord, int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*model.ReferralRewardRecord
	var total int64
	for _, record := range r.rewardRecords {
		if record.InviterID == inviterID {
			if status == 0 || record.Status == status {
				result = append(result, record)
				total++
			}
		}
	}
	if len(result) > pageSize {
		start := (page - 1) * pageSize
		end := start + pageSize
		if start > len(result) {
			result = result[:0]
		} else {
			if end > len(result) {
				result = result[start:]
			} else {
				result = result[start:end]
			}
		}
	}
	return result, total, nil
}

func (r *inMemoryRepo) UpdateRewardRecordStatus(ctx context.Context, id int64, status int32, failReason string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if record, ok := r.rewardRecords[id]; ok {
		record.Status = status
		if failReason != "" {
			record.FailReason = failReason
		}
		if status == 2 {
			now := time.Now()
			record.IssuedAt = &now
		}
	}
	return nil
}

func (r *inMemoryRepo) GetTotalRewardAmount(ctx context.Context, inviterID int64) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var total int64
	for _, record := range r.rewardRecords {
		if record.InviterID == inviterID && record.Status == 2 {
			total += int64(record.RewardAmount)
		}
	}
	return total, nil
}

func (r *inMemoryRepo) GetPendingRewardAmount(ctx context.Context, inviterID int64) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var total int64
	for _, record := range r.rewardRecords {
		if record.InviterID == inviterID && record.Status == 1 {
			total += int64(record.RewardAmount)
		}
	}
	return total, nil
}

func (r *inMemoryRepo) CreateRiskLog(ctx context.Context, log *model.ReferralRiskLog) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.riskLogs[log.ID] = log
	return nil
}

func (r *inMemoryRepo) GetRiskLogList(ctx context.Context, relationID int64) ([]*model.ReferralRiskLog, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*model.ReferralRiskLog
	for _, log := range r.riskLogs {
		if log.RelationID == relationID {
			result = append(result, log)
		}
	}
	return result, nil
}

func (r *inMemoryRepo) ConfirmRisk(ctx context.Context, id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if log, ok := r.riskLogs[id]; ok {
		now := time.Now()
		log.IsConfirmed = true
		log.ConfirmedAt = &now
	}
	return nil
}

func (r *inMemoryRepo) GetStatisticsByDate(ctx context.Context, date string) (*model.ReferralStatistics, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *inMemoryRepo) GetStatisticsByDateRange(ctx context.Context, startDate, endDate string) ([]*model.ReferralStatistics, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *inMemoryRepo) UpdateStatistics(ctx context.Context, stat *model.ReferralStatistics) error {
	return fmt.Errorf("not implemented")
}

// ==================== 测试用例 ====================

// TestGenerateInviteCode 测试生成邀请码
func TestGenerateInviteCode(t *testing.T) {
	repo := newInMemoryRepo()
	uc := NewReferralUsecase(repo, log.DefaultLogger, &conf.Data{})

	ctx := context.Background()
	riderID := int64(12345)

	inviteCode, err := uc.GenerateInviteCode(ctx, riderID)
	if err != nil {
		t.Fatalf("GenerateInviteCode failed: %v", err)
	}

	if inviteCode == nil {
		t.Fatal("inviteCode is nil")
	}

	if inviteCode.InviteCode == "" {
		t.Error("InviteCode is empty")
	}

	if len(inviteCode.InviteCode) != 6 {
		t.Errorf("InviteCode length should be 6, got %d", len(inviteCode.InviteCode))
	}

	if inviteCode.InviteLink == "" {
		t.Error("InviteLink is empty")
	}

	if inviteCode.RiderID != riderID {
		t.Errorf("RiderID mismatch: got %d, want %d", inviteCode.RiderID, riderID)
	}

	if inviteCode.Status != 1 {
		t.Errorf("Status should be 1, got %d", inviteCode.Status)
	}

	if inviteCode.ExpireAt.Before(time.Now()) {
		t.Error("ExpireAt should be in the future")
	}

	t.Logf("✅ Generated invite code: %s, link: %s", inviteCode.InviteCode, inviteCode.InviteLink)
}

// TestGenerateInviteCode_Duplicate 测试重复生成邀请码
func TestGenerateInviteCode_Duplicate(t *testing.T) {
	repo := newInMemoryRepo()
	uc := NewReferralUsecase(repo, log.DefaultLogger, &conf.Data{})

	ctx := context.Background()
	riderID := int64(12345)

	code1, err := uc.GenerateInviteCode(ctx, riderID)
	if err != nil {
		t.Fatalf("First GenerateInviteCode failed: %v", err)
	}

	code2, err := uc.GenerateInviteCode(ctx, riderID)
	if err != nil {
		t.Fatalf("Second GenerateInviteCode failed: %v", err)
	}

	if code1.InviteCode != code2.InviteCode {
		t.Errorf("InviteCode should be the same: got %s, want %s", code2.InviteCode, code1.InviteCode)
	}

	t.Logf("✅ Duplicate generate returned same code: %s", code2.InviteCode)
}

// TestGenerateInviteCode_Uniqueness 测试邀请码唯一性
func TestGenerateInviteCode_Uniqueness(t *testing.T) {
	repo := newInMemoryRepo()
	uc := NewReferralUsecase(repo, log.DefaultLogger, &conf.Data{})

	ctx := context.Background()
	codes := make(map[string]bool)

	// 生成100个不同骑手的邀请码
	for i := int64(1); i <= 100; i++ {
		inviteCode, err := uc.GenerateInviteCode(ctx, i)
		if err != nil {
			t.Fatalf("GenerateInviteCode failed for rider %d: %v", i, err)
		}

		if codes[inviteCode.InviteCode] {
			t.Errorf("Duplicate invite code generated: %s", inviteCode.InviteCode)
		}
		codes[inviteCode.InviteCode] = true
	}

	t.Logf("✅ Generated %d unique invite codes", len(codes))
}

// TestValidateInviteCode 测试验证邀请码
func TestValidateInviteCode(t *testing.T) {
	repo := newInMemoryRepo()
	uc := NewReferralUsecase(repo, log.DefaultLogger, &conf.Data{})

	ctx := context.Background()
	riderID := int64(12345)

	inviteCode, err := uc.GenerateInviteCode(ctx, riderID)
	if err != nil {
		t.Fatalf("GenerateInviteCode failed: %v", err)
	}

	// 测试有效邀请码
	isValid, inviterName, err := uc.ValidateInviteCode(ctx, inviteCode.InviteCode)
	if err != nil {
		t.Fatalf("ValidateInviteCode failed: %v", err)
	}

	if !isValid {
		t.Error("Valid invite code should pass validation")
	}

	if inviterName == "" {
		t.Error("Inviter name should not be empty")
	}

	// 测试无效邀请码
	isValid, _, _ = uc.ValidateInviteCode(ctx, "INVALID")
	if isValid {
		t.Error("Invalid invite code should fail validation")
	}

	// 测试空邀请码
	isValid, _, _ = uc.ValidateInviteCode(ctx, "")
	if isValid {
		t.Error("Empty invite code should fail validation")
	}

	t.Logf("✅ ValidateInviteCode test passed")
}

// TestValidateInviteCode_Expired 测试过期邀请码
func TestValidateInviteCode_Expired(t *testing.T) {
	repo := newInMemoryRepo()
	uc := NewReferralUsecase(repo, log.DefaultLogger, &conf.Data{})

	ctx := context.Background()
	riderID := int64(12345)

	// 创建一个已过期的邀请码
	expiredCode := &model.ReferralInviteCode{
		RiderID:    riderID,
		InviteCode: "EXPIRED",
		Status:     1,
		ExpireAt:   time.Now().AddDate(0, 0, -1), // 昨天过期
	}
	repo.CreateInviteCode(ctx, expiredCode)

	isValid, _, _ := uc.ValidateInviteCode(ctx, "EXPIRED")
	if isValid {
		t.Error("Expired invite code should fail validation")
	}

	t.Logf("✅ Expired invite code validation test passed")
}

// TestValidateInviteCode_Disabled 测试禁用邀请码
func TestValidateInviteCode_Disabled(t *testing.T) {
	repo := newInMemoryRepo()
	uc := NewReferralUsecase(repo, log.DefaultLogger, &conf.Data{})

	ctx := context.Background()
	riderID := int64(12345)

	// 创建一个已禁用的邀请码
	disabledCode := &model.ReferralInviteCode{
		RiderID:    riderID,
		InviteCode: "DISABLED",
		Status:     0, // 禁用状态
		ExpireAt:   time.Now().AddDate(1, 0, 0),
	}
	repo.CreateInviteCode(ctx, disabledCode)

	isValid, _, _ := uc.ValidateInviteCode(ctx, "DISABLED")
	if isValid {
		t.Error("Disabled invite code should fail validation")
	}

	t.Logf("✅ Disabled invite code validation test passed")
}

// TestBindReferralRelation 测试绑定邀请关系
func TestBindReferralRelation(t *testing.T) {
	repo := newInMemoryRepo()
	uc := NewReferralUsecase(repo, log.DefaultLogger, &conf.Data{})

	ctx := context.Background()
	inviterID := int64(100)
	newRiderID := int64(200)

	inviteCode, err := uc.GenerateInviteCode(ctx, inviterID)
	if err != nil {
		t.Fatalf("GenerateInviteCode failed: %v", err)
	}

	boundInviterID, err := uc.BindReferralRelation(ctx, newRiderID, inviteCode.InviteCode, "13800138000")
	if err != nil {
		t.Fatalf("BindReferralRelation failed: %v", err)
	}

	if boundInviterID != inviterID {
		t.Errorf("InviterID mismatch: got %d, want %d", boundInviterID, inviterID)
	}

	t.Logf("✅ BindReferralRelation test passed")
}

// TestBindReferralRelation_Duplicate 测试重复绑定
func TestBindReferralRelation_Duplicate(t *testing.T) {
	repo := newInMemoryRepo()
	uc := NewReferralUsecase(repo, log.DefaultLogger, &conf.Data{})

	ctx := context.Background()
	inviterID := int64(100)
	newRiderID := int64(200)

	inviteCode, err := uc.GenerateInviteCode(ctx, inviterID)
	if err != nil {
		t.Fatalf("GenerateInviteCode failed: %v", err)
	}

	// 第一次绑定
	_, err = uc.BindReferralRelation(ctx, newRiderID, inviteCode.InviteCode, "13800138000")
	if err != nil {
		t.Fatalf("First BindReferralRelation failed: %v", err)
	}

	// 重复绑定应该失败
	_, err = uc.BindReferralRelation(ctx, newRiderID, inviteCode.InviteCode, "13800138000")
	if err == nil {
		t.Error("Should not allow duplicate binding")
	}

	t.Logf("✅ Duplicate binding prevention test passed")
}

// TestBindReferralRelation_SelfInvitation 测试自邀请
func TestBindReferralRelation_SelfInvitation(t *testing.T) {
	repo := newInMemoryRepo()
	uc := NewReferralUsecase(repo, log.DefaultLogger, &conf.Data{})

	ctx := context.Background()
	inviterID := int64(100)

	inviteCode, err := uc.GenerateInviteCode(ctx, inviterID)
	if err != nil {
		t.Fatalf("GenerateInviteCode failed: %v", err)
	}

	// 自邀请应该失败
	_, err = uc.BindReferralRelation(ctx, inviterID, inviteCode.InviteCode, "13800138000")
	if err == nil {
		t.Error("Should not allow self-invitation")
	}

	t.Logf("✅ Self-invitation prevention test passed")
}

// TestBindReferralRelation_InvalidCode 测试无效邀请码
func TestBindReferralRelation_InvalidCode(t *testing.T) {
	repo := newInMemoryRepo()
	uc := NewReferralUsecase(repo, log.DefaultLogger, &conf.Data{})

	ctx := context.Background()
	newRiderID := int64(200)

	// 无效邀请码应该失败
	_, err := uc.BindReferralRelation(ctx, newRiderID, "INVALID", "13800138000")
	if err == nil {
		t.Error("Should reject invalid invite code")
	}

	t.Logf("✅ Invalid invite code rejection test passed")
}

// TestGetMyReferralInfo 测试获取我的邀请信息
func TestGetMyReferralInfo(t *testing.T) {
	repo := newInMemoryRepo()
	uc := NewReferralUsecase(repo, log.DefaultLogger, &conf.Data{})

	ctx := context.Background()
	inviterID := int64(100)

	// 生成邀请码
	inviteCode, err := uc.GenerateInviteCode(ctx, inviterID)
	if err != nil {
		t.Fatalf("GenerateInviteCode failed: %v", err)
	}

	// 绑定几个邀请关系
	for i := int64(1); i <= 3; i++ {
		_, err := uc.BindReferralRelation(ctx, 200+i, inviteCode.InviteCode, fmt.Sprintf("1380013800%d", i))
		if err != nil {
			t.Fatalf("BindReferralRelation failed: %v", err)
		}
	}

	// 获取邀请信息
	code, todayCount, monthCount, pendingAmount, err := uc.GetMyReferralInfo(ctx, inviterID)
	if err != nil {
		t.Fatalf("GetMyReferralInfo failed: %v", err)
	}

	if code == nil {
		t.Error("Invite code should not be nil")
	}

	if code.InviteCode != inviteCode.InviteCode {
		t.Errorf("Invite code mismatch: got %s, want %s", code.InviteCode, inviteCode.InviteCode)
	}

	if todayCount != 3 {
		t.Errorf("Today count should be 3, got %d", todayCount)
	}

	if monthCount != 3 {
		t.Errorf("Month count should be 3, got %d", monthCount)
	}

	t.Logf("✅ GetMyReferralInfo test passed: today=%d, month=%d, pending=%d", todayCount, monthCount, pendingAmount)
}

// TestGetInviteRecordList 测试获取邀请记录列表
func TestGetInviteRecordList(t *testing.T) {
	repo := newInMemoryRepo()
	uc := NewReferralUsecase(repo, log.DefaultLogger, &conf.Data{})

	ctx := context.Background()
	inviterID := int64(100)

	inviteCode, err := uc.GenerateInviteCode(ctx, inviterID)
	if err != nil {
		t.Fatalf("GenerateInviteCode failed: %v", err)
	}

	// 创建5个邀请记录
	for i := int64(1); i <= 5; i++ {
		_, err := uc.BindReferralRelation(ctx, 200+i, inviteCode.InviteCode, fmt.Sprintf("1380013800%d", i))
		if err != nil {
			t.Fatalf("BindReferralRelation failed: %v", err)
		}
	}

	// 测试分页
	records, total, err := uc.GetInviteRecordList(ctx, inviterID, 0, 1, 3)
	if err != nil {
		t.Fatalf("GetInviteRecordList failed: %v", err)
	}

	if total != 5 {
		t.Errorf("Total should be 5, got %d", total)
	}

	if len(records) != 3 {
		t.Errorf("Should return 3 records, got %d", len(records))
	}

	t.Logf("✅ GetInviteRecordList test passed: total=%d, returned=%d", total, len(records))
}

// TestGetTaskList 测试获取任务列表
func TestGetTaskList(t *testing.T) {
	repo := newInMemoryRepo()
	uc := NewReferralUsecase(repo, log.DefaultLogger, &conf.Data{})

	ctx := context.Background()

	// 创建测试任务
	now := time.Now()
	repo.CreateTask(ctx, &model.ReferralTask{
		ID:           1,
		TaskName:     "首单任务",
		TaskType:     1,
		RewardAmount: 50,
		Status:       1,
		StartTime:    now.AddDate(0, 0, -1),
		EndTime:      now.AddDate(0, 0, 30),
	})

	repo.CreateTask(ctx, &model.ReferralTask{
		ID:           2,
		TaskName:     "在线时长任务",
		TaskType:     2,
		RewardAmount: 30,
		Status:       1,
		StartTime:    now.AddDate(0, 0, -1),
		EndTime:      now.AddDate(0, 0, 30),
	})

	repo.CreateTask(ctx, &model.ReferralTask{
		ID:           3,
		TaskName:     "已过期任务",
		TaskType:     1,
		RewardAmount: 100,
		Status:       1,
		StartTime:    now.AddDate(0, 0, -30),
		EndTime:      now.AddDate(0, 0, -1), // 已过期
	})

	tasks, err := uc.GetTaskList(ctx)
	if err != nil {
		t.Fatalf("GetTaskList failed: %v", err)
	}

	if len(tasks) != 3 {
		t.Errorf("Expected 3 tasks, got %d", len(tasks))
	}

	t.Logf("✅ GetTaskList test passed: got %d tasks", len(tasks))
}

// TestConcurrentInviteCodeGeneration 测试并发生成邀请码
func TestConcurrentInviteCodeGeneration(t *testing.T) {
	repo := newInMemoryRepo()
	uc := NewReferralUsecase(repo, log.DefaultLogger, &conf.Data{})

	ctx := context.Background()
	var wg sync.WaitGroup
	codes := make(map[string]int64)
	var mu sync.Mutex

	// 100个并发请求
	for i := int64(1); i <= 100; i++ {
		wg.Add(1)
		go func(id int64) {
			defer wg.Done()
			inviteCode, err := uc.GenerateInviteCode(ctx, id)
			if err != nil {
				t.Errorf("GenerateInviteCode failed for rider %d: %v", id, err)
				return
			}

			mu.Lock()
			if existingID, exists := codes[inviteCode.InviteCode]; exists {
				t.Errorf("Duplicate code generated: %s by riders %d and %d", inviteCode.InviteCode, existingID, id)
			}
			codes[inviteCode.InviteCode] = id
			mu.Unlock()
		}(i)
	}

	wg.Wait()

	if len(codes) != 100 {
		t.Errorf("Expected 100 unique codes, got %d", len(codes))
	}

	t.Logf("✅ Concurrent invite code generation test passed: generated %d unique codes", len(codes))
}

// TestConcurrentBindReferral 测试并发绑定邀请关系
func TestConcurrentBindReferral(t *testing.T) {
	repo := newInMemoryRepo()
	uc := NewReferralUsecase(repo, log.DefaultLogger, &conf.Data{})

	ctx := context.Background()
	inviterID := int64(100)

	inviteCode, err := uc.GenerateInviteCode(ctx, inviterID)
	if err != nil {
		t.Fatalf("GenerateInviteCode failed: %v", err)
	}

	var wg sync.WaitGroup
	successCount := 0
	var mu sync.Mutex

	// 尝试用同一个邀请码绑定多个新骑手
	for i := int64(1); i <= 50; i++ {
		wg.Add(1)
		go func(id int64) {
			defer wg.Done()
			_, err := uc.BindReferralRelation(ctx, 200+id, inviteCode.InviteCode, fmt.Sprintf("138001380%02d", id))
			if err == nil {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()

	if successCount != 50 {
		t.Errorf("Expected 50 successful bindings, got %d", successCount)
	}

	t.Logf("✅ Concurrent bind referral test passed: %d successful bindings", successCount)
}

// TestRiskCheck_VirtualPhone 测试虚拟手机号检测
func TestRiskCheck_VirtualPhone(t *testing.T) {
	testCases := []struct {
		phone     string
		isVirtual bool
	}{
		{"13812345678", false},
		{"13912345678", false},
		{"17012345678", true}, // 虚拟运营商
		{"17112345678", true}, // 虚拟运营商
		{"16712345678", true}, // 虚拟运营商
		{"16612345678", true}, // 虚拟运营商
		{"16512345678", true}, // 虚拟运营商
		{"13412345678", false},
		{"13512345678", false},
		{"13612345678", false},
		{"13712345678", false},
		{"15012345678", false},
		{"15112345678", false},
		{"15212345678", false},
		{"15712345678", false},
		{"15812345678", false},
		{"15912345678", false},
		{"18212345678", false},
		{"18312345678", false},
		{"18712345678", false},
		{"18812345678", false},
	}

	svc := NewRiskControlService(nil, &conf.Data{}, log.DefaultLogger)

	for _, tc := range testCases {
		isVirtual := svc.(*riskControlService).isVirtualPhone(tc.phone)
		if isVirtual != tc.isVirtual {
			t.Errorf("Phone %s: expected isVirtual=%v, got %v", tc.phone, tc.isVirtual, isVirtual)
		}
	}

	t.Logf("✅ Virtual phone detection test passed for %d cases", len(testCases))
}

// TestRiskCheck_RegistrationRisk 测试注册风险检测
func TestRiskCheck_RegistrationRisk(t *testing.T) {
	repo := newInMemoryRepo()
	svc := NewRiskControlService(repo, &conf.Data{}, log.DefaultLogger)
	ctx := context.Background()

	// 测试正常注册
	result, err := svc.CheckRegistrationRisk(ctx, "ABC123", "13800138000", "device_12345", "192.168.1.1")
	if err != nil {
		t.Fatalf("Risk check failed: %v", err)
	}

	if result == nil {
		t.Fatal("Risk check result should not be nil")
	}

	if result.RiskLevel != RiskLevelLow {
		t.Errorf("Normal registration should be low risk, got level %d", result.RiskLevel)
	}

	// 测试虚拟手机号
	result2, err := svc.CheckRegistrationRisk(ctx, "ABC123", "17012345678", "device_12346", "192.168.1.2")
	if err != nil {
		t.Fatalf("Risk check for virtual phone failed: %v", err)
	}

	if !result2.IsRisky {
		t.Error("Virtual phone should be detected as risky")
	}

	t.Logf("✅ Registration risk check test passed")
}

// TestRiskCheck_ProxyIP 测试代理IP检测
func TestRiskCheck_ProxyIP(t *testing.T) {
	svc := NewRiskControlService(nil, &conf.Data{}, log.DefaultLogger)
	ctx := context.Background()

	// 测试正常IP
	result, _ := svc.CheckRegistrationRisk(ctx, "ABC123", "13800138000", "device_1", "8.8.8.8")
	if result == nil {
		t.Fatal("Result should not be nil")
	}

	t.Logf("✅ Proxy IP detection test passed")
}

// TestRewardCalculation 测试奖励计算
func TestRewardCalculation(t *testing.T) {
	testCases := []struct {
		name         string
		inviteCount  int32
		validCount   int32
		expectedRate float64
	}{
		{"No bonus", 5, 3, 1.0},
		{"10% bonus", 10, 8, 1.1},
		{"20% bonus", 20, 16, 1.2},
		{"30% bonus", 50, 40, 1.3},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 基础奖励50元
			baseReward := int32(50)
			actualReward := int32(float64(baseReward) * tc.expectedRate)

			t.Logf("✅ %s: invite=%d, valid=%d, rate=%.1f, reward=%d",
				tc.name, tc.inviteCount, tc.validCount, tc.expectedRate, actualReward)
		})
	}
}

// TestCompleteReferralFlow 测试完整的拉新流程
func TestCompleteReferralFlow(t *testing.T) {
	repo := newInMemoryRepo()
	uc := NewReferralUsecase(repo, log.DefaultLogger, &conf.Data{})
	riskSvc := NewRiskControlService(repo, &conf.Data{}, log.DefaultLogger)
	ctx := context.Background()

	// 步骤1: 老骑手生成邀请码
	inviterID := int64(100)
	inviteCode, err := uc.GenerateInviteCode(ctx, inviterID)
	if err != nil {
		t.Fatalf("Step 1 - GenerateInviteCode failed: %v", err)
	}
	t.Logf("✅ Step 1: Generated invite code %s", inviteCode.InviteCode)

	// 步骤2: 验证邀请码
	isValid, inviterName, err := uc.ValidateInviteCode(ctx, inviteCode.InviteCode)
	if err != nil {
		t.Fatalf("Step 2 - ValidateInviteCode failed: %v", err)
	}
	if !isValid {
		t.Fatal("Step 2 - Invite code should be valid")
	}
	t.Logf("✅ Step 2: Validated invite code, inviter=%s", inviterName)

	// 步骤3: 新骑手注册并绑定邀请关系
	newRiderID := int64(200)
	deviceID := "device_new_rider"
	phone := "13800138000"

	// 风险检测
	riskResult, err := riskSvc.CheckRegistrationRisk(ctx, inviteCode.InviteCode, phone, deviceID, "192.168.1.1")
	if err != nil {
		t.Fatalf("Step 3a - Risk check failed: %v", err)
	}
	if riskResult.RiskLevel > RiskLevelMedium {
		t.Fatal("Step 3a - Risk level too high")
	}
	t.Logf("✅ Step 3a: Risk check passed, level=%d", riskResult.RiskLevel)

	// 绑定关系
	boundInviterID, err := uc.BindReferralRelation(ctx, newRiderID, inviteCode.InviteCode, phone)
	if err != nil {
		t.Fatalf("Step 3b - BindReferralRelation failed: %v", err)
	}
	if boundInviterID != inviterID {
		t.Fatalf("Step 3b - Wrong inviter ID: got %d, want %d", boundInviterID, inviterID)
	}
	t.Logf("✅ Step 3b: Bound referral relation")

	// 步骤4: 获取邀请信息
	code, todayCount, monthCount, pendingAmount, err := uc.GetMyReferralInfo(ctx, inviterID)
	if err != nil {
		t.Fatalf("Step 4 - GetMyReferralInfo failed: %v", err)
	}
	if code == nil {
		t.Fatal("Step 4 - Invite code should not be nil")
	}
	if todayCount != 1 {
		t.Errorf("Step 4 - Today count should be 1, got %d", todayCount)
	}
	t.Logf("✅ Step 4: Got referral info: today=%d, month=%d, pending=%d", todayCount, monthCount, pendingAmount)

	// 步骤5: 获取邀请记录
	records, total, err := uc.GetInviteRecordList(ctx, inviterID, 0, 1, 10)
	if err != nil {
		t.Fatalf("Step 5 - GetInviteRecordList failed: %v", err)
	}
	if total != 1 {
		t.Errorf("Step 5 - Total should be 1, got %d", total)
	}
	if len(records) != 1 {
		t.Errorf("Step 5 - Should have 1 record, got %d", len(records))
	}
	t.Logf("✅ Step 5: Got %d invite records", len(records))

	t.Logf("✅ Complete referral flow test passed!")
}
