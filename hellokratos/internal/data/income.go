package data

import (
	"context"
	"hellokratos/internal/data/model"

	"gorm.io/gorm"
)

// IncomeRepo 收入相关的数据访问接口
type IncomeRepo interface {
	// CreateIncome 创建收入记录
	CreateIncome(ctx context.Context, income *model.Income) error
	// GetIncomeByID 根据ID获取收入记录
	GetIncomeByID(ctx context.Context, id int64) (*model.Income, error)
	// GetIncomesByRiderID 获取骑手的收入列表
	GetIncomesByRiderID(ctx context.Context, riderID int64, limit int) ([]*model.Income, error)
	// GetTotalIncomeByRiderID 获取骑手的总收入
	GetTotalIncomeByRiderID(ctx context.Context, riderID int64) (float64, error)
	// UpdateIncomeStatus 更新收入状态
	UpdateIncomeStatus(ctx context.Context, incomeID int64, status int32) error
	// UpdateIncomeStatusByAmount 根据金额批量更新收入状态（用于提现）
	UpdateIncomeStatusByAmount(ctx context.Context, riderID int64, amount float64, status int32) error
}

// WithdrawalRepo 提现相关的数据访问接口
type WithdrawalRepo interface {
	// CreateWithdrawal 创建提现记录
	CreateWithdrawal(ctx context.Context, withdrawal *model.Withdrawal) error
	// GetWithdrawalByID 根据ID获取提现记录
	GetWithdrawalByID(ctx context.Context, id int64) (*model.Withdrawal, error)
	// GetWithdrawalsByRiderID 获取骑手的提现列表
	GetWithdrawalsByRiderID(ctx context.Context, riderID int64, limit int) ([]*model.Withdrawal, error)
	// GetTotalWithdrawalByRiderID 获取骑手的已提现总金额
	GetTotalWithdrawalByRiderID(ctx context.Context, riderID int64) (float64, error)
	// UpdateWithdrawalStatus 更新提现状态
	UpdateWithdrawalStatus(ctx context.Context, withdrawalID int64, status int32, transactionID string) error
}

// incomeRepo 收入相关的数据访问实现
type incomeRepo struct {
	db *gorm.DB
}

// NewIncomeRepo 创建收入数据访问实例
func NewIncomeRepo(data *Data) IncomeRepo {
	return &incomeRepo{db: data.db}
}

// CreateIncome 创建收入记录
func (r *incomeRepo) CreateIncome(ctx context.Context, income *model.Income) error {
	return r.db.WithContext(ctx).Create(income).Error
}

// GetIncomeByID 根据ID获取收入记录
func (r *incomeRepo) GetIncomeByID(ctx context.Context, id int64) (*model.Income, error) {
	var income model.Income
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&income).Error
	if err != nil {
		return nil, err
	}
	return &income, nil
}

// GetIncomesByRiderID 获取骑手的收入列表
func (r *incomeRepo) GetIncomesByRiderID(ctx context.Context, riderID int64, limit int) ([]*model.Income, error) {
	var incomes []*model.Income
	err := r.db.WithContext(ctx).Where("rider_id = ?", riderID).Order("created_at desc").Limit(limit).Find(&incomes).Error
	if err != nil {
		return nil, err
	}
	return incomes, nil
}

// GetTotalIncomeByRiderID 获取骑手的总收入（统计所有状态的收入）
func (r *incomeRepo) GetTotalIncomeByRiderID(ctx context.Context, riderID int64) (float64, error) {
	type Result struct {
		Total float64
	}
	var result Result
	err := r.db.WithContext(ctx).Model(&model.Income{}).Where("rider_id = ?", riderID).Select("COALESCE(SUM(amount), 0) as total").Scan(&result).Error
	if err != nil {
		return 0, err
	}
	return result.Total, nil
}

// UpdateIncomeStatus 更新收入状态
func (r *incomeRepo) UpdateIncomeStatus(ctx context.Context, incomeID int64, status int32) error {
	return r.db.WithContext(ctx).Model(&model.Income{}).Where("id = ?", incomeID).Update("status", status).Error
}

// UpdateIncomeStatusByAmount 根据金额批量更新收入状态（用于提现）
func (r *incomeRepo) UpdateIncomeStatusByAmount(ctx context.Context, riderID int64, amount float64, status int32) error {
	// 获取已结算的收入记录，按创建时间排序
	var incomes []*model.Income
	err := r.db.WithContext(ctx).Where("rider_id = ? AND status = 1", riderID).Order("created_at asc").Find(&incomes).Error
	if err != nil {
		return err
	}

	// 逐个更新收入状态，直到累计金额达到提现金额
	remaining := amount
	for _, income := range incomes {
		if remaining <= 0 {
			break
		}
		if income.Amount <= remaining {
			// 整个收入记录都用于提现
			err = r.db.WithContext(ctx).Model(&model.Income{}).Where("id = ?", income.ID).Update("status", status).Error
			if err != nil {
				return err
			}
			remaining -= income.Amount
		} else {
			// 部分提现：将当前收入记录拆分为两部分
			// 1. 已提现部分：更新原记录金额为(原金额-剩余金额)，状态改为已提现
			// 2. 剩余部分：创建新记录，金额为剩余金额，状态保持已结算

			// 更新原记录：金额减少，状态改为已提现
			newAmount := income.Amount - remaining
			err = r.db.WithContext(ctx).Model(&model.Income{}).Where("id = ?", income.ID).Updates(map[string]interface{}{
				"amount": newAmount,
				"status": status,
			}).Error
			if err != nil {
				return err
			}

			// 创建新记录：剩余金额，状态保持已结算
			newIncome := &model.Income{
				RiderID: income.RiderID,
				OrderID: income.OrderID,
				Amount:  remaining,
				Type:    income.Type,
				Status:  1, // 已结算
			}
			err = r.db.WithContext(ctx).Create(newIncome).Error
			if err != nil {
				return err
			}

			remaining = 0
		}
	}

	return nil
}

// withdrawalRepo 提现相关的数据访问实现
type withdrawalRepo struct {
	db *gorm.DB
}

// NewWithdrawalRepo 创建提现数据访问实例
func NewWithdrawalRepo(data *Data) WithdrawalRepo {
	return &withdrawalRepo{db: data.db}
}

// CreateWithdrawal 创建提现记录
func (r *withdrawalRepo) CreateWithdrawal(ctx context.Context, withdrawal *model.Withdrawal) error {
	return r.db.WithContext(ctx).Create(withdrawal).Error
}

// GetWithdrawalByID 根据ID获取提现记录
func (r *withdrawalRepo) GetWithdrawalByID(ctx context.Context, id int64) (*model.Withdrawal, error) {
	var withdrawal model.Withdrawal
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&withdrawal).Error
	if err != nil {
		return nil, err
	}
	return &withdrawal, nil
}

// GetWithdrawalsByRiderID 获取骑手的提现列表
func (r *withdrawalRepo) GetWithdrawalsByRiderID(ctx context.Context, riderID int64, limit int) ([]*model.Withdrawal, error) {
	var withdrawals []*model.Withdrawal
	err := r.db.WithContext(ctx).Where("rider_id = ?", riderID).Order("created_at desc").Limit(limit).Find(&withdrawals).Error
	if err != nil {
		return nil, err
	}
	return withdrawals, nil
}

// GetTotalWithdrawalByRiderID 获取骑手的已提现总金额
func (r *withdrawalRepo) GetTotalWithdrawalByRiderID(ctx context.Context, riderID int64) (float64, error) {
	type Result struct {
		Total float64
	}
	var result Result
	err := r.db.WithContext(ctx).Model(&model.Withdrawal{}).Where("rider_id = ? AND status IN (1,2)", riderID).Select("COALESCE(SUM(amount), 0) as total").Scan(&result).Error
	if err != nil {
		return 0, err
	}
	return result.Total, nil
}

// UpdateWithdrawalStatus 更新提现状态
func (r *withdrawalRepo) UpdateWithdrawalStatus(ctx context.Context, withdrawalID int64, status int32, transactionID string) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if transactionID != "" {
		updates["transaction_id"] = transactionID
	}
	return r.db.WithContext(ctx).Model(&model.Withdrawal{}).Where("id = ?", withdrawalID).Updates(updates).Error
}
