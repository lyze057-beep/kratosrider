package ai

import (
	"context"
	"fmt"
	"strings"
	"time"

	"hellokratos/internal/data"
)

// IncomeSkill 收入助手技能
// 处理所有与收入相关的查询和统计
type IncomeSkill struct {
	incomeRepo data.IncomeRepo
}

// NewIncomeSkill 创建收入助手技能
func NewIncomeSkill(incomeRepo data.IncomeRepo) Skill {
	return &IncomeSkill{incomeRepo: incomeRepo}
}

// Name 技能名称
func (s *IncomeSkill) Name() string {
	return "income_assistant"
}

// Description 技能描述
func (s *IncomeSkill) Description() string {
	return "处理收入查询、收入统计、提现咨询等收入相关业务"
}

// CanHandle 判断是否可以处理该请求
func (s *IncomeSkill) CanHandle(ctx context.Context, intent string, entities map[string]string) bool {
	// 意图匹配
	incomeIntents := []string{
		"query_income",
		"today_income",
		"month_income",
		"income_stats",
		"withdraw",
		"income_detail",
	}

	for _, i := range incomeIntents {
		if intent == i {
			return true
		}
	}

	// 关键词匹配
	keywords := []string{"收入", "工资", "提现", "结算", " earnings", "reward", "bonus"}
	for _, keyword := range keywords {
		if strings.Contains(intent, keyword) {
			return true
		}
	}

	return false
}

// Execute 执行技能
func (s *IncomeSkill) Execute(ctx context.Context, params SkillParams) (*SkillResult, error) {
	switch params.Intent {
	case "query_income", "today_income":
		return s.handleQueryTodayIncome(ctx, params)
	case "month_income":
		return s.handleQueryMonthIncome(ctx, params)
	case "income_stats":
		return s.handleIncomeStats(ctx, params)
	case "withdraw":
		return s.handleWithdraw(ctx, params)
	default:
		return s.handleDefault(ctx, params)
	}
}

// handleQueryTodayIncome 查询今日收入
func (s *IncomeSkill) handleQueryTodayIncome(ctx context.Context, params SkillParams) (*SkillResult, error) {
	riderID := params.RiderID

	// 获取总收入
	totalIncome, err := s.incomeRepo.GetTotalIncomeByRiderID(ctx, riderID)
	if err != nil {
		return &SkillResult{
			Success:  false,
			Response: fmt.Sprintf("查询收入失败：%v", err),
		}, nil
	}

	// 获取最近收入记录
	incomes, err := s.incomeRepo.GetIncomesByRiderID(ctx, riderID, 5)
	if err != nil {
		return &SkillResult{
			Success:  false,
			Response: fmt.Sprintf("查询收入明细失败：%v", err),
		}, nil
	}

	// 计算今日收入（简化处理，实际应该按日期筛选）
	todayIncome := 0.0
	for _, income := range incomes {
		if income.CreatedAt.Format("2006-01-02") == time.Now().Format("2006-01-02") {
			todayIncome += income.Amount
		}
	}

	response := fmt.Sprintf(
		"📊 您的收入情况：\n\n"+
			"💰 累计总收入：¥%.2f\n"+
			"📅 今日收入：¥%.2f\n"+
			"📋 最近收入记录（5笔）：\n",
		totalIncome,
		todayIncome,
	)

	for i, income := range incomes {
		response += fmt.Sprintf("%d. %s +¥%.2f\n", i+1, income.CreatedAt.Format("01-02 15:04"), income.Amount)
	}

	response += "\n💡 提示：收入会在订单完成后结算，具体到账时间取决于结算周期。"

	return &SkillResult{
		Success:  true,
		Response: response,
		Data: map[string]interface{}{
			"total_income":   totalIncome,
			"today_income":   todayIncome,
			"recent_records": incomes,
		},
		SuggestedActions: []Action{
			{Type: "view_income_detail", Name: "查看明细", Params: `{}`},
			{Type: "withdraw", Name: "提现", Params: `{}`},
			{Type: "income_stats", Name: "收入统计", Params: `{}`},
		},
	}, nil
}

// handleQueryMonthIncome 查询本月收入
func (s *IncomeSkill) handleQueryMonthIncome(ctx context.Context, params SkillParams) (*SkillResult, error) {
	// 这里可以实现更复杂的月度统计逻辑
	return &SkillResult{
		Success:  true,
		Response: "📊 本月收入统计功能开发中...\n\n您可以通过以下方式查看：\n1. 打开骑手APP\n2. 进入[我的]-[收入明细]\n3. 选择本月查看详细统计",
		SuggestedActions: []Action{
			{Type: "open_app", Name: "打开APP", Params: `{}`},
		},
	}, nil
}

// handleIncomeStats 收入统计
func (s *IncomeSkill) handleIncomeStats(ctx context.Context, params SkillParams) (*SkillResult, error) {
	return &SkillResult{
		Success: true,
		Response: "📈 收入统计功能：\n\n" +
			"1. **日统计** - 查看每日收入趋势\n" +
			"2. **周统计** - 查看本周收入汇总\n" +
			"3. **月统计** - 查看月度收入报表\n" +
			"4. **收入来源分析** - 配送费、奖励、补贴占比\n\n" +
			"请告诉我您想查看哪种统计？",
		SuggestedActions: []Action{
			{Type: "daily_stats", Name: "日统计", Params: `{}`},
			{Type: "weekly_stats", Name: "周统计", Params: `{}`},
			{Type: "monthly_stats", Name: "月统计", Params: `{}`},
		},
	}, nil
}

// handleWithdraw 提现
func (s *IncomeSkill) handleWithdraw(ctx context.Context, params SkillParams) (*SkillResult, error) {
	return &SkillResult{
		Success: true,
		Response: "💳 提现功能：\n\n" +
			"提现流程：\n" +
			"1. 确认可提现金额\n" +
			"2. 选择提现方式（银行卡/支付宝）\n" +
			"3. 确认提现金额\n" +
			"4. 等待到账（通常1-3个工作日）\n\n" +
			"⚠️ 注意：\n" +
			"- 单笔提现最低10元\n" +
			"- 每日最多提现3次\n" +
			"- 节假日可能延迟到账",
		SuggestedActions: []Action{
			{Type: "withdraw_now", Name: "立即提现", Params: `{}`},
			{Type: "withdraw_history", Name: "提现记录", Params: `{}`},
			{Type: "bind_account", Name: "绑定账户", Params: `{}`},
		},
	}, nil
}

// handleDefault 默认处理
func (s *IncomeSkill) handleDefault(ctx context.Context, params SkillParams) (*SkillResult, error) {
	return &SkillResult{
		Success: true,
		Response: "💰 我可以帮您处理以下收入相关事务：\n\n" +
			"1. 查询今日/累计收入\n" +
			"2. 查看收入明细\n" +
			"3. 收入统计分析\n" +
			"4. 提现操作\n" +
			"5. 收入问题咨询\n\n" +
			"请告诉我您想了解什么？",
		SuggestedActions: []Action{
			{Type: "query_income", Name: "查询收入", Params: `{}`},
			{Type: "income_detail", Name: "收入明细", Params: `{}`},
			{Type: "withdraw", Name: "提现", Params: `{}`},
		},
	}, nil
}

// GetExamples 获取示例
func (s *IncomeSkill) GetExamples() []SkillExample {
	return []SkillExample{
		{
			Input:    "我今天赚了多少钱",
			Intent:   "today_income",
			Entities: map[string]string{},
			Response: "您今日收入为 ¥156.50，已完成8单配送...",
		},
		{
			Input:    "我想提现",
			Intent:   "withdraw",
			Entities: map[string]string{},
			Response: "您的可提现金额为 ¥1,250.00，请选择提现方式...",
		},
		{
			Input:    "查看收入统计",
			Intent:   "income_stats",
			Entities: map[string]string{},
			Response: "本月您已完成156单，总收入 ¥5,280.00...",
		},
	}
}
