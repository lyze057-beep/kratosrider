package biz

import (
	"context"
	"encoding/json"
	"fmt"
	"hellokratos/internal/data"
	"hellokratos/internal/data/model"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

// AIAgentUsecase AI客服业务逻辑接口
type AIAgentUsecase interface {
	// 消息处理
	SendMessage(ctx context.Context, riderID int64, content string, messageType int32) (*AIResponse, error)
	GetChatHistory(ctx context.Context, riderID int64, lastID int64, limit int) ([]*ChatMessage, error)

	// FAQ
	GetFAQList(ctx context.Context, category string, limit int) ([]*FAQItem, error)

	// 会话管理
	GetOrCreateSession(ctx context.Context, riderID int64) (int64, error)
	EndSession(ctx context.Context, sessionID int64) error
	RateSession(ctx context.Context, sessionID int64, rating int32, feedback string) error

	// 智能推荐
	GetSuggestedQuestions(ctx context.Context, riderID int64, context string) ([]string, error)
}

// AIResponse AI响应
type AIResponse struct {
	Content      string
	ResponseType int32
	Actions      []Action
}

// Action 可执行操作
type Action struct {
	ActionType   string
	ActionName   string
	ActionParams string
}

// ChatMessage 聊天消息
type ChatMessage struct {
	ID          int64
	Content     string
	MessageType int32
	ContentType int32
	CreatedAt   string
}

// FAQItem FAQ项
type FAQItem struct {
	ID        int64
	Question  string
	Answer    string
	Category  string
	ViewCount int32
}

// aiAgentUsecase AI客服业务逻辑实现
type aiAgentUsecase struct {
	aiAgentRepo data.AIAgentRepo
	orderRepo   data.OrderRepo
	incomeRepo  data.IncomeRepo
	log         *log.Helper
}

// NewAIAgentUsecase 创建AI客服业务逻辑实例
func NewAIAgentUsecase(aiAgentRepo data.AIAgentRepo, orderRepo data.OrderRepo, incomeRepo data.IncomeRepo, logger log.Logger) AIAgentUsecase {
	return &aiAgentUsecase{
		aiAgentRepo: aiAgentRepo,
		orderRepo:   orderRepo,
		incomeRepo:  incomeRepo,
		log:         log.NewHelper(logger),
	}
}

// SendMessage 发送消息给AI客服
func (uc *aiAgentUsecase) SendMessage(ctx context.Context, riderID int64, content string, messageType int32) (*AIResponse, error) {
	// 1. 获取或创建会话
	sessionID, err := uc.GetOrCreateSession(ctx, riderID)
	if err != nil {
		return nil, err
	}

	// 2. 保存用户消息
	userMessage := &model.AIAgentMessage{
		RiderID:     riderID,
		Content:     content,
		MessageType: 0, // 用户消息
		ContentType: messageType,
		SessionID:   sessionID,
	}
	err = uc.aiAgentRepo.CreateMessage(ctx, userMessage)
	if err != nil {
		uc.log.Error("create user message failed", "err", err)
		return nil, err
	}

	// 3. AI处理消息
	aiResponse, err := uc.processMessage(ctx, riderID, content)
	if err != nil {
		uc.log.Error("process message failed", "err", err)
		return nil, err
	}

	// 4. 保存AI回复
	aiMessage := &model.AIAgentMessage{
		RiderID:     riderID,
		Content:     aiResponse.Content,
		MessageType: 1, // AI回复
		ContentType: 0, // 文本
		SessionID:   sessionID,
	}
	err = uc.aiAgentRepo.CreateMessage(ctx, aiMessage)
	if err != nil {
		uc.log.Error("create ai message failed", "err", err)
		return nil, err
	}

	return aiResponse, nil
}

// processMessage 处理用户消息
func (uc *aiAgentUsecase) processMessage(ctx context.Context, riderID int64, content string) (*AIResponse, error) {
	// 简单的关键词匹配（实际项目中可以使用NLP或调用AI服务）
	response := &AIResponse{
		ResponseType: 0, // 默认文本回复
		Actions:      []Action{},
	}

	switch {
	case contains(content, []string{"订单", "order"}):
		// 查询订单
		orders, _, err := uc.orderRepo.GetOrdersByRiderID(ctx, riderID, 1, 1, 5) // status=1配送中, page=1, pageSize=5
		if err != nil {
			response.Content = "查询订单信息失败，请稍后重试。"
			return response, nil
		}

		if len(orders) == 0 {
			response.Content = "您当前没有进行中的订单。"
		} else {
			response.Content = fmt.Sprintf("您有 %d 个进行中的订单。", len(orders))
			response.ResponseType = 1 // 订单查询结果
			response.Actions = append(response.Actions, Action{
				ActionType:   "view_order",
				ActionName:   "查看订单详情",
				ActionParams: "{}",
			})
		}

	case contains(content, []string{"收入", "收益", "income", "money"}):
		// 查询收入
		totalAmount, err := uc.incomeRepo.GetTotalIncomeByRiderID(ctx, riderID)
		if err != nil {
			response.Content = "查询收入信息失败，请稍后重试。"
			return response, nil
		}

		response.Content = fmt.Sprintf("您的总收入为 %.2f 元。", totalAmount)
		response.ResponseType = 2 // 收入查询结果
		response.Actions = append(response.Actions, Action{
			ActionType:   "view_income",
			ActionName:   "查看收入明细",
			ActionParams: "{}",
		})

	case contains(content, []string{"人工", "客服", "help", "人工客服"}):
		response.Content = "正在为您转接人工客服，请稍候..."
		response.ResponseType = 4 // 转人工
		response.Actions = append(response.Actions, Action{
			ActionType:   "contact_human",
			ActionName:   "联系人工客服",
			ActionParams: "{}",
		})

	case contains(content, []string{"规则", "rule", "规定"}):
		response.Content = "平台规则说明：\n1. 骑手需通过实名认证\n2. 配送过程中需遵守交通规则\n3. 订单需在约定时间内完成\n4. 保持良好的服务态度"
		response.Actions = append(response.Actions, Action{
			ActionType:   "view_rules",
			ActionName:   "查看完整规则",
			ActionParams: "{}",
		})

	default:
		response.Content = "您好！我是智能客服助手。您可以问我关于订单、收入、规则等问题，或者输入人工转接人工客服。"
	}

	return response, nil
}

// GetChatHistory 获取聊天历史
func (uc *aiAgentUsecase) GetChatHistory(ctx context.Context, riderID int64, lastID int64, limit int) ([]*ChatMessage, error) {
	messages, err := uc.aiAgentRepo.GetMessagesByRiderID(ctx, riderID, lastID, limit)
	if err != nil {
		return nil, err
	}

	var result []*ChatMessage
	for _, msg := range messages {
		result = append(result, &ChatMessage{
			ID:          msg.ID,
			Content:     msg.Content,
			MessageType: msg.MessageType,
			ContentType: msg.ContentType,
			CreatedAt:   msg.CreatedAt.Format(time.RFC3339),
		})
	}

	return result, nil
}

// GetFAQList 获取FAQ列表
func (uc *aiAgentUsecase) GetFAQList(ctx context.Context, category string, limit int) ([]*FAQItem, error) {
	faqs, err := uc.aiAgentRepo.GetFAQList(ctx, category, limit)
	if err != nil {
		return nil, err
	}

	var result []*FAQItem
	for _, faq := range faqs {
		result = append(result, &FAQItem{
			ID:        faq.ID,
			Question:  faq.Question,
			Answer:    faq.Answer,
			Category:  faq.Category,
			ViewCount: faq.ViewCount,
		})

		// 异步增加查看次数
		go uc.aiAgentRepo.IncrementFAQViewCount(ctx, faq.ID)
	}

	return result, nil
}

// GetOrCreateSession 获取或创建会话
func (uc *aiAgentUsecase) GetOrCreateSession(ctx context.Context, riderID int64) (int64, error) {
	// 查找活跃会话
	session, err := uc.aiAgentRepo.GetActiveSessionByRiderID(ctx, riderID)
	if err != nil {
		return 0, err
	}

	if session != nil {
		return session.ID, nil
	}

	// 创建新会话
	newSession := &model.AIAgentSession{
		RiderID: riderID,
		Status:  0, // 进行中
	}
	err = uc.aiAgentRepo.CreateSession(ctx, newSession)
	if err != nil {
		return 0, err
	}

	return newSession.ID, nil
}

// EndSession 结束会话
func (uc *aiAgentUsecase) EndSession(ctx context.Context, sessionID int64) error {
	return uc.aiAgentRepo.EndSession(ctx, sessionID)
}

// RateSession 评价会话
func (uc *aiAgentUsecase) RateSession(ctx context.Context, sessionID int64, rating int32, feedback string) error {
	return uc.aiAgentRepo.RateSession(ctx, sessionID, rating, feedback)
}

// GetSuggestedQuestions 获取智能推荐问题
func (uc *aiAgentUsecase) GetSuggestedQuestions(ctx context.Context, riderID int64, contextStr string) ([]string, error) {
	// 根据上下文返回推荐问题
	switch contextStr {
	case "order":
		return []string{
			"如何查看我的订单？",
			"订单超时怎么办？",
			"如何取消订单？",
			"订单收入怎么计算？",
		}, nil
	case "income":
		return []string{
			"我的收入在哪里查看？",
			"如何申请提现？",
			"提现多久到账？",
			"收入计算公式是什么？",
		}, nil
	case "profile":
		return []string{
			"如何修改个人信息？",
			"实名认证怎么操作？",
			"如何更换手机号？",
			"忘记密码怎么办？",
		}, nil
	default:
		return []string{
			"如何查看我的订单？",
			"我的收入在哪里查看？",
			"平台规则有哪些？",
			"如何联系人工客服？",
		}, nil
	}
}

// contains 检查字符串是否包含关键词
func contains(content string, keywords []string) bool {
	for _, keyword := range keywords {
		if len(content) >= len(keyword) {
			// 简单包含检查
			for i := 0; i <= len(content)-len(keyword); i++ {
				if content[i:i+len(keyword)] == keyword {
					return true
				}
			}
		}
	}
	return false
}

// toJSON 转换为JSON字符串
func toJSON(v interface{}) string {
	data, _ := json.Marshal(v)
	return string(data)
}
