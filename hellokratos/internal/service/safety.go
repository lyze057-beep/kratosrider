package service

import (
	"context"
	v1 "hellokratos/api/rider/v1"
	"hellokratos/internal/biz"
	"hellokratos/internal/data/model"

	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// SafetyService 安全服务
type SafetyService struct {
	v1.UnimplementedSafetyServer
	safetyUsecase biz.SafetyUsecase
	log           *log.Helper
}

// NewSafetyService 创建安全服务实例
func NewSafetyService(safetyUsecase biz.SafetyUsecase, logger log.Logger) *SafetyService {
	return &SafetyService{
		safetyUsecase: safetyUsecase,
		log:           log.NewHelper(logger),
	}
}

// EmergencyHelp 发起紧急求助
func (s *SafetyService) EmergencyHelp(ctx context.Context, req *v1.EmergencyHelpRequest) (*v1.EmergencyHelpReply, error) {
	// 参数校验
	if req.RiderId <= 0 {
		return &v1.EmergencyHelpReply{
			Success: false,
			Message: "无效的骑手ID",
		}, nil
	}
	if req.HelpType == "" {
		return &v1.EmergencyHelpReply{
			Success: false,
			Message: "求助类型不能为空",
		}, nil
	}

	// 调用业务逻辑
	help, err := s.safetyUsecase.EmergencyHelp(
		ctx,
		req.RiderId,
		req.OrderId,
		req.HelpType,
		req.Description,
		req.Latitude,
		req.Longitude,
		req.Address,
		req.Images,
	)
	if err != nil {
		s.log.Error("发起紧急求助失败", "err", err)
		return &v1.EmergencyHelpReply{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &v1.EmergencyHelpReply{
		Success:     true,
		Message:     "紧急求助已发起，客服人员将尽快联系您",
		EmergencyId: help.ID,
		TicketNo:    help.TicketNo,
	}, nil
}

// CancelEmergency 取消紧急求助
func (s *SafetyService) CancelEmergency(ctx context.Context, req *v1.CancelEmergencyRequest) (*v1.CancelEmergencyReply, error) {
	if req.EmergencyId <= 0 {
		return &v1.CancelEmergencyReply{
			Success: false,
			Message: "无效的求助ID",
		}, nil
	}

	err := s.safetyUsecase.CancelEmergency(ctx, req.EmergencyId, req.RiderId, req.CancelReason)
	if err != nil {
		s.log.Error("取消紧急求助失败", "err", err)
		return &v1.CancelEmergencyReply{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &v1.CancelEmergencyReply{
		Success: true,
		Message: "紧急求助已取消",
	}, nil
}

// GetEmergencyRecords 获取紧急求助记录
func (s *SafetyService) GetEmergencyRecords(ctx context.Context, req *v1.GetEmergencyRecordsRequest) (*v1.GetEmergencyRecordsReply, error) {
	if req.RiderId <= 0 {
		return &v1.GetEmergencyRecordsReply{
			Success: false,
			Message: "无效的骑手ID",
		}, nil
	}

	records, total, err := s.safetyUsecase.GetEmergencyRecords(
		ctx,
		req.RiderId,
		req.Status,
		int(req.Page),
		int(req.PageSize),
	)
	if err != nil {
		s.log.Error("获取紧急求助记录失败", "err", err)
		return &v1.GetEmergencyRecordsReply{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	// 转换数据
	recordInfos := make([]*v1.EmergencyRecord, 0, len(records))
	for _, record := range records {
		info := &v1.EmergencyRecord{
			EmergencyId:  record.ID,
			TicketNo:     record.TicketNo,
			HelpType:     record.HelpType,
			HelpTypeName: getHelpTypeName(record.HelpType),
			Description:  record.Description,
			Latitude:     record.Latitude,
			Longitude:    record.Longitude,
			Address:      record.Address,
			Status:       record.Status,
			StatusName:   getEmergencyStatusName(record.Status),
			CreatedAt:    timestamppb.New(record.CreatedAt),
			HandlerName:  record.HandlerName,
			Result:       record.Result,
		}
		if record.ResolvedAt != nil {
			info.ResolvedAt = timestamppb.New(*record.ResolvedAt)
		}
		recordInfos = append(recordInfos, info)
	}

	return &v1.GetEmergencyRecordsReply{
		Success: true,
		Message: "获取成功",
		Records: recordInfos,
		Total:   int32(total),
	}, nil
}

// SubmitInsuranceClaim 提交保险理赔申请
func (s *SafetyService) SubmitInsuranceClaim(ctx context.Context, req *v1.SubmitInsuranceClaimRequest) (*v1.SubmitInsuranceClaimReply, error) {
	if req.RiderId <= 0 {
		return &v1.SubmitInsuranceClaimReply{
			Success: false,
			Message: "无效的骑手ID",
		}, nil
	}
	if req.IncidentType == "" {
		return &v1.SubmitInsuranceClaimReply{
			Success: false,
			Message: "事故类型不能为空",
		}, nil
	}

	incidentTime := ""
	if req.IncidentTime != nil {
		incidentTime = req.IncidentTime.AsTime().Format("2006-01-02 15:04:05")
	}

	claim, err := s.safetyUsecase.SubmitInsuranceClaim(
		ctx,
		req.RiderId,
		req.OrderId,
		req.IncidentType,
		incidentTime,
		req.IncidentLocation,
		req.Latitude,
		req.Longitude,
		req.IncidentDesc,
		req.InjurySituation,
		req.MedicalExpense,
		req.EvidenceImages,
		req.MedicalRecords,
		req.ContactName,
		req.ContactPhone,
		req.BankName,
		req.BankAccount,
		req.AccountName,
	)
	if err != nil {
		s.log.Error("提交保险理赔申请失败", "err", err)
		return &v1.SubmitInsuranceClaimReply{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &v1.SubmitInsuranceClaimReply{
		Success: true,
		Message: "理赔申请已提交，请耐心等待审核",
		ClaimId: claim.ID,
		ClaimNo: claim.ClaimNo,
	}, nil
}

// GetInsuranceClaims 获取保险理赔列表
func (s *SafetyService) GetInsuranceClaims(ctx context.Context, req *v1.GetInsuranceClaimsRequest) (*v1.GetInsuranceClaimsReply, error) {
	if req.RiderId <= 0 {
		return &v1.GetInsuranceClaimsReply{
			Success: false,
			Message: "无效的骑手ID",
		}, nil
	}

	claims, total, err := s.safetyUsecase.GetInsuranceClaims(
		ctx,
		req.RiderId,
		req.Status,
		int(req.Page),
		int(req.PageSize),
	)
	if err != nil {
		s.log.Error("获取保险理赔列表失败", "err", err)
		return &v1.GetInsuranceClaimsReply{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	claimInfos := make([]*v1.InsuranceClaimInfo, 0, len(claims))
	for _, claim := range claims {
		info := &v1.InsuranceClaimInfo{
			ClaimId:          claim.ID,
			ClaimNo:          claim.ClaimNo,
			IncidentType:     claim.IncidentType,
			IncidentTypeName: getIncidentTypeName(claim.IncidentType),
			IncidentTime:     timestamppb.New(claim.IncidentTime),
			IncidentLocation: claim.IncidentLocation,
			Status:           claim.Status,
			StatusName:       getInsuranceClaimStatusName(claim.Status),
			ClaimAmount:      claim.ClaimAmount,
			PaidAmount:       claim.PaidAmount,
			CreatedAt:        timestamppb.New(claim.CreatedAt),
			UpdatedAt:        timestamppb.New(claim.UpdatedAt),
		}
		claimInfos = append(claimInfos, info)
	}

	return &v1.GetInsuranceClaimsReply{
		Success: true,
		Message: "获取成功",
		Claims:  claimInfos,
		Total:   int32(total),
	}, nil
}

// GetInsuranceClaimDetail 获取保险理赔详情
func (s *SafetyService) GetInsuranceClaimDetail(ctx context.Context, req *v1.GetInsuranceClaimDetailRequest) (*v1.GetInsuranceClaimDetailReply, error) {
	if req.ClaimId <= 0 {
		return &v1.GetInsuranceClaimDetailReply{
			Success: false,
			Message: "无效的理赔ID",
		}, nil
	}

	claim, err := s.safetyUsecase.GetInsuranceClaimDetail(ctx, req.ClaimId, req.RiderId)
	if err != nil {
		s.log.Error("获取保险理赔详情失败", "err", err)
		return &v1.GetInsuranceClaimDetailReply{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	detail := &v1.InsuranceClaimDetail{
		ClaimId:          claim.ID,
		ClaimNo:          claim.ClaimNo,
		OrderId:          claim.OrderID,
		IncidentType:     claim.IncidentType,
		IncidentTypeName: getIncidentTypeName(claim.IncidentType),
		IncidentTime:     timestamppb.New(claim.IncidentTime),
		IncidentLocation: claim.IncidentLocation,
		Latitude:         claim.Latitude,
		Longitude:        claim.Longitude,
		IncidentDesc:     claim.IncidentDesc,
		InjurySituation:  claim.InjurySituation,
		MedicalExpense:   claim.MedicalExpense,
		ContactName:      claim.ContactName,
		ContactPhone:     claim.ContactPhone,
		BankName:         claim.BankName,
		BankAccount:      claim.BankAccount,
		AccountName:      claim.AccountName,
		Status:           claim.Status,
		StatusName:       getInsuranceClaimStatusName(claim.Status),
		ClaimAmount:      claim.ClaimAmount,
		PaidAmount:       claim.PaidAmount,
		RejectReason:     claim.RejectReason,
		HandlerName:      claim.HandlerName,
		Remark:           claim.Remark,
		CreatedAt:        timestamppb.New(claim.CreatedAt),
		UpdatedAt:        timestamppb.New(claim.UpdatedAt),
	}

	// 解析图片
	if claim.EvidenceImages != "" {
		detail.EvidenceImages = parseImages(claim.EvidenceImages)
	}
	if claim.MedicalRecords != "" {
		detail.MedicalRecords = parseImages(claim.MedicalRecords)
	}

	if claim.PaidAt != nil {
		detail.PaidAt = timestamppb.New(*claim.PaidAt)
	}

	return &v1.GetInsuranceClaimDetailReply{
		Success: true,
		Message: "获取成功",
		Claim:   detail,
	}, nil
}

// GetInsuranceInfo 获取保险信息
func (s *SafetyService) GetInsuranceInfo(ctx context.Context, req *v1.GetInsuranceInfoRequest) (*v1.GetInsuranceInfoReply, error) {
	if req.RiderId <= 0 {
		return &v1.GetInsuranceInfoReply{
			Success: false,
			Message: "无效的骑手ID",
		}, nil
	}

	insurance, err := s.safetyUsecase.GetInsuranceInfo(ctx, req.RiderId)
	if err != nil {
		s.log.Error("获取保险信息失败", "err", err)
		return &v1.GetInsuranceInfoReply{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &v1.GetInsuranceInfoReply{
		Success: true,
		Message: "获取成功",
		Info: &v1.InsuranceInfo{
			InsuranceNo:      insurance.InsuranceNo,
			InsuranceType:    insurance.InsuranceType,
			CoverageDesc:     insurance.CoverageDesc,
			CoverageAmount:   insurance.CoverageAmount,
			EffectiveDate:    timestamppb.New(insurance.EffectiveDate),
			ExpireDate:       timestamppb.New(insurance.ExpireDate),
			Status:           getInsuranceStatusName(insurance.Status),
			ClaimCount:       insurance.ClaimCount,
			TotalClaimAmount: insurance.TotalClaimAmount,
		},
	}, nil
}

// ReportSafetyEvent 上报安全事件
func (s *SafetyService) ReportSafetyEvent(ctx context.Context, req *v1.ReportSafetyEventRequest) (*v1.ReportSafetyEventReply, error) {
	if req.RiderId <= 0 {
		return &v1.ReportSafetyEventReply{
			Success: false,
			Message: "无效的骑手ID",
		}, nil
	}
	if req.EventType == "" {
		return &v1.ReportSafetyEventReply{
			Success: false,
			Message: "事件类型不能为空",
		}, nil
	}

	event, err := s.safetyUsecase.ReportSafetyEvent(
		ctx,
		req.RiderId,
		req.OrderId,
		req.EventType,
		req.Description,
		req.Latitude,
		req.Longitude,
		req.Address,
		req.Images,
		req.NeedHelp,
	)
	if err != nil {
		s.log.Error("上报安全事件失败", "err", err)
		return &v1.ReportSafetyEventReply{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &v1.ReportSafetyEventReply{
		Success: true,
		Message: "安全事件已上报",
		EventId: event.ID,
	}, nil
}

// GetSafetyTips 获取安全提示
func (s *SafetyService) GetSafetyTips(ctx context.Context, req *v1.GetSafetyTipsRequest) (*v1.GetSafetyTipsReply, error) {
	tips, total, err := s.safetyUsecase.GetSafetyTips(
		ctx,
		req.TipType,
		int(req.Page),
		int(req.PageSize),
	)
	if err != nil {
		s.log.Error("获取安全提示失败", "err", err)
		return &v1.GetSafetyTipsReply{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	tipInfos := make([]*v1.SafetyTip, 0, len(tips))
	for _, tip := range tips {
		tipInfos = append(tipInfos, &v1.SafetyTip{
			Id:          tip.ID,
			Title:       tip.Title,
			Content:     tip.Content,
			TipType:     tip.TipType,
			Priority:    tip.Priority,
			PublishTime: timestamppb.New(tip.PublishTime),
		})
	}

	return &v1.GetSafetyTipsReply{
		Success: true,
		Message: "获取成功",
		Tips:    tipInfos,
		Total:   int32(total),
	}, nil
}

// GetEmergencyContacts 获取紧急联系人
func (s *SafetyService) GetEmergencyContacts(ctx context.Context, req *v1.GetEmergencyContactsRequest) (*v1.GetEmergencyContactsReply, error) {
	if req.RiderId <= 0 {
		return &v1.GetEmergencyContactsReply{
			Success: false,
			Message: "无效的骑手ID",
		}, nil
	}

	contacts, err := s.safetyUsecase.GetEmergencyContacts(ctx, req.RiderId)
	if err != nil {
		s.log.Error("获取紧急联系人失败", "err", err)
		return &v1.GetEmergencyContactsReply{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	contactInfos := make([]*v1.EmergencyContact, 0, len(contacts))
	for _, contact := range contacts {
		contactInfos = append(contactInfos, &v1.EmergencyContact{
			Id:        contact.ID,
			Name:      contact.Name,
			Phone:     contact.Phone,
			Relation:  contact.Relation,
			IsPrimary: contact.IsPrimary,
		})
	}

	return &v1.GetEmergencyContactsReply{
		Success:  true,
		Message:  "获取成功",
		Contacts: contactInfos,
	}, nil
}

// UpdateEmergencyContacts 更新紧急联系人
func (s *SafetyService) UpdateEmergencyContacts(ctx context.Context, req *v1.UpdateEmergencyContactsRequest) (*v1.UpdateEmergencyContactsReply, error) {
	if req.RiderId <= 0 {
		return &v1.UpdateEmergencyContactsReply{
			Success: false,
			Message: "无效的骑手ID",
		}, nil
	}

	// 转换联系人数据
	contacts := make([]*model.EmergencyContact, 0, len(req.Contacts))
	for _, c := range req.Contacts {
		contacts = append(contacts, &model.EmergencyContact{
			ID:        c.Id,
			Name:      c.Name,
			Phone:     c.Phone,
			Relation:  c.Relation,
			IsPrimary: c.IsPrimary,
		})
	}

	err := s.safetyUsecase.UpdateEmergencyContacts(ctx, req.RiderId, contacts)
	if err != nil {
		s.log.Error("更新紧急联系人失败", "err", err)
		return &v1.UpdateEmergencyContactsReply{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &v1.UpdateEmergencyContactsReply{
		Success: true,
		Message: "紧急联系人已更新",
	}, nil
}

// 辅助函数
func getHelpTypeName(helpType string) string {
	switch helpType {
	case "accident":
		return "交通事故"
	case "illness":
		return "突发疾病"
	case "danger":
		return "人身安全"
	case "other":
		return "其他紧急情况"
	default:
		return "未知类型"
	}
}

func getEmergencyStatusName(status int32) string {
	switch status {
	case 1:
		return "处理中"
	case 2:
		return "已响应"
	case 3:
		return "已处理"
	case 4:
		return "已取消"
	default:
		return "未知状态"
	}
}

func getIncidentTypeName(incidentType string) string {
	switch incidentType {
	case "accident":
		return "意外事故"
	case "injury":
		return "人身伤害"
	case "third_party":
		return "第三方责任"
	default:
		return "其他"
	}
}

func getInsuranceClaimStatusName(status int32) string {
	switch status {
	case 1:
		return "待审核"
	case 2:
		return "审核中"
	case 3:
		return "已通过"
	case 4:
		return "已拒绝"
	case 5:
		return "已赔付"
	default:
		return "未知状态"
	}
}

func getInsuranceStatusName(status int32) string {
	switch status {
	case 1:
		return "有效"
	case 2:
		return "已过期"
	case 3:
		return "已终止"
	default:
		return "未知"
	}
}
