package ai

import (
	"context"
	"encoding/json"
)

// Skill 技能接口定义
// 每个 Skill 都是一个独立的业务功能模块，可以处理特定类型的用户请求
type Skill interface {
	// Name 返回技能名称
	Name() string
	// Description 返回技能描述
	Description() string
	// CanHandle 判断是否可以处理该请求
	CanHandle(ctx context.Context, intent string, entities map[string]string) bool
	// Execute 执行技能
	Execute(ctx context.Context, params SkillParams) (*SkillResult, error)
	// GetExamples 获取示例用法（用于Few-shot学习）
	GetExamples() []SkillExample
}

// SkillParams 技能执行参数
type SkillParams struct {
	// 用户原始输入
	RawInput string
	// 解析后的意图
	Intent string
	// 提取的实体
	Entities map[string]string
	// 上下文信息
	Context map[string]interface{}
	// 骑手ID
	RiderID int64
}

// SkillResult 技能执行结果
type SkillResult struct {
	// 是否成功
	Success bool `json:"success"`
	// 回复内容
	Response string `json:"response"`
	// 数据结果
	Data interface{} `json:"data,omitempty"`
	// 建议的操作
	SuggestedActions []Action `json:"suggested_actions,omitempty"`
	// 需要进一步确认的信息
	NeedConfirm bool `json:"need_confirm,omitempty"`
	// 确认提示
	ConfirmPrompt string `json:"confirm_prompt,omitempty"`
}

// Action 建议的操作
type Action struct {
	Type   string `json:"type"`
	Name   string `json:"name"`
	Params string `json:"params"` // JSON格式
}

// SkillExample 技能示例
type SkillExample struct {
	Input    string `json:"input"`
	Intent   string `json:"intent"`
	Entities map[string]string `json:"entities"`
	Response string `json:"response"`
}

// SkillRegistry 技能注册表
type SkillRegistry struct {
	skills []Skill
}

// NewSkillRegistry 创建技能注册表
func NewSkillRegistry() *SkillRegistry {
	return &SkillRegistry{
		skills: make([]Skill, 0),
	}
}

// Register 注册技能
func (r *SkillRegistry) Register(skill Skill) {
	r.skills = append(r.skills, skill)
}

// FindSkill 查找可以处理请求的技能
func (r *SkillRegistry) FindSkill(ctx context.Context, intent string, entities map[string]string) Skill {
	for _, skill := range r.skills {
		if skill.CanHandle(ctx, intent, entities) {
			return skill
		}
	}
	return nil
}

// GetAllSkills 获取所有技能
func (r *SkillRegistry) GetAllSkills() []Skill {
	return r.skills
}

// GetSkillDescriptions 获取所有技能描述（用于Prompt）
func (r *SkillRegistry) GetSkillDescriptions() []map[string]string {
	var descriptions []map[string]string
	for _, skill := range r.skills {
		descriptions = append(descriptions, map[string]string{
			"name":        skill.Name(),
			"description": skill.Description(),
		})
	}
	return descriptions
}

// MarshalResult 将结果序列化为JSON
func (r *SkillResult) MarshalResult() string {
	data, _ := json.Marshal(r)
	return string(data)
}
