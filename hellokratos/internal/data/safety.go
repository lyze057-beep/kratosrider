package data

import (
	"context"
	"fmt"
	"time"

	"hellokratos/internal/data/model"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

// SafetyRepo 安全数据访问接口
type SafetyRepo interface {
	// 紧急求助相关
	CreateEmergencyHelp(ctx context.Context, help *model.EmergencyHelp) error
	GetEmergencyHelpByID(ctx context.Context, id int64) (*model.EmergencyHelp, error)
	GetEmergencyHelpsByRiderID(ctx context.Context, riderID int64, status int32, page, pageSize int) ([]*model.EmergencyHelp, int64, error)
	CancelEmergencyHelp(ctx context.Context, id int64, cancelReason string) error

	// 保险理赔相关
	CreateInsuranceClaim(ctx context.Context, claim *model.InsuranceClaim) error
	GetInsuranceClaimByID(ctx context.Context, id int64) (*model.InsuranceClaim, error)
	GetInsuranceClaimsByRiderID(ctx context.Context, riderID int64, status int32, page, pageSize int) ([]*model.InsuranceClaim, int64, error)

	// 保险信息相关
	GetRiderInsurance(ctx context.Context, riderID int64) (*model.RiderInsurance, error)
	UpdateInsuranceClaimStats(ctx context.Context, riderID int64, claimAmount string) error

	// 安全事件相关
	CreateSafetyEvent(ctx context.Context, event *model.SafetyEvent) error

	// 安全提示相关
	GetSafetyTips(ctx context.Context, tipType string, page, pageSize int) ([]*model.SafetyTip, int64, error)

	// 紧急联系人相关
	GetEmergencyContacts(ctx context.Context, riderID int64) ([]*model.EmergencyContact, error)
	SaveEmergencyContacts(ctx context.Context, riderID int64, contacts []*model.EmergencyContact) error
}

// safetyRepo 安全数据访问实现
type safetyRepo struct {
	data *Data
	log  *log.Helper
}

// NewSafetyRepo 创建安全数据访问实例
func NewSafetyRepo(data *Data, logger log.Logger) SafetyRepo {
	return &safetyRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// CreateEmergencyHelp 创建紧急求助
func (r *safetyRepo) CreateEmergencyHelp(ctx context.Context, help *model.EmergencyHelp) error {
	return r.data.db.WithContext(ctx).Create(help).Error
}

// GetEmergencyHelpByID 根据ID获取紧急求助
func (r *safetyRepo) GetEmergencyHelpByID(ctx context.Context, id int64) (*model.EmergencyHelp, error) {
	var help model.EmergencyHelp
	err := r.data.db.WithContext(ctx).First(&help, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &help, nil
}

// GetEmergencyHelpsByRiderID 获取骑手的紧急求助列表
func (r *safetyRepo) GetEmergencyHelpsByRiderID(ctx context.Context, riderID int64, status int32, page, pageSize int) ([]*model.EmergencyHelp, int64, error) {
	var helps []*model.EmergencyHelp
	var total int64

	query := r.data.db.WithContext(ctx).Model(&model.EmergencyHelp{}).Where("rider_id = ?", riderID)
	if status > 0 {
		query = query.Where("status = ?", status)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&helps).Error
	if err != nil {
		return nil, 0, err
	}

	return helps, total, nil
}

// CancelEmergencyHelp 取消紧急求助
func (r *safetyRepo) CancelEmergencyHelp(ctx context.Context, id int64, cancelReason string) error {
	now := time.Now()
	return r.data.db.WithContext(ctx).Model(&model.EmergencyHelp{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":         4, // 已取消
		"cancel_reason":  cancelReason,
		"resolved_at":    &now,
	}).Error
}

// CreateInsuranceClaim 创建保险理赔申请
func (r *safetyRepo) CreateInsuranceClaim(ctx context.Context, claim *model.InsuranceClaim) error {
	return r.data.db.WithContext(ctx).Create(claim).Error
}

// GetInsuranceClaimByID 根据ID获取保险理赔
func (r *safetyRepo) GetInsuranceClaimByID(ctx context.Context, id int64) (*model.InsuranceClaim, error) {
	var claim model.InsuranceClaim
	err := r.data.db.WithContext(ctx).First(&claim, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &claim, nil
}

// GetInsuranceClaimsByRiderID 获取骑手的保险理赔列表
func (r *safetyRepo) GetInsuranceClaimsByRiderID(ctx context.Context, riderID int64, status int32, page, pageSize int) ([]*model.InsuranceClaim, int64, error) {
	var claims []*model.InsuranceClaim
	var total int64

	query := r.data.db.WithContext(ctx).Model(&model.InsuranceClaim{}).Where("rider_id = ?", riderID)
	if status > 0 {
		query = query.Where("status = ?", status)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&claims).Error
	if err != nil {
		return nil, 0, err
	}

	return claims, total, nil
}

// GetRiderInsurance 获取骑手保险信息
func (r *safetyRepo) GetRiderInsurance(ctx context.Context, riderID int64) (*model.RiderInsurance, error) {
	var insurance model.RiderInsurance
	err := r.data.db.WithContext(ctx).Where("rider_id = ?", riderID).First(&insurance).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &insurance, nil
}

// UpdateInsuranceClaimStats 更新保险理赔统计
func (r *safetyRepo) UpdateInsuranceClaimStats(ctx context.Context, riderID int64, claimAmount string) error {
	return r.data.db.WithContext(ctx).Model(&model.RiderInsurance{}).Where("rider_id = ?", riderID).Updates(map[string]interface{}{
		"claim_count":       gorm.Expr("claim_count + 1"),
		"total_claim_amount": gorm.Expr("total_claim_amount + ?", claimAmount),
	}).Error
}

// CreateSafetyEvent 创建安全事件
func (r *safetyRepo) CreateSafetyEvent(ctx context.Context, event *model.SafetyEvent) error {
	return r.data.db.WithContext(ctx).Create(event).Error
}

// GetSafetyTips 获取安全提示列表
func (r *safetyRepo) GetSafetyTips(ctx context.Context, tipType string, page, pageSize int) ([]*model.SafetyTip, int64, error) {
	var tips []*model.SafetyTip
	var total int64

	now := time.Now()
	query := r.data.db.WithContext(ctx).Model(&model.SafetyTip{}).Where("is_active = ?", true)
	query = query.Where("(expire_time IS NULL OR expire_time > ?)", now)

	if tipType != "" {
		query = query.Where("tip_type = ?", tipType)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Order("priority DESC, publish_time DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&tips).Error
	if err != nil {
		return nil, 0, err
	}

	return tips, total, nil
}

// GetEmergencyContacts 获取紧急联系人
func (r *safetyRepo) GetEmergencyContacts(ctx context.Context, riderID int64) ([]*model.EmergencyContact, error) {
	var contacts []*model.EmergencyContact
	err := r.data.db.WithContext(ctx).Where("rider_id = ?", riderID).Order("is_primary DESC, sort_order ASC").Find(&contacts).Error
	return contacts, err
}

// SaveEmergencyContacts 保存紧急联系人
func (r *safetyRepo) SaveEmergencyContacts(ctx context.Context, riderID int64, contacts []*model.EmergencyContact) error {
	return r.data.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 删除原有联系人
		if err := tx.Where("rider_id = ?", riderID).Delete(&model.EmergencyContact{}).Error; err != nil {
			return err
		}
		// 创建新联系人
		for _, contact := range contacts {
			contact.RiderID = riderID
			if err := tx.Create(contact).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// GenerateClaimNo 生成理赔编号
func GenerateClaimNo() string {
	return fmt.Sprintf("CL%s%d", time.Now().Format("20060102"), time.Now().UnixNano()%1000000)
}
