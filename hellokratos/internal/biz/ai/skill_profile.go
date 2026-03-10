package ai

import (
	"context"
	"fmt"
	"strings"

	"hellokratos/internal/data"
)

// ProfileSkill 骑手信息助手技能
// 处理骑手个人信息查询、资质状态、账户设置等
type ProfileSkill struct {
	authRepo          data.AuthRepo
	qualificationRepo data.QualificationRepo
}

// NewProfileSkill 创建骑手信息助手技能
func NewProfileSkill(authRepo data.AuthRepo, qualificationRepo data.QualificationRepo) Skill {
	return &ProfileSkill{
		authRepo:          authRepo,
		qualificationRepo: qualificationRepo,
	}
}

// Name 技能名称
func (s *ProfileSkill) Name() string {
	return "profile_assistant"
}

// Description 技能描述
func (s *ProfileSkill) Description() string {
	return "处理骑手个人信息查询、资质认证状态、账户设置等"
}

// CanHandle 判断是否可以处理该请求
func (s *ProfileSkill) CanHandle(ctx context.Context, intent string, entities map[string]string) bool {
	// 意图匹配
	profileIntents := []string{
		"query_profile",
		"query_qualification",
		"update_profile",
		"account_settings",
	}

	for _, i := range profileIntents {
		if intent == i {
			return true
		}
	}

	// 关键词匹配
	keywords := []string{"信息", "资料", "资质", "认证", "身份证", "驾驶证", "健康证", "账户", "设置"}
	for _, keyword := range keywords {
		if strings.Contains(intent, keyword) {
			return true
		}
	}

	return false
}

// Execute 执行技能
func (s *ProfileSkill) Execute(ctx context.Context, params SkillParams) (*SkillResult, error) {
	switch params.Intent {
	case "query_profile":
		return s.handleQueryProfile(ctx, params)
	case "query_qualification":
		return s.handleQueryQualification(ctx, params)
	case "update_profile":
		return s.handleUpdateProfile(ctx, params)
	case "account_settings":
		return s.handleAccountSettings(ctx, params)
	default:
		return s.handleDefault(ctx, params)
	}
}

// handleQueryProfile 查询个人信息
func (s *ProfileSkill) handleQueryProfile(ctx context.Context, params SkillParams) (*SkillResult, error) {
	riderID := params.RiderID

	// 查询用户信息
	user, err := s.authRepo.GetUserByID(ctx, riderID)
	if err != nil {
		return &SkillResult{
			Success:  false,
			Response: fmt.Sprintf("查询个人信息失败：%v", err),
		}, nil
	}

	// 构建回复
	response := fmt.Sprintf(
		"📋 您的个人信息：\n\n"+
			"👤 昵称：%s\n"+
			"📱 手机号：%s\n"+
			"🆔 用户ID：%d\n"+
			"📅 注册时间：%s",
		user.Nickname,
		maskPhone(user.Phone),
		user.ID,
		user.CreatedAt.Format("2006-01-02"),
	)

	return &SkillResult{
		Success:  true,
		Response: response,
		Data: map[string]interface{}{
			"nickname":      user.Nickname,
			"phone":         user.Phone,
			"register_time": user.CreatedAt,
		},
		SuggestedActions: []Action{
			{Type: "update_profile", Name: "修改信息", Params: `{}`},
			{Type: "query_qualification", Name: "查看资质", Params: `{}`},
			{Type: "change_password", Name: "修改密码", Params: `{}`},
		},
	}, nil
}

// handleQueryQualification 查询资质认证状态
func (s *ProfileSkill) handleQueryQualification(ctx context.Context, params SkillParams) (*SkillResult, error) {
	riderID := params.RiderID

	// 查询总体资质状态
	qual, err := s.qualificationRepo.GetQualificationByRiderID(ctx, riderID)
	if err != nil {
		return &SkillResult{
			Success:  false,
			Response: fmt.Sprintf("查询资质状态失败：%v", err),
		}, nil
	}

	if qual == nil {
		return &SkillResult{
			Success:  true,
			Response: "您尚未提交资质认证信息。请完成以下认证：\n\n1. 身份证认证\n2. 驾驶证认证\n3. 健康证认证\n\n认证通过后即可开始接单。",
			SuggestedActions: []Action{
				{Type: "start_qualification", Name: "开始认证", Params: `{}`},
			},
		}, nil
	}

	// 查询各项认证详情
	idCard, _ := s.qualificationRepo.GetIDCardVerification(ctx, riderID)
	driverLicense, _ := s.qualificationRepo.GetDriverLicenseVerification(ctx, riderID)
	healthCert, _ := s.qualificationRepo.GetHealthCertificateVerification(ctx, riderID)

	// 构建状态文本
	statusMap := map[int32]string{
		0: "⏳ 未提交",
		1: "🔄 审核中",
		2: "✅ 已通过",
		3: "❌ 已拒绝",
	}

	response := "📋 您的资质认证状态：\n\n"
	response += fmt.Sprintf("总体状态：%s\n\n", statusMap[qual.OverallStatus])

	if idCard != nil {
		response += fmt.Sprintf("🆔 身份证认证：%s\n", statusMap[idCard.Status])
		if idCard.Status == 3 && idCard.RejectReason != "" {
			response += fmt.Sprintf("   拒绝原因：%s\n", idCard.RejectReason)
		}
	}

	if driverLicense != nil {
		response += fmt.Sprintf("🚗 驾驶证认证：%s\n", statusMap[driverLicense.Status])
		if driverLicense.Status == 3 && driverLicense.RejectReason != "" {
			response += fmt.Sprintf("   拒绝原因：%s\n", driverLicense.RejectReason)
		}
	}

	if healthCert != nil {
		response += fmt.Sprintf("🏥 健康证认证：%s\n", statusMap[healthCert.Status])
		if healthCert.Status == 3 && healthCert.RejectReason != "" {
			response += fmt.Sprintf("   拒绝原因：%s\n", healthCert.RejectReason)
		}
	}

	// 根据状态添加提示
	if qual.OverallStatus == 2 {
		response += "\n🎉 恭喜！您的资质认证已通过，可以正常接单了。"
	} else if qual.OverallStatus == 1 {
		response += "\n⏳ 您的资质正在审核中，请耐心等待。"
	} else if qual.OverallStatus == 3 {
		response += "\n⚠️ 您的资质认证未通过，请根据拒绝原因修改后重新提交。"
	}

	return &SkillResult{
		Success:  true,
		Response: response,
		Data: map[string]interface{}{
			"overall_status": qual.OverallStatus,
			"id_card":        idCard,
			"driver_license": driverLicense,
			"health_cert":    healthCert,
		},
		SuggestedActions: []Action{
			{Type: "upload_idcard", Name: "上传身份证", Params: `{}`},
			{Type: "upload_license", Name: "上传驾驶证", Params: `{}`},
			{Type: "upload_health", Name: "上传健康证", Params: `{}`},
		},
	}, nil
}

// handleUpdateProfile 修改个人信息
func (s *ProfileSkill) handleUpdateProfile(ctx context.Context, params SkillParams) (*SkillResult, error) {
	return &SkillResult{
		Success:       true,
		Response:      "您想修改哪项信息？\n\n1. 修改姓名\n2. 修改手机号\n3. 修改邮箱\n4. 修改密码",
		NeedConfirm:   true,
		ConfirmPrompt: "请选择要修改的信息",
		SuggestedActions: []Action{
			{Type: "update_name", Name: "修改姓名", Params: `{}`},
			{Type: "update_phone", Name: "修改手机号", Params: `{}`},
			{Type: "update_email", Name: "修改邮箱", Params: `{}`},
			{Type: "update_password", Name: "修改密码", Params: `{}`},
		},
	}, nil
}

// handleAccountSettings 账户设置
func (s *ProfileSkill) handleAccountSettings(ctx context.Context, params SkillParams) (*SkillResult, error) {
	return &SkillResult{
		Success: true,
		Response: "⚙️ 账户设置：\n\n" +
			"1. 消息通知设置\n" +
			"2. 隐私设置\n" +
			"3. 安全设置\n" +
			"4. 注销账户\n\n" +
			"请告诉我您想设置哪项？",
		SuggestedActions: []Action{
			{Type: "notification_settings", Name: "通知设置", Params: `{}`},
			{Type: "privacy_settings", Name: "隐私设置", Params: `{}`},
			{Type: "security_settings", Name: "安全设置", Params: `{}`},
		},
	}, nil
}

// handleDefault 默认处理
func (s *ProfileSkill) handleDefault(ctx context.Context, params SkillParams) (*SkillResult, error) {
	return &SkillResult{
		Success: true,
		Response: "👤 我可以帮您处理以下个人信息相关事务：\n\n" +
			"1. 查询个人资料\n" +
			"2. 查看资质认证状态\n" +
			"3. 修改个人信息\n" +
			"4. 账户设置\n\n" +
			"请告诉我您想了解什么？",
		SuggestedActions: []Action{
			{Type: "query_profile", Name: "查看资料", Params: `{}`},
			{Type: "query_qualification", Name: "查看资质", Params: `{}`},
			{Type: "update_profile", Name: "修改信息", Params: `{}`},
		},
	}, nil
}

// GetExamples 获取示例
func (s *ProfileSkill) GetExamples() []SkillExample {
	return []SkillExample{
		{
			Input:    "查看我的信息",
			Intent:   "query_profile",
			Entities: map[string]string{},
			Response: "您的个人信息：姓名：张三，手机号：138****8888...",
		},
		{
			Input:    "我的资质认证通过了吗",
			Intent:   "query_qualification",
			Entities: map[string]string{},
			Response: "您的资质认证状态：身份证：已通过，驾驶证：审核中...",
		},
		{
			Input:    "我想修改手机号",
			Intent:   "update_profile",
			Entities: map[string]string{"field": "phone"},
			Response: "请输入新的手机号，我们将发送验证码进行验证...",
		},
	}
}

// maskPhone 隐藏手机号中间4位
func maskPhone(phone string) string {
	if len(phone) != 11 {
		return phone
	}
	return phone[:3] + "****" + phone[7:]
}
