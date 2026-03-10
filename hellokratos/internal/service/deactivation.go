package service

import (
	"context"

	"google.golang.org/protobuf/types/known/timestamppb"

	pb "hellokratos/api/rider/v1"
	"hellokratos/internal/biz"
	"hellokratos/internal/data/model"
)

// DeactivationService 骑手注销服务
type DeactivationService struct {
	pb.UnimplementedDeactivationServiceServer
	uc *biz.DeactivationUsecase
}

// NewDeactivationService 创建注销服务
func NewDeactivationService(uc *biz.DeactivationUsecase) *DeactivationService {
	return &DeactivationService{uc: uc}
}

// ApplyDeactivation 申请注销
func (s *DeactivationService) ApplyDeactivation(ctx context.Context, req *pb.ApplyDeactivationRequest) (*pb.ApplyDeactivationReply, error) {
	result, err := s.uc.ApplyDeactivation(ctx, &biz.ApplyDeactivationRequest{
		RiderID:      req.RiderId,
		ReasonType:   model.DeactivationReason(req.ReasonType),
		ReasonDetail: req.ReasonDetail,
		Phone:        req.Phone,
		IP:           req.Ip,
		DeviceID:     req.DeviceId,
	})
	if err != nil {
		return nil, err
	}

	return &pb.ApplyDeactivationReply{
		DeactivationId: result.DeactivationID,
		Status:         pb.DeactivationStatus(result.Status),
		Checklist:      convertCheckResult(result.Checklist),
		Message:        result.Message,
	}, nil
}

// GetDeactivationStatus 获取注销状态
func (s *DeactivationService) GetDeactivationStatus(ctx context.Context, req *pb.GetDeactivationStatusRequest) (*pb.GetDeactivationStatusReply, error) {
	deactivation, checklist, err := s.uc.GetDeactivationStatus(ctx, req.RiderId)
	if err != nil {
		return nil, err
	}

	if deactivation == nil {
		return &pb.GetDeactivationStatusReply{}, nil
	}

	reply := &pb.GetDeactivationStatusReply{
		Deactivation: convertDeactivationInfo(deactivation),
		Checklist:    convertCheckResult(convertModelChecklist(checklist)),
	}

	return reply, nil
}

// CancelDeactivation 取消注销申请
func (s *DeactivationService) CancelDeactivation(ctx context.Context, req *pb.CancelDeactivationRequest) (*pb.CancelDeactivationReply, error) {
	err := s.uc.CancelDeactivation(ctx, req.RiderId, req.DeactivationId)
	if err != nil {
		return &pb.CancelDeactivationReply{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.CancelDeactivationReply{
		Success: true,
		Message: "取消成功",
	}, nil
}

// GetDeactivationReasons 获取注销原因列表
func (s *DeactivationService) GetDeactivationReasons(ctx context.Context, req *pb.GetDeactivationReasonsRequest) (*pb.GetDeactivationReasonsReply, error) {
	reasons := []*pb.ReasonItem{
		{ReasonType: pb.DeactivationReason_REASON_PERSONAL, ReasonName: "个人原因", Description: "因个人原因需要注销账号"},
		{ReasonType: pb.DeactivationReason_REASON_HEALTH, ReasonName: "健康原因", Description: "身体健康原因无法继续配送"},
		{ReasonType: pb.DeactivationReason_REASON_OTHER_JOB, ReasonName: "找到其他工作", Description: "已找到其他全职工作"},
		{ReasonType: pb.DeactivationReason_REASON_INCOME, ReasonName: "收入不满意", Description: "对当前收入不满意"},
		{ReasonType: pb.DeactivationReason_REASON_TIME, ReasonName: "时间不合适", Description: "配送时间与个人安排冲突"},
		{ReasonType: pb.DeactivationReason_REASON_OTHER, ReasonName: "其他原因", Description: "其他未列出的原因"},
	}

	return &pb.GetDeactivationReasonsReply{Reasons: reasons}, nil
}

// ==================== 管理员接口 ====================

// GetDeactivationList 获取注销申请列表
func (s *DeactivationService) GetDeactivationList(ctx context.Context, req *pb.GetDeactivationListRequest) (*pb.GetDeactivationListReply, error) {
	list, total, err := s.uc.GetDeactivationList(ctx, model.DeactivationStatus(req.Status), int(req.Page), int(req.PageSize))
	if err != nil {
		return nil, err
	}

	var pbList []*pb.DeactivationInfo
	for _, item := range list {
		pbList = append(pbList, convertDeactivationInfo(item))
	}

	return &pb.GetDeactivationListReply{
		List:  pbList,
		Total: total,
	}, nil
}

// ReviewDeactivation 审核注销申请
func (s *DeactivationService) ReviewDeactivation(ctx context.Context, req *pb.ReviewDeactivationRequest) (*pb.ReviewDeactivationReply, error) {
	err := s.uc.ReviewDeactivation(ctx, &biz.ReviewDeactivationRequest{
		DeactivationID: req.DeactivationId,
		ReviewerID:     req.ReviewerId,
		Action:         req.Action,
		Remark:         req.Remark,
	})
	if err != nil {
		return &pb.ReviewDeactivationReply{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	var newStatus pb.DeactivationStatus
	if req.Action == "APPROVE" {
		newStatus = pb.DeactivationStatus_STATUS_APPROVED
	} else {
		newStatus = pb.DeactivationStatus_STATUS_REJECTED
	}

	return &pb.ReviewDeactivationReply{
		Success:    true,
		NewStatus:  newStatus,
		Message:    "审核完成",
	}, nil
}

// GetDeactivationDetail 获取注销详情
func (s *DeactivationService) GetDeactivationDetail(ctx context.Context, req *pb.GetDeactivationDetailRequest) (*pb.GetDeactivationDetailReply, error) {
	// 这里简化处理，实际应该查询历史记录
	return &pb.GetDeactivationDetailReply{}, nil
}

// GetDeactivationStatistics 获取注销统计
func (s *DeactivationService) GetDeactivationStatistics(ctx context.Context, req *pb.GetDeactivationStatisticsRequest) (*pb.GetDeactivationStatisticsReply, error) {
	stat, err := s.uc.GetDeactivationStatistics(ctx, req.StartDate, req.EndDate)
	if err != nil {
		return nil, err
	}

	return &pb.GetDeactivationStatisticsReply{
		TotalApply:      stat.TotalApply,
		PendingCount:    stat.PendingCount,
		ApprovedCount:   stat.ApprovedCount,
		RejectedCount:   stat.RejectedCount,
		CompletedCount:  stat.CompletedCount,
		AvgProcessTime:  stat.AvgProcessTime,
		TotalSettlement: stat.TotalSettlement,
	}, nil
}

// ==================== 转换函数 ====================

func convertDeactivationInfo(d *model.RiderDeactivation) *pb.DeactivationInfo {
	info := &pb.DeactivationInfo{
		Id:              d.ID,
		RiderId:         d.RiderID,
		ReasonType:      pb.DeactivationReason(d.ReasonType),
		ReasonDetail:    d.ReasonDetail,
		Status:          pb.DeactivationStatus(d.Status),
		ApplyTime:       timestamppb.New(d.ApplyTime),
		ReviewerId:      d.ReviewerID,
		ReviewRemark:    d.ReviewRemark,
		ClearanceStatus: d.ClearanceStatus,
		PendingIncome:   d.PendingIncome,
		PendingPenalty:  d.PendingPenalty,
		FinalSettlement: d.FinalSettlement,
	}

	if d.ReviewTime != nil {
		info.ReviewTime = timestamppb.New(*d.ReviewTime)
	}
	if d.CompleteTime != nil {
		info.CompleteTime = timestamppb.New(*d.CompleteTime)
	}

	return info
}

func convertCheckResult(c *biz.DeactivationCheckResult) *pb.DeactivationCheckResult {
	if c == nil {
		return nil
	}
	return &pb.DeactivationCheckResult{
		CanDeactivate:       c.CanDeactivate,
		HasPendingOrders:    c.HasPendingOrders,
		HasPendingIncome:    c.HasPendingIncome,
		HasPendingComplaint: c.HasPendingComplaint,
		HasEquipmentDebt:    c.HasEquipmentDebt,
		HasViolationRecord:  c.HasViolationRecord,
		IsBalanceClear:      c.IsBalanceClear,
		BlockingItems:       c.BlockingItems,
	}
}

func convertModelChecklist(c *model.DeactivationChecklist) *biz.DeactivationCheckResult {
	if c == nil {
		return nil
	}
	return &biz.DeactivationCheckResult{
		CanDeactivate:       c.AllChecksPassed,
		HasPendingOrders:    c.HasPendingOrders,
		HasPendingIncome:    c.HasPendingIncome,
		HasPendingComplaint: c.HasPendingComplaint,
		HasEquipmentDebt:    c.HasEquipmentDebt,
		HasViolationRecord:  c.HasViolationRecord,
		IsBalanceClear:      c.IsBalanceClear,
	}
}
