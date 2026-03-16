package service

import (
	"context"
	v1 "hellokratos/api/rider/v1"
	"hellokratos/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// AppealService 申诉服务
type AppealService struct {
	v1.UnimplementedAppealServer
	appealUsecase biz.AppealUsecase
	log           *log.Helper
}

// NewAppealService 创建申诉服务实例
func NewAppealService(appealUsecase biz.AppealUsecase, logger log.Logger) *AppealService {
	return &AppealService{
		appealUsecase: appealUsecase,
		log:           log.NewHelper(logger),
	}
}

// SubmitOrderAppeal 提交订单申诉
func (s *AppealService) SubmitOrderAppeal(ctx context.Context, req *v1.SubmitOrderAppealRequest) (*v1.SubmitOrderAppealReply, error) {
	// 参数校验
	if req.RiderId <= 0 || req.OrderId <= 0 {
		return &v1.SubmitOrderAppealReply{
			Success: false,
			Message: "无效的骑手或订单ID",
		}, nil
	}
	if req.Reason == "" {
		return &v1.SubmitOrderAppealReply{
			Success: false,
			Message: "申诉原因不能为空",
		}, nil
	}

	// 调用业务逻辑
	appeal, err := s.appealUsecase.SubmitOrderAppeal(
		ctx,
		req.RiderId,
		req.OrderId,
		req.AppealType,
		req.Reason,
		req.Description,
		req.EvidenceImages,
		req.ContactPhone,
	)
	if err != nil {
		s.log.Error("提交订单申诉失败", "err", err)
		return &v1.SubmitOrderAppealReply{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &v1.SubmitOrderAppealReply{
		Success:  true,
		Message:  "申诉提交成功",
		AppealId: appeal.ID,
		TicketNo: appeal.TicketNo,
	}, nil
}

// SubmitPenaltyAppeal 提交处罚申诉
func (s *AppealService) SubmitPenaltyAppeal(ctx context.Context, req *v1.SubmitPenaltyAppealRequest) (*v1.SubmitPenaltyAppealReply, error) {
	// 参数校验
	if req.RiderId <= 0 {
		return &v1.SubmitPenaltyAppealReply{
			Success: false,
			Message: "无效的骑手ID",
		}, nil
	}
	if req.PenaltyType == "" || req.PenaltyId == "" {
		return &v1.SubmitPenaltyAppealReply{
			Success: false,
			Message: "处罚类型和处罚ID不能为空",
		}, nil
	}
	if req.Reason == "" {
		return &v1.SubmitPenaltyAppealReply{
			Success: false,
			Message: "申诉原因不能为空",
		}, nil
	}

	// 调用业务逻辑
	appeal, err := s.appealUsecase.SubmitPenaltyAppeal(
		ctx,
		req.RiderId,
		req.PenaltyType,
		req.PenaltyId,
		req.AppealType,
		req.Reason,
		req.Description,
		req.EvidenceImages,
		req.ContactPhone,
	)
	if err != nil {
		s.log.Error("提交处罚申诉失败", "err", err)
		return &v1.SubmitPenaltyAppealReply{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &v1.SubmitPenaltyAppealReply{
		Success:  true,
		Message:  "申诉提交成功",
		AppealId: appeal.ID,
		TicketNo: appeal.TicketNo,
	}, nil
}

// GetAppeals 获取申诉列表
func (s *AppealService) GetAppeals(ctx context.Context, req *v1.GetAppealsRequest) (*v1.GetAppealsReply, error) {
	if req.RiderId <= 0 {
		return &v1.GetAppealsReply{
			Success: false,
			Message: "无效的骑手ID",
		}, nil
	}

	appeals, total, err := s.appealUsecase.GetAppeals(
		ctx,
		req.RiderId,
		req.Status,
		req.AppealType,
		int(req.Page),
		int(req.PageSize),
	)
	if err != nil {
		s.log.Error("获取申诉列表失败", "err", err)
		return &v1.GetAppealsReply{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	// 转换数据
	appealInfos := make([]*v1.AppealInfo, 0, len(appeals))
	for _, appeal := range appeals {
		info := &v1.AppealInfo{
			AppealId:       appeal.ID,
			TicketNo:       appeal.TicketNo,
			AppealType:     appeal.AppealType,
			AppealTypeName: getAppealTypeName(appeal.AppealType),
			OrderId:        appeal.OrderID,
			PenaltyType:    appeal.PenaltyType,
			Reason:         appeal.Reason,
			Description:    appeal.Description,
			Status:         appeal.Status,
			StatusName:     getAppealStatusName(appeal.Status),
			Result:         appeal.Result,
			Reply:          appeal.Reply,
			CreatedAt:      timestamppb.New(appeal.CreatedAt),
			UpdatedAt:      timestamppb.New(appeal.UpdatedAt),
		}
		if appeal.ResolvedAt != nil {
			info.ResolvedAt = timestamppb.New(*appeal.ResolvedAt)
		}
		appealInfos = append(appealInfos, info)
	}

	return &v1.GetAppealsReply{
		Success:  true,
		Message:  "获取成功",
		Appeals:  appealInfos,
		Total:    int32(total),
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// GetAppealDetail 获取申诉详情
func (s *AppealService) GetAppealDetail(ctx context.Context, req *v1.GetAppealDetailRequest) (*v1.GetAppealDetailReply, error) {
	if req.AppealId <= 0 {
		return &v1.GetAppealDetailReply{
			Success: false,
			Message: "无效的申诉ID",
		}, nil
	}

	appeal, err := s.appealUsecase.GetAppealDetail(ctx, req.AppealId, req.RiderId)
	if err != nil {
		s.log.Error("获取申诉详情失败", "err", err)
		return &v1.GetAppealDetailReply{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	detail := &v1.AppealDetail{
		AppealId:       appeal.ID,
		TicketNo:       appeal.TicketNo,
		AppealType:     appeal.AppealType,
		AppealTypeName: getAppealTypeName(appeal.AppealType),
		OrderId:        appeal.OrderID,
		PenaltyType:    appeal.PenaltyType,
		PenaltyId:      appeal.PenaltyID,
		Reason:         appeal.Reason,
		Description:    appeal.Description,
		ContactPhone:   appeal.ContactPhone,
		Status:         appeal.Status,
		StatusName:     getAppealStatusName(appeal.Status),
		Result:         appeal.Result,
		Reply:          appeal.Reply,
		HandlerName:    appeal.HandlerName,
		CreatedAt:      timestamppb.New(appeal.CreatedAt),
		UpdatedAt:      timestamppb.New(appeal.UpdatedAt),
	}

	// 解析证据图片
	if appeal.EvidenceImages != "" {
		detail.EvidenceImages = parseImages(appeal.EvidenceImages)
	}

	if appeal.ResolvedAt != nil {
		detail.ResolvedAt = timestamppb.New(*appeal.ResolvedAt)
	}

	return &v1.GetAppealDetailReply{
		Success: true,
		Message: "获取成功",
		Appeal:  detail,
	}, nil
}

// CancelAppeal 取消申诉
func (s *AppealService) CancelAppeal(ctx context.Context, req *v1.CancelAppealRequest) (*v1.CancelAppealReply, error) {
	if req.AppealId <= 0 {
		return &v1.CancelAppealReply{
			Success: false,
			Message: "无效的申诉ID",
		}, nil
	}

	err := s.appealUsecase.CancelAppeal(ctx, req.AppealId, req.RiderId, req.CancelReason)
	if err != nil {
		s.log.Error("取消申诉失败", "err", err)
		return &v1.CancelAppealReply{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &v1.CancelAppealReply{
		Success: true,
		Message: "申诉已取消",
	}, nil
}

// GetAppealTypes 获取申诉类型列表
func (s *AppealService) GetAppealTypes(ctx context.Context, req *v1.GetAppealTypesRequest) (*v1.GetAppealTypesReply, error) {
	types, err := s.appealUsecase.GetAppealTypes(ctx, req.AppealCategory)
	if err != nil {
		s.log.Error("获取申诉类型失败", "err", err)
		return &v1.GetAppealTypesReply{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	typeInfos := make([]*v1.AppealType, 0, len(types))
	for _, t := range types {
		typeInfos = append(typeInfos, &v1.AppealType{
			TypeCode:    t.TypeCode,
			TypeName:    t.TypeName,
			Description: t.Description,
			SortOrder:   int32(t.SortOrder),
		})
	}

	return &v1.GetAppealTypesReply{
		Success: true,
		Message: "获取成功",
		Types:   typeInfos,
	}, nil
}

// GetExceptionOrders 获取异常订单列表
func (s *AppealService) GetExceptionOrders(ctx context.Context, req *v1.GetExceptionOrdersRequest) (*v1.GetExceptionOrdersReply, error) {
	if req.RiderId <= 0 {
		return &v1.GetExceptionOrdersReply{
			Success: false,
			Message: "无效的骑手ID",
		}, nil
	}

	orders, total, err := s.appealUsecase.GetExceptionOrders(
		ctx,
		req.RiderId,
		req.StartDate,
		req.EndDate,
		int(req.Page),
		int(req.PageSize),
	)
	if err != nil {
		s.log.Error("获取异常订单失败", "err", err)
		return &v1.GetExceptionOrdersReply{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	orderInfos := make([]*v1.ExceptionOrder, 0, len(orders))
	for _, order := range orders {
		orderInfos = append(orderInfos, &v1.ExceptionOrder{
			OrderId:       order.OrderID,
			OrderNo:       "", // 需要从订单表获取
			ExceptionType: order.ExceptionType,
			ExceptionDesc: order.ExceptionDesc,
			OccurredAt:    timestamppb.New(order.OccurredAt),
			Status:        getExceptionOrderStatusName(order.Status),
		})
	}

	return &v1.GetExceptionOrdersReply{
		Success: true,
		Message: "获取成功",
		Orders:  orderInfos,
		Total:   int32(total),
	}, nil
}

// SubmitExceptionReport 提交异常报备
func (s *AppealService) SubmitExceptionReport(ctx context.Context, req *v1.SubmitExceptionReportRequest) (*v1.SubmitExceptionReportReply, error) {
	if req.RiderId <= 0 {
		return &v1.SubmitExceptionReportReply{
			Success: false,
			Message: "无效的骑手ID",
		}, nil
	}
	if req.ExceptionType == "" {
		return &v1.SubmitExceptionReportReply{
			Success: false,
			Message: "异常类型不能为空",
		}, nil
	}

	report, err := s.appealUsecase.SubmitExceptionReport(
		ctx,
		req.RiderId,
		req.OrderId,
		req.ExceptionType,
		req.Description,
		req.Images,
		req.Location,
		req.Latitude,
		req.Longitude,
	)
	if err != nil {
		s.log.Error("提交异常报备失败", "err", err)
		return &v1.SubmitExceptionReportReply{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &v1.SubmitExceptionReportReply{
		Success:  true,
		Message:  "异常报备提交成功",
		ReportId: report.ID,
	}, nil
}

// 辅助函数
func getAppealTypeName(appealType int32) string {
	switch appealType {
	case 1:
		return "订单申诉"
	case 2:
		return "处罚申诉"
	default:
		return "未知类型"
	}
}

func getAppealStatusName(status int32) string {
	switch status {
	case 1:
		return "待处理"
	case 2:
		return "处理中"
	case 3:
		return "已通过"
	case 4:
		return "已驳回"
	case 5:
		return "已取消"
	default:
		return "未知状态"
	}
}

func getExceptionOrderStatusName(status int32) string {
	switch status {
	case 1:
		return "未处理"
	case 2:
		return "已申诉"
	case 3:
		return "已处理"
	default:
		return "未知状态"
	}
}
