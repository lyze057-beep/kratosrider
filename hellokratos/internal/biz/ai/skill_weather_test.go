package ai

import (
	"context"
	"fmt"
	"hellokratos/internal/conf"
	"testing"
)

func TestWeatherSkill_RealTimeWeather(t *testing.T) {
	cfg := &conf.Data{}
	skill := NewWeatherSkill(cfg)

	params := SkillParams{
		Intent: "query_weather",
		Entities: map[string]string{
			"location": "北京",
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

	fmt.Println("实时天气测试通过")
	fmt.Println("响应:", result.Response)

	if result.Data == nil {
		t.Error("缺少天气数据")
	}
}

func TestWeatherSkill_WeatherForecast(t *testing.T) {
	cfg := &conf.Data{}
	skill := NewWeatherSkill(cfg)

	params := SkillParams{
		Intent: "weather_forecast",
		Entities: map[string]string{
			"location": "北京",
			"days":     "3",
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

	fmt.Println("天气预报测试通过")
	fmt.Println("响应:", result.Response)
}

func TestWeatherSkill_DeliveryWeather(t *testing.T) {
	cfg := &conf.Data{}
	skill := NewWeatherSkill(cfg)

	params := SkillParams{
		Intent: "delivery_weather",
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

	fmt.Println("配送天气测试通过")
	fmt.Println("响应:", result.Response)
}

func TestWeatherSkill_WeatherAlert(t *testing.T) {
	cfg := &conf.Data{}
	skill := NewWeatherSkill(cfg)

	params := SkillParams{
		Intent: "weather_alert",
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

	fmt.Println("天气预警测试通过")
	fmt.Println("响应:", result.Response)
}
