package service

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/types/known/timestamppb"

	pb "hellokratos/api/rider/v1"
	"hellokratos/internal/biz"
	"hellokratos/internal/data/model"
)

// ReferralService 拉新服务
type ReferralService struct {
	pb.UnimplementedReferralServer
	uc *biz.ReferralUsecase
}

// NewReferralService 创建拉新服务
func NewReferralService(uc *biz.ReferralUsecase) *ReferralService {
	return &ReferralService{
		uc: uc,
	}
}

// GenerateInviteCode 生成邀请码
func (s *ReferralService) GenerateInviteCode(ctx context.Context, req *pb.GenerateInviteCodeRequest) (*pb.GenerateInviteCodeReply, error) {
	inviteCode, err := s.uc.GenerateInviteCode(ctx, req.RiderId)
	if err != nil {
		return nil, fmt.Errorf("生成邀请码失败: %w", err)
	}

	return &pb.GenerateInviteCodeReply{
		InviteCodeInfo: &pb.InviteCodeInfo{
			InviteCode:   inviteCode.InviteCode,
			InviteLink:   inviteCode.InviteLink,
			CreatedAt:    timestamppb.New(inviteCode.CreatedAt),
			ExpireAt:     timestamppb.New(inviteCode.ExpireAt),
			TotalInvited: inviteCode.TotalInvited,
			ValidInvited: inviteCode.ValidInvited,
			TotalRewards: inviteCode.TotalRewards,
		},
	}, nil
}

// GetMyReferralInfo 获取我的邀请信息
func (s *ReferralService) GetMyReferralInfo(ctx context.Context, req *pb.GetMyReferralInfoRequest) (*pb.GetMyReferralInfoReply, error) {
	inviteCode, todayInvited, monthInvited, pendingRewards, err := s.uc.GetMyReferralInfo(ctx, req.RiderId)
	if err != nil {
		return nil, fmt.Errorf("获取邀请信息失败: %w", err)
	}

	reply := &pb.GetMyReferralInfoReply{
		TodayInvited:   todayInvited,
		MonthInvited:   monthInvited,
		PendingRewards: int32(pendingRewards),
	}

	if inviteCode != nil {
		reply.InviteCodeInfo = &pb.InviteCodeInfo{
			InviteCode:   inviteCode.InviteCode,
			InviteLink:   inviteCode.InviteLink,
			CreatedAt:    timestamppb.New(inviteCode.CreatedAt),
			ExpireAt:     timestamppb.New(inviteCode.ExpireAt),
			TotalInvited: inviteCode.TotalInvited,
			ValidInvited: inviteCode.ValidInvited,
			TotalRewards: inviteCode.TotalRewards,
		}
	}

	return reply, nil
}

// GetInviteRecordList 获取邀请记录列表
func (s *ReferralService) GetInviteRecordList(ctx context.Context, req *pb.GetInviteRecordListRequest) (*pb.GetInviteRecordListReply, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	relations, total, err := s.uc.GetInviteRecordList(ctx, req.RiderId, req.Status, int(req.Page), int(req.PageSize))
	if err != nil {
		return nil, fmt.Errorf("获取邀请记录失败: %w", err)
	}

	records := make([]*pb.InvitedRiderInfo, 0, len(relations))
	for _, relation := range relations {
		info := &pb.InvitedRiderInfo{
			RiderId: relation.InvitedID,
			Status:  relation.Status,
		}
		if relation.RegisterAt.Unix() > 0 {
			info.RegisterAt = timestamppb.New(relation.RegisterAt)
		}
		if relation.FirstOrderAt != nil {
			info.FirstOrderAt = timestamppb.New(*relation.FirstOrderAt)
		}
		if relation.TaskCompletedAt != nil {
			info.TaskCompletedAt = timestamppb.New(*relation.TaskCompletedAt)
		}
		info.RewardAmount = relation.RewardAmount
		records = append(records, info)
	}

	return &pb.GetInviteRecordListReply{
		Records:  records,
		Total:    int32(total),
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// ValidateInviteCode 验证邀请码
func (s *ReferralService) ValidateInviteCode(ctx context.Context, req *pb.ValidateInviteCodeRequest) (*pb.ValidateInviteCodeReply, error) {
	isValid, inviterName, err := s.uc.ValidateInviteCode(ctx, req.InviteCode)
	if err != nil {
		return nil, fmt.Errorf("验证邀请码失败: %w", err)
	}

	message := ""
	if !isValid {
		message = "邀请码无效或已过期"
	} else {
		message = "邀请码有效"
	}

	return &pb.ValidateInviteCodeReply{
		IsValid:     isValid,
		InviterName: inviterName,
		Message:     message,
	}, nil
}

// BindReferralRelation 绑定邀请关系
func (s *ReferralService) BindReferralRelation(ctx context.Context, req *pb.BindReferralRelationRequest) (*pb.BindReferralRelationReply, error) {
	inviterID, err := s.uc.BindReferralRelation(ctx, req.NewRiderId, req.InviteCode, req.Phone)
	if err != nil {
		return &pb.BindReferralRelationReply{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.BindReferralRelationReply{
		Success:   true,
		Message:   "绑定成功",
		InviterId: inviterID,
	}, nil
}

// GetTaskList 获取任务列表
func (s *ReferralService) GetTaskList(ctx context.Context, req *pb.GetTaskListRequest) (*pb.GetTaskListReply, error) {
	tasks, err := s.uc.GetTaskList(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取任务列表失败: %w", err)
	}

	taskInfos := make([]*pb.TaskInfo, 0, len(tasks))
	for _, task := range tasks {
		taskInfos = append(taskInfos, &pb.TaskInfo{
			TaskId:        task.ID,
			TaskName:      task.TaskName,
			TaskDesc:      task.TaskDesc,
			TaskType:      task.TaskType,
			TargetValue:   task.TargetValue,
			RewardAmount:  task.RewardAmount,
			TimeLimitDays: task.TimeLimitDays,
			StartTime:     timestamppb.New(task.StartTime),
			EndTime:       timestamppb.New(task.EndTime),
		})
	}

	return &pb.GetTaskListReply{
		Tasks: taskInfos,
	}, nil
}

// GetTaskProgress 获取任务进度
func (s *ReferralService) GetTaskProgress(ctx context.Context, req *pb.GetTaskProgressRequest) (*pb.GetTaskProgressReply, error) {
	progressList, err := s.uc.GetTaskProgress(ctx, req.RiderId)
	if err != nil {
		return nil, fmt.Errorf("获取任务进度失败: %w", err)
	}

	// 查找指定任务的进度
	var targetProgress *model.ReferralTaskProgress
	for _, p := range progressList {
		if p.TaskID == req.TaskId {
			targetProgress = p
			break
		}
	}

	if targetProgress == nil {
		return nil, fmt.Errorf("任务不存在")
	}

	// 计算进度百分比
	progressPercent := int32(0)
	if targetProgress.TargetValue > 0 {
		progressPercent = targetProgress.CurrentValue * 100 / targetProgress.TargetValue
	}

	return &pb.GetTaskProgressReply{
		Progress: &pb.TaskProgressInfo{
			TaskId:          targetProgress.TaskID,
			CurrentValue:    targetProgress.CurrentValue,
			TargetValue:     targetProgress.TargetValue,
			ProgressPercent: progressPercent,
			Status:          targetProgress.Status,
			Deadline:        timestamppb.New(targetProgress.Deadline),
			IsClaimed:       targetProgress.IsClaimed,
		},
	}, nil
}

// ClaimTaskReward 领取任务奖励
func (s *ReferralService) ClaimTaskReward(ctx context.Context, req *pb.ClaimTaskRewardRequest) (*pb.ClaimTaskRewardReply, error) {
	rewardAmount, err := s.uc.ClaimTaskReward(ctx, req.RiderId, req.TaskId)
	if err != nil {
		return &pb.ClaimTaskRewardReply{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.ClaimTaskRewardReply{
		Success:      true,
		RewardAmount: rewardAmount,
		Message:      "领取成功",
	}, nil
}

// GetRewardRecordList 获取奖励记录列表
func (s *ReferralService) GetRewardRecordList(ctx context.Context, req *pb.GetRewardRecordListRequest) (*pb.GetRewardRecordListReply, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	records, total, totalAmount, err := s.uc.GetRewardRecordList(ctx, req.RiderId, req.Status, int(req.Page), int(req.PageSize))
	if err != nil {
		return nil, fmt.Errorf("获取奖励记录失败: %w", err)
	}

	recordList := make([]*pb.RewardRecord, 0, len(records))
	for _, record := range records {
		r := &pb.RewardRecord{
			RecordId:     record.ID,
			TaskId:       record.TaskID,
			RewardAmount: record.RewardAmount,
			RewardType:   record.RewardType,
			Status:       record.Status,
			CreatedAt:    timestamppb.New(record.CreatedAt),
		}
		if record.IssuedAt != nil {
			r.IssuedAt = timestamppb.New(*record.IssuedAt)
		}
		recordList = append(recordList, r)
	}

	return &pb.GetRewardRecordListReply{
		Records:     recordList,
		Total:       int32(total),
		TotalAmount: int32(totalAmount),
	}, nil
}

// GetReferralStatistics 获取拉新统计
func (s *ReferralService) GetReferralStatistics(ctx context.Context, req *pb.GetReferralStatisticsRequest) (*pb.GetReferralStatisticsReply, error) {
	stats, err := s.uc.GetReferralStatistics(ctx, req.StartDate, req.EndDate)
	if err != nil {
		return nil, fmt.Errorf("获取统计数据失败: %w", err)
	}

	var totalInvited, validInvited, totalRewards int32
	dailyStats := make([]*pb.DailyStatistics, 0, len(stats))

	for _, stat := range stats {
		totalInvited += stat.NewInvited
		validInvited += stat.ValidInvited
		totalRewards += stat.RewardsAmount

		dailyStats = append(dailyStats, &pb.DailyStatistics{
			Date:          stat.StatDate,
			NewInvited:    stat.NewInvited,
			ValidInvited:  stat.ValidInvited,
			RewardsAmount: stat.RewardsAmount,
		})
	}

	// 计算转化率
	conversionRate := int32(0)
	if totalInvited > 0 {
		conversionRate = validInvited * 100 / totalInvited
	}

	return &pb.GetReferralStatisticsReply{
		TotalInvited:   totalInvited,
		ValidInvited:   validInvited,
		TotalRewards:   totalRewards,
		ConversionRate: conversionRate,
		DailyStats:     dailyStats,
	}, nil
}
