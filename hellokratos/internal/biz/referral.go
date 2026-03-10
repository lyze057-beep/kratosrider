package biz

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"

	"hellokratos/internal/conf"
	"hellokratos/internal/data"
	"hellokratos/internal/data/model"
)

// ReferralUsecase 拉新业务用例
type ReferralUsecase struct {
	repo data.ReferralRepo
	log  *log.Helper
	conf *conf.Data
}

// NewReferralUsecase 创建拉新业务用例
func NewReferralUsecase(repo data.ReferralRepo, logger log.Logger, conf *conf.Data) *ReferralUsecase {
	return &ReferralUsecase{
		repo: repo,
		log:  log.NewHelper(logger),
		conf: conf,
	}
}

// GenerateInviteCode 生成邀请码
func (uc *ReferralUsecase) GenerateInviteCode(ctx context.Context, riderID int64) (*model.ReferralInviteCode, error) {
	// 检查是否已有有效邀请码
	existingCode, err := uc.repo.GetInviteCodeByRiderID(ctx, riderID)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("查询邀请码失败: %w", err)
	}

	if existingCode != nil && existingCode.Status == 1 && existingCode.ExpireAt.After(time.Now()) {
		// 返回已有的有效邀请码
		return existingCode, nil
	}

	// 生成新的邀请码
	code, err := uc.generateUniqueCode(ctx)
	if err != nil {
		return nil, fmt.Errorf("生成邀请码失败: %w", err)
	}

	// 创建邀请码记录
	inviteCode := &model.ReferralInviteCode{
		RiderID:    riderID,
		InviteCode: code,
		InviteLink: fmt.Sprintf("https://rider.example.com/register?invite_code=%s", code),
		Status:     1,
		ExpireAt:   time.Now().AddDate(1, 0, 0), // 1年有效期
	}

	if err := uc.repo.CreateInviteCode(ctx, inviteCode); err != nil {
		return nil, fmt.Errorf("保存邀请码失败: %w", err)
	}

	uc.log.Infof("生成邀请码成功: rider_id=%d, code=%s", riderID, code)
	return inviteCode, nil
}

// generateUniqueCode 生成唯一邀请码
func (uc *ReferralUsecase) generateUniqueCode(ctx context.Context) (string, error) {
	const maxRetries = 10
	for i := 0; i < maxRetries; i++ {
		// 生成6位随机码
		code := uc.generateRandomCode(6)

		// 检查是否已存在
		_, err := uc.repo.GetInviteCodeByCode(ctx, code)
		if err == gorm.ErrRecordNotFound {
			return code, nil
		}
		if err != nil {
			return "", err
		}
		// 已存在，重新生成
	}
	return "", fmt.Errorf("无法生成唯一邀请码")
}

// generateRandomCode 生成随机码
func (uc *ReferralUsecase) generateRandomCode(length int) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	bytes := make([]byte, length)
	rand.Read(bytes)
	for i := range bytes {
		bytes[i] = charset[bytes[i]%byte(len(charset))]
	}
	return string(bytes)
}

// GetMyReferralInfo 获取我的邀请信息
func (uc *ReferralUsecase) GetMyReferralInfo(ctx context.Context, riderID int64) (*model.ReferralInviteCode, int32, int32, int64, error) {
	// 获取邀请码信息
	inviteCode, err := uc.repo.GetInviteCodeByRiderID(ctx, riderID)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, 0, 0, 0, fmt.Errorf("查询邀请码失败: %w", err)
	}

	// 获取今日邀请数
	today := time.Now().Format("2006-01-02")
	todayStart, _ := time.Parse("2006-01-02", today)
	todayEnd := todayStart.AddDate(0, 0, 1)
	todayCount, err := uc.repo.GetReferralRelationCount(ctx, riderID, todayStart, todayEnd)
	if err != nil {
		return nil, 0, 0, 0, fmt.Errorf("查询今日邀请数失败: %w", err)
	}

	// 获取本月邀请数
	monthStart := time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.Local)
	monthEnd := monthStart.AddDate(0, 1, 0)
	monthCount, err := uc.repo.GetReferralRelationCount(ctx, riderID, monthStart, monthEnd)
	if err != nil {
		return nil, 0, 0, 0, fmt.Errorf("查询本月邀请数失败: %w", err)
	}

	// 获取待领取奖励金额
	pendingAmount, err := uc.repo.GetPendingRewardAmount(ctx, riderID)
	if err != nil {
		return nil, 0, 0, 0, fmt.Errorf("查询待领取奖励失败: %w", err)
	}

	return inviteCode, int32(todayCount), int32(monthCount), pendingAmount, nil
}

// GetInviteRecordList 获取邀请记录列表
func (uc *ReferralUsecase) GetInviteRecordList(ctx context.Context, riderID int64, status int32, page, pageSize int) ([]*model.ReferralRelation, int64, error) {
	return uc.repo.GetReferralRelationList(ctx, riderID, status, page, pageSize)
}

// ValidateInviteCode 验证邀请码
func (uc *ReferralUsecase) ValidateInviteCode(ctx context.Context, code string) (bool, string, error) {
	inviteCode, err := uc.repo.GetInviteCodeByCode(ctx, code)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, "", nil
		}
		return false, "", fmt.Errorf("查询邀请码失败: %w", err)
	}

	// 检查状态
	if inviteCode.Status != 1 {
		return false, "", nil
	}

	// 检查是否过期
	if inviteCode.ExpireAt.Before(time.Now()) {
		return false, "", nil
	}

	// TODO: 获取邀请人名称（脱敏）
	inviterName := "骑手" + fmt.Sprintf("%d", inviteCode.RiderID%10000)

	return true, inviterName, nil
}

// BindReferralRelation 绑定邀请关系
func (uc *ReferralUsecase) BindReferralRelation(ctx context.Context, newRiderID int64, inviteCode, phone string) (int64, error) {
	// 验证邀请码
	inviteCodeInfo, err := uc.repo.GetInviteCodeByCode(ctx, inviteCode)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, fmt.Errorf("邀请码无效")
		}
		return 0, fmt.Errorf("查询邀请码失败: %w", err)
	}

	// 检查邀请码状态
	if inviteCodeInfo.Status != 1 {
		return 0, fmt.Errorf("邀请码已失效")
	}

	if inviteCodeInfo.ExpireAt.Before(time.Now()) {
		return 0, fmt.Errorf("邀请码已过期")
	}

	// 不能邀请自己
	if inviteCodeInfo.RiderID == newRiderID {
		return 0, fmt.Errorf("不能邀请自己")
	}

	// 检查是否已被邀请过
	existingRelation, err := uc.repo.GetReferralRelationByInvitedID(ctx, newRiderID)
	if err != nil && err != gorm.ErrRecordNotFound {
		return 0, fmt.Errorf("查询邀请关系失败: %w", err)
	}
	if existingRelation != nil {
		return 0, fmt.Errorf("您已被邀请过，不能重复绑定")
	}

	// 创建邀请关系
	relation := &model.ReferralRelation{
		InviterID:  inviteCodeInfo.RiderID,
		InvitedID:  newRiderID,
		InviteCode: inviteCode,
		Status:     1, // 已注册
		RegisterAt: time.Now(),
	}

	if err := uc.repo.CreateReferralRelation(ctx, relation); err != nil {
		return 0, fmt.Errorf("创建邀请关系失败: %w", err)
	}

	// 更新邀请码统计
	if err := uc.repo.UpdateInviteCodeStats(ctx, inviteCode, inviteCodeInfo.TotalInvited+1, inviteCodeInfo.ValidInvited, inviteCodeInfo.TotalRewards); err != nil {
		uc.log.Warnf("更新邀请码统计失败: %v", err)
	}

	// 初始化任务进度
	if err := uc.initTaskProgress(ctx, relation.ID, inviteCodeInfo.RiderID, newRiderID); err != nil {
		uc.log.Warnf("初始化任务进度失败: %v", err)
	}

	uc.log.Infof("绑定邀请关系成功: inviter_id=%d, invited_id=%d", inviteCodeInfo.RiderID, newRiderID)
	return inviteCodeInfo.RiderID, nil
}

// initTaskProgress 初始化任务进度
func (uc *ReferralUsecase) initTaskProgress(ctx context.Context, relationID, inviterID, invitedID int64) error {
	// 获取有效任务列表
	tasks, err := uc.repo.GetTaskList(ctx, 1)
	if err != nil {
		return fmt.Errorf("获取任务列表失败: %w", err)
	}

	now := time.Now()
	for _, task := range tasks {
		// 检查任务是否在有效期内
		if now.Before(task.StartTime) || now.After(task.EndTime) {
			continue
		}

		progress := &model.ReferralTaskProgress{
			RelationID:   relationID,
			TaskID:       task.ID,
			InviterID:    inviterID,
			InvitedID:    invitedID,
			CurrentValue: 0,
			TargetValue:  task.TargetValue,
			Status:       1, // 进行中
			StartTime:    now,
			Deadline:     now.AddDate(0, 0, int(task.TimeLimitDays)),
		}

		if err := uc.repo.CreateTaskProgress(ctx, progress); err != nil {
			uc.log.Warnf("创建任务进度失败: task_id=%d, err=%v", task.ID, err)
		}
	}

	return nil
}

// GetTaskList 获取任务列表
func (uc *ReferralUsecase) GetTaskList(ctx context.Context) ([]*model.ReferralTask, error) {
	return uc.repo.GetTaskList(ctx, 1)
}

// GetTaskProgress 获取任务进度
func (uc *ReferralUsecase) GetTaskProgress(ctx context.Context, riderID int64) ([]*model.ReferralTaskProgress, error) {
	return uc.repo.GetTaskProgressList(ctx, riderID)
}

// ClaimTaskReward 领取任务奖励
func (uc *ReferralUsecase) ClaimTaskReward(ctx context.Context, riderID, taskID int64) (int32, error) {
	// 获取任务进度
	progress, err := uc.repo.GetTaskProgress(ctx, 0, taskID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, fmt.Errorf("任务不存在")
		}
		return 0, fmt.Errorf("查询任务进度失败: %w", err)
	}

	// 检查是否是该骑手的任务
	if progress.InviterID != riderID {
		return 0, fmt.Errorf("无权领取该任务奖励")
	}

	// 检查任务是否已完成
	if progress.Status != 2 {
		return 0, fmt.Errorf("任务未完成")
	}

	// 检查是否已领取
	if progress.IsClaimed {
		return 0, fmt.Errorf("奖励已领取")
	}

	// 获取任务信息
	task, err := uc.repo.GetTaskByID(ctx, taskID)
	if err != nil {
		return 0, fmt.Errorf("获取任务信息失败: %w", err)
	}

	// 创建奖励记录
	record := &model.ReferralRewardRecord{
		RelationID:   progress.RelationID,
		TaskID:       taskID,
		InviterID:    riderID,
		InvitedID:    progress.InvitedID,
		RewardAmount: task.RewardAmount,
		RewardType:   task.RewardType,
		Status:       1, // 待发放
	}

	if err := uc.repo.CreateRewardRecord(ctx, record); err != nil {
		return 0, fmt.Errorf("创建奖励记录失败: %w", err)
	}

	// 标记任务已领取
	if err := uc.repo.ClaimTaskReward(ctx, progress.ID); err != nil {
		uc.log.Warnf("标记任务已领取失败: %v", err)
	}

	uc.log.Infof("领取任务奖励成功: rider_id=%d, task_id=%d, amount=%d", riderID, taskID, task.RewardAmount)
	return task.RewardAmount, nil
}

// GetRewardRecordList 获取奖励记录列表
func (uc *ReferralUsecase) GetRewardRecordList(ctx context.Context, riderID int64, status int32, page, pageSize int) ([]*model.ReferralRewardRecord, int64, int64, error) {
	records, total, err := uc.repo.GetRewardRecordList(ctx, riderID, status, page, pageSize)
	if err != nil {
		return nil, 0, 0, err
	}

	totalAmount, err := uc.repo.GetTotalRewardAmount(ctx, riderID)
	if err != nil {
		return nil, 0, 0, err
	}

	return records, total, totalAmount, nil
}

// CheckRisk 检查风险
func (uc *ReferralUsecase) CheckRisk(ctx context.Context, relationID int64, deviceID, ipAddress string) ([]*model.ReferralRiskLog, error) {
	// TODO: 实现风控检测逻辑
	// 1. 检查设备是否重复
	// 2. 检查IP是否异常
	// 3. 检查行为是否异常
	// 4. 检查是否虚假注册

	return uc.repo.GetRiskLogList(ctx, relationID)
}

// ProcessFirstOrder 处理首单完成
func (uc *ReferralUsecase) ProcessFirstOrder(ctx context.Context, riderID int64) error {
	// 获取邀请关系
	relation, err := uc.repo.GetReferralRelationByInvitedID(ctx, riderID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 不是被邀请的骑手，直接返回
			return nil
		}
		return fmt.Errorf("查询邀请关系失败: %w", err)
	}

	// 更新首单时间
	if err := uc.repo.UpdateReferralRelationFirstOrder(ctx, riderID); err != nil {
		return fmt.Errorf("更新首单时间失败: %w", err)
	}

	// 更新任务进度（首单任务）
	// TODO: 根据实际任务类型更新进度

	uc.log.Infof("处理首单完成: rider_id=%d, relation_id=%d", riderID, relation.ID)
	return nil
}

// UpdateTaskProgress 更新任务进度
func (uc *ReferralUsecase) UpdateTaskProgress(ctx context.Context, invitedID int64, taskType int32, increment int32) error {
	// 获取邀请关系
	relation, err := uc.repo.GetReferralRelationByInvitedID(ctx, invitedID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil
		}
		return fmt.Errorf("查询邀请关系失败: %w", err)
	}

	// 获取任务进度列表
	progressList, err := uc.repo.GetTaskProgressList(ctx, relation.InviterID)
	if err != nil {
		return fmt.Errorf("获取任务进度失败: %w", err)
	}

	for _, progress := range progressList {
		// 只处理进行中的任务
		if progress.Status != 1 {
			continue
		}

		// 检查是否过期
		if time.Now().After(progress.Deadline) {
			uc.repo.UpdateTaskProgress(ctx, progress.ID, progress.CurrentValue, 3) // 已过期
			continue
		}

		// 更新进度
		newValue := progress.CurrentValue + increment
		if newValue >= progress.TargetValue {
			// 任务完成
			if err := uc.repo.CompleteTaskProgress(ctx, progress.ID); err != nil {
				uc.log.Warnf("完成任务进度失败: %v", err)
				continue
			}

			// 更新邀请关系状态
			task, err := uc.repo.GetTaskByID(ctx, progress.TaskID)
			if err != nil {
				uc.log.Warnf("获取任务信息失败: %v", err)
				continue
			}

			if err := uc.repo.UpdateReferralRelationTaskCompleted(ctx, invitedID, task.RewardAmount); err != nil {
				uc.log.Warnf("更新邀请关系任务完成状态失败: %v", err)
			}
		} else {
			// 更新进度
			if err := uc.repo.UpdateTaskProgress(ctx, progress.ID, newValue, 1); err != nil {
				uc.log.Warnf("更新任务进度失败: %v", err)
			}
		}
	}

	return nil
}

// IssueReward 发放奖励
func (uc *ReferralUsecase) IssueReward(ctx context.Context, recordID int64) error {
	// TODO: 实现实际发放逻辑（调用支付系统或积分系统）
	// 这里先模拟发放成功

	if err := uc.repo.UpdateRewardRecordStatus(ctx, recordID, 2, ""); err != nil {
		return fmt.Errorf("更新奖励记录状态失败: %w", err)
	}

	uc.log.Infof("发放奖励成功: record_id=%d", recordID)
	return nil
}

// GetReferralStatistics 获取拉新统计
func (uc *ReferralUsecase) GetReferralStatistics(ctx context.Context, startDate, endDate string) ([]*model.ReferralStatistics, error) {
	return uc.repo.GetStatisticsByDateRange(ctx, startDate, endDate)
}
