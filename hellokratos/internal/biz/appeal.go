package biz

import (
	"context"
	"errors"
	"hellokratos/internal/data"
	"hellokratos/internal/data/model"

	"github.com/go-kratos/kratos/v2/log"
)

// AppealUsecase 申诉业务逻辑接口
type AppealUsecase interface {
	// 申诉相关
	SubmitOrderAppeal(ctx context.Context, riderID, orderID int64, appealType, reason, description string, evidenceImages []string, contactPhone string) (*model.Appeal, error)
	SubmitPenaltyAppeal(ctx context.Context, riderID int64, penaltyType, penaltyID, appealType, reason, description string, evidenceImages []string, contactPhone string) (*model.Appeal, error)
	GetAppeals(ctx context.Context, riderID int64, status int32, appealType string, page, pageSize int) ([]*model.Appeal, int64, error)
	GetAppealDetail(ctx context.Context, appealID, riderID int64) (*model.Appeal, error)
	CancelAppeal(ctx context.Context, appealID, riderID int64, cancelReason string) error

	// 申诉类型相关
	GetAppealTypes(ctx context.Context, category int32) ([]*model.AppealType, error)

	// 异常报备相关
	SubmitExceptionReport(ctx context.Context, riderID, orderID int64, exceptionType, description string, images []string, location string, latitude, longitude float64) (*model.ExceptionReport, error)
	GetExceptionOrders(ctx context.Context, riderID int64, startDate, endDate string, page, pageSize int) ([]*model.ExceptionOrder, int64, error)
}

// appealUsecase 申诉业务逻辑实现
type appealUsecase struct {
	appealRepo data.AppealRepo
	orderRepo  data.OrderRepo
	log        *log.Helper
}

// NewAppealUsecase 创建申诉业务逻辑实例
func NewAppealUsecase(appealRepo data.AppealRepo, orderRepo data.OrderRepo, logger log.Logger) AppealUsecase {
	return &appealUsecase{
		appealRepo: appealRepo,
		orderRepo:  orderRepo,
		log:        log.NewHelper(logger),
	}
}

// SubmitOrderAppeal 提交订单申诉
func (uc *appealUsecase) SubmitOrderAppeal(ctx context.Context, riderID, orderID int64, appealType, reason, description string, evidenceImages []string, contactPhone string) (*model.Appeal, error) {
	// 1. 参数校验
	if riderID <= 0 || orderID <= 0 {
		return nil, errors.New("无效的骑手或订单ID")
	}
	if reason == "" {
		return nil, errors.New("申诉原因不能为空")
	}

	// 2. 检查订单是否存在
	order, err := uc.orderRepo.GetOrderByID(ctx, orderID)
	if err != nil {
		uc.log.Error("查询订单失败", "err", err)
		return nil, errors.New("订单不存在")
	}
	if order.RiderID != riderID {
		return nil, errors.New("无权申诉该订单")
	}

	// 3. 创建申诉
	appeal := &model.Appeal{
		TicketNo:       generateTicketNo(),
		RiderID:        riderID,
		AppealType:     1, // 订单申诉
		OrderID:        orderID,
		Reason:         reason,
		Description:    description,
		EvidenceImages: stringifySlice(evidenceImages),
		ContactPhone:   contactPhone,
		Status:         1, // 待处理
	}

	err = uc.appealRepo.CreateAppeal(ctx, appeal)
	if err != nil {
		uc.log.Error("创建申诉失败", "err", err)
		return nil, errors.New("提交申诉失败")
	}

	uc.log.Info("订单申诉提交成功", "appeal_id", appeal.ID, "ticket_no", appeal.TicketNo)
	return appeal, nil
}

// SubmitPenaltyAppeal 提交处罚申诉
func (uc *appealUsecase) SubmitPenaltyAppeal(ctx context.Context, riderID int64, penaltyType, penaltyID, appealType, reason, description string, evidenceImages []string, contactPhone string) (*model.Appeal, error) {
	// 1. 参数校验
	if riderID <= 0 {
		return nil, errors.New("无效的骑手ID")
	}
	if penaltyType == "" || penaltyID == "" {
		return nil, errors.New("处罚类型和处罚ID不能为空")
	}
	if reason == "" {
		return nil, errors.New("申诉原因不能为空")
	}

	// 2. 创建申诉
	appeal := &model.Appeal{
		TicketNo:       generateTicketNo(),
		RiderID:        riderID,
		AppealType:     2, // 处罚申诉
		PenaltyType:    penaltyType,
		PenaltyID:      penaltyID,
		Reason:         reason,
		Description:    description,
		EvidenceImages: stringifySlice(evidenceImages),
		ContactPhone:   contactPhone,
		Status:         1, // 待处理
	}

	err := uc.appealRepo.CreateAppeal(ctx, appeal)
	if err != nil {
		uc.log.Error("创建申诉失败", "err", err)
		return nil, errors.New("提交申诉失败")
	}

	uc.log.Info("处罚申诉提交成功", "appeal_id", appeal.ID, "ticket_no", appeal.TicketNo)
	return appeal, nil
}

// GetAppeals 获取申诉列表
func (uc *appealUsecase) GetAppeals(ctx context.Context, riderID int64, status int32, appealType string, page, pageSize int) ([]*model.Appeal, int64, error) {
	if riderID <= 0 {
		return nil, 0, errors.New("无效的骑手ID")
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	return uc.appealRepo.GetAppealsByRiderID(ctx, riderID, status, appealType, page, pageSize)
}

// GetAppealDetail 获取申诉详情
func (uc *appealUsecase) GetAppealDetail(ctx context.Context, appealID, riderID int64) (*model.Appeal, error) {
	if appealID <= 0 {
		return nil, errors.New("无效的申诉ID")
	}

	appeal, err := uc.appealRepo.GetAppealByID(ctx, appealID)
	if err != nil {
		uc.log.Error("查询申诉失败", "err", err)
		return nil, errors.New("查询申诉失败")
	}
	if appeal == nil {
		return nil, errors.New("申诉不存在")
	}
	if appeal.RiderID != riderID {
		return nil, errors.New("无权查看该申诉")
	}

	return appeal, nil
}

// CancelAppeal 取消申诉
func (uc *appealUsecase) CancelAppeal(ctx context.Context, appealID, riderID int64, cancelReason string) error {
	if appealID <= 0 {
		return errors.New("无效的申诉ID")
	}

	// 查询申诉
	appeal, err := uc.appealRepo.GetAppealByID(ctx, appealID)
	if err != nil {
		uc.log.Error("查询申诉失败", "err", err)
		return errors.New("查询申诉失败")
	}
	if appeal == nil {
		return errors.New("申诉不存在")
	}
	if appeal.RiderID != riderID {
		return errors.New("无权取消该申诉")
	}
	if appeal.Status != 1 && appeal.Status != 2 { // 只有待处理和处理中的可以取消
		return errors.New("当前状态无法取消申诉")
	}

	err = uc.appealRepo.CancelAppeal(ctx, appealID, cancelReason)
	if err != nil {
		uc.log.Error("取消申诉失败", "err", err)
		return errors.New("取消申诉失败")
	}

	uc.log.Info("申诉取消成功", "appeal_id", appealID)
	return nil
}

// GetAppealTypes 获取申诉类型列表
func (uc *appealUsecase) GetAppealTypes(ctx context.Context, category int32) ([]*model.AppealType, error) {
	return uc.appealRepo.GetAppealTypes(ctx, category)
}

// SubmitExceptionReport 提交异常报备
func (uc *appealUsecase) SubmitExceptionReport(ctx context.Context, riderID, orderID int64, exceptionType, description string, images []string, location string, latitude, longitude float64) (*model.ExceptionReport, error) {
	// 1. 参数校验
	if riderID <= 0 {
		return nil, errors.New("无效的骑手ID")
	}
	if exceptionType == "" {
		return nil, errors.New("异常类型不能为空")
	}

	// 2. 如果有关联订单，验证订单
	if orderID > 0 {
		order, err := uc.orderRepo.GetOrderByID(ctx, orderID)
		if err != nil {
			uc.log.Error("查询订单失败", "err", err)
			return nil, errors.New("订单不存在")
		}
		if order.RiderID != riderID {
			return nil, errors.New("无权报备该订单")
		}
	}

	// 3. 创建异常报备
	report := &model.ExceptionReport{
		RiderID:       riderID,
		OrderID:       orderID,
		ExceptionType: exceptionType,
		Description:   description,
		Images:        stringifySlice(images),
		Location:      location,
		Latitude:      latitude,
		Longitude:     longitude,
		Status:        1, // 待处理
	}

	err := uc.appealRepo.CreateExceptionReport(ctx, report)
	if err != nil {
		uc.log.Error("创建异常报备失败", "err", err)
		return nil, errors.New("提交异常报备失败")
	}

	uc.log.Info("异常报备提交成功", "report_id", report.ID)
	return report, nil
}

// GetExceptionOrders 获取异常订单列表
func (uc *appealUsecase) GetExceptionOrders(ctx context.Context, riderID int64, startDate, endDate string, page, pageSize int) ([]*model.ExceptionOrder, int64, error) {
	if riderID <= 0 {
		return nil, 0, errors.New("无效的骑手ID")
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	return uc.appealRepo.GetExceptionOrdersByRiderID(ctx, riderID, startDate, endDate, page, pageSize)
}
