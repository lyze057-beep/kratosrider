package biz

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"hellokratos/internal/data"
)

// Tool 工具定义
type Tool struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Parameters  ToolParams `json:"parameters"`
}

// ToolParams 工具参数定义
type ToolParams struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties"`
	Required   []string            `json:"required"`
}

// Property 参数属性
type Property struct {
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Enum        []string `json:"enum,omitempty"`
}

// ToolCall 工具调用请求
type ToolCall struct {
	ID       string           `json:"id"`
	Type     string           `json:"type"`
	Function ToolCallFunction `json:"function"`
}

// ToolCallFunction 工具调用函数
type ToolCallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// ToolResult 工具执行结果
type ToolResult struct {
	ToolCallID string      `json:"tool_call_id"`
	Name       string      `json:"name"`
	Result     interface{} `json:"result"`
	Error      string      `json:"error,omitempty"`
}

// ToolRegistry 工具注册表
type ToolRegistry struct {
	tools    map[string]Tool
	handlers map[string]ToolHandler
}

// ToolHandler 工具处理函数类型
type ToolHandler func(ctx context.Context, args map[string]interface{}) (interface{}, error)

// NewToolRegistry 创建工具注册表
func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools:    make(map[string]Tool),
		handlers: make(map[string]ToolHandler),
	}
}

// Register 注册工具
func (r *ToolRegistry) Register(tool Tool, handler ToolHandler) {
	r.tools[tool.Name] = tool
	r.handlers[tool.Name] = handler
}

// GetTools 获取所有工具定义（用于传给LLM）
func (r *ToolRegistry) GetTools() []Tool {
	var tools []Tool
	for _, tool := range r.tools {
		tools = append(tools, tool)
	}
	return tools
}

// Execute 执行工具调用
func (r *ToolRegistry) Execute(ctx context.Context, toolCall ToolCall) ToolResult {
	handler, exists := r.handlers[toolCall.Function.Name]
	if !exists {
		return ToolResult{
			ToolCallID: toolCall.ID,
			Name:       toolCall.Function.Name,
			Error:      fmt.Sprintf("tool not found: %s", toolCall.Function.Name),
		}
	}

	// 解析参数
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
		return ToolResult{
			ToolCallID: toolCall.ID,
			Name:       toolCall.Function.Name,
			Error:      fmt.Sprintf("parse arguments failed: %v", err),
		}
	}

	// 执行工具
	result, err := handler(ctx, args)
	if err != nil {
		return ToolResult{
			ToolCallID: toolCall.ID,
			Name:       toolCall.Function.Name,
			Error:      err.Error(),
		}
	}

	return ToolResult{
		ToolCallID: toolCall.ID,
		Name:       toolCall.Function.Name,
		Result:     result,
	}
}

// CreateRiderTools 创建骑手相关的工具集
func CreateRiderTools(
	orderRepo data.OrderRepo,
	incomeRepo data.IncomeRepo,
) *ToolRegistry {
	registry := NewToolRegistry()

	// 1. 查询订单详情
	registry.Register(Tool{
		Name:        "query_order_detail",
		Description: "查询指定订单的详细信息",
		Parameters: ToolParams{
			Type: "object",
			Properties: map[string]Property{
				"order_id": {
					Type:        "string",
					Description: "订单ID",
				},
			},
			Required: []string{"order_id"},
		},
	}, func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
		orderIDStr, ok := args["order_id"].(string)
		if !ok {
			return nil, fmt.Errorf("order_id must be string")
		}
		orderID, err := strconv.ParseInt(orderIDStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid order_id: %v", err)
		}

		order, err := orderRepo.GetOrderByID(ctx, orderID)
		if err != nil {
			return nil, err
		}

		return map[string]interface{}{
			"order_id":    order.ID,
			"status":      order.Status,
			"origin":      order.Origin,
			"destination": order.Destination,
			"amount":      order.Amount,
			"distance":    order.Distance,
			"create_time": order.CreatedAt.Format("2006-01-02 15:04:05"),
		}, nil
	})

	// 2. 查询骑手收入统计
	registry.Register(Tool{
		Name:        "query_income_stats",
		Description: "查询骑手的收入统计信息",
		Parameters: ToolParams{
			Type:       "object",
			Properties: map[string]Property{},
			Required:   []string{},
		},
	}, func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
		riderID, ok := ctx.Value("rider_id").(int64)
		if !ok {
			return nil, fmt.Errorf("rider_id not found in context")
		}

		totalIncome, err := incomeRepo.GetTotalIncomeByRiderID(ctx, riderID)
		if err != nil {
			return nil, err
		}

		incomes, err := incomeRepo.GetIncomesByRiderID(ctx, riderID, 10)
		if err != nil {
			return nil, err
		}

		return map[string]interface{}{
			"total_income": totalIncome,
			"recent_count": len(incomes),
		}, nil
	})

	// 3. 查询订单列表
	registry.Register(Tool{
		Name:        "query_order_list",
		Description: "查询骑手的订单列表",
		Parameters: ToolParams{
			Type: "object",
			Properties: map[string]Property{
				"limit": {
					Type:        "integer",
					Description: "返回数量限制，默认10条",
				},
			},
			Required: []string{},
		},
	}, func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
		riderID, ok := ctx.Value("rider_id").(int64)
		if !ok {
			return nil, fmt.Errorf("rider_id not found in context")
		}

		limit := 10
		if l, ok := args["limit"].(float64); ok {
			limit = int(l)
		}

		// 使用现有的接口查询订单
		orders, _, err := orderRepo.GetOrdersByRiderID(ctx, riderID, 0, 1, limit)
		if err != nil {
			return nil, err
		}

		var result []map[string]interface{}
		for _, order := range orders {
			result = append(result, map[string]interface{}{
				"order_id":    order.ID,
				"status":      order.Status,
				"origin":      order.Origin,
				"destination": order.Destination,
				"amount":      order.Amount,
				"create_time": order.CreatedAt.Format("2006-01-02 15:04:05"),
			})
		}

		return result, nil
	})

	// 4. 上报异常
	registry.Register(Tool{
		Name:        "report_exception",
		Description: "上报配送过程中的异常情况",
		Parameters: ToolParams{
			Type: "object",
			Properties: map[string]Property{
				"order_id": {
					Type:        "string",
					Description: "订单ID",
				},
				"exception_type": {
					Type:        "string",
					Description: "异常类型",
					Enum:        []string{"merchant_delay", "customer_unreachable", "address_error", "accident", "other"},
				},
				"description": {
					Type:        "string",
					Description: "异常描述",
				},
			},
			Required: []string{"order_id", "exception_type", "description"},
		},
	}, func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
		riderID, ok := ctx.Value("rider_id").(int64)
		if !ok {
			return nil, fmt.Errorf("rider_id not found in context")
		}

		orderIDStr, _ := args["order_id"].(string)
		orderID, _ := strconv.ParseInt(orderIDStr, 10, 64)

		// 这里应该调用异常上报接口
		// TODO: 实现异常上报逻辑

		return map[string]interface{}{
			"success":   true,
			"message":   "异常已上报，客服将尽快处理",
			"ticket_id": fmt.Sprintf("TK%d%d", riderID, orderID),
		}, nil
	})

	return registry
}
