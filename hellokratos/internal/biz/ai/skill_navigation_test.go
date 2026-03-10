package ai

import (
	"context"
	"fmt"
	"hellokratos/internal/conf"
	"testing"
)

func TestNavigationSkill_RoutePlanning(t *testing.T) {
	// 创建测试配置
	cfg := &conf.Data{}
	skill := NewNavigationSkill(cfg)

	// 测试路线规划
	params := SkillParams{
		Intent: "route_planning",
		Entities: map[string]string{
			"origin":      "北京天安门",
			"destination": "北京故宫",
		},
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

	fmt.Println("路线规划测试通过")
	fmt.Println("响应:", result.Response)

	if result.Data == nil {
		t.Error("缺少路线数据")
	}
}

func TestNavigationSkill_NearbySearch(t *testing.T) {
	cfg := &conf.Data{}
	skill := NewNavigationSkill(cfg)

	params := SkillParams{
		Intent: "nearby_search",
		Entities: map[string]string{
			"keyword": "加油站",
		},
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

	fmt.Println("附近搜索测试通过")
	fmt.Println("响应:", result.Response)
}

func TestNavigationSkill_CalculateDistance(t *testing.T) {
	cfg := &conf.Data{}
	skill := NewNavigationSkill(cfg)

	params := SkillParams{
		Intent: "calculate_distance",
		Entities: map[string]string{
			"origin":      "北京天安门",
			"destination": "北京西站",
		},
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

	fmt.Println("距离计算测试通过")
	fmt.Println("响应:", result.Response)
}

func TestNavigationSkill_EstimateTime(t *testing.T) {
	cfg := &conf.Data{}
	skill := NewNavigationSkill(cfg)

	params := SkillParams{
		Intent: "estimate_time",
		Entities: map[string]string{
			"destination": "北京首都国际机场",
		},
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

	fmt.Println("预计时间测试通过")
	fmt.Println("响应:", result.Response)
}
