package biz

import (
	"context"
	"errors"
	"hellokratos/internal/data"
	"hellokratos/internal/data/model"

	"github.com/go-kratos/kratos/v2/log"
)

// SafetyUsecase 安全业务逻辑接口
type SafetyUsecase interface {
	// 紧急求助相关
	EmergencyHelp(ctx context.Context, riderID, orderID int64, helpType, description string, latitude, longitude float64, address string, images []string) (*model.EmergencyHelp, error)
	CancelEmergency(ctx context.Context, emergencyID, riderID int64, cancelReason string) error
	GetEmergencyRecords(ctx context.Context, riderID int64, status int32, page, pageSize int) ([]*model.EmergencyHelp, int64, error)

	// 保险理赔相关
	SubmitInsuranceClaim(ctx context.Context, riderID, orderID int64, incidentType string, incidentTime string, incidentLocation string, latitude, longitude float64, incidentDesc, injurySituation, medicalExpense string, evidenceImages, medicalRecords []string, contactName, contactPhone, bankName, bankAccount, accountName string) (*model.InsuranceClaim, error)
	GetInsuranceClaims(ctx context.Context, riderID int64, status int32, page, pageSize int) ([]*model.InsuranceClaim, int64, error)
	GetInsuranceClaimDetail(ctx context.Context, claimID, riderID int64) (*model.InsuranceClaim, error)
	GetInsuranceInfo(ctx context.Context, riderID int64) (*model.RiderInsurance, error)

	// 安全事件相关
	ReportSafetyEvent(ctx context.Context, riderID, orderID int64, eventType, description string, latitude, longitude float64, address string, images []string, needHelp bool) (*model.SafetyEvent, error)

	// 安全提示相关
	GetSafetyTips(ctx context.Context, tipType string, page, pageSize int) ([]*model.SafetyTip, int64, error)

	// 紧急联系人相关
	GetEmergencyContacts(ctx context.Context, riderID int64) ([]*model.EmergencyContact, error)
	UpdateEmergencyContacts(ctx context.Context, riderID int64, contacts []*model.EmergencyContact) error
}

// safetyUsecase 安全业务逻辑实现
type safetyUsecase struct {
	safetyRepo data.SafetyRepo
	orderRepo  data.OrderRepo
	log        *log.Helper
}

// NewSafetyUsecase 创建安全业务逻辑实例
func NewSafetyUsecase(safetyRepo data.SafetyRepo, orderRepo data.OrderRepo, logger log.Logger) SafetyUsecase {
	return &safetyUsecase{
		safetyRepo: safetyRepo,
		orderRepo:  orderRepo,
		log:        log.NewHelper(logger),
	}
}

// EmergencyHelp 发起紧急求助
func (uc *safetyUsecase) EmergencyHelp(ctx context.Context, riderID, orderID int64, helpType, description string, latitude, longitude float64, address string, images []string) (*model.EmergencyHelp, error) {
	// 1. 参数校验
	if riderID <= 0 {
		return nil, errors.New("无效的骑手ID")
	}
	if helpType == "" {
		return nil, errors.New("求助类型不能为空")
	}

	// 2. 验证订单（如果提供了订单ID）
	if orderID > 0 {
		order, err := uc.orderRepo.GetOrderByID(ctx, orderID)
		if err != nil {
			uc.log.Error("查询订单失败", "err", err)
			return nil, errors.New("订单不存在")
		}
		if order.RiderID != riderID {
			return nil, errors.New("无权操作该订单")
		}
	}

	// 3. 创建紧急求助记录
	help := &model.EmergencyHelp{
		TicketNo:    generateTicketNo(),
		RiderID:     riderID,
		OrderID:     orderID,
		HelpType:    helpType,
		Description: description,
		Latitude:    latitude,
		Longitude:   longitude,
		Address:     address,
		Images:      stringifySlice(images),
		Status:      1, // 处理中
	}

	err := uc.safetyRepo.CreateEmergencyHelp(ctx, help)
	if err != nil {
		uc.log.Error("创建紧急求助失败", "err", err)
		return nil, errors.New("发起紧急求助失败")
	}

	// 4. TODO: 发送紧急通知给平台和紧急联系人
	// 这里应该调用消息服务发送短信/推送

	uc.log.Info("紧急求助已发起", "help_id", help.ID, "ticket_no", help.TicketNo, "rider_id", riderID)
	return help, nil
}

// CancelEmergency 取消紧急求助
func (uc *safetyUsecase) CancelEmergency(ctx context.Context, emergencyID, riderID int64, cancelReason string) error {
	if emergencyID <= 0 {
		return errors.New("无效的求助ID")
	}

	// 查询求助记录
	help, err := uc.safetyRepo.GetEmergencyHelpByID(ctx, emergencyID)
	if err != nil {
		uc.log.Error("查询紧急求助失败", "err", err)
		return errors.New("查询紧急求助失败")
	}
	if help == nil {
		return errors.New("紧急求助记录不存在")
	}
	if help.RiderID != riderID {
		return errors.New("无权操作该求助记录")
	}
	if help.Status != 1 { // 只有处理中的可以取消
		return errors.New("当前状态无法取消")
	}

	err = uc.safetyRepo.CancelEmergencyHelp(ctx, emergencyID, cancelReason)
	if err != nil {
		uc.log.Error("取消紧急求助失败", "err", err)
		return errors.New("取消紧急求助失败")
	}

	uc.log.Info("紧急求助已取消", "help_id", emergencyID)
	return nil
}

// GetEmergencyRecords 获取紧急求助记录
func (uc *safetyUsecase) GetEmergencyRecords(ctx context.Context, riderID int64, status int32, page, pageSize int) ([]*model.EmergencyHelp, int64, error) {
	if riderID <= 0 {
		return nil, 0, errors.New("无效的骑手ID")
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	return uc.safetyRepo.GetEmergencyHelpsByRiderID(ctx, riderID, status, page, pageSize)
}

// SubmitInsuranceClaim 提交保险理赔申请
func (uc *safetyUsecase) SubmitInsuranceClaim(ctx context.Context, riderID, orderID int64, incidentType string, incidentTime string, incidentLocation string, latitude, longitude float64, incidentDesc, injurySituation, medicalExpense string, evidenceImages, medicalRecords []string, contactName, contactPhone, bankName, bankAccount, accountName string) (*model.InsuranceClaim, error) {
	// 1. 参数校验
	if riderID <= 0 {
		return nil, errors.New("无效的骑手ID")
	}
	if incidentType == "" {
		return nil, errors.New("事故类型不能为空")
	}
	if incidentTime == "" {
		return nil, errors.New("事故时间不能为空")
	}

	// 2. 检查保险是否有效
	insurance, err := uc.safetyRepo.GetRiderInsurance(ctx, riderID)
	if err != nil {
		uc.log.Error("查询保险信息失败", "err", err)
		return nil, errors.New("查询保险信息失败")
	}
	if insurance == nil {
		return nil, errors.New("未找到保险信息")
	}
	if insurance.Status != 1 {
		return nil, errors.New("保险已过期或无效")
	}

	// 3. 验证订单（如果提供了订单ID）
	if orderID > 0 {
		order, err := uc.orderRepo.GetOrderByID(ctx, orderID)
		if err != nil {
			uc.log.Error("查询订单失败", "err", err)
			return nil, errors.New("订单不存在")
		}
		if order.RiderID != riderID {
			return nil, errors.New("无权操作该订单")
		}
	}

	// 4. 创建理赔申请
	claim := &model.InsuranceClaim{
		ClaimNo:          generateClaimNo(),
		RiderID:          riderID,
		OrderID:          orderID,
		IncidentType:     incidentType,
		IncidentLocation: incidentLocation,
		Latitude:         latitude,
		Longitude:        longitude,
		IncidentDesc:     incidentDesc,
		InjurySituation:  injurySituation,
		MedicalExpense:   medicalExpense,
		EvidenceImages:   stringifySlice(evidenceImages),
		MedicalRecords:   stringifySlice(medicalRecords),
		ContactName:      contactName,
		ContactPhone:     contactPhone,
		BankName:         bankName,
		BankAccount:      bankAccount,
		AccountName:      accountName,
		Status:           1, // 待审核
	}

	err = uc.safetyRepo.CreateInsuranceClaim(ctx, claim)
	if err != nil {
		uc.log.Error("创建理赔申请失败", "err", err)
		return nil, errors.New("提交理赔申请失败")
	}

	// 5. 更新保险理赔统计
	uc.safetyRepo.UpdateInsuranceClaimStats(ctx, riderID, medicalExpense)

	uc.log.Info("保险理赔申请已提交", "claim_id", claim.ID, "claim_no", claim.ClaimNo)
	return claim, nil
}

// GetInsuranceClaims 获取保险理赔列表
func (uc *safetyUsecase) GetInsuranceClaims(ctx context.Context, riderID int64, status int32, page, pageSize int) ([]*model.InsuranceClaim, int64, error) {
	if riderID <= 0 {
		return nil, 0, errors.New("无效的骑手ID")
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	return uc.safetyRepo.GetInsuranceClaimsByRiderID(ctx, riderID, status, page, pageSize)
}

// GetInsuranceClaimDetail 获取保险理赔详情
func (uc *safetyUsecase) GetInsuranceClaimDetail(ctx context.Context, claimID, riderID int64) (*model.InsuranceClaim, error) {
	if claimID <= 0 {
		return nil, errors.New("无效的理赔ID")
	}

	claim, err := uc.safetyRepo.GetInsuranceClaimByID(ctx, claimID)
	if err != nil {
		uc.log.Error("查询理赔申请失败", "err", err)
		return nil, errors.New("查询理赔申请失败")
	}
	if claim == nil {
		return nil, errors.New("理赔申请不存在")
	}
	if claim.RiderID != riderID {
		return nil, errors.New("无权查看该理赔申请")
	}

	return claim, nil
}

// GetInsuranceInfo 获取保险信息
func (uc *safetyUsecase) GetInsuranceInfo(ctx context.Context, riderID int64) (*model.RiderInsurance, error) {
	if riderID <= 0 {
		return nil, errors.New("无效的骑手ID")
	}

	insurance, err := uc.safetyRepo.GetRiderInsurance(ctx, riderID)
	if err != nil {
		uc.log.Error("查询保险信息失败", "err", err)
		return nil, errors.New("查询保险信息失败")
	}
	if insurance == nil {
		return nil, errors.New("未找到保险信息")
	}

	return insurance, nil
}

// ReportSafetyEvent 上报安全事件
func (uc *safetyUsecase) ReportSafetyEvent(ctx context.Context, riderID, orderID int64, eventType, description string, latitude, longitude float64, address string, images []string, needHelp bool) (*model.SafetyEvent, error) {
	// 1. 参数校验
	if riderID <= 0 {
		return nil, errors.New("无效的骑手ID")
	}
	if eventType == "" {
		return nil, errors.New("事件类型不能为空")
	}

	// 2. 验证订单（如果提供了订单ID）
	if orderID > 0 {
		order, err := uc.orderRepo.GetOrderByID(ctx, orderID)
		if err != nil {
			uc.log.Error("查询订单失败", "err", err)
			return nil, errors.New("订单不存在")
		}
		if order.RiderID != riderID {
			return nil, errors.New("无权操作该订单")
		}
	}

	// 3. 创建安全事件
	event := &model.SafetyEvent{
		RiderID:     riderID,
		OrderID:     orderID,
		EventType:   eventType,
		Description: description,
		Latitude:    latitude,
		Longitude:   longitude,
		Address:     address,
		Images:      stringifySlice(images),
		NeedHelp:    needHelp,
		Status:      1, // 待处理
	}

	err := uc.safetyRepo.CreateSafetyEvent(ctx, event)
	if err != nil {
		uc.log.Error("创建安全事件失败", "err", err)
		return nil, errors.New("上报安全事件失败")
	}

	// 4. 如果需要帮助，自动创建紧急求助
	if needHelp {
		_, err = uc.EmergencyHelp(ctx, riderID, orderID, "other", description, latitude, longitude, address, images)
		if err != nil {
			uc.log.Error("自动创建紧急求助失败", "err", err)
		}
	}

	uc.log.Info("安全事件已上报", "event_id", event.ID, "rider_id", riderID)
	return event, nil
}

// GetSafetyTips 获取安全提示
func (uc *safetyUsecase) GetSafetyTips(ctx context.Context, tipType string, page, pageSize int) ([]*model.SafetyTip, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	return uc.safetyRepo.GetSafetyTips(ctx, tipType, page, pageSize)
}

// GetEmergencyContacts 获取紧急联系人
func (uc *safetyUsecase) GetEmergencyContacts(ctx context.Context, riderID int64) ([]*model.EmergencyContact, error) {
	if riderID <= 0 {
		return nil, errors.New("无效的骑手ID")
	}

	return uc.safetyRepo.GetEmergencyContacts(ctx, riderID)
}

// UpdateEmergencyContacts 更新紧急联系人
func (uc *safetyUsecase) UpdateEmergencyContacts(ctx context.Context, riderID int64, contacts []*model.EmergencyContact) error {
	if riderID <= 0 {
		return errors.New("无效的骑手ID")
	}

	// 校验联系人信息
	if len(contacts) > 5 {
		return errors.New("紧急联系人最多5个")
	}
	for i, contact := range contacts {
		if contact.Name == "" {
			return errors.New("联系人姓名不能为空")
		}
		if contact.Phone == "" {
			return errors.New("联系人电话不能为空")
		}
		contact.SortOrder = int32(i)
	}

	return uc.safetyRepo.SaveEmergencyContacts(ctx, riderID, contacts)
}
