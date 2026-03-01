package service

import (
	"context"
	v1 "hellokratos/api/rider/v1"
	"hellokratos/internal/biz"
	"hellokratos/internal/data/model"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

// IncomeService 收入服务
type IncomeService struct {
	v1.UnimplementedIncomeServer
	incomeUsecase biz.IncomeUsecase
	log           *log.Helper
}

// NewIncomeService 创建收入服务实例
func NewIncomeService(incomeUsecase biz.IncomeUsecase, logger log.Logger) *IncomeService {
	return &IncomeService{
		incomeUsecase: incomeUsecase,
		log:           log.NewHelper(logger),
	}
}

// CreateIncome 创建收入记录
func (s *IncomeService) CreateIncome(ctx context.Context, req *v1.CreateIncomeRequest) (*v1.CreateIncomeReply, error) {
	result, err := s.incomeUsecase.CreateIncomeFromOrder(ctx, req.OrderId, req.IncomeType, req.IncomeStatus)
	if err != nil {
		s.log.Error("create income failed", "err", err)
		return &v1.CreateIncomeReply{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &v1.CreateIncomeReply{
		Success:  true,
		Message:  "收入记录创建成功",
		IncomeId: result.Income.ID,
		Amount:   float32(result.Income.Amount),
		Distance: float32(result.Distance),
	}, nil
}

// GetIncomeList 获取收入列表
func (s *IncomeService) GetIncomeList(ctx context.Context, req *v1.GetIncomeListRequest) (*v1.GetIncomeListReply, error) {
	incomes, err := s.incomeUsecase.GetIncomes(ctx, req.RiderId, int(req.Limit))
	if err != nil {
		s.log.Error("get income list failed", "err", err)
		return nil, err
	}

	incomeInfos := make([]*v1.IncomeInfo, 0, len(incomes))
	for _, income := range incomes {
		incomeInfos = append(incomeInfos, s.convertIncomeToInfo(income))
	}

	return &v1.GetIncomeListReply{
		Incomes: incomeInfos,
		Total:   int32(len(incomeInfos)),
	}, nil
}

// GetTotalIncome 获取总收入
func (s *IncomeService) GetTotalIncome(ctx context.Context, req *v1.GetTotalIncomeRequest) (*v1.GetTotalIncomeReply, error) {
	total, err := s.incomeUsecase.GetTotalIncome(ctx, req.RiderId)
	if err != nil {
		s.log.Error("get total income failed", "err", err)
		return nil, err
	}
	return &v1.GetTotalIncomeReply{
		Total: float32(total),
	}, nil
}

// ApplyWithdrawal 申请提现
func (s *IncomeService) ApplyWithdrawal(ctx context.Context, req *v1.ApplyWithdrawalRequest) (*v1.ApplyWithdrawalReply, error) {
	withdrawal := &model.Withdrawal{
		RiderID:  req.RiderId,
		Amount:   float64(req.Amount),
		Platform: req.Platform,
		Account:  req.Account,
		Status:   0, // 待处理
	}

	err := s.incomeUsecase.CreateWithdrawal(ctx, withdrawal)
	if err != nil {
		s.log.Error("apply withdrawal failed", "err", err)
		return &v1.ApplyWithdrawalReply{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &v1.ApplyWithdrawalReply{
		Success:      true,
		Message:      "申请成功",
		WithdrawalId: withdrawal.ID,
	}, nil
}

// GetWithdrawalList 获取提现列表
func (s *IncomeService) GetWithdrawalList(ctx context.Context, req *v1.GetWithdrawalListRequest) (*v1.GetWithdrawalListReply, error) {
	withdrawals, err := s.incomeUsecase.GetWithdrawals(ctx, req.RiderId, int(req.Limit))
	if err != nil {
		s.log.Error("get withdrawal list failed", "err", err)
		return nil, err
	}

	withdrawalInfos := make([]*v1.WithdrawalInfo, 0, len(withdrawals))
	for _, withdrawal := range withdrawals {
		withdrawalInfos = append(withdrawalInfos, s.convertWithdrawalToInfo(withdrawal))
	}

	return &v1.GetWithdrawalListReply{
		Withdrawals: withdrawalInfos,
		Total:       int32(len(withdrawalInfos)),
	}, nil
}

// convertIncomeToInfo 将收入模型转换为收入信息
func (s *IncomeService) convertIncomeToInfo(income *model.Income) *v1.IncomeInfo {
	return &v1.IncomeInfo{
		Id:           income.ID,
		RiderId:      income.RiderID,
		OrderId:      income.OrderID,
		Amount:       float32(income.Amount),
		IncomeType:   int32(income.Type),
		IncomeStatus: int32(income.Status),
		CreatedAt:    income.CreatedAt.Format(time.RFC3339),
	}
}

// convertWithdrawalToInfo 将提现模型转换为提现信息
func (s *IncomeService) convertWithdrawalToInfo(withdrawal *model.Withdrawal) *v1.WithdrawalInfo {
	return &v1.WithdrawalInfo{
		Id:            withdrawal.ID,
		RiderId:       withdrawal.RiderID,
		Amount:        float32(withdrawal.Amount),
		Platform:      withdrawal.Platform,
		Account:       withdrawal.Account,
		Status:        int32(withdrawal.Status),
		TransactionId: withdrawal.TransactionID,
		CreatedAt:     withdrawal.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     withdrawal.UpdatedAt.Format(time.RFC3339),
	}
}

// UpdateWithdrawalStatus 更新提现状态
func (s *IncomeService) UpdateWithdrawalStatus(ctx context.Context, req *v1.UpdateWithdrawalStatusRequest) (*v1.UpdateWithdrawalStatusReply, error) {
	err := s.incomeUsecase.UpdateWithdrawalStatus(ctx, req.WithdrawalId, req.Status, req.TransactionId)
	if err != nil {
		s.log.Error("update withdrawal status failed", "err", err)
		return &v1.UpdateWithdrawalStatusReply{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &v1.UpdateWithdrawalStatusReply{
		Success: true,
		Message: "提现状态更新成功",
	}, nil
}
