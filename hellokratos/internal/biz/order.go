package biz

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"hellokratos/internal/data"
	"hellokratos/internal/data/model"
	"math"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// OrderUsecase 订单相关的业务逻辑接口
type OrderUsecase interface {
	// CreateOrder 创建订单
	CreateOrder(ctx context.Context, order *model.Order) error
	// GetOrderByID 根据ID获取订单
	GetOrderByID(ctx context.Context, id int64) (*model.Order, error)
	// UpdateOrderStatus 更新订单状态
	UpdateOrderStatus(ctx context.Context, orderID int64, status int32) error
	// GetPendingOrders 获取待接单订单
	GetPendingOrders(ctx context.Context, limit int) ([]*model.Order, error)
	// GetOrdersByRiderID 获取骑手的订单列表
	GetOrdersByRiderID(ctx context.Context, riderID int64, status int32, page int, pageSize int) ([]*model.Order, int64, error)
	// AcceptOrder 接单
	AcceptOrder(ctx context.Context, orderID int64, riderID int64) error
	// CalculateDistance 计算两点之间的距离
	CalculateDistance(lat1, lng1, lat2, lng2 float64) float64
}

// orderUsecase 订单相关的业务逻辑实现
type orderUsecase struct {
	orderRepo            data.OrderRepo             // 订单数据访问接口
	rdb                  *redis.Client              // Redis客户端
	orderMessageProducer *data.OrderMessageProducer // 订单消息生产者
	log                  *log.Helper                // 日志记录器
}

// NewOrderUsecase 创建订单业务逻辑实例
//
// 参数:
//
//	orderRepo: 订单数据访问接口
//	rdb: Redis客户端
//	orderMessageProducer: 订单消息生产者
//	log: 日志记录器
//
// 返回值:
//
//	OrderUsecase: 订单业务逻辑接口
func NewOrderUsecase(orderRepo data.OrderRepo, rdb *redis.Client, orderMessageProducer *data.OrderMessageProducer, logger log.Logger) OrderUsecase {
	return &orderUsecase{
		orderRepo:            orderRepo,
		rdb:                  rdb,
		orderMessageProducer: orderMessageProducer,
		log:                  log.NewHelper(logger),
	}
}

// CreateOrder 创建订单
//
// 参数:
//
//	ctx: 上下文
//	order: 订单信息
//
// 返回值:
//
//	error: 错误信息
func (uc *orderUsecase) CreateOrder(ctx context.Context, order *model.Order) error {
	// 生成订单号
	order.OrderNo = fmt.Sprintf("ORD%s%d", time.Now().Format("20060102150405"), time.Now().UnixNano()%1000)
	// 设置默认状态
	order.Status = 0
	// 计算距离
	order.Distance = uc.CalculateDistance(order.OriginLat, order.OriginLng, order.DestLat, order.DestLng)
	// 保存订单
	err := uc.orderRepo.CreateOrder(ctx, order)
	if err != nil {
		uc.log.Error("failed to create order", "err", err)
		return err
	}
	// TODO: 发送订单广播
	uc.log.Info("order created", "order_no", order.OrderNo)
	return nil
}

// GetOrderByID 根据ID获取订单
//
// 参数:
//
//	ctx: 上下文
//	id: 订单ID
//
// 返回值:
//
//	*model.Order: 订单信息
//	error: 错误信息
func (uc *orderUsecase) GetOrderByID(ctx context.Context, id int64) (*model.Order, error) {
	// 优化：添加Redis缓存
	cacheKey := fmt.Sprintf("order:%d", id)

	// 1. 先查缓存
	cached, err := uc.rdb.Get(ctx, cacheKey).Result()
	if err == nil {
		var order model.Order
		if err := json.Unmarshal([]byte(cached), &order); err == nil {
			return &order, nil
		}
	}

	// 2. 缓存未命中，查数据库
	order, err := uc.orderRepo.GetOrderByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 3. 写入缓存（5分钟过期）
	data, _ := json.Marshal(order)
	uc.rdb.Set(ctx, cacheKey, data, 5*time.Minute)

	return order, nil
}

// UpdateOrderStatus 更新订单状态
//
// 参数:
//
//	ctx: 上下文
//	orderID: 订单ID
//	status: 订单状态
//
// 返回值:
//
//	error: 错误信息
func (uc *orderUsecase) UpdateOrderStatus(ctx context.Context, orderID int64, status int32) error {
	order, err := uc.orderRepo.GetOrderByID(ctx, orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("订单不存在")
		}
		return err
	}
	order.Status = status
	return uc.orderRepo.UpdateOrder(ctx, order)
}

// GetPendingOrders 获取待接单订单
//
// 参数:
//
//	ctx: 上下文
//	limit: 限制数量
//
// 返回值:
//
//	[]*model.Order: 待接单订单列表
//	error: 错误信息
func (uc *orderUsecase) GetPendingOrders(ctx context.Context, limit int) ([]*model.Order, error) {
	return uc.orderRepo.GetPendingOrders(ctx, limit)
}

// GetOrdersByRiderID 获取骑手的订单列表
//
// 参数:
//
//	ctx: 上下文
//	riderID: 骑手ID
//	status: 订单状态（-1表示全部）
//	page: 页码
//	pageSize: 每页大小
//
// 返回值:
//
//	[]*model.Order: 订单列表
//	int64: 总记录数
//	error: 错误信息
func (uc *orderUsecase) GetOrdersByRiderID(ctx context.Context, riderID int64, status int32, page int, pageSize int) ([]*model.Order, int64, error) {
	return uc.orderRepo.GetOrdersByRiderID(ctx, riderID, status, page, pageSize)
}

// AcceptOrder 接单（优化版）
//
// 参数:
//
//	ctx: 上下文
//	orderID: 订单ID
//	riderID: 骑手ID
//
// 返回值:
//
//	error: 错误信息
func (uc *orderUsecase) AcceptOrder(ctx context.Context, orderID int64, riderID int64) error {
	// 优化1：使用Pipeline批量执行Redis操作
	lockKey := fmt.Sprintf("order:lock:%d", orderID)
	lockValue := fmt.Sprintf("%d", riderID)

	// 使用Pipeline批量执行Redis操作，减少网络往返
	pipe := uc.rdb.Pipeline()
	pipe.SetNX(ctx, lockKey, lockValue, 5*time.Second)
	pipe.Expire(ctx, lockKey, 5*time.Second)
	_, err := pipe.Exec(ctx)
	if err != nil {
		uc.log.Error("failed to acquire lock with pipeline", "err", err)
		return err
	}

	// 释放锁
	defer uc.rdb.Del(ctx, lockKey)

	// 优化2：使用单次数据库查询更新订单（避免先查询再更新）
	result := uc.orderRepo.UpdateOrderStatusWithRider(ctx, orderID, riderID)
	if result != nil {
		uc.log.Error("failed to update order with rider", "err", result)
		return result
	}

	// 优化3：异步发布消息，不阻塞主流程
	if uc.orderMessageProducer != nil {
		go func() {
			msg := &data.OrderMessage{
				OrderID: orderID,
				RiderID: riderID,
				Action:  "accept",
				Status:  1,
			}
			err := uc.orderMessageProducer.PublishOrderMessage(ctx, msg)
			if err != nil {
				uc.log.Error("failed to publish order accept message", "err", err)
			}
		}()
	}

	return nil
}

// CalculateDistance 计算两点之间的距离（单位：公里）
//
// 参数:
//
//	lat1: 第一个点的纬度
//	lng1: 第一个点的经度
//	lat2: 第二个点的纬度
//	lng2: 第二个点的经度
//
// 返回值:
//
//	float64: 两点之间的距离（公里）
func (uc *orderUsecase) CalculateDistance(lat1, lng1, lat2, lng2 float64) float64 {
	const earthRadius = 6371.0 // 地球半径（公里）

	// 将角度转换为弧度
	lat1Rad := lat1 * math.Pi / 180.0
	lng1Rad := lng1 * math.Pi / 180.0
	lat2Rad := lat2 * math.Pi / 180.0
	lng2Rad := lng2 * math.Pi / 180.0

	// 计算差值
	dLat := lat2Rad - lat1Rad
	dLng := lng2Rad - lng1Rad

	// 使用Haversine公式计算距离
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(dLng/2)*math.Sin(dLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	distance := earthRadius * c

	return math.Round(distance*100) / 100 // 保留两位小数
}
