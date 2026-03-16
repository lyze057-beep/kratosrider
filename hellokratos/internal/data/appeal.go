package data

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"hellokratos/internal/data/model"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

// AppealRepo 申诉数据访问接口
type AppealRepo interface {
	// 申诉相关
	CreateAppeal(ctx context.Context, appeal *model.Appeal) error
	GetAppealByID(ctx context.Context, id int64) (*model.Appeal, error)
	GetAppealsByRiderID(ctx context.Context, riderID int64, status int32, appealType string, page, pageSize int) ([]*model.Appeal, int64, error)
	UpdateAppealStatus(ctx context.Context, id int64, status int32, result, reply string) error
	CancelAppeal(ctx context.Context, id int64, cancelReason string) error

	// 申诉类型相关
	GetAppealTypes(ctx context.Context, category int32) ([]*model.AppealType, error)

	// 异常报备相关
	CreateExceptionReport(ctx context.Context, report *model.ExceptionReport) error
	GetExceptionReportsByRiderID(ctx context.Context, riderID int64, page, pageSize int) ([]*model.ExceptionReport, int64, error)

	// 异常订单相关
	GetExceptionOrdersByRiderID(ctx context.Context, riderID int64, startDate, endDate string, page, pageSize int) ([]*model.ExceptionOrder, int64, error)
}

// appealRepo 申诉数据访问实现
type appealRepo struct {
	data *Data
	log  *log.Helper
}

// NewAppealRepo 创建申诉数据访问实例
func NewAppealRepo(data *Data, logger log.Logger) AppealRepo {
	return &appealRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// CreateAppeal 创建申诉
func (r *appealRepo) CreateAppeal(ctx context.Context, appeal *model.Appeal) error {
	return r.data.db.WithContext(ctx).Create(appeal).Error
}

// GetAppealByID 根据ID获取申诉
func (r *appealRepo) GetAppealByID(ctx context.Context, id int64) (*model.Appeal, error) {
	var appeal model.Appeal
	err := r.data.db.WithContext(ctx).First(&appeal, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &appeal, nil
}

// GetAppealsByRiderID 获取骑手的申诉列表
func (r *appealRepo) GetAppealsByRiderID(ctx context.Context, riderID int64, status int32, appealType string, page, pageSize int) ([]*model.Appeal, int64, error) {
	var appeals []*model.Appeal
	var total int64

	query := r.data.db.WithContext(ctx).Model(&model.Appeal{}).Where("rider_id = ?", riderID)

	if status > 0 {
		query = query.Where("status = ?", status)
	}
	if appealType != "" {
		query = query.Where("appeal_type = ?", appealType)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&appeals).Error
	if err != nil {
		return nil, 0, err
	}

	return appeals, total, nil
}

// UpdateAppealStatus 更新申诉状态
func (r *appealRepo) UpdateAppealStatus(ctx context.Context, id int64, status int32, result, reply string) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if result != "" {
		updates["result"] = result
	}
	if reply != "" {
		updates["reply"] = reply
	}
	if status == 3 || status == 4 { // 已通过或已驳回
		now := time.Now()
		updates["resolved_at"] = &now
	}

	return r.data.db.WithContext(ctx).Model(&model.Appeal{}).Where("id = ?", id).Updates(updates).Error
}

// CancelAppeal 取消申诉
func (r *appealRepo) CancelAppeal(ctx context.Context, id int64, cancelReason string) error {
	now := time.Now()
	return r.data.db.WithContext(ctx).Model(&model.Appeal{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":         5, // 已取消
		"cancel_reason":  cancelReason,
		"resolved_at":    &now,
	}).Error
}

// GetAppealTypes 获取申诉类型列表
func (r *appealRepo) GetAppealTypes(ctx context.Context, category int32) ([]*model.AppealType, error) {
	var types []*model.AppealType
	query := r.data.db.WithContext(ctx).Where("is_active = ?", true)
	if category > 0 {
		query = query.Where("appeal_category = ?", category)
	}
	err := query.Order("sort_order ASC").Find(&types).Error
	return types, err
}

// CreateExceptionReport 创建异常报备
func (r *appealRepo) CreateExceptionReport(ctx context.Context, report *model.ExceptionReport) error {
	return r.data.db.WithContext(ctx).Create(report).Error
}

// GetExceptionReportsByRiderID 获取骑手的异常报备列表
func (r *appealRepo) GetExceptionReportsByRiderID(ctx context.Context, riderID int64, page, pageSize int) ([]*model.ExceptionReport, int64, error) {
	var reports []*model.ExceptionReport
	var total int64

	query := r.data.db.WithContext(ctx).Model(&model.ExceptionReport{}).Where("rider_id = ?", riderID)

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&reports).Error
	if err != nil {
		return nil, 0, err
	}

	return reports, total, nil
}

// GetExceptionOrdersByRiderID 获取骑手的异常订单列表
func (r *appealRepo) GetExceptionOrdersByRiderID(ctx context.Context, riderID int64, startDate, endDate string, page, pageSize int) ([]*model.ExceptionOrder, int64, error) {
	var orders []*model.ExceptionOrder
	var total int64

	query := r.data.db.WithContext(ctx).Model(&model.ExceptionOrder{}).Where("rider_id = ?", riderID)

	if startDate != "" && endDate != "" {
		query = query.Where("occurred_at BETWEEN ? AND ?", startDate, endDate)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Order("occurred_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&orders).Error
	if err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

// GenerateTicketNo 生成工单号
func GenerateTicketNo() string {
	return fmt.Sprintf("AP%s%d", time.Now().Format("20060102"), time.Now().UnixNano()%1000000)
}

// StringifySlice 将字符串切片转换为JSON字符串
func StringifySlice(slice []string) string {
	if len(slice) == 0 {
		return "[]"
	}
	data, _ := json.Marshal(slice)
	return string(data)
}

// ParseSlice 将JSON字符串解析为字符串切片
func ParseSlice(data string) []string {
	var slice []string
	json.Unmarshal([]byte(data), &slice)
	return slice
}
