package data

import (
	"context"
	"time"

	"hellokratos/internal/data/model"

	"gorm.io/gorm"
)

// ReferralTaskSeeder 拉新任务种子数据
type ReferralTaskSeeder struct {
	db *gorm.DB
}

// NewReferralTaskSeeder 创建任务种子数据管理器
func NewReferralTaskSeeder(db *gorm.DB) *ReferralTaskSeeder {
	return &ReferralTaskSeeder{db: db}
}

// SeedTasks 初始化拉新任务数据
func (s *ReferralTaskSeeder) SeedTasks(ctx context.Context) error {
	now := time.Now()
	oneYearLater := now.AddDate(1, 0, 0)

	tasks := []*model.ReferralTask{
		{
			TaskName:      "首单任务",
			TaskDesc:      "新骑手完成首笔订单，邀请人可获得奖励",
			TaskType:      1,
			TargetValue:   1,
			RewardAmount:  50,
			RewardType:    1,
			TimeLimitDays: 7,
			StartTime:     now,
			EndTime:       oneYearLater,
			Status:        1,
		},
		{
			TaskName:      "在线时长任务",
			TaskDesc:      "新骑手累计在线满10小时，邀请人可获得奖励",
			TaskType:      2,
			TargetValue:   600,
			RewardAmount:  30,
			RewardType:    1,
			TimeLimitDays: 14,
			StartTime:     now,
			EndTime:       oneYearLater,
			Status:        1,
		},
		{
			TaskName:      "完成单量任务",
			TaskDesc:      "新骑手完成20笔订单，邀请人可获得奖励",
			TaskType:      3,
			TargetValue:   20,
			RewardAmount:  100,
			RewardType:    1,
			TimeLimitDays: 30,
			StartTime:     now,
			EndTime:       oneYearLater,
			Status:        1,
		},
		{
			TaskName:      "新手任务",
			TaskDesc:      "新骑手完成5笔订单，邀请人可获得奖励",
			TaskType:      3,
			TargetValue:   5,
			RewardAmount:  20,
			RewardType:    1,
			TimeLimitDays: 7,
			StartTime:     now,
			EndTime:       oneYearLater,
			Status:        1,
		},
	}

	for _, task := range tasks {
		var existingTask model.ReferralTask
		err := s.db.WithContext(ctx).Where("task_name = ? AND task_type = ?", task.TaskName, task.TaskType).First(&existingTask).Error
		if err == nil {
			continue
		}
		if err := s.db.WithContext(ctx).Create(task).Error; err != nil {
			return err
		}
	}

	return nil
}

// SeedReferralStatistics 初始化统计数据（如果不存在）
func (s *ReferralTaskSeeder) SeedReferralStatistics(ctx context.Context) error {
	today := time.Now().Format("2006-01-02")

	var existingStat model.ReferralStatistics
	err := s.db.WithContext(ctx).Where("stat_date = ?", today).First(&existingStat).Error
	if err == nil {
		return nil
	}

	stat := &model.ReferralStatistics{
		StatDate:      today,
		NewInvited:    0,
		ValidInvited:  0,
		RewardsAmount: 0,
		CheatingCount: 0,
	}

	return s.db.WithContext(ctx).Create(stat).Error
}
