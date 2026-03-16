package biz

import (
	"context"
	"hellokratos/internal/data/model"

	"github.com/stretchr/testify/mock"
)

// MockOrderRepo 订单仓库mock
type MockOrderRepo struct {
	mock.Mock
}

func (m *MockOrderRepo) GetOrderByID(ctx context.Context, id int64) (*model.Order, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Order), args.Error(1)
}

func (m *MockOrderRepo) CreateOrder(ctx context.Context, order *model.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *MockOrderRepo) UpdateOrderStatus(ctx context.Context, orderID int64, status int32) error {
	args := m.Called(ctx, orderID, status)
	return args.Error(0)
}

func (m *MockOrderRepo) GetPendingOrders(ctx context.Context, limit int) ([]*model.Order, error) {
	args := m.Called(ctx, limit)
	return args.Get(0).([]*model.Order), args.Error(1)
}

func (m *MockOrderRepo) GetOrdersByRiderID(ctx context.Context, riderID int64, status int32, page int, pageSize int) ([]*model.Order, int64, error) {
	args := m.Called(ctx, riderID, status, page, pageSize)
	return args.Get(0).([]*model.Order), args.Get(1).(int64), args.Error(2)
}

func (m *MockOrderRepo) GetOrderByOrderNo(ctx context.Context, orderNo string) (*model.Order, error) {
	args := m.Called(ctx, orderNo)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Order), args.Error(1)
}

func (m *MockOrderRepo) UpdateOrder(ctx context.Context, order *model.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *MockOrderRepo) UpdateOrderStatusWithRider(ctx context.Context, orderID int64, riderID int64) error {
	args := m.Called(ctx, orderID, riderID)
	return args.Error(0)
}
