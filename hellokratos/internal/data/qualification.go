package data

import (
	"context"
	"hellokratos/internal/data/model"

	"gorm.io/gorm"
)

// QualificationRepo 资质验证数据访问接口
type QualificationRepo interface {
	// 总体资质
	GetQualificationByRiderID(ctx context.Context, riderID int64) (*model.RiderQualification, error)
	CreateQualification(ctx context.Context, qualification *model.RiderQualification) error
	UpdateOverallStatus(ctx context.Context, riderID int64, status int32) error

	// 身份证验证
	GetIDCardVerification(ctx context.Context, riderID int64) (*model.IDCardVerification, error)
	CreateIDCardVerification(ctx context.Context, verification *model.IDCardVerification) error
	UpdateIDCardStatus(ctx context.Context, riderID int64, status int32, rejectReason string) error
	UpdateIDCardVerifyResult(ctx context.Context, riderID int64, verifyResult string) error
	UpdateIDCardOverallStatus(ctx context.Context, riderID int64, status int32) error

	// 驾驶证验证
	GetDriverLicenseVerification(ctx context.Context, riderID int64) (*model.DriverLicenseVerification, error)
	CreateDriverLicenseVerification(ctx context.Context, verification *model.DriverLicenseVerification) error
	UpdateDriverLicenseStatus(ctx context.Context, riderID int64, status int32, rejectReason string) error
	UpdateDriverLicenseVerifyResult(ctx context.Context, riderID int64, verifyResult string) error
	UpdateDriverLicenseOverallStatus(ctx context.Context, riderID int64, status int32) error

	// 健康证验证
	GetHealthCertificateVerification(ctx context.Context, riderID int64) (*model.HealthCertificateVerification, error)
	CreateHealthCertificateVerification(ctx context.Context, verification *model.HealthCertificateVerification) error
	UpdateHealthCertificateStatus(ctx context.Context, riderID int64, status int32, rejectReason string) error
	UpdateHealthCertificateVerifyResult(ctx context.Context, riderID int64, verifyResult string) error
	UpdateHealthCertificateOverallStatus(ctx context.Context, riderID int64, status int32) error
}

// qualificationRepo 资质验证数据访问实现
type qualificationRepo struct {
	data *Data
}

// NewQualificationRepo 创建资质验证数据访问实例
func NewQualificationRepo(data *Data) QualificationRepo {
	return &qualificationRepo{
		data: data,
	}
}

// GetQualificationByRiderID 获取骑手的资质信息
func (r *qualificationRepo) GetQualificationByRiderID(ctx context.Context, riderID int64) (*model.RiderQualification, error) {
	var qualification model.RiderQualification
	err := r.data.db.WithContext(ctx).Where("rider_id = ?", riderID).First(&qualification).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &qualification, err
}

// CreateQualification 创建资质记录
func (r *qualificationRepo) CreateQualification(ctx context.Context, qualification *model.RiderQualification) error {
	return r.data.db.WithContext(ctx).Create(qualification).Error
}

// UpdateOverallStatus 更新总体状态
func (r *qualificationRepo) UpdateOverallStatus(ctx context.Context, riderID int64, status int32) error {
	return r.data.db.WithContext(ctx).Model(&model.RiderQualification{}).
		Where("rider_id = ?", riderID).
		Update("overall_status", status).Error
}

// GetIDCardVerification 获取身份证验证信息
func (r *qualificationRepo) GetIDCardVerification(ctx context.Context, riderID int64) (*model.IDCardVerification, error) {
	var verification model.IDCardVerification
	err := r.data.db.WithContext(ctx).Where("rider_id = ?", riderID).First(&verification).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &verification, err
}

// CreateIDCardVerification 创建身份证验证记录
func (r *qualificationRepo) CreateIDCardVerification(ctx context.Context, verification *model.IDCardVerification) error {
	return r.data.db.WithContext(ctx).Create(verification).Error
}

// UpdateIDCardStatus 更新身份证验证状态
func (r *qualificationRepo) UpdateIDCardStatus(ctx context.Context, riderID int64, status int32, rejectReason string) error {
	updates := map[string]interface{}{
		"status":      status,
		"verify_time": gorm.Expr("NOW()"),
	}
	if rejectReason != "" {
		updates["reject_reason"] = rejectReason
	}

	return r.data.db.WithContext(ctx).Model(&model.IDCardVerification{}).
		Where("rider_id = ?", riderID).
		Updates(updates).Error
}

// GetDriverLicenseVerification 获取驾驶证验证信息
func (r *qualificationRepo) GetDriverLicenseVerification(ctx context.Context, riderID int64) (*model.DriverLicenseVerification, error) {
	var verification model.DriverLicenseVerification
	err := r.data.db.WithContext(ctx).Where("rider_id = ?", riderID).First(&verification).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &verification, err
}

// CreateDriverLicenseVerification 创建驾驶证验证记录
func (r *qualificationRepo) CreateDriverLicenseVerification(ctx context.Context, verification *model.DriverLicenseVerification) error {
	return r.data.db.WithContext(ctx).Create(verification).Error
}

// UpdateDriverLicenseStatus 更新驾驶证验证状态
func (r *qualificationRepo) UpdateDriverLicenseStatus(ctx context.Context, riderID int64, status int32, rejectReason string) error {
	updates := map[string]interface{}{
		"status":      status,
		"verify_time": gorm.Expr("NOW()"),
	}
	if rejectReason != "" {
		updates["reject_reason"] = rejectReason
	}

	return r.data.db.WithContext(ctx).Model(&model.DriverLicenseVerification{}).
		Where("rider_id = ?", riderID).
		Updates(updates).Error
}

// GetHealthCertificateVerification 获取健康证验证信息
func (r *qualificationRepo) GetHealthCertificateVerification(ctx context.Context, riderID int64) (*model.HealthCertificateVerification, error) {
	var verification model.HealthCertificateVerification
	err := r.data.db.WithContext(ctx).Where("rider_id = ?", riderID).First(&verification).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &verification, err
}

// CreateHealthCertificateVerification 创建健康证验证记录
func (r *qualificationRepo) CreateHealthCertificateVerification(ctx context.Context, verification *model.HealthCertificateVerification) error {
	return r.data.db.WithContext(ctx).Create(verification).Error
}

// UpdateHealthCertificateStatus 更新健康证验证状态
func (r *qualificationRepo) UpdateHealthCertificateStatus(ctx context.Context, riderID int64, status int32, rejectReason string) error {
	updates := map[string]interface{}{
		"status":      status,
		"verify_time": gorm.Expr("NOW()"),
	}
	if rejectReason != "" {
		updates["reject_reason"] = rejectReason
	}

	return r.data.db.WithContext(ctx).Model(&model.HealthCertificateVerification{}).
		Where("rider_id = ?", riderID).
		Updates(updates).Error
}

// UpdateIDCardVerifyResult 更新身份证验证结果
func (r *qualificationRepo) UpdateIDCardVerifyResult(ctx context.Context, riderID int64, verifyResult string) error {
	return r.data.db.WithContext(ctx).Model(&model.IDCardVerification{}).
		Where("rider_id = ?", riderID).
		Update("verify_result", verifyResult).Error
}

// UpdateDriverLicenseVerifyResult 更新驾驶证验证结果
func (r *qualificationRepo) UpdateDriverLicenseVerifyResult(ctx context.Context, riderID int64, verifyResult string) error {
	return r.data.db.WithContext(ctx).Model(&model.DriverLicenseVerification{}).
		Where("rider_id = ?", riderID).
		Update("verify_result", verifyResult).Error
}

// UpdateHealthCertificateVerifyResult 更新健康证验证结果
func (r *qualificationRepo) UpdateHealthCertificateVerifyResult(ctx context.Context, riderID int64, verifyResult string) error {
	return r.data.db.WithContext(ctx).Model(&model.HealthCertificateVerification{}).
		Where("rider_id = ?", riderID).
		Update("verify_result", verifyResult).Error
}

// UpdateIDCardOverallStatus 更新总体资质表中的身份证状态
func (r *qualificationRepo) UpdateIDCardOverallStatus(ctx context.Context, riderID int64, status int32) error {
	// 先更新身份证状态
	err := r.data.db.WithContext(ctx).Model(&model.RiderQualification{}).
		Where("rider_id = ?", riderID).
		Update("id_card_status", status).Error
	if err != nil {
		return err
	}

	// 重新查询并更新总体状态
	return r.updateOverallStatusByRiderID(ctx, riderID)
}

// UpdateDriverLicenseOverallStatus 更新总体资质表中的驾驶证状态
func (r *qualificationRepo) UpdateDriverLicenseOverallStatus(ctx context.Context, riderID int64, status int32) error {
	// 先更新驾驶证状态
	err := r.data.db.WithContext(ctx).Model(&model.RiderQualification{}).
		Where("rider_id = ?", riderID).
		Update("license_status", status).Error
	if err != nil {
		return err
	}

	// 重新查询并更新总体状态
	return r.updateOverallStatusByRiderID(ctx, riderID)
}

// UpdateHealthCertificateOverallStatus 更新总体资质表中的健康证状态
func (r *qualificationRepo) UpdateHealthCertificateOverallStatus(ctx context.Context, riderID int64, status int32) error {
	// 先更新健康证状态
	err := r.data.db.WithContext(ctx).Model(&model.RiderQualification{}).
		Where("rider_id = ?", riderID).
		Update("health_status", status).Error
	if err != nil {
		return err
	}

	// 重新查询并更新总体状态
	return r.updateOverallStatusByRiderID(ctx, riderID)
}

// updateOverallStatusByRiderID 根据骑手ID更新总体状态
func (r *qualificationRepo) updateOverallStatusByRiderID(ctx context.Context, riderID int64) error {
	// 使用原生SQL查询获取最新数据，避免GORM缓存
	var idCardStatus, licenseStatus, healthStatus int32
	err := r.data.db.WithContext(ctx).Raw(
		"SELECT id_card_status, license_status, health_status FROM rider_qualification WHERE rider_id = ?",
		riderID,
	).Row().Scan(&idCardStatus, &licenseStatus, &healthStatus)

	if err != nil {
		return err
	}

	// 计算总体状态
	var overallStatus int32
	if idCardStatus == 2 && licenseStatus == 2 && healthStatus == 2 {
		overallStatus = 2 // 认证通过
	} else if idCardStatus == 3 || licenseStatus == 3 || healthStatus == 3 {
		overallStatus = 3 // 认证失败
	} else {
		overallStatus = 1 // 认证中
	}

	// 更新总体状态
	return r.data.db.WithContext(ctx).Model(&model.RiderQualification{}).
		Where("rider_id = ?", riderID).
		Update("overall_status", overallStatus).Error
}
