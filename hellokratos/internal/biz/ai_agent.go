package biz

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"hellokratos/internal/biz/ai"
	"hellokratos/internal/data"
	"hellokratos/internal/data/model"

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
	aiAgentRepo     data.AIAgentRepo
	orderRepo       data.OrderRepo
	incomeRepo      data.IncomeRepo
	vectorDBService VectorDBService
	llmService      LLMService
	ocrService      ai.OCRService
	asrService      ai.ASRService
	toolRegistry    *ToolRegistry
	skillManager    *SkillManager
	log             *log.Helper
}

// NewAIAgentUsecase 创建AI客服业务逻辑实例
func NewAIAgentUsecase(aiAgentRepo data.AIAgentRepo, orderRepo data.OrderRepo, incomeRepo data.IncomeRepo, vectorDBService VectorDBService, llmService LLMService, ocrService ai.OCRService, asrService ai.ASRService, skillManager *SkillManager, logger log.Logger) AIAgentUsecase {
	// 创建工具注册表
	toolRegistry := CreateRiderTools(orderRepo, incomeRepo)

	return &aiAgentUsecase{
		aiAgentRepo:     aiAgentRepo,
		orderRepo:       orderRepo,
		incomeRepo:      incomeRepo,
		vectorDBService: vectorDBService,
		llmService:      llmService,
		ocrService:      ocrService,
		asrService:      asrService,
		toolRegistry:    toolRegistry,
		skillManager:    skillManager,
		log:             log.NewHelper(logger),
	}
}

// SendMessage 发送消息给AI客服
func (uc *aiAgentUsecase) SendMessage(ctx context.Context, riderID int64, content string, messageType int32) (*AIResponse, error) {
	// 0. 检查请求频率限制
	allowed, err := uc.aiAgentRepo.CheckRateLimit(ctx, riderID)
	if err != nil {
		uc.log.Warn("check rate limit failed", "err", err)
		// 继续处理，不阻止用户
	} else if !allowed {
		return &AIResponse{
			Content:      "您的消息发送过于频繁，请稍后再试",
			ResponseType: 1,
			Actions:      []Action{},
		}, nil
	}

	// 1. 获取或创建会话
	sessionID, err := uc.GetOrCreateSession(ctx, riderID)
	if err != nil {
		return nil, err
	}

	// 处理不同模式的输入内容
	processedContent := content
	if messageType == 1 {
		// 图片输入 - 调用 OCR
		ocrText, err := uc.ocrService.Recognize(ctx, content)
		if err != nil {
			uc.log.Errorf("OCR recognition failed: %v", err)
			processedContent = "[骑手上传了一张图片，但系统识别失败]"
		} else {
			processedContent = "[骑手上传图片识别内容]: " + ocrText
		}
	} else if messageType == 2 {
		// 语音输入 - 调用 ASR
		asrText, err := uc.asrService.Recognize(ctx, content)
		if err != nil {
			uc.log.Errorf("ASR recognition failed: %v", err)
			processedContent = "[骑手发送了一段语音，但系统识别失败]"
		} else {
			processedContent = "[骑手发送语音识别内容]: " + asrText
		}
	}

	// 2. 保存用户消息
	userMessage := &model.AIAgentMessage{
		RiderID:     riderID,
		Content:     processedContent,
		MessageType: 0, // 用户消息
		ContentType: messageType,
		SessionID:   sessionID,
	}
	err = uc.aiAgentRepo.CreateMessage(ctx, userMessage)
	if err != nil {
		uc.log.Error("create user message failed", "err", err)
		return nil, err
	}

	// 3. AI处理消息 (统一传入处理好的文本内容)
	aiResponse, err := uc.processMessage(ctx, riderID, processedContent)
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

	// 5. 更新限流计数
	if err := uc.aiAgentRepo.UpdateRateLimit(ctx, riderID); err != nil {
		uc.log.Warn("update rate limit failed", "err", err)
	}

	return aiResponse, nil
}

// processMessage 处理用户消息（基于RAG+向量数据库+大模型+Skill）
func (uc *aiAgentUsecase) processMessage(ctx context.Context, riderID int64, content string) (*AIResponse, error) {
	response := &AIResponse{
		ResponseType: 0, // 默认文本回复
		Actions:      []Action{},
	}

	// 1. 获取相关上下文信息
	contextInfo, err := uc.getContextInfo(ctx, riderID, content)
	if err != nil {
		uc.log.Error("get context info failed", "err", err)
		response.Content = "获取上下文信息失败，请稍后重试。"
		return response, nil
	}

	// 2. 先尝试使用 Skill 处理
	if uc.skillManager != nil {
		skillResult, err := uc.skillManager.Process(ctx, riderID, content, contextInfo)
		if err != nil {
			uc.log.Error("skill process failed", "err", err)
		} else if skillResult != nil && skillResult.Success {
			// Skill 成功处理，直接返回结果
			response.Content = skillResult.Response
			response.ResponseType = 0 // 文本回复

			// 转换 Skill 的 Actions 为 AIResponse 的 Actions
			for _, action := range skillResult.SuggestedActions {
				response.Actions = append(response.Actions, Action{
					ActionType:   action.Type,
					ActionName:   action.Name,
					ActionParams: action.Params,
				})
			}

			return response, nil
		}
	}

	// 3. Skill 未处理或处理失败，使用 RAG + LLM 生成回复
	reply, err := uc.generateAIResponse(ctx, content, contextInfo)
	if err != nil {
		uc.log.Error("generate AI response failed", "err", err)
		response.Content = "AI处理失败，请稍后重试。"
		return response, nil
	}

	response.Content = reply

	// 4. 提取可执行操作
	actions := uc.extractActions(content, reply)
	response.Actions = actions

	// 5. 确定响应类型
	response.ResponseType = uc.determineResponseType(content, reply)

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

		// 异步增加查看次数（使用新的context，避免原context被取消）
		go func(faqID int64) {
			bgCtx := context.Background()
			if err := uc.aiAgentRepo.IncrementFAQViewCount(bgCtx, faqID); err != nil {
				uc.log.Warn("increment FAQ view count failed", "err", err)
			}
		}(faq.ID)
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

	// 检查会话是否超时（30分钟）
	if session != nil {
		timeout := 30 * time.Minute
		lastMessage, err := uc.aiAgentRepo.GetLastMessageBySessionID(ctx, session.ID)
		if err == nil && lastMessage != nil {
			if time.Since(lastMessage.CreatedAt) > timeout {
				// 会话已超时，结束它
				if err := uc.aiAgentRepo.EndSession(ctx, session.ID); err != nil {
					uc.log.Warn("end timeout session failed", "err", err)
				}
				session = nil // 标记会话为无效，创建新会话
			}
		}
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

// getContextInfo 获取相关上下文信息
func (uc *aiAgentUsecase) getContextInfo(ctx context.Context, riderID int64, content string) (map[string]interface{}, error) {
	contextInfo := make(map[string]interface{})

	// 1. 获取用户基本信息
	contextInfo["rider_id"] = riderID

	// 2. 获取最近订单信息
	orders, _, err := uc.orderRepo.GetOrdersByRiderID(ctx, riderID, 0, 1, 5)
	if err != nil {
		uc.log.Warn("get orders failed", "err", err)
	} else {
		contextInfo["orders"] = orders
	}

	// 3. 获取收入信息
	totalAmount, err := uc.incomeRepo.GetTotalIncomeByRiderID(ctx, riderID)
	if err != nil {
		uc.log.Warn("get income failed", "err", err)
	} else {
		contextInfo["total_income"] = totalAmount
	}

	// 4. 获取最近聊天历史
	messages, err := uc.aiAgentRepo.GetMessagesByRiderID(ctx, riderID, 0, 10)
	if err != nil {
		uc.log.Warn("get chat history failed", "err", err)
	} else {
		contextInfo["chat_history"] = messages
	}

	// 5. 获取相关FAQ
	faqs, err := uc.aiAgentRepo.GetFAQList(ctx, "", 5)
	if err != nil {
		uc.log.Warn("get faq failed", "err", err)
	} else {
		contextInfo["faqs"] = faqs
	}

	return contextInfo, nil
}

// generateAIResponse 生成AI回复（使用RAG+向量数据库+大模型）
func (uc *aiAgentUsecase) generateAIResponse(ctx context.Context, content string, contextInfo map[string]interface{}) (string, error) {
	// 1. 向向量数据库查询相关信息
	relevantInfo, err := uc.vectorDBService.Search(ctx, content, 5)
	if err != nil {
		uc.log.Warn("vectorDB search failed", "err", err)
	}

	// 2. 构建历史对话上下文
	history := uc.buildChatHistory(contextInfo)

	// 3. 构建提示词
	prompt := uc.buildPrompt(content, contextInfo, relevantInfo)

	// 4. 获取可用工具
	tools := uc.toolRegistry.GetTools()

	// 5. 调用大模型生成回复（支持工具调用）
	reply, toolCalls, err := uc.llmService.GenerateResponseWithTools(ctx, prompt, history, tools)
	if err != nil {
		uc.log.Warn("LLM generate failed", "err", err)
		// 降级为本地处理
		return uc.fallbackGenerateResponse(content, contextInfo), nil
	}

	// 6. 处理工具调用
	if len(toolCalls) > 0 {
		// 将 rider_id 添加到 context
		if riderID, ok := contextInfo["rider_id"].(int64); ok {
			ctx = context.WithValue(ctx, "rider_id", riderID)
		}

		// 执行工具调用
		var toolResults []ToolResult
		for _, toolCall := range toolCalls {
			result := uc.toolRegistry.Execute(ctx, toolCall)
			toolResults = append(toolResults, result)
		}

		// 将工具结果添加到历史对话
		history = append(history, map[string]string{
			"role":    "assistant",
			"content": reply,
		})

		// 添加工具调用结果到历史
		for _, result := range toolResults {
			resultJSON, _ := json.Marshal(result.Result)
			history = append(history, map[string]string{
				"role":    "tool",
				"content": string(resultJSON),
			})
		}

		// 再次调用LLM生成最终回复
		finalReply, _, err := uc.llmService.GenerateResponseWithTools(ctx, "根据查询结果回答用户问题", history, nil)
		if err != nil {
			uc.log.Warn("LLM final generate failed", "err", err)
			return reply, nil
		}
		return finalReply, nil
	}

	return reply, nil
}

// buildChatHistory 构建历史对话上下文
func (uc *aiAgentUsecase) buildChatHistory(contextInfo map[string]interface{}) []map[string]string {
	var history []map[string]string

	// 从contextInfo中获取聊天历史
	if messages, ok := contextInfo["chat_history"].([]*model.AIAgentMessage); ok && len(messages) > 0 {
		// 取最近10条消息作为上下文
		start := 0
		if len(messages) > 10 {
			start = len(messages) - 10
		}

		for _, msg := range messages[start:] {
			role := "user"
			if msg.MessageType == 1 { // AI回复
				role = "assistant"
			}
			history = append(history, map[string]string{
				"role":    role,
				"content": msg.Content,
			})
		}
	}

	return history
}

// buildPrompt 构建提示词
func (uc *aiAgentUsecase) buildPrompt(content string, contextInfo map[string]interface{}, relevantInfo []map[string]interface{}) string {
	prompt := fmt.Sprintf("用户问题：%s\n\n", content)

	// 添加上下文信息
	prompt += "上下文信息：\n"

	// 订单信息
	if orders, ok := contextInfo["orders"].([]*model.Order); ok && len(orders) > 0 {
		prompt += "最近订单：\n"
		for i, order := range orders[:min(3, len(orders))] {
			prompt += fmt.Sprintf("%d. 订单号：%s，状态：%s\n", i+1, order.OrderNo, getOrderStatusText(order.Status))
		}
	}

	// 收入信息
	if totalIncome, ok := contextInfo["total_income"].(float64); ok {
		prompt += fmt.Sprintf("总收入：%.2f 元\n", totalIncome)
	}

	// 相关FAQ
	if faqs, ok := contextInfo["faqs"].([]*model.AIAgentFAQ); ok && len(faqs) > 0 {
		prompt += "相关常见问题：\n"
		for i, faq := range faqs[:min(3, len(faqs))] {
			prompt += fmt.Sprintf("%d. %s\n", i+1, faq.Question)
		}
	}

	// 相关文档（从向量数据库检索）
	if len(relevantInfo) > 0 {
		prompt += "相关文档：\n"
		for i, info := range relevantInfo[:min(3, len(relevantInfo))] {
			if content, ok := info["content"].(string); ok {
				prompt += fmt.Sprintf("%d. %s\n", i+1, content)
			}
		}
	}

	prompt += "\n请根据以上信息，生成一个友好、准确的回复。回复应该：\n"
	prompt += "1. 直接回答用户问题\n"
	prompt += "2. 基于提供的上下文信息\n"
	prompt += "3. 使用自然、友好的语言\n"
	prompt += "4. 避免使用专业术语\n"
	prompt += "5. 如果无法回答，请诚实告知\n"

	return prompt
}

// fallbackGenerateResponse 降级生成回复
func (uc *aiAgentUsecase) fallbackGenerateResponse(content string, contextInfo map[string]interface{}) string {
	// 基于上下文生成回复
	switch {
	case containsKeywords(content, []string{"订单", "order", "配送"}):
		orders, ok := contextInfo["orders"].([]*model.Order)
		if ok && len(orders) > 0 {
			return fmt.Sprintf("您好，您当前有 %d 个订单。最近的订单是 %s，状态为 %s。", len(orders), orders[0].OrderNo, getOrderStatusText(orders[0].Status))
		} else {
			return "您好，您当前没有订单。"
		}

	case containsKeywords(content, []string{"收入", "收益", "income", "money"}):
		totalIncome, ok := contextInfo["total_income"].(float64)
		if ok {
			return fmt.Sprintf("您好，您的总收入为 %.2f 元。", totalIncome)
		} else {
			return "您好，暂时无法查询您的收入信息。"
		}

	case containsKeywords(content, []string{"人工", "客服", "help", "human"}):
		return "正在为您转接人工客服，请稍候..."

	case containsKeywords(content, []string{"规则", "rule", "规定"}):
		return "平台规则说明：\n1. 骑手需通过实名认证\n2. 配送过程中需遵守交通规则\n3. 订单需在约定时间内完成\n4. 保持良好的服务态度"

	default:
		return "您好！我是智能客服助手。您可以问我关于订单、收入、规则等问题，或者输入人工转接人工客服。"
	}
}

// min 返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// extractActions 提取可执行操作
func (uc *aiAgentUsecase) extractActions(content string, reply string) []Action {
	actions := []Action{}

	if containsKeywords(content, []string{"订单", "order"}) {
		actions = append(actions, Action{
			ActionType:   "view_order",
			ActionName:   "查看订单详情",
			ActionParams: "{}",
		})
	}

	if containsKeywords(content, []string{"收入", "收益", "income"}) {
		actions = append(actions, Action{
			ActionType:   "view_income",
			ActionName:   "查看收入明细",
			ActionParams: "{}",
		})
	}

	if containsKeywords(content, []string{"人工", "客服"}) {
		actions = append(actions, Action{
			ActionType:   "contact_human",
			ActionName:   "联系人工客服",
			ActionParams: "{}",
		})
	}

	if containsKeywords(content, []string{"规则", "rule"}) {
		actions = append(actions, Action{
			ActionType:   "view_rules",
			ActionName:   "查看完整规则",
			ActionParams: "{}",
		})
	}

	return actions
}

// determineResponseType 确定响应类型
func (uc *aiAgentUsecase) determineResponseType(content string, reply string) int32 {
	if containsKeywords(content, []string{"订单", "order"}) {
		return 1 // 订单查询结果
	}

	if containsKeywords(content, []string{"收入", "收益", "income"}) {
		return 2 // 收入查询结果
	}

	if containsKeywords(content, []string{"人工", "客服"}) {
		return 4 // 转人工
	}

	return 0 // 文本回复
}

// containsKeywords 检查字符串是否包含关键词（辅助方法）
func containsKeywords(content string, keywords []string) bool {
	contentLower := strings.ToLower(content)
	for _, keyword := range keywords {
		if strings.Contains(contentLower, strings.ToLower(keyword)) {
			return true
		}
	}
	return false
}

// getOrderStatusText 获取订单状态文本
func getOrderStatusText(status int32) string {
	switch status {
	case 0:
		return "待支付"
	case 1:
		return "配送中"
	case 2:
		return "已完成"
	case 3:
		return "已取消"
	default:
		return "未知"
	}
}

// toJSON 转换为JSON字符串
func toJSON(v interface{}) string {
	data, _ := json.Marshal(v)
	return string(data)
}
