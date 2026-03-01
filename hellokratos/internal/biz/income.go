package biz

import (
	"context"
	"errors"
	"hellokratos/internal/data"
	"hellokratos/internal/data/model"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"
)

const (
	BaseFee       = 3.0 // 基础费用（元）
	PerKmFee      = 2.0 // 每公里费用（元）
	MinDistanceKm = 1.0 // 最小计算距离（公里）
)

type CreateIncomeResult struct {
	Income   *model.Income
	Distance float64
}

// IncomeUsecase 收入相关的业务逻辑接口
type IncomeUsecase interface {
	// CreateIncomeFromOrder 根据订单创建收入记录（自动计算）
	CreateIncomeFromOrder(ctx context.Context, orderID int64, incomeType int32, incomeStatus int32) (*CreateIncomeResult, error)
	// CreateIncome 创建收入记录
	CreateIncome(ctx context.Context, income *model.Income) error
	// GetIncomes 获取骑手的收入列表
	GetIncomes(ctx context.Context, riderID int64, limit int) ([]*model.Income, error)
	// GetTotalIncome 获取骑手的总收入
	GetTotalIncome(ctx context.Context, riderID int64) (float64, error)
	// CreateWithdrawal 创建提现记录
	CreateWithdrawal(ctx context.Context, withdrawal *model.Withdrawal) error
	// GetWithdrawals 获取骑手的提现列表
	GetWithdrawals(ctx context.Context, riderID int64, limit int) ([]*model.Withdrawal, error)
	// UpdateWithdrawalStatus 更新提现状态
	UpdateWithdrawalStatus(ctx context.Context, withdrawalID int64, status int32, transactionID string) error
}

// incomeUsecase 收入相关的业务逻辑实现
type incomeUsecase struct {
	incomeRepo     data.IncomeRepo     // 收入数据访问接口
	withdrawalRepo data.WithdrawalRepo // 提现数据访问接口
	orderRepo      data.OrderRepo      // 订单数据访问接口
	rdb            *redis.Client       // Redis客户端
	log            *log.Helper         // 日志记录器
}

// NewIncomeUsecase 创建收入业务逻辑实例
//
// 参数:
//
//	incomeRepo: 收入数据访问接口
//	withdrawalRepo: 提现数据访问接口
//	orderRepo: 订单数据访问接口
//	rdb: Redis客户端
//	log: 日志记录器
//
// 返回值:
//
//	IncomeUsecase: 收入业务逻辑接口
func NewIncomeUsecase(incomeRepo data.IncomeRepo, withdrawalRepo data.WithdrawalRepo, orderRepo data.OrderRepo, rdb *redis.Client, logger log.Logger) IncomeUsecase {
	return &incomeUsecase{
		incomeRepo:     incomeRepo,
		withdrawalRepo: withdrawalRepo,
		orderRepo:      orderRepo,
		rdb:            rdb,
		log:            log.NewHelper(logger),
	}
}

// CalculateIncomeFromDistance 根据距离计算收入
// 公式：骑手收入 = 基础费 3元 + 距离 × 2元/公里
func (uc *incomeUsecase) CalculateIncomeFromDistance(distance float64) float64 {
	if distance < MinDistanceKm {
		distance = MinDistanceKm
	}
	return BaseFee + distance*PerKmFee
}

// CreateIncomeFromOrder 根据订单创建收入记录（自动计算）
func (uc *incomeUsecase) CreateIncomeFromOrder(ctx context.Context, orderID int64, incomeType int32, incomeStatus int32) (*CreateIncomeResult, error) {
	order, err := uc.orderRepo.GetOrderByID(ctx, orderID)
	if err != nil {
		uc.log.Error("failed to get order", "order_id", orderID, "err", err)
		return nil, errors.New("订单不存在")
	}

	if order.RiderID == 0 {
		return nil, errors.New("订单尚未分配骑手")
	}

	distance := order.Distance
	if distance <= 0 {
		distance = MinDistanceKm
	}

	amount := uc.CalculateIncomeFromDistance(distance)

	income := &model.Income{
		RiderID: order.RiderID,
		OrderID: order.ID,
		Amount:  amount,
		Type:    incomeType,
		Status:  incomeStatus,
	}

	err = uc.incomeRepo.CreateIncome(ctx, income)
	if err != nil {
		uc.log.Error("failed to create income", "err", err)
		return nil, err
	}

	uc.log.Info("income created from order", "order_id", orderID, "rider_id", order.RiderID, "distance", distance, "amount", amount)

	return &CreateIncomeResult{
		Income:   income,
		Distance: distance,
	}, nil
}

// CreateIncome 创建收入记录
func (uc *incomeUsecase) CreateIncome(ctx context.Context, income *model.Income) error {
	return uc.incomeRepo.CreateIncome(ctx, income)
}

// GetIncomes 获取骑手的收入列表
//
// 参数:
//
//	ctx: 上下文
//	riderID: 骑手ID
//	limit: 限制数量
//
// 返回值:
//
//	[]*model.Income: 收入列表
//	error: 错误信息
func (uc *incomeUsecase) GetIncomes(ctx context.Context, riderID int64, limit int) ([]*model.Income, error) {
	return uc.incomeRepo.GetIncomesByRiderID(ctx, riderID, limit)
}

// GetTotalIncome 获取骑手的总收入
//
// 参数:
//
//	ctx: 上下文
//	riderID: 骑手ID
//
// 返回值:
//
//	float64: 总收入
//	error: 错误信息
func (uc *incomeUsecase) GetTotalIncome(ctx context.Context, riderID int64) (float64, error) {
	return uc.incomeRepo.GetTotalIncomeByRiderID(ctx, riderID)
}

// CreateWithdrawal 创建提现记录
//
// 参数:
//
//	ctx: 上下文
//	withdrawal: 提现信息
//
// 返回值:
//
//	error: 错误信息
func (uc *incomeUsecase) CreateWithdrawal(ctx context.Context, withdrawal *model.Withdrawal) error {
	// 检查骑手的已结算总收入是否足够
	totalIncome, err := uc.incomeRepo.GetTotalIncomeByRiderID(ctx, withdrawal.RiderID)
	if err != nil {
		uc.log.Error("failed to get total income", "err", err)
		return err
	}
	if totalIncome < withdrawal.Amount {
		return errors.New("余额不足")
	}

	// 创建提现记录
	err = uc.withdrawalRepo.CreateWithdrawal(ctx, withdrawal)
	if err != nil {
		uc.log.Error("failed to create withdrawal", "err", err)
		return err
	}

	// 更新收入状态为已提现
	err = uc.incomeRepo.UpdateIncomeStatusByAmount(ctx, withdrawal.RiderID, withdrawal.Amount, 2)
	if err != nil {
		uc.log.Error("failed to update income status", "err", err)
		return err
	}

	uc.log.Info("withdrawal created", "rider_id", withdrawal.RiderID, "amount", withdrawal.Amount)

	return nil
}

// GetWithdrawals 获取骑手的提现列表
//
// 参数:
//
//	ctx: 上下文
//	riderID: 骑手ID
//	limit: 限制数量
//
// 返回值:
//
//	[]*model.Withdrawal: 提现列表
//	error: 错误信息
func (uc *incomeUsecase) GetWithdrawals(ctx context.Context, riderID int64, limit int) ([]*model.Withdrawal, error) {
	return uc.withdrawalRepo.GetWithdrawalsByRiderID(ctx, riderID, limit)
}

// UpdateWithdrawalStatus 更新提现状态
//
// 参数:
//
//	ctx: 上下文
//	withdrawalID: 提现ID
//	status: 提现状态
//	transactionID: 交易ID
//
// 返回值:
//
//	error: 错误信息
func (uc *incomeUsecase) UpdateWithdrawalStatus(ctx context.Context, withdrawalID int64, status int32, transactionID string) error {
	// 先获取提现记录
	withdrawal, err := uc.withdrawalRepo.GetWithdrawalByID(ctx, withdrawalID)
	if err != nil {
		uc.log.Error("failed to get withdrawal", "err", err)
		return err
	}

	// 更新提现状态
	err = uc.withdrawalRepo.UpdateWithdrawalStatus(ctx, withdrawalID, status, transactionID)
	if err != nil {
		uc.log.Error("failed to update withdrawal status", "err", err)
		return err
	}

	// 如果提现状态更新为成功，更新收入状态为已提现
	if status == 2 {
		err = uc.incomeRepo.UpdateIncomeStatusByAmount(ctx, withdrawal.RiderID, withdrawal.Amount, 2)
		if err != nil {
			uc.log.Error("failed to update income status", "err", err)
			return err
		}
		uc.log.Info("income status updated to withdrawn", "rider_id", withdrawal.RiderID, "amount", withdrawal.Amount)
	}

	return nil
}
