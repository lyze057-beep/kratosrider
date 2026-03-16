package data

import (
	"context"
	"hellokratos/internal/data/model"

	"gorm.io/gorm"
)

// OrderRepo 订单相关的数据访问接口
type OrderRepo interface {
	// CreateOrder 创建订单
	CreateOrder(ctx context.Context, order *model.Order) error
	// GetOrderByID 根据ID获取订单
	GetOrderByID(ctx context.Context, id int64) (*model.Order, error)
	// GetOrderByOrderNo 根据订单号获取订单
	GetOrderByOrderNo(ctx context.Context, orderNo string) (*model.Order, error)
	// UpdateOrder 更新订单信息
	UpdateOrder(ctx context.Context, order *model.Order) error
	// GetPendingOrders 获取待接单订单
	GetPendingOrders(ctx context.Context, limit int) ([]*model.Order, error)
	// GetOrdersByRiderID 获取骑手的订单列表
	GetOrdersByRiderID(ctx context.Context, riderID int64, status int32, page int, pageSize int) ([]*model.Order, int64, error)
	// UpdateOrderStatusWithRider 使用单次查询更新订单状态和骑手ID（优化版）
	UpdateOrderStatusWithRider(ctx context.Context, orderID int64, riderID int64) error
}

// orderRepo 订单相关的数据访问实现
type orderRepo struct {
	db *gorm.DB
}

// NewOrderRepo 创建订单数据访问实例
func NewOrderRepo(data *Data) OrderRepo {
	return &orderRepo{db: data.db}
}

// CreateOrder 创建订单
func (r *orderRepo) CreateOrder(ctx context.Context, order *model.Order) error {
	return r.db.WithContext(ctx).Create(order).Error
}

// GetOrderByID 根据ID获取订单
func (r *orderRepo) GetOrderByID(ctx context.Context, id int64) (*model.Order, error) {
	var order model.Order
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// GetOrderByOrderNo 根据订单号获取订单
func (r *orderRepo) GetOrderByOrderNo(ctx context.Context, orderNo string) (*model.Order, error) {
	var order model.Order
	err := r.db.WithContext(ctx).Where("order_no = ?", orderNo).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// UpdateOrder 更新订单信息
func (r *orderRepo) UpdateOrder(ctx context.Context, order *model.Order) error {
	return r.db.WithContext(ctx).Save(order).Error
}

// GetPendingOrders 获取待接单订单
func (r *orderRepo) GetPendingOrders(ctx context.Context, limit int) ([]*model.Order, error) {
	var orders []*model.Order
	err := r.db.WithContext(ctx).Where("status = 0").Order("created_at desc").Limit(limit).Find(&orders).Error
	if err != nil {
		return nil, err
	}
	return orders, nil
}

// GetOrdersByRiderID 获取骑手的订单列表
func (r *orderRepo) GetOrdersByRiderID(ctx context.Context, riderID int64, status int32, page int, pageSize int) ([]*model.Order, int64, error) {
	var orders []*model.Order
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Order{}).Where("rider_id = ?", riderID)
	if status >= 0 {
		query = query.Where("status = ?", status)
	}

	// 计算总记录数
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err = query.Order("created_at desc").Offset(offset).Limit(pageSize).Find(&orders).Error
	if err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

// UpdateOrderStatusWithRider 使用单次查询更新订单状态和骑手ID（优化版）
// 避免先查询再更新，直接使用UPDATE语句
func (r *orderRepo) UpdateOrderStatusWithRider(ctx context.Context, orderID int64, riderID int64) error {
	result := r.db.WithContext(ctx).Model(&model.Order{}).
		Where("id = ? AND status = 0", orderID).
		Updates(map[string]interface{}{
			"status":   1,
			"rider_id": riderID,
		})
	
	if result.Error != nil {
		return result.Error
	}
	
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	
	return nil
}
