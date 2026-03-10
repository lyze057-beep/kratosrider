package ai

import (
	"context"
	"fmt"
	"testing"
)

func TestSupportSkill_Complaint(t *testing.T) {
	skill := NewSupportSkill()

	params := SkillParams{
		Intent: "complaint",
		Entities: map[string]string{
			"complaint_type": "merchant",
			"content":        "商家出餐太慢，已经等了40分钟",
		},
		RiderID: 123456,
	}

	ctx := context.Background()
	result, err := skill.Execute(ctx, params)

	if err != nil {
		t.Errorf("执行失败: %v", err)
		return
	}

	if !result.Success {
		t.Errorf("执行不成功: %s", result.Response)
		return
	}

	fmt.Println("投诉工单测试通过")
	fmt.Println("响应:", result.Response)

	if result.Data == nil {
		t.Error("缺少工单数据")
	}
}

func TestSupportSkill_Suggestion(t *testing.T) {
	skill := NewSupportSkill()

	params := SkillParams{
		Intent: "suggestion",
		Entities: map[string]string{
			"content": "希望增加夜间配送补贴",
		},
		RiderID: 123456,
	}

	ctx := context.Background()
	result, err := skill.Execute(ctx, params)

	if err != nil {
		t.Errorf("执行失败: %v", err)
		return
	}

	if !result.Success {
		t.Errorf("执行不成功: %s", result.Response)
		return
	}

	fmt.Println("建议工单测试通过")
	fmt.Println("响应:", result.Response)
}

func TestSupportSkill_TransferToHuman(t *testing.T) {
	skill := NewSupportSkill()

	params := SkillParams{
		Intent: "transfer_to_human",
	}

	ctx := context.Background()
	result, err := skill.Execute(ctx, params)

	if err != nil {
		t.Errorf("执行失败: %v", err)
		return
	}

	if !result.Success {
		t.Errorf("执行不成功: %s", result.Response)
		return
	}

	fmt.Println("转人工客服测试通过")
	fmt.Println("响应:", result.Response)
}

func TestSupportSkill_Default(t *testing.T) {
	skill := NewSupportSkill()

	params := SkillParams{
		Intent: "default",
	}

	ctx := context.Background()
	result, err := skill.Execute(ctx, params)

	if err != nil {
		t.Errorf("执行失败: %v", err)
		return
	}

	if !result.Success {
		t.Errorf("执行不成功: %s", result.Response)
		return
	}

	fmt.Println("客服中心默认测试通过")
	fmt.Println("响应:", result.Response)
}
