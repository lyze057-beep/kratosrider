package data

import (
	"context"
	"time"

	"hellokratos/internal/data/model"
)

// ReferralRepo 拉新仓库接口
type ReferralRepo interface {
	// 邀请码相关
	GetInviteCodeByRiderID(ctx context.Context, riderID int64) (*model.ReferralInviteCode, error)
	GetInviteCodeByCode(ctx context.Context, code string) (*model.ReferralInviteCode, error)
	CreateInviteCode(ctx context.Context, code *model.ReferralInviteCode) error
	UpdateInviteCodeStats(ctx context.Context, code string, totalInvited, validInvited, totalRewards int32) error

	// 邀请关系相关
	GetReferralRelationByInvitedID(ctx context.Context, invitedID int64) (*model.ReferralRelation, error)
	CreateReferralRelation(ctx context.Context, relation *model.ReferralRelation) error
	UpdateReferralRelationStatus(ctx context.Context, id int64, status int32) error
	UpdateReferralRelationFirstOrder(ctx context.Context, invitedID int64) error
	UpdateReferralRelationTaskCompleted(ctx context.Context, invitedID int64, rewardAmount int32) error
	GetReferralRelationList(ctx context.Context, inviterID int64, status int32, page, pageSize int) ([]*model.ReferralRelation, int64, error)
	GetReferralRelationCount(ctx context.Context, inviterID int64, startTime, endTime time.Time) (int64, error)

	// 任务相关
	GetTaskList(ctx context.Context, status int32) ([]*model.ReferralTask, error)
	GetTaskByID(ctx context.Context, taskID int64) (*model.ReferralTask, error)

	// 任务进度相关
	GetTaskProgress(ctx context.Context, relationID, taskID int64) (*model.ReferralTaskProgress, error)
	CreateTaskProgress(ctx context.Context, progress *model.ReferralTaskProgress) error
	UpdateTaskProgress(ctx context.Context, id int64, currentValue int32, status int32) error
	CompleteTaskProgress(ctx context.Context, id int64) error
	ClaimTaskReward(ctx context.Context, id int64) error
	GetTaskProgressList(ctx context.Context, inviterID int64) ([]*model.ReferralTaskProgress, error)

	// 奖励记录相关
	CreateRewardRecord(ctx context.Context, record *model.ReferralRewardRecord) error
	GetRewardRecordList(ctx context.Context, inviterID int64, status int32, page, pageSize int) ([]*model.ReferralRewardRecord, int64, error)
	UpdateRewardRecordStatus(ctx context.Context, id int64, status int32, failReason string) error
	GetTotalRewardAmount(ctx context.Context, inviterID int64) (int64, error)
	GetPendingRewardAmount(ctx context.Context, inviterID int64) (int64, error)

	// 风控相关
	CreateRiskLog(ctx context.Context, log *model.ReferralRiskLog) error
	GetRiskLogList(ctx context.Context, relationID int64) ([]*model.ReferralRiskLog, error)
	ConfirmRisk(ctx context.Context, id int64) error

	// 统计相关
	GetStatisticsByDate(ctx context.Context, date string) (*model.ReferralStatistics, error)
	GetStatisticsByDateRange(ctx context.Context, startDate, endDate string) ([]*model.ReferralStatistics, error)
	UpdateStatistics(ctx context.Context, stat *model.ReferralStatistics) error
}

type referralRepoImpl struct {
	data *Data
}

// NewReferralRepo 创建拉新仓库
func NewReferralRepo(data *Data) ReferralRepo {
	return &referralRepoImpl{
		data: data,
	}
}

// GetInviteCodeByRiderID 根据骑手ID获取邀请码
func (r *referralRepoImpl) GetInviteCodeByRiderID(ctx context.Context, riderID int64) (*model.ReferralInviteCode, error) {
	var code model.ReferralInviteCode
	err := r.data.db.WithContext(ctx).Where("rider_id = ?", riderID).First(&code).Error
	if err != nil {
		return nil, err
	}
	return &code, nil
}

// GetInviteCodeByCode 根据邀请码获取
func (r *referralRepoImpl) GetInviteCodeByCode(ctx context.Context, code string) (*model.ReferralInviteCode, error) {
	var inviteCode model.ReferralInviteCode
	err := r.data.db.WithContext(ctx).Where("invite_code = ?", code).First(&inviteCode).Error
	if err != nil {
		return nil, err
	}
	return &inviteCode, nil
}

// CreateInviteCode 创建邀请码
func (r *referralRepoImpl) CreateInviteCode(ctx context.Context, code *model.ReferralInviteCode) error {
	return r.data.db.WithContext(ctx).Create(code).Error
}

// UpdateInviteCodeStats 更新邀请码统计
func (r *referralRepoImpl) UpdateInviteCodeStats(ctx context.Context, code string, totalInvited, validInvited, totalRewards int32) error {
	return r.data.db.WithContext(ctx).Model(&model.ReferralInviteCode{}).
		Where("invite_code = ?", code).
		Updates(map[string]interface{}{
			"total_invited": totalInvited,
			"valid_invited": validInvited,
			"total_rewards": totalRewards,
		}).Error
}

// GetReferralRelationByInvitedID 根据被邀请人ID获取邀请关系
func (r *referralRepoImpl) GetReferralRelationByInvitedID(ctx context.Context, invitedID int64) (*model.ReferralRelation, error) {
	var relation model.ReferralRelation
	err := r.data.db.WithContext(ctx).Where("invited_id = ?", invitedID).First(&relation).Error
	if err != nil {
		return nil, err
	}
	return &relation, nil
}

// CreateReferralRelation 创建邀请关系
func (r *referralRepoImpl) CreateReferralRelation(ctx context.Context, relation *model.ReferralRelation) error {
	return r.data.db.WithContext(ctx).Create(relation).Error
}

// UpdateReferralRelationStatus 更新邀请关系状态
func (r *referralRepoImpl) UpdateReferralRelationStatus(ctx context.Context, id int64, status int32) error {
	return r.data.db.WithContext(ctx).Model(&model.ReferralRelation{}).
		Where("id = ?", id).
		Update("status", status).Error
}

// UpdateReferralRelationFirstOrder 更新首单时间
func (r *referralRepoImpl) UpdateReferralRelationFirstOrder(ctx context.Context, invitedID int64) error {
	now := time.Now()
	return r.data.db.WithContext(ctx).Model(&model.ReferralRelation{}).
		Where("invited_id = ?", invitedID).
		Updates(map[string]interface{}{
			"first_order_at": &now,
			"status":         2,
		}).Error
}

// UpdateReferralRelationTaskCompleted 更新任务完成状态
func (r *referralRepoImpl) UpdateReferralRelationTaskCompleted(ctx context.Context, invitedID int64, rewardAmount int32) error {
	now := time.Now()
	return r.data.db.WithContext(ctx).Model(&model.ReferralRelation{}).
		Where("invited_id = ?", invitedID).
		Updates(map[string]interface{}{
			"task_completed_at": &now,
			"status":            3,
			"reward_amount":     rewardAmount,
			"reward_status":     1,
		}).Error
}

// GetReferralRelationList 获取邀请关系列表
func (r *referralRepoImpl) GetReferralRelationList(ctx context.Context, inviterID int64, status int32, page, pageSize int) ([]*model.ReferralRelation, int64, error) {
	var relations []*model.ReferralRelation
	var total int64

	db := r.data.db.WithContext(ctx).Where("inviter_id = ?", inviterID)
	if status > 0 {
		db = db.Where("status = ?", status)
	}

	if err := db.Model(&model.ReferralRelation{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&relations).Error; err != nil {
		return nil, 0, err
	}

	return relations, total, nil
}

// GetReferralRelationCount 获取邀请关系数量
func (r *referralRepoImpl) GetReferralRelationCount(ctx context.Context, inviterID int64, startTime, endTime time.Time) (int64, error) {
	var count int64
	err := r.data.db.WithContext(ctx).Model(&model.ReferralRelation{}).
		Where("inviter_id = ? AND created_at >= ? AND created_at < ?", inviterID, startTime, endTime).
		Count(&count).Error
	return count, err
}

// GetTaskList 获取任务列表
func (r *referralRepoImpl) GetTaskList(ctx context.Context, status int32) ([]*model.ReferralTask, error) {
	var tasks []*model.ReferralTask
	db := r.data.db.WithContext(ctx)
	if status > 0 {
		db = db.Where("status = ?", status)
	}
	err := db.Find(&tasks).Error
	return tasks, err
}

// GetTaskByID 根据ID获取任务
func (r *referralRepoImpl) GetTaskByID(ctx context.Context, taskID int64) (*model.ReferralTask, error) {
	var task model.ReferralTask
	err := r.data.db.WithContext(ctx).Where("id = ?", taskID).First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// GetTaskProgress 获取任务进度
func (r *referralRepoImpl) GetTaskProgress(ctx context.Context, relationID, taskID int64) (*model.ReferralTaskProgress, error) {
	var progress model.ReferralTaskProgress
	db := r.data.db.WithContext(ctx)
	if relationID > 0 {
		db = db.Where("relation_id = ?", relationID)
	}
	if taskID > 0 {
		db = db.Where("task_id = ?", taskID)
	}
	err := db.First(&progress).Error
	if err != nil {
		return nil, err
	}
	return &progress, nil
}

// CreateTaskProgress 创建任务进度
func (r *referralRepoImpl) CreateTaskProgress(ctx context.Context, progress *model.ReferralTaskProgress) error {
	return r.data.db.WithContext(ctx).Create(progress).Error
}

// UpdateTaskProgress 更新任务进度
func (r *referralRepoImpl) UpdateTaskProgress(ctx context.Context, id int64, currentValue int32, status int32) error {
	return r.data.db.WithContext(ctx).Model(&model.ReferralTaskProgress{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"current_value": currentValue,
			"status":        status,
		}).Error
}

// CompleteTaskProgress 完成任务进度
func (r *referralRepoImpl) CompleteTaskProgress(ctx context.Context, id int64) error {
	now := time.Now()
	return r.data.db.WithContext(ctx).Model(&model.ReferralTaskProgress{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":       2,
			"completed_at": &now,
		}).Error
}

// ClaimTaskReward 领取任务奖励
func (r *referralRepoImpl) ClaimTaskReward(ctx context.Context, id int64) error {
	now := time.Now()
	return r.data.db.WithContext(ctx).Model(&model.ReferralTaskProgress{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_claimed": true,
			"claimed_at": &now,
		}).Error
}

// GetTaskProgressList 获取任务进度列表
func (r *referralRepoImpl) GetTaskProgressList(ctx context.Context, inviterID int64) ([]*model.ReferralTaskProgress, error) {
	var progressList []*model.ReferralTaskProgress
	err := r.data.db.WithContext(ctx).Where("inviter_id = ?", inviterID).Find(&progressList).Error
	return progressList, err
}

// CreateRewardRecord 创建奖励记录
func (r *referralRepoImpl) CreateRewardRecord(ctx context.Context, record *model.ReferralRewardRecord) error {
	return r.data.db.WithContext(ctx).Create(record).Error
}

// GetRewardRecordList 获取奖励记录列表
func (r *referralRepoImpl) GetRewardRecordList(ctx context.Context, inviterID int64, status int32, page, pageSize int) ([]*model.ReferralRewardRecord, int64, error) {
	var records []*model.ReferralRewardRecord
	var total int64

	db := r.data.db.WithContext(ctx).Where("inviter_id = ?", inviterID)
	if status > 0 {
		db = db.Where("status = ?", status)
	}

	if err := db.Model(&model.ReferralRewardRecord{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&records).Error; err != nil {
		return nil, 0, err
	}

	return records, total, nil
}

// UpdateRewardRecordStatus 更新奖励记录状态
func (r *referralRepoImpl) UpdateRewardRecordStatus(ctx context.Context, id int64, status int32, failReason string) error {
	now := time.Now()
	updates := map[string]interface{}{
		"status": status,
	}
	if status == 2 {
		updates["issued_at"] = &now
	}
	if failReason != "" {
		updates["fail_reason"] = failReason
	}
	return r.data.db.WithContext(ctx).Model(&model.ReferralRewardRecord{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// GetTotalRewardAmount 获取总奖励金额
func (r *referralRepoImpl) GetTotalRewardAmount(ctx context.Context, inviterID int64) (int64, error) {
	var total int64
	err := r.data.db.WithContext(ctx).Model(&model.ReferralRewardRecord{}).
		Where("inviter_id = ? AND status = 2", inviterID).
		Select("COALESCE(SUM(reward_amount), 0)").
		Scan(&total).Error
	return total, err
}

// GetPendingRewardAmount 获取待领取奖励金额
func (r *referralRepoImpl) GetPendingRewardAmount(ctx context.Context, inviterID int64) (int64, error) {
	var total int64
	err := r.data.db.WithContext(ctx).Model(&model.ReferralRewardRecord{}).
		Where("inviter_id = ? AND status = 1", inviterID).
		Select("COALESCE(SUM(reward_amount), 0)").
		Scan(&total).Error
	return total, err
}

// CreateRiskLog 创建风控日志
func (r *referralRepoImpl) CreateRiskLog(ctx context.Context, log *model.ReferralRiskLog) error {
	return r.data.db.WithContext(ctx).Create(log).Error
}

// GetRiskLogList 获取风控日志列表
func (r *referralRepoImpl) GetRiskLogList(ctx context.Context, relationID int64) ([]*model.ReferralRiskLog, error) {
	var logs []*model.ReferralRiskLog
	err := r.data.db.WithContext(ctx).Where("relation_id = ?", relationID).Find(&logs).Error
	return logs, err
}

// ConfirmRisk 确认风险
func (r *referralRepoImpl) ConfirmRisk(ctx context.Context, id int64) error {
	now := time.Now()
	return r.data.db.WithContext(ctx).Model(&model.ReferralRiskLog{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_confirmed": true,
			"confirmed_at": &now,
		}).Error
}

// GetStatisticsByDate 根据日期获取统计
func (r *referralRepoImpl) GetStatisticsByDate(ctx context.Context, date string) (*model.ReferralStatistics, error) {
	var stat model.ReferralStatistics
	err := r.data.db.WithContext(ctx).Where("stat_date = ?", date).First(&stat).Error
	if err != nil {
		return nil, err
	}
	return &stat, nil
}

// GetStatisticsByDateRange 根据日期范围获取统计
func (r *referralRepoImpl) GetStatisticsByDateRange(ctx context.Context, startDate, endDate string) ([]*model.ReferralStatistics, error) {
	var stats []*model.ReferralStatistics
	err := r.data.db.WithContext(ctx).
		Where("stat_date >= ? AND stat_date <= ?", startDate, endDate).
		Order("stat_date ASC").
		Find(&stats).Error
	return stats, err
}

// UpdateStatistics 更新统计
func (r *referralRepoImpl) UpdateStatistics(ctx context.Context, stat *model.ReferralStatistics) error {
	return r.data.db.WithContext(ctx).Save(stat).Error
}
