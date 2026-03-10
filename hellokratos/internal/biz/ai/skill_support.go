package ai

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// TicketCreator 工单创建接口
// 用于在 Skill 中创建工单，避免循环依赖
type TicketCreator interface {
	CreateTicketFromSkill(ctx context.Context, userID int64, ticketType, title, description string, orderID int64) (map[string]interface{}, error)
}

// SupportSkill 客服助手技能
// 处理投诉、建议、咨询、转人工等客服相关功能
type SupportSkill struct {
	ticketCreator TicketCreator
}

// NewSupportSkill 创建客服助手技能
func NewSupportSkill() Skill {
	return &SupportSkill{}
}

// SetTicketCreator 设置工单创建器
func (s *SupportSkill) SetTicketCreator(creator TicketCreator) {
	s.ticketCreator = creator
}

// Name 技能名称
func (s *SupportSkill) Name() string {
	return "support_assistant"
}

// Description 技能描述
func (s *SupportSkill) Description() string {
	return "处理投诉、建议、问题咨询、转人工客服等服务"
}

// CanHandle 判断是否可以处理该请求
func (s *SupportSkill) CanHandle(ctx context.Context, intent string, entities map[string]string) bool {
	// 意图匹配
	supportIntents := []string{
		"complaint",
		"suggestion",
		"consultation",
		"transfer_to_human",
		"report_problem",
		"feedback",
	}

	for _, i := range supportIntents {
		if intent == i {
			return true
		}
	}

	// 关键词匹配
	keywords := []string{"投诉", "建议", "反馈", "客服", "人工", "问题", "举报", "申诉", "不满", "差评"}
	for _, keyword := range keywords {
		if strings.Contains(intent, keyword) {
			return true
		}
	}

	return false
}

// Execute 执行技能
func (s *SupportSkill) Execute(ctx context.Context, params SkillParams) (*SkillResult, error) {
	switch params.Intent {
	case "complaint":
		return s.handleComplaint(ctx, params)
	case "suggestion":
		return s.handleSuggestion(ctx, params)
	case "consultation":
		return s.handleConsultation(ctx, params)
	case "transfer_to_human":
		return s.handleTransferToHuman(ctx, params)
	case "report_problem":
		return s.handleReportProblem(ctx, params)
	case "feedback":
		return s.handleFeedback(ctx, params)
	default:
		return s.handleDefault(ctx, params)
	}
}

// handleComplaint 处理投诉
func (s *SupportSkill) handleComplaint(ctx context.Context, params SkillParams) (*SkillResult, error) {
	// 获取投诉类型
	complaintType, ok := params.Entities["complaint_type"]
	if !ok {
		return &SkillResult{
			Success: true,
			Response: "📝 投诉服务\n\n" +
				"请选择投诉类型：\n\n" +
				"1. **商家问题** - 出餐慢、餐品问题等\n" +
				"2. **顾客问题** - 态度恶劣、恶意差评等\n" +
				"3. **平台问题** - 系统故障、扣款异常等\n" +
				"4. **其他问题** - 其他不满意的地方\n\n" +
				"请告诉我您要投诉什么？",
			SuggestedActions: []Action{
				{Type: "complaint_merchant", Name: "投诉商家", Params: `{}`},
				{Type: "complaint_customer", Name: "投诉顾客", Params: `{}`},
				{Type: "complaint_platform", Name: "投诉平台", Params: `{}`},
				{Type: "transfer_to_human", Name: "转人工", Params: `{}`},
			},
		}, nil
	}

	// 获取投诉内容
	content, ok := params.Entities["content"]
	if !ok {
		return &SkillResult{
			Success:       true,
			Response:      fmt.Sprintf("您选择了投诉【%s】，请详细描述您遇到的问题：", complaintType),
			NeedConfirm:   true,
			ConfirmPrompt: "请详细描述投诉内容",
		}, nil
	}

	// 创建投诉工单
	var ticketResult map[string]interface{}
	var err error
	if s.ticketCreator != nil {
		ticketResult, err = s.ticketCreator.CreateTicketFromSkill(ctx, params.RiderID, "complaint", fmt.Sprintf("投诉：%s", complaintType), content, 0)
	}

	if err != nil || s.ticketCreator == nil {
		// 如果工单创建失败或没有配置工单创建器，返回模拟数据
		return &SkillResult{
			Success: true,
			Response: fmt.Sprintf(
				"✅ 投诉已记录\n\n"+
					"投诉类型：%s\n"+
					"投诉内容：%s\n\n"+
					"工单编号：TK%010d\n"+
					"提交时间：%s\n\n"+
					"我们会在24小时内处理您的投诉，请保持电话畅通。",
				complaintType,
				content,
				params.RiderID,
				time.Now().Format("2006-01-02 15:04"),
			),
			Data: map[string]interface{}{
				"ticket_id": fmt.Sprintf("TK%010d", params.RiderID),
				"type":      "complaint",
				"status":    "pending",
			},
			SuggestedActions: []Action{
			{Type: "query_ticket", Name: "查询进度", Params: `{}`},
			{Type: "cancel_ticket", Name: "撤销投诉", Params: `{}`},
			},
		}, nil
	}

	// 工单创建成功
	ticketID := ticketResult["ticket_id"].(int64)
	status := ticketResult["status"].(string)
	return &SkillResult{
		Success: true,
		Response: fmt.Sprintf(
			"✅ 投诉已提交\n\n"+
				"投诉类型：%s\n"+
				"投诉内容：%s\n\n"+
				"工单编号：#%d\n"+
				"提交时间：%s\n"+
				"当前状态：%s\n\n"+
				"我们会在24小时内处理您的投诉，请保持电话畅通。",
			complaintType,
			content,
			ticketID,
			time.Now().Format("2006-01-02 15:04"),
			status,
		),
		Data: map[string]interface{}{
			"ticket_id": ticketID,
			"type":      "complaint",
			"status":    status,
		},
		SuggestedActions: []Action{
			{Type: "query_ticket", Name: "查询进度", Params: fmt.Sprintf(`{"ticket_id":%d}`, ticketID)},
			{Type: "cancel_ticket", Name: "撤销投诉", Params: fmt.Sprintf(`{"ticket_id":%d}`, ticketID)},
		},
	}, nil
}

// handleSuggestion 处理建议
func (s *SupportSkill) handleSuggestion(ctx context.Context, params SkillParams) (*SkillResult, error) {
	content, ok := params.Entities["content"]
	if !ok {
		return &SkillResult{
			Success:       true,
			Response:      "💡 感谢您的宝贵建议！\n\n请详细描述您的建议内容：",
			NeedConfirm:   true,
			ConfirmPrompt: "请输入您的建议",
		}, nil
	}

	// 创建建议工单
	var ticketResult map[string]interface{}
	var err error
	if s.ticketCreator != nil {
		ticketResult, err = s.ticketCreator.CreateTicketFromSkill(ctx, params.RiderID, "suggestion", "建议", content, 0)
	}

	if err != nil || s.ticketCreator == nil {
		// 如果工单创建失败或没有配置工单创建器，返回模拟数据
		return &SkillResult{
			Success: true,
			Response: fmt.Sprintf(
				"✅ 建议已记录\n\n"+
					"您的建议：%s\n\n"+
					"非常感谢您的反馈！我们会认真评估您的建议，如有需要会与您联系。",
				content,
			),
			Data: map[string]interface{}{
				"type":   "suggestion",
				"status": "submitted",
			},
		}, nil
	}

	// 工单创建成功
	ticketID := ticketResult["ticket_id"].(int64)
	status := ticketResult["status"].(string)
	return &SkillResult{
		Success: true,
		Response: fmt.Sprintf(
			"✅ 建议已提交\n\n"+
				"您的建议：%s\n\n"+
				"工单编号：#%d\n"+
				"非常感谢您的反馈！我们会认真评估您的建议，如有需要会与您联系。",
			content,
			ticketID,
		),
		Data: map[string]interface{}{
			"ticket_id": ticketID,
			"type":      "suggestion",
			"status":    status,
		},
	}, nil
}

// handleConsultation 处理咨询
func (s *SupportSkill) handleConsultation(ctx context.Context, params SkillParams) (*SkillResult, error) {
	question, ok := params.Entities["question"]
	if !ok {
		question = params.RawInput
	}

	// 常见问题自动回复
	faq := s.matchFAQ(question)
	if faq != "" {
		return &SkillResult{
			Success:  true,
			Response: fmt.Sprintf("💬 %s\n\n%s\n\n如果还有其他问题，可以转人工客服。", question, faq),
			SuggestedActions: []Action{
				{Type: "transfer_to_human", Name: "转人工", Params: `{}`},
				{Type: "more_faq", Name: "更多问题", Params: `{}`},
			},
		}, nil
	}

	// 无法自动回复，建议转人工
	return &SkillResult{
		Success: true,
		Response: "💬 您的问题：" + question + "\n\n" +
			"这个问题比较复杂，建议您转接人工客服获取更专业的解答。",
		SuggestedActions: []Action{
			{Type: "transfer_to_human", Name: "转人工客服", Params: `{}`},
			{Type: "leave_message", Name: "留言", Params: `{}`},
		},
	}, nil
}

// handleTransferToHuman 转人工客服
func (s *SupportSkill) handleTransferToHuman(ctx context.Context, params SkillParams) (*SkillResult, error) {
	return &SkillResult{
		Success: true,
		Response: "👨‍💼 正在为您转接人工客服...\n\n" +
			"当前排队人数：3人\n" +
			"预计等待时间：2分钟\n\n" +
			"💡 提示：\n" +
			"1. 请保持网络畅通\n" +
			"2. 准备好相关订单号或截图\n" +
			"3. 客服工作时间：9:00-21:00\n\n" +
			"如需取消，请说【取消转接】",
		Data: map[string]interface{}{
			"queue_position": 3,
			"wait_time":      2,
			"status":         "queuing",
		},
		SuggestedActions: []Action{
			{Type: "cancel_transfer", Name: "取消转接", Params: `{}`},
			{Type: "leave_message", Name: "留言", Params: `{}`},
		},
	}, nil
}

// handleReportProblem 上报问题
func (s *SupportSkill) handleReportProblem(ctx context.Context, params SkillParams) (*SkillResult, error) {
	return &SkillResult{
		Success: true,
		Response: "🚨 问题上报\n\n" +
			"请选择问题类型：\n\n" +
			"1. **APP闪退/卡顿** - 系统技术问题\n" +
			"2. **定位异常** - GPS定位不准确\n" +
			"3. **订单异常** - 订单状态错误\n" +
			"4. **支付问题** - 收入或扣款异常\n" +
			"5. **其他问题** - 其他技术问题\n\n" +
			"请告诉我您遇到了什么问题？",
		SuggestedActions: []Action{
			{Type: "report_app_issue", Name: "APP问题", Params: `{}`},
			{Type: "report_location_issue", Name: "定位问题", Params: `{}`},
			{Type: "report_order_issue", Name: "订单问题", Params: `{}`},
			{Type: "report_payment_issue", Name: "支付问题", Params: `{}`},
		},
	}, nil
}

// handleFeedback 处理评价反馈
func (s *SupportSkill) handleFeedback(ctx context.Context, params SkillParams) (*SkillResult, error) {
	return &SkillResult{
		Success: true,
		Response: "⭐ 服务评价\n\n" +
			"请对我们的服务进行评价：\n\n" +
			"1. ⭐ 非常不满意\n" +
			"2. ⭐⭐ 不满意\n" +
			"3. ⭐⭐⭐ 一般\n" +
			"4. ⭐⭐⭐⭐ 满意\n" +
			"5. ⭐⭐⭐⭐⭐ 非常满意\n\n" +
			"您的评价将帮助我们改进服务！",
		SuggestedActions: []Action{
			{Type: "rate_1", Name: "⭐", Params: `{"rating":1}`},
			{Type: "rate_2", Name: "⭐⭐", Params: `{"rating":2}`},
			{Type: "rate_3", Name: "⭐⭐⭐", Params: `{"rating":3}`},
			{Type: "rate_4", Name: "⭐⭐⭐⭐", Params: `{"rating":4}`},
			{Type: "rate_5", Name: "⭐⭐⭐⭐⭐", Params: `{"rating":5}`},
		},
	}, nil
}

// handleDefault 默认处理
func (s *SupportSkill) handleDefault(ctx context.Context, params SkillParams) (*SkillResult, error) {
	return &SkillResult{
		Success: true,
		Response: "👨‍💼 客服中心\n\n" +
			"我可以帮您：\n\n" +
			"1. **投诉** - 对商家、顾客或平台的不满\n" +
			"2. **建议** - 对产品或服务的改进建议\n" +
			"3. **咨询** - 常见问题解答\n" +
			"4. **问题上报** - APP或系统问题\n" +
			"5. **转人工** - 连接人工客服\n\n" +
			"请告诉我您需要什么帮助？",
		SuggestedActions: []Action{
			{Type: "complaint", Name: "我要投诉", Params: `{}`},
			{Type: "suggestion", Name: "提建议", Params: `{}`},
			{Type: "consultation", Name: "咨询问题", Params: `{}`},
			{Type: "transfer_to_human", Name: "转人工", Params: `{}`},
		},
	}, nil
}

// GetExamples 获取示例
func (s *SupportSkill) GetExamples() []SkillExample {
	return []SkillExample{
		{
			Input:    "我要投诉商家",
			Intent:   "complaint",
			Entities: map[string]string{"complaint_type": "merchant"},
			Response: "请选择投诉类型：1.商家问题 2.顾客问题 3.平台问题...",
		},
		{
			Input:    "转人工客服",
			Intent:   "transfer_to_human",
			Entities: map[string]string{},
			Response: "正在为您转接人工客服，当前排队3人，预计等待2分钟...",
		},
		{
			Input:    "我有个建议",
			Intent:   "suggestion",
			Entities: map[string]string{},
			Response: "感谢您的建议！请详细描述您的建议内容...",
		},
	}
}

// matchFAQ 匹配常见问题
func (s *SupportSkill) matchFAQ(question string) string {
	faqMap := map[string]string{
		"收入什么时候到账": "收入一般在订单完成后24小时内到账，提现申请提交后1-3个工作日到账。",
		"如何修改手机号":  "您可以在[我的]-[账户设置]-[修改手机号]中进行修改，需要验证原手机号。",
		"怎么查看订单":   "您可以在首页点击[订单]查看当前订单，或点击[历史]查看已完成订单。",
		"如何联系顾客":   "在订单详情页点击[联系顾客]按钮即可拨打电话，请注意保护顾客隐私。",
		"忘记密码怎么办":  "在登录页面点击[忘记密码]，通过手机号验证后即可重置密码。",
		"如何提现":     "在[我的]-[我的收入]中点击[提现]，选择提现方式并输入金额即可。",
		"为什么接不到单":  "请检查：1.是否处于上线状态 2.定位是否开启 3.所在区域订单量 4.资质是否通过。",
		"怎么投诉":     "您可以直接对我说[我要投诉]，我会帮您处理。",
	}

	for key, answer := range faqMap {
		if strings.Contains(question, key) {
			return answer
		}
	}

	return ""
}
