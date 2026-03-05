package biz

import (
	"context"
	"errors"
	"hellokratos/internal/data"
	"hellokratos/internal/data/model"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

// QualificationUsecase 资质验证业务逻辑接口
type QualificationUsecase interface {
	// 身份证验证
	SubmitIDCard(ctx context.Context, riderID int64, realName, idCardNumber, idCardFront, idCardBack, idCardHandheld string) error

	// 驾驶证验证
	SubmitDriverLicense(ctx context.Context, riderID int64, licenseNumber, licenseType, issueDate, expiryDate, licenseFront, licenseBack string) error

	// 健康证验证
	SubmitHealthCertificate(ctx context.Context, riderID int64, certificateNumber, issueDate, expiryDate, certificateImage, hospitalName string) error

	// 查询
	GetVerificationStatus(ctx context.Context, riderID int64) (*VerificationStatus, error)
	GetVerificationDetail(ctx context.Context, riderID int64, itemType string) (*VerificationDetail, error)

	// 重新提交
	ResubmitVerification(ctx context.Context, riderID int64, itemType, reason string) error

	// 审核验证（管理员接口）
	VerifyQualification(ctx context.Context, riderID int64, itemType string, status int32, rejectReason, verifyResult string) error
}

// VerificationStatus 验证状态
type VerificationStatus struct {
	OverallStatus OverallStatus
	Items         []VerificationItem
}

// OverallStatus 总体状态
type OverallStatus struct {
	Status     int32
	StatusText string
	Message    string
}

// VerificationItem 验证项
type VerificationItem struct {
	ItemType     string
	ItemName     string
	Status       int32
	StatusText   string
	SubmitTime   string
	VerifyTime   string
	RejectReason string
	ExpiryDate   string
	DaysToExpiry int32
}

// VerificationDetail 验证详情
type VerificationDetail struct {
	ID           int64
	RiderID      int64
	ItemType     string
	ItemName     string
	Status       int32
	StatusText   string
	SubmitData   string
	VerifyResult string
	RejectReason string
	SubmitTime   string
	VerifyTime   string
}

// qualificationUsecase 资质验证业务逻辑实现
type qualificationUsecase struct {
	qualificationRepo data.QualificationRepo
	log               *log.Helper
}

// NewQualificationUsecase 创建资质验证业务逻辑实例
func NewQualificationUsecase(qualificationRepo data.QualificationRepo, logger log.Logger) QualificationUsecase {
	return &qualificationUsecase{
		qualificationRepo: qualificationRepo,
		log:               log.NewHelper(logger),
	}
}

// SubmitIDCard 提交身份证信息
func (uc *qualificationUsecase) SubmitIDCard(ctx context.Context, riderID int64, realName, idCardNumber, idCardFront, idCardBack, idCardHandheld string) error {
	// 1. 检查是否已提交
	existing, err := uc.qualificationRepo.GetIDCardVerification(ctx, riderID)
	if err != nil {
		return err
	}

	if existing != nil && existing.Status == 2 {
		return errors.New("身份证已通过验证，无需重复提交")
	}

	// 2. 创建或更新验证记录
	verification := &model.IDCardVerification{
		RiderID:        riderID,
		RealName:       realName,
		IDCardNumber:   idCardNumber,
		IDCardFront:    idCardFront,
		IDCardBack:     idCardBack,
		IDCardHandheld: idCardHandheld,
		Status:         1, // 审核中
	}

	if existing != nil {
		// 更新已有记录
		err = uc.qualificationRepo.UpdateIDCardStatus(ctx, riderID, 1, "")
	} else {
		// 创建新记录
		err = uc.qualificationRepo.CreateIDCardVerification(ctx, verification)
	}

	if err != nil {
		return err
	}

	// 3. 更新总体资质状态
	err = uc.updateOverallStatus(ctx, riderID)
	return err
}

// SubmitDriverLicense 提交驾驶证信息
func (uc *qualificationUsecase) SubmitDriverLicense(ctx context.Context, riderID int64, licenseNumber, licenseType, issueDate, expiryDate, licenseFront, licenseBack string) error {
	// 1. 检查是否已提交
	existing, err := uc.qualificationRepo.GetDriverLicenseVerification(ctx, riderID)
	if err != nil {
		return err
	}

	if existing != nil && existing.Status == 2 {
		return errors.New("驾驶证已通过验证，无需重复提交")
	}

	// 2. 创建或更新验证记录
	verification := &model.DriverLicenseVerification{
		RiderID:       riderID,
		LicenseNumber: licenseNumber,
		LicenseType:   licenseType,
		IssueDate:     issueDate,
		ExpiryDate:    expiryDate,
		LicenseFront:  licenseFront,
		LicenseBack:   licenseBack,
		Status:        1, // 审核中
	}

	if existing != nil {
		err = uc.qualificationRepo.UpdateDriverLicenseStatus(ctx, riderID, 1, "")
	} else {
		err = uc.qualificationRepo.CreateDriverLicenseVerification(ctx, verification)
	}

	if err != nil {
		return err
	}

	// 3. 更新总体资质状态
	err = uc.updateOverallStatus(ctx, riderID)
	return err
}

// SubmitHealthCertificate 提交健康证信息
func (uc *qualificationUsecase) SubmitHealthCertificate(ctx context.Context, riderID int64, certificateNumber, issueDate, expiryDate, certificateImage, hospitalName string) error {
	// 1. 检查是否已提交
	existing, err := uc.qualificationRepo.GetHealthCertificateVerification(ctx, riderID)
	if err != nil {
		return err
	}

	if existing != nil && existing.Status == 2 {
		return errors.New("健康证已通过验证，无需重复提交")
	}

	// 2. 创建或更新验证记录
	verification := &model.HealthCertificateVerification{
		RiderID:           riderID,
		CertificateNumber: certificateNumber,
		IssueDate:         issueDate,
		ExpiryDate:        expiryDate,
		CertificateImage:  certificateImage,
		HospitalName:      hospitalName,
		Status:            1, // 审核中
	}

	if existing != nil {
		err = uc.qualificationRepo.UpdateHealthCertificateStatus(ctx, riderID, 1, "")
	} else {
		err = uc.qualificationRepo.CreateHealthCertificateVerification(ctx, verification)
	}

	if err != nil {
		return err
	}

	// 3. 更新总体资质状态
	err = uc.updateOverallStatus(ctx, riderID)
	return err
}

// GetVerificationStatus 获取验证状态
func (uc *qualificationUsecase) GetVerificationStatus(ctx context.Context, riderID int64) (*VerificationStatus, error) {
	// 1. 获取总体资质
	qualification, err := uc.qualificationRepo.GetQualificationByRiderID(ctx, riderID)
	if err != nil {
		return nil, err
	}

	if qualification == nil {
		qualification = &model.RiderQualification{
			RiderID:       riderID,
			OverallStatus: 0,
			IDCardStatus:  0,
			LicenseStatus: 0,
			HealthStatus:  0,
		}
	}

	// 2. 获取各项验证详情
	idCard, _ := uc.qualificationRepo.GetIDCardVerification(ctx, riderID)
	license, _ := uc.qualificationRepo.GetDriverLicenseVerification(ctx, riderID)
	health, _ := uc.qualificationRepo.GetHealthCertificateVerification(ctx, riderID)

	// 3. 构建响应
	result := &VerificationStatus{
		OverallStatus: OverallStatus{
			Status:     qualification.OverallStatus,
			StatusText: uc.getOverallStatusText(qualification.OverallStatus),
			Message:    uc.getOverallStatusMessage(qualification.OverallStatus),
		},
		Items: []VerificationItem{},
	}

	// 身份证
	item := VerificationItem{
		ItemType:   "idcard",
		ItemName:   "身份证",
		Status:     qualification.IDCardStatus,
		StatusText: uc.getItemStatusText(qualification.IDCardStatus),
	}
	if idCard != nil {
		item.SubmitTime = idCard.SubmitTime.Format(time.RFC3339)
		if idCard.VerifyTime != nil {
			item.VerifyTime = idCard.VerifyTime.Format(time.RFC3339)
		}
		item.RejectReason = idCard.RejectReason
	}
	result.Items = append(result.Items, item)

	// 驾驶证
	item = VerificationItem{
		ItemType:   "driver_license",
		ItemName:   "驾驶证",
		Status:     qualification.LicenseStatus,
		StatusText: uc.getItemStatusText(qualification.LicenseStatus),
	}
	if license != nil {
		item.SubmitTime = license.SubmitTime.Format(time.RFC3339)
		if license.VerifyTime != nil {
			item.VerifyTime = license.VerifyTime.Format(time.RFC3339)
		}
		item.RejectReason = license.RejectReason
		item.ExpiryDate = license.ExpiryDate
		item.DaysToExpiry = uc.calculateDaysToExpiry(license.ExpiryDate)
	}
	result.Items = append(result.Items, item)

	// 健康证
	item = VerificationItem{
		ItemType:   "health_certificate",
		ItemName:   "健康证",
		Status:     qualification.HealthStatus,
		StatusText: uc.getItemStatusText(qualification.HealthStatus),
	}
	if health != nil {
		item.SubmitTime = health.SubmitTime.Format(time.RFC3339)
		if health.VerifyTime != nil {
			item.VerifyTime = health.VerifyTime.Format(time.RFC3339)
		}
		item.RejectReason = health.RejectReason
		item.ExpiryDate = health.ExpiryDate
		item.DaysToExpiry = uc.calculateDaysToExpiry(health.ExpiryDate)
	}
	result.Items = append(result.Items, item)

	return result, nil
}

// GetVerificationDetail 获取验证详情
func (uc *qualificationUsecase) GetVerificationDetail(ctx context.Context, riderID int64, itemType string) (*VerificationDetail, error) {
	var detail *VerificationDetail

	switch itemType {
	case "idcard":
		verification, err := uc.qualificationRepo.GetIDCardVerification(ctx, riderID)
		if err != nil {
			return nil, err
		}
		if verification == nil {
			return nil, errors.New("未找到身份证验证记录")
		}
		detail = &VerificationDetail{
			ID:           verification.ID,
			RiderID:      verification.RiderID,
			ItemType:     "idcard",
			ItemName:     "身份证",
			Status:       verification.Status,
			StatusText:   uc.getItemStatusText(verification.Status),
			SubmitData:   verification.RealName,
			VerifyResult: verification.VerifyResult,
			RejectReason: verification.RejectReason,
			SubmitTime:   verification.SubmitTime.Format(time.RFC3339),
		}
		if verification.VerifyTime != nil {
			detail.VerifyTime = verification.VerifyTime.Format(time.RFC3339)
		}

	case "driver_license":
		verification, err := uc.qualificationRepo.GetDriverLicenseVerification(ctx, riderID)
		if err != nil {
			return nil, err
		}
		if verification == nil {
			return nil, errors.New("未找到驾驶证验证记录")
		}
		detail = &VerificationDetail{
			ID:           verification.ID,
			RiderID:      verification.RiderID,
			ItemType:     "driver_license",
			ItemName:     "驾驶证",
			Status:       verification.Status,
			StatusText:   uc.getItemStatusText(verification.Status),
			SubmitData:   verification.LicenseNumber + " (" + verification.LicenseType + ")",
			VerifyResult: verification.VerifyResult,
			RejectReason: verification.RejectReason,
			SubmitTime:   verification.SubmitTime.Format(time.RFC3339),
		}
		if verification.VerifyTime != nil {
			detail.VerifyTime = verification.VerifyTime.Format(time.RFC3339)
		}

	case "health_certificate":
		verification, err := uc.qualificationRepo.GetHealthCertificateVerification(ctx, riderID)
		if err != nil {
			return nil, err
		}
		if verification == nil {
			return nil, errors.New("未找到健康证验证记录")
		}
		detail = &VerificationDetail{
			ID:           verification.ID,
			RiderID:      verification.RiderID,
			ItemType:     "health_certificate",
			ItemName:     "健康证",
			Status:       verification.Status,
			StatusText:   uc.getItemStatusText(verification.Status),
			SubmitData:   verification.CertificateNumber,
			VerifyResult: verification.VerifyResult,
			RejectReason: verification.RejectReason,
			SubmitTime:   verification.SubmitTime.Format(time.RFC3339),
		}
		if verification.VerifyTime != nil {
			detail.VerifyTime = verification.VerifyTime.Format(time.RFC3339)
		}

	default:
		return nil, errors.New("未知的验证类型")
	}

	return detail, nil
}

// ResubmitVerification 重新提交验证
func (uc *qualificationUsecase) ResubmitVerification(ctx context.Context, riderID int64, itemType, reason string) error {
	// 重新提交时，将状态重置为未提交
	switch itemType {
	case "idcard":
		return uc.qualificationRepo.UpdateIDCardStatus(ctx, riderID, 0, reason)
	case "driver_license":
		return uc.qualificationRepo.UpdateDriverLicenseStatus(ctx, riderID, 0, reason)
	case "health_certificate":
		return uc.qualificationRepo.UpdateHealthCertificateStatus(ctx, riderID, 0, reason)
	default:
		return errors.New("未知的验证类型")
	}
}

// VerifyQualification 审核验证（管理员接口）
func (uc *qualificationUsecase) VerifyQualification(ctx context.Context, riderID int64, itemType string, status int32, rejectReason, verifyResult string) error {
	// 验证状态值
	if status != 2 && status != 3 {
		return errors.New("无效的审核状态，必须为 2（通过）或 3（拒绝）")
	}

	// 如果拒绝，必须提供拒绝原因
	if status == 3 && rejectReason == "" {
		return errors.New("拒绝审核时必须提供拒绝原因")
	}

	// 执行审核操作
	switch itemType {
	case "idcard":
		err := uc.qualificationRepo.UpdateIDCardStatus(ctx, riderID, status, rejectReason)
		if err != nil {
			return err
		}
		// 更新验证结果
		err = uc.qualificationRepo.UpdateIDCardVerifyResult(ctx, riderID, verifyResult)
		if err != nil {
			return err
		}
		// 更新总体资质表中的身份证状态
		err = uc.qualificationRepo.UpdateIDCardOverallStatus(ctx, riderID, status)
		if err != nil {
			return err
		}
	case "driver_license":
		err := uc.qualificationRepo.UpdateDriverLicenseStatus(ctx, riderID, status, rejectReason)
		if err != nil {
			return err
		}
		// 更新验证结果
		err = uc.qualificationRepo.UpdateDriverLicenseVerifyResult(ctx, riderID, verifyResult)
		if err != nil {
			return err
		}
		// 更新总体资质表中的驾驶证状态
		err = uc.qualificationRepo.UpdateDriverLicenseOverallStatus(ctx, riderID, status)
		if err != nil {
			return err
		}
	case "health_certificate":
		err := uc.qualificationRepo.UpdateHealthCertificateStatus(ctx, riderID, status, rejectReason)
		if err != nil {
			return err
		}
		// 更新验证结果
		err = uc.qualificationRepo.UpdateHealthCertificateVerifyResult(ctx, riderID, verifyResult)
		if err != nil {
			return err
		}
		// 更新总体资质表中的健康证状态
		err = uc.qualificationRepo.UpdateHealthCertificateOverallStatus(ctx, riderID, status)
		if err != nil {
			return err
		}
	default:
		return errors.New("未知的验证类型")
	}

	return nil
}

// updateOverallStatus 更新总体状态
func (uc *qualificationUsecase) updateOverallStatus(ctx context.Context, riderID int64) error {
	// 重新从数据库获取最新状态（确保读取到最新的更新值）
	qualification, err := uc.qualificationRepo.GetQualificationByRiderID(ctx, riderID)
	if err != nil {
		return err
	}

	if qualification == nil {
		// 创建新的资质记录
		qualification = &model.RiderQualification{
			RiderID:       riderID,
			OverallStatus: 1, // 认证中
		}
		return uc.qualificationRepo.CreateQualification(ctx, qualification)
	}

	// 检查所有项是否都已通过
	if qualification.IDCardStatus == 2 && qualification.LicenseStatus == 2 && qualification.HealthStatus == 2 {
		return uc.qualificationRepo.UpdateOverallStatus(ctx, riderID, 2) // 认证通过
	}

	// 检查是否有被拒绝的项
	if qualification.IDCardStatus == 3 || qualification.LicenseStatus == 3 || qualification.HealthStatus == 3 {
		return uc.qualificationRepo.UpdateOverallStatus(ctx, riderID, 3) // 认证失败
	}

	// 否则为认证中
	return uc.qualificationRepo.UpdateOverallStatus(ctx, riderID, 1)
}

// getOverallStatusText 获取总体状态文本
func (uc *qualificationUsecase) getOverallStatusText(status int32) string {
	switch status {
	case 0:
		return "未认证"
	case 1:
		return "认证中"
	case 2:
		return "认证通过"
	case 3:
		return "认证失败"
	default:
		return "未知"
	}
}

// getOverallStatusMessage 获取总体状态提示信息
func (uc *qualificationUsecase) getOverallStatusMessage(status int32) string {
	switch status {
	case 0:
		return "请先完成实名认证"
	case 1:
		return "您的资料正在审核中，请耐心等待"
	case 2:
		return "恭喜您，认证已通过"
	case 3:
		return "认证未通过，请查看拒绝原因并重新提交"
	default:
		return ""
	}
}

// getItemStatusText 获取单项状态文本
func (uc *qualificationUsecase) getItemStatusText(status int32) string {
	switch status {
	case 0:
		return "未提交"
	case 1:
		return "审核中"
	case 2:
		return "已通过"
	case 3:
		return "已拒绝"
	default:
		return "未知"
	}
}

// calculateDaysToExpiry 计算距离过期天数
func (uc *qualificationUsecase) calculateDaysToExpiry(expiryDate string) int32 {
	if expiryDate == "" {
		return -1
	}

	expiry, err := time.Parse("2006-01-02", expiryDate)
	if err != nil {
		return -1
	}

	days := int32(expiry.Sub(time.Now()).Hours() / 24)
	if days < 0 {
		return 0
	}
	return days
}
