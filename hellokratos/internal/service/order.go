package service

import (
	"context"
	"fmt"
	v1 "hellokratos/api/rider/v1"
	"hellokratos/internal/biz"
	"hellokratos/internal/data/model"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

// OrderService 订单服务
type OrderService struct {
	v1.UnimplementedOrderServer
	orderUsecase biz.OrderUsecase
	log          *log.Helper
}

// NewOrderService 创建订单服务实例
func NewOrderService(orderUsecase biz.OrderUsecase, logger log.Logger) *OrderService {
	return &OrderService{
		orderUsecase: orderUsecase,
		log:          log.NewHelper(logger),
	}
}

// GetOrderList 获取订单列表
func (s *OrderService) GetOrderList(ctx context.Context, req *v1.GetOrderListRequest) (*v1.GetOrderListReply, error) {
	// 确保页码和每页大小有效
	page := int(req.Page)
	if page < 1 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize < 1 || pageSize > 50 {
		pageSize = 10
	}

	var orders []*model.Order
	var total int64
	var err error

	// status=0 查询所有待接单订单（不限制骑手）
	// status=-1 查询所有订单
	// status>0 查询特定状态的订单
	if req.Status == 0 {
		// 查询待接单订单（rider_id=0 表示未分配骑手）
		orders, total, err = s.orderUsecase.GetOrdersByRiderID(ctx, 0, 0, page, pageSize)
	} else {
		// 从上下文获取骑手ID（实际应用中需要从token中解析）
		riderID := int64(1) // 这里暂时硬编码，实际应从认证信息中获取
		orders, total, err = s.orderUsecase.GetOrdersByRiderID(ctx, riderID, req.Status, page, pageSize)
	}

	if err != nil {
		s.log.Error("get order list failed", "err", err)
		return nil, err
	}

	// 转换为API响应格式
	orderInfos := make([]*v1.OrderInfo, len(orders))
	for i, order := range orders {
		orderInfos[i] = s.convertOrderToInfo(order)
	}

	return &v1.GetOrderListReply{
		Orders:   orderInfos,
		Total:    int32(total),
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// GetOrderDetail 获取订单详情
func (s *OrderService) GetOrderDetail(ctx context.Context, req *v1.GetOrderDetailRequest) (*v1.GetOrderDetailReply, error) {
	order, err := s.orderUsecase.GetOrderByID(ctx, req.OrderId)
	if err != nil {
		s.log.Error("get order detail failed", "err", err)
		return nil, err
	}
	return &v1.GetOrderDetailReply{
		Order: s.convertOrderToInfo(order),
	}, nil
}

// AcceptOrder 接单
func (s *OrderService) AcceptOrder(ctx context.Context, req *v1.AcceptOrderRequest) (*v1.AcceptOrderReply, error) {
	err := s.orderUsecase.AcceptOrder(ctx, req.OrderId, req.RiderId)
	if err != nil {
		s.log.Error("accept order failed", "err", err)
		return &v1.AcceptOrderReply{
			Success: false,
			Message: err.Error(),
		}, nil
	}
	return &v1.AcceptOrderReply{
		Success: true,
		Message: "接单成功",
	}, nil
}

// UpdateOrderStatus 更新订单状态
func (s *OrderService) UpdateOrderStatus(ctx context.Context, req *v1.UpdateOrderStatusRequest) (*v1.UpdateOrderStatusReply, error) {
	err := s.orderUsecase.UpdateOrderStatus(ctx, req.OrderId, req.Status)
	if err != nil {
		s.log.Error("update order status failed", "err", err)
		return &v1.UpdateOrderStatusReply{
			Success: false,
			Message: err.Error(),
		}, nil
	}
	return &v1.UpdateOrderStatusReply{
		Success: true,
		Message: "状态更新成功",
	}, nil
}

// CreateOrder 创建订单
func (s *OrderService) CreateOrder(ctx context.Context, req *v1.CreateOrderRequest) (*v1.CreateOrderReply, error) {
	// 创建订单模型
	order := &model.Order{
		Origin:      req.Origin,
		Destination: req.Destination,
		OriginLat:   float64(req.OriginLat),
		OriginLng:   float64(req.OriginLng),
		DestLat:     float64(req.DestLat),
		DestLng:     float64(req.DestLng),
		Amount:      float64(req.Amount),
		RiderID:     0, // 初始无骑手
	}

	// 调用业务逻辑创建订单
	err := s.orderUsecase.CreateOrder(ctx, order)
	if err != nil {
		s.log.Error("create order failed", "err", err)
		return &v1.CreateOrderReply{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &v1.CreateOrderReply{
		Success: true,
		Message: "订单创建成功",
		Order:   s.convertOrderToInfo(order),
	}, nil
}

// SubscribeOrders 实时订单推送（WebSocket）
func (s *OrderService) SubscribeOrders(req *v1.SubscribeOrdersRequest, stream v1.Order_SubscribeOrdersServer) error {
	s.log.Info("rider subscribed orders", "rider_id", req.RiderId, "lat", req.Lat, "lng", req.Lng)

	// 模拟推送订单和状态更新
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// 模拟订单ID计数器
	orderIDCounter := int64(1000)

	for {
		select {
		case <-ticker.C:
			// 模拟获取待接单订单
			orders, err := s.orderUsecase.GetPendingOrders(context.Background(), 5)
			if err != nil {
				s.log.Error("get pending orders failed", "err", err)
				// 模拟创建一些订单用于测试
				if len(orders) == 0 {
					// 模拟创建5个订单
					for i := 0; i < 5; i++ {
						orderIDCounter++
						// 生成随机位置（围绕骑手位置）
						offsetLat := (float64(i) - 2.5) * 0.01 // 约1公里
						offsetLng := (float64(i) - 2.5) * 0.01

						mockOrder := &model.Order{
							ID:          orderIDCounter,
							OrderNo:     fmt.Sprintf("ORD%d", orderIDCounter),
							Status:      0,
							Origin:      fmt.Sprintf("起点%d", i+1),
							Destination: fmt.Sprintf("终点%d", i+1),
							OriginLat:   float64(req.Lat) + offsetLat,
							OriginLng:   float64(req.Lng) + offsetLng,
							DestLat:     float64(req.Lat) + offsetLat + 0.02,
							DestLng:     float64(req.Lng) + offsetLng + 0.02,
							Distance:    2.5,
							Amount:      15.0 + float64(i)*2,
							RiderID:     0,
							CreatedAt:   time.Now(),
							UpdatedAt:   time.Now(),
						}
						orders = append(orders, mockOrder)
					}
				}
			}

			for _, order := range orders {
				// 计算骑手与订单起点的距离
				distance := s.orderUsecase.CalculateDistance(
					float64(req.Lat), float64(req.Lng),
					order.OriginLat, order.OriginLng,
				)

				// 只推送5公里内的订单
				if distance <= 5 {
					err := stream.Send(&v1.SubscribeOrdersReply{
						Order: s.convertOrderToInfo(order),
						Type:  1, // 新订单
					})
					if err != nil {
						s.log.Error("send order failed", "err", err)
						return err
					}
					s.log.Info("pushed order to rider", "order_id", order.ID, "distance", distance)
				}
			}
		case <-stream.Context().Done():
			s.log.Info("rider unsubscribed orders", "rider_id", req.RiderId)
			return nil
		}
	}
}

// convertOrderToInfo 将订单模型转换为订单信息
func (s *OrderService) convertOrderToInfo(order *model.Order) *v1.OrderInfo {
	return &v1.OrderInfo{
		Id:          order.ID,
		OrderNo:     order.OrderNo,
		Status:      int32(order.Status),
		Origin:      order.Origin,
		Destination: order.Destination,
		OriginLat:   float32(order.OriginLat),
		OriginLng:   float32(order.OriginLng),
		DestLat:     float32(order.DestLat),
		DestLng:     float32(order.DestLng),
		Distance:    float32(order.Distance),
		Amount:      float32(order.Amount),
		RiderId:     order.RiderID,
		CreatedAt:   order.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   order.UpdatedAt.Format(time.RFC3339),
	}
}
