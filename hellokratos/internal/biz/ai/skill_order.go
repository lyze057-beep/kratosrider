package ai

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"hellokratos/internal/data"
	"hellokratos/internal/data/model"
)

// OrderSkill 订单助手技能
// 处理所有与订单相关的查询和操作
type OrderSkill struct {
	orderRepo data.OrderRepo
}

// NewOrderSkill 创建订单助手技能
func NewOrderSkill(orderRepo data.OrderRepo) Skill {
	return &OrderSkill{orderRepo: orderRepo}
}

// Name 技能名称
func (s *OrderSkill) Name() string {
	return "order_assistant"
}

// Description 技能描述
func (s *OrderSkill) Description() string {
	return "处理订单查询、订单状态跟踪、配送问题等订单相关业务"
}

// CanHandle 判断是否可以处理该请求
func (s *OrderSkill) CanHandle(ctx context.Context, intent string, entities map[string]string) bool {
	// 意图匹配
	orderIntents := []string{
		"query_order",
		"track_order",
		"order_status",
		"delivery_problem",
		"cancel_order",
		"order_detail",
	}

	for _, i := range orderIntents {
		if intent == i {
			return true
		}
	}

	// 关键词匹配
	keywords := []string{"订单", "配送", "商家", "顾客", "取餐", "送达"}
	for _, keyword := range keywords {
		if strings.Contains(intent, keyword) {
			return true
		}
	}

	// 实体匹配
	if _, hasOrderID := entities["order_id"]; hasOrderID {
		return true
	}

	return false
}

// Execute 执行技能
func (s *OrderSkill) Execute(ctx context.Context, params SkillParams) (*SkillResult, error) {
	switch params.Intent {
	case "query_order", "order_detail":
		return s.handleQueryOrder(ctx, params)
	case "track_order", "order_status":
		return s.handleTrackOrder(ctx, params)
	case "delivery_problem":
		return s.handleDeliveryProblem(ctx, params)
	case "cancel_order":
		return s.handleCancelOrder(ctx, params)
	default:
		// 默认处理：尝试识别意图
		return s.handleDefault(ctx, params)
	}
}

// handleQueryOrder 处理订单查询
func (s *OrderSkill) handleQueryOrder(ctx context.Context, params SkillParams) (*SkillResult, error) {
	// 获取订单ID
	orderIDStr, ok := params.Entities["order_id"]
	if !ok {
		return &SkillResult{
			Success:       false,
			Response:      "请提供订单号，例如：查询订单 12345",
			NeedConfirm:   true,
			ConfirmPrompt: "请输入订单号",
		}, nil
	}

	orderID, err := strconv.ParseInt(orderIDStr, 10, 64)
	if err != nil {
		return &SkillResult{
			Success:  false,
			Response: "订单号格式不正确，请提供数字订单号",
		}, nil
	}

	// 查询订单
	order, err := s.orderRepo.GetOrderByID(ctx, orderID)
	if err != nil {
		return &SkillResult{
			Success:  false,
			Response: fmt.Sprintf("查询订单失败：%v", err),
		}, nil
	}

	// 构建回复
	statusMap := map[int32]string{
		0: "待接单",
		1: "已接单",
		2: "配送中",
		3: "已完成",
		4: "已取消",
	}

	response := fmt.Sprintf(
		"订单 %d 信息如下：\n"+
			"📋 状态：%s\n"+
			"🏪 起点：%s\n"+
			"📍 终点：%s\n"+
			"💰 金额：¥%.2f\n"+
			"📏 距离：%.2f公里\n"+
			"🕐 创建时间：%s",
		order.ID,
		statusMap[order.Status],
		order.Origin,
		order.Destination,
		order.Amount,
		order.Distance,
		order.CreatedAt.Format("2006-01-02 15:04"),
	)

	// 根据状态添加建议操作
	var actions []Action
	switch order.Status {
	case 0:
		actions = append(actions, Action{
			Type:   "accept_order",
			Name:   "接单",
			Params: fmt.Sprintf(`{"order_id":%d}`, order.ID),
		})
	case 1, 2:
		actions = append(actions,
			Action{
				Type:   "navigate",
				Name:   "导航",
				Params: fmt.Sprintf(`{"lat":%f,"lng":%f}`, order.DestLat, order.DestLng),
			},
			Action{
				Type:   "contact_customer",
				Name:   "联系顾客",
				Params: fmt.Sprintf(`{"order_id":%d}`, order.ID),
			},
		)
	}

	return &SkillResult{
		Success:          true,
		Response:         response,
		Data:             order,
		SuggestedActions: actions,
	}, nil
}

// handleTrackOrder 处理订单跟踪
func (s *OrderSkill) handleTrackOrder(ctx context.Context, params SkillParams) (*SkillResult, error) {
	riderID := params.RiderID

	// 获取骑手最近的订单
	orders, _, err := s.orderRepo.GetOrdersByRiderID(ctx, riderID, 0, 1, 5)
	if err != nil {
		return &SkillResult{
			Success:  false,
			Response: fmt.Sprintf("查询订单失败：%v", err),
		}, nil
	}

	if len(orders) == 0 {
		return &SkillResult{
			Success:  true,
			Response: "您当前没有进行中的订单",
		}, nil
	}

	// 找到进行中的订单
	var activeOrders []*model.Order
	for _, order := range orders {
		if order.Status == 1 || order.Status == 2 {
			activeOrders = append(activeOrders, order)
		}
	}

	if len(activeOrders) == 0 {
		return &SkillResult{
			Success:  true,
			Response: "您当前没有进行中的订单，最近完成的订单可以通过历史记录查看",
		}, nil
	}

	// 构建回复
	response := "您当前有 " + strconv.Itoa(len(activeOrders)) + " 个进行中的订单：\n\n"
	for i, order := range activeOrders {
		statusText := "配送中"
		if order.Status == 1 {
			statusText = "已接单"
		}
		response += fmt.Sprintf(
			"%d. 订单 %d - %s\n   从 %s 到 %s\n   金额：¥%.2f\n\n",
			i+1,
			order.ID,
			statusText,
			order.Origin,
			order.Destination,
			order.Amount,
		)
	}

	return &SkillResult{
		Success:  true,
		Response: response,
		Data:     activeOrders,
	}, nil
}

// handleDeliveryProblem 处理配送问题
func (s *OrderSkill) handleDeliveryProblem(ctx context.Context, params SkillParams) (*SkillResult, error) {
	response := "遇到配送问题了吗？我可以帮您：\n\n" +
		"1. **商家出餐慢** - 我可以帮您联系商家催促\n" +
		"2. **联系不上顾客** - 提供备用联系方式或上报异常\n" +
		"3. **地址错误** - 协助修改地址或联系顾客确认\n" +
		"4. **交通事故/意外** - 紧急上报并启动保险理赔\n" +
		"5. **其他问题** - 转接人工客服\n\n" +
		"请告诉我具体情况，我会为您提供解决方案。"

	return &SkillResult{
		Success:  true,
		Response: response,
		SuggestedActions: []Action{
			{Type: "report_delay", Name: "商家延迟", Params: `{}`},
			{Type: "report_unreachable", Name: "联系不上", Params: `{}`},
			{Type: "report_accident", Name: "交通事故", Params: `{}`},
			{Type: "contact_human", Name: "人工客服", Params: `{}`},
		},
	}, nil
}

// handleCancelOrder 处理取消订单
func (s *OrderSkill) handleCancelOrder(ctx context.Context, params SkillParams) (*SkillResult, error) {
	return &SkillResult{
		Success:       true,
		Response:      "取消订单需要先上报异常原因，确认后系统会为您处理。",
		NeedConfirm:   true,
		ConfirmPrompt: "请确认取消原因：1.商家问题 2.顾客问题 3.个人原因 4.其他",
		SuggestedActions: []Action{
			{Type: "cancel_merchant", Name: "商家问题", Params: `{}`},
			{Type: "cancel_customer", Name: "顾客问题", Params: `{}`},
			{Type: "cancel_personal", Name: "个人原因", Params: `{}`},
		},
	}, nil
}

// handleDefault 默认处理
func (s *OrderSkill) handleDefault(ctx context.Context, params SkillParams) (*SkillResult, error) {
	return &SkillResult{
		Success:  true,
		Response: "我可以帮您处理以下订单相关事务：\n\n1. 查询订单详情\n2. 跟踪订单状态\n3. 处理配送问题\n4. 取消订单\n\n请告诉我您需要做什么？",
		SuggestedActions: []Action{
			{Type: "query_order", Name: "查询订单", Params: `{}`},
			{Type: "track_order", Name: "跟踪订单", Params: `{}`},
			{Type: "delivery_help", Name: "配送帮助", Params: `{}`},
		},
	}, nil
}

// GetExamples 获取示例
func (s *OrderSkill) GetExamples() []SkillExample {
	return []SkillExample{
		{
			Input:    "帮我查一下订单 12345",
			Intent:   "query_order",
			Entities: map[string]string{"order_id": "12345"},
			Response: "订单 12345 信息如下：状态：配送中，起点：XX商家，终点：XX小区...",
		},
		{
			Input:    "我现在的订单在哪",
			Intent:   "track_order",
			Entities: map[string]string{},
			Response: "您当前有 2 个进行中的订单：1. 订单 123 - 配送中...",
		},
		{
			Input:    "商家出餐太慢了怎么办",
			Intent:   "delivery_problem",
			Entities: map[string]string{"problem_type": "merchant_delay"},
			Response: "遇到商家出餐慢的情况，我可以帮您联系商家催促...",
		},
	}
}
