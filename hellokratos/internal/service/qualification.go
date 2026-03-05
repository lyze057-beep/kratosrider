package service

import (
	"context"
	v1 "hellokratos/api/rider/v1"
	"hellokratos/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
)

// QualificationService 资质验证服务
type QualificationService struct {
	v1.UnimplementedQualificationServer

	qualificationUsecase biz.QualificationUsecase
	log                  *log.Helper
}

// NewQualificationService 创建资质验证服务实例
func NewQualificationService(qualificationUsecase biz.QualificationUsecase, logger log.Logger) *QualificationService {
	return &QualificationService{
		qualificationUsecase: qualificationUsecase,
		log:                  log.NewHelper(logger),
	}
}

// SubmitIDCard 提交身份证信息
func (s *QualificationService) SubmitIDCard(ctx context.Context, req *v1.SubmitIDCardRequest) (*v1.SubmitIDCardReply, error) {
	err := s.qualificationUsecase.SubmitIDCard(ctx, req.RiderId, req.RealName, req.IdCardNumber, req.IdCardFront, req.IdCardBack, req.IdCardHandheld)
	if err != nil {
		s.log.Error("submit id card failed", "err", err)
		return &v1.SubmitIDCardReply{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &v1.SubmitIDCardReply{
		Success:        true,
		Message:        "身份证信息提交成功，请等待审核",
		VerificationId: 0, // 实际应该从usecase返回
	}, nil
}

// SubmitDriverLicense 提交驾驶证信息
func (s *QualificationService) SubmitDriverLicense(ctx context.Context, req *v1.SubmitDriverLicenseRequest) (*v1.SubmitDriverLicenseReply, error) {
	err := s.qualificationUsecase.SubmitDriverLicense(ctx, req.RiderId, req.LicenseNumber, req.LicenseType, req.IssueDate, req.ExpiryDate, req.LicenseFront, req.LicenseBack)
	if err != nil {
		s.log.Error("submit driver license failed", "err", err)
		return &v1.SubmitDriverLicenseReply{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &v1.SubmitDriverLicenseReply{
		Success:        true,
		Message:        "驾驶证信息提交成功，请等待审核",
		VerificationId: 0,
	}, nil
}

// SubmitHealthCertificate 提交健康证信息
func (s *QualificationService) SubmitHealthCertificate(ctx context.Context, req *v1.SubmitHealthCertificateRequest) (*v1.SubmitHealthCertificateReply, error) {
	err := s.qualificationUsecase.SubmitHealthCertificate(ctx, req.RiderId, req.CertificateNumber, req.IssueDate, req.ExpiryDate, req.CertificateImage, req.HospitalName)
	if err != nil {
		s.log.Error("submit health certificate failed", "err", err)
		return &v1.SubmitHealthCertificateReply{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &v1.SubmitHealthCertificateReply{
		Success:        true,
		Message:        "健康证信息提交成功，请等待审核",
		VerificationId: 0,
	}, nil
}

// GetVerificationStatus 获取验证状态
func (s *QualificationService) GetVerificationStatus(ctx context.Context, req *v1.GetVerificationStatusRequest) (*v1.GetVerificationStatusReply, error) {
	status, err := s.qualificationUsecase.GetVerificationStatus(ctx, req.RiderId)
	if err != nil {
		s.log.Error("get verification status failed", "err", err)
		return nil, err
	}

	// 转换总体状态
	overallStatus := &v1.OverallStatus{
		Status:     status.OverallStatus.Status,
		StatusText: status.OverallStatus.StatusText,
		Message:    status.OverallStatus.Message,
	}

	// 转换各项状态
	var items []*v1.VerificationItem
	for _, item := range status.Items {
		items = append(items, &v1.VerificationItem{
			ItemType:     item.ItemType,
			ItemName:     item.ItemName,
			Status:       item.Status,
			StatusText:   item.StatusText,
			SubmitTime:   item.SubmitTime,
			VerifyTime:   item.VerifyTime,
			RejectReason: item.RejectReason,
			ExpiryDate:   item.ExpiryDate,
			DaysToExpiry: item.DaysToExpiry,
		})
	}

	return &v1.GetVerificationStatusReply{
		OverallStatus: overallStatus,
		Items:         items,
	}, nil
}

// GetVerificationDetail 获取验证详情
func (s *QualificationService) GetVerificationDetail(ctx context.Context, req *v1.GetVerificationDetailRequest) (*v1.GetVerificationDetailReply, error) {
	detail, err := s.qualificationUsecase.GetVerificationDetail(ctx, req.RiderId, req.ItemType)
	if err != nil {
		s.log.Error("get verification detail failed", "err", err)
		return nil, err
	}

	return &v1.GetVerificationDetailReply{
		Detail: &v1.VerificationDetail{
			Id:           detail.ID,
			RiderId:      detail.RiderID,
			ItemType:     detail.ItemType,
			ItemName:     detail.ItemName,
			Status:       detail.Status,
			StatusText:   detail.StatusText,
			SubmitData:   detail.SubmitData,
			VerifyResult: detail.VerifyResult,
			RejectReason: detail.RejectReason,
			SubmitTime:   detail.SubmitTime,
			VerifyTime:   detail.VerifyTime,
		},
	}, nil
}

// ResubmitVerification 重新提交验证材料
func (s *QualificationService) ResubmitVerification(ctx context.Context, req *v1.ResubmitVerificationRequest) (*v1.ResubmitVerificationReply, error) {
	err := s.qualificationUsecase.ResubmitVerification(ctx, req.RiderId, req.ItemType, req.Reason)
	if err != nil {
		s.log.Error("resubmit verification failed", "err", err)
		return &v1.ResubmitVerificationReply{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &v1.ResubmitVerificationReply{
		Success: true,
		Message: "已重置验证状态，请重新提交材料",
	}, nil
}

// VerifyQualification 审核验证（管理员接口）
func (s *QualificationService) VerifyQualification(ctx context.Context, req *v1.VerifyQualificationRequest) (*v1.VerifyQualificationReply, error) {
	err := s.qualificationUsecase.VerifyQualification(ctx, req.RiderId, req.ItemType, req.Status, req.RejectReason, req.VerifyResult)
	if err != nil {
		s.log.Error("verify qualification failed", "err", err)
		return &v1.VerifyQualificationReply{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &v1.VerifyQualificationReply{
		Success: true,
		Message: "审核操作成功",
	}, nil
}
