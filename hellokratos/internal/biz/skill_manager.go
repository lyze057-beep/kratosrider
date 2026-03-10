package biz

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-kratos/kratos/v2/log"

	"hellokratos/internal/biz/ai"
)

// SkillManager 技能管理器
// 负责管理所有 Skill 的注册、路由和执行
type SkillManager struct {
	registry   *ai.SkillRegistry
	llmService LLMService // 用于意图识别
	log        *log.Helper
}

// NewSkillManager 创建技能管理器
func NewSkillManager(llmService LLMService, logger log.Logger) *SkillManager {
	return &SkillManager{
		registry:   ai.NewSkillRegistry(),
		llmService: llmService,
		log:        log.NewHelper(logger),
	}
}

// RegisterSkill 注册技能
func (m *SkillManager) RegisterSkill(skill ai.Skill) {
	m.registry.Register(skill)
	m.log.Infof("Skill registered: %s", skill.Name())
}

// Process 处理用户请求
// 1. 意图识别
// 2. 实体提取
// 3. Skill 路由
// 4. 执行 Skill
func (m *SkillManager) Process(ctx context.Context, riderID int64, input string, contextInfo map[string]interface{}) (*ai.SkillResult, error) {
	m.log.Infof("Processing request: rider=%d, input=%s", riderID, input)

	// 1. 意图识别
	intent, entities, err := m.recognizeIntent(ctx, input)
	if err != nil {
		m.log.Warn("Intent recognition failed", "err", err)
		// 降级处理：使用关键词匹配
		intent, entities = m.fallbackIntentRecognition(input)
	}

	m.log.Infof("Intent recognized: %s, entities: %v", intent, entities)

	// 2. 查找合适的 Skill
	skill := m.registry.FindSkill(ctx, intent, entities)
	if skill == nil {
		// 没有找到合适的 Skill，使用通用回复
		return m.handleGeneralQuery(ctx, input, contextInfo)
	}

	// 3. 构建 Skill 参数
	params := ai.SkillParams{
		RawInput: input,
		Intent:   intent,
		Entities: entities,
		Context:  contextInfo,
		RiderID:  riderID,
	}

	// 4. 执行 Skill
	result, err := skill.Execute(ctx, params)
	if err != nil {
		m.log.Error("Skill execution failed", "skill", skill.Name(), "err", err)
		return &ai.SkillResult{
			Success:  false,
			Response: "处理请求时出现错误，请稍后重试",
		}, nil
	}

	return result, nil
}

// recognizeIntent 使用 LLM 进行意图识别
func (m *SkillManager) recognizeIntent(ctx context.Context, input string) (string, map[string]string, error) {
	// 构建意图识别 Prompt
	prompt := m.buildIntentRecognitionPrompt(input)

	response, err := m.llmService.GenerateResponse(ctx, prompt, nil)
	if err != nil {
		return "", nil, err
	}

	// 解析 LLM 返回的 JSON
	var result struct {
		Intent   string            `json:"intent"`
		Entities map[string]string `json:"entities"`
	}

	if err := json.Unmarshal([]byte(response), &result); err != nil {
		// 尝试从文本中提取
		return m.extractIntentFromText(response), nil, nil
	}

	return result.Intent, result.Entities, nil
}

// buildIntentRecognitionPrompt 构建意图识别 Prompt
func (m *SkillManager) buildIntentRecognitionPrompt(input string) string {
	skills := m.registry.GetSkillDescriptions()

	prompt := "你是一个意图识别助手。请分析用户的输入，识别意图和提取关键实体。\n\n"
	prompt += "可用技能列表：\n"
	for _, skill := range skills {
		prompt += fmt.Sprintf("- %s: %s\n", skill["name"], skill["description"])
	}

	prompt += "\n意图列表：\n"
	prompt += "- query_order: 查询订单\n"
	prompt += "- track_order: 跟踪订单\n"
	prompt += "- delivery_problem: 配送问题\n"
	prompt += "- query_income: 查询收入\n"
	prompt += "- withdraw: 提现\n"
	prompt += "- general: 通用问题\n"

	prompt += "\n请以下面 JSON 格式返回结果：\n"
	prompt += `{"intent": "识别的意图", "entities": {"key": "value"}}` + "\n\n"
	prompt += "用户输入：" + input + "\n"
	prompt += "识别结果："

	return prompt
}

// fallbackIntentRecognition 降级意图识别（关键词匹配）
func (m *SkillManager) fallbackIntentRecognition(input string) (string, map[string]string) {
	input = strings.ToLower(input)
	entities := make(map[string]string)

	// 订单相关
	if strings.Contains(input, "订单") {
		if strings.Contains(input, "查") || strings.Contains(input, "看") {
			// 尝试提取订单号
			orderID := extractOrderID(input)
			if orderID != "" {
				entities["order_id"] = orderID
				return "query_order", entities
			}
			return "track_order", entities
		}
		if strings.Contains(input, "问题") || strings.Contains(input, "异常") {
			return "delivery_problem", entities
		}
	}

	// 收入相关
	if strings.Contains(input, "收入") || strings.Contains(input, "钱") || strings.Contains(input, "赚") {
		if strings.Contains(input, "今天") || strings.Contains(input, "今日") {
			entities["time"] = "today"
		}
		return "query_income", entities
	}

	if strings.Contains(input, "提现") {
		return "withdraw", entities
	}

	return "general", entities
}

// extractOrderID 从文本中提取订单号
func extractOrderID(input string) string {
	// 简单的数字提取逻辑
	parts := strings.Fields(input)
	for _, part := range parts {
		// 检查是否为纯数字
		isNumber := true
		for _, ch := range part {
			if ch < '0' || ch > '9' {
				isNumber = false
				break
			}
		}
		if isNumber && len(part) > 3 {
			return part
		}
	}
	return ""
}

// extractIntentFromText 从文本中提取意图
func (m *SkillManager) extractIntentFromText(text string) string {
	text = strings.ToLower(text)
	if strings.Contains(text, "query_order") {
		return "query_order"
	}
	if strings.Contains(text, "track_order") {
		return "track_order"
	}
	if strings.Contains(text, "income") {
		return "query_income"
	}
	if strings.Contains(text, "withdraw") {
		return "withdraw"
	}
	return "general"
}

// handleGeneralQuery 处理通用查询（没有匹配到 Skill）
func (m *SkillManager) handleGeneralQuery(ctx context.Context, input string, contextInfo map[string]interface{}) (*ai.SkillResult, error) {
	// 使用 LLM 生成通用回复
	prompt := "用户问题：" + input + "\n\n"
	prompt += "你是一个专业的骑手客服助手。请回答用户的问题。如果问题涉及订单或收入，请引导用户使用相应的功能。"

	response, err := m.llmService.GenerateResponse(ctx, prompt, nil)
	if err != nil {
		return &ai.SkillResult{
			Success:  false,
			Response: "抱歉，我暂时无法处理这个问题",
		}, nil
	}

	return &ai.SkillResult{
		Success:  true,
		Response: response,
		SuggestedActions: []ai.Action{
			{Type: "query_order", Name: "查询订单", Params: `{}`},
			{Type: "query_income", Name: "查询收入", Params: `{}`},
			{Type: "contact_human", Name: "人工客服", Params: `{}`},
		},
	}, nil
}

// GetAllSkills 获取所有技能信息
func (m *SkillManager) GetAllSkills() []map[string]string {
	return m.registry.GetSkillDescriptions()
}
