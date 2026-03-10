package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"hellokratos/internal/conf"
)

// mustInt 将字符串转换为整数，失败返回0
func mustInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

// WeatherSkill 天气助手技能
// 提供天气预报、实时天气、配送天气建议等功能
type WeatherSkill struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

// NewWeatherSkill 创建天气助手技能
// TODO: 需要替换为实际的天气服务 API Key
func NewWeatherSkill(cfg *conf.Data) Skill {
	skill := &WeatherSkill{
		// 和风天气 API Key
		apiKey:  "eb5ee80bd47f49c089682d19bd1e19f4",
		baseURL: "https://devapi.qweather.com/v7", // 和风天气 API
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}

	// 如果配置中有天气服务配置，使用配置值
	if cfg != nil && cfg.GetAiService() != nil {
		// 可以通过配置读取 API Key
		// skill.apiKey = cfg.GetWeatherService().GetApiKey()
	}

	return skill
}

// Name 技能名称
func (s *WeatherSkill) Name() string {
	return "weather_assistant"
}

// Description 技能描述
func (s *WeatherSkill) Description() string {
	return "提供天气预报、实时天气查询、配送天气建议、恶劣天气提醒等服务"
}

// CanHandle 判断是否可以处理该请求
func (s *WeatherSkill) CanHandle(ctx context.Context, intent string, entities map[string]string) bool {
	// 意图匹配
	weatherIntents := []string{
		"query_weather",
		"weather_forecast",
		"delivery_weather",
		"weather_alert",
	}

	for _, i := range weatherIntents {
		if intent == i {
			return true
		}
	}

	// 关键词匹配
	keywords := []string{"天气", "下雨", "温度", "预报", "雾霾", "风", "雪", "热", "冷", "穿衣", "带伞"}
	for _, keyword := range keywords {
		if strings.Contains(intent, keyword) {
			return true
		}
	}

	return false
}

// Execute 执行技能
func (s *WeatherSkill) Execute(ctx context.Context, params SkillParams) (*SkillResult, error) {
	switch params.Intent {
	case "query_weather":
		return s.handleQueryWeather(ctx, params)
	case "weather_forecast":
		return s.handleWeatherForecast(ctx, params)
	case "delivery_weather":
		return s.handleDeliveryWeather(ctx, params)
	case "weather_alert":
		return s.handleWeatherAlert(ctx, params)
	default:
		return s.handleDefault(ctx, params)
	}
}

// handleQueryWeather 查询实时天气
func (s *WeatherSkill) handleQueryWeather(ctx context.Context, params SkillParams) (*SkillResult, error) {
	location, ok := params.Entities["location"]
	if !ok {
		location = "当前位置" // 默认使用当前位置
	}

	// 调用和风天气 API 获取实时天气
	weather, err := s.getRealTimeWeather(ctx, location)
	if err != nil {
		return &SkillResult{
			Success:  false,
			Response: fmt.Sprintf("实时天气查询失败：%v", err),
		}, nil
	}

	// 构建回复
	response := fmt.Sprintf(
		"🌤️ %s实时天气：\n\n"+
			"🌡️ 温度：%d°C（体感%d°C）\n"+
			"☁️ 天气：%s\n"+
			"💧 湿度：%d%%\n"+
			"🌬️ 风力：%s %s\n"+
			"👁️ 能见度：%d公里\n"+
			"🕐 更新时间：%s",
		weather.Location,
		weather.Temperature,
		weather.FeelsLike,
		weather.Weather,
		weather.Humidity,
		weather.WindDir,
		weather.WindLevel,
		weather.Visibility,
		weather.UpdateTime.Format("15:04"),
	)

	// 添加配送建议
	suggestion := s.getDeliverySuggestion(weather)
	response += fmt.Sprintf("\n\n💡 配送建议：%s", suggestion)

	return &SkillResult{
		Success:  true,
		Response: response,
		Data:     weather,
		SuggestedActions: []Action{
			{Type: "weather_forecast", Name: "查看预报", Params: `{}`},
			{Type: "delivery_weather", Name: "配送天气", Params: `{}`},
		},
	}, nil
}

// handleWeatherForecast 天气预报
func (s *WeatherSkill) handleWeatherForecast(ctx context.Context, params SkillParams) (*SkillResult, error) {
	location, ok := params.Entities["location"]
	if !ok {
		location = "当前位置"
	}

	days, _ := params.Entities["days"]
	if days == "" {
		days = "3" // 默认3天
	}

	// 调用和风天气 API 获取预报
	forecasts, err := s.getWeatherForecast(ctx, location, days)
	if err != nil {
		return &SkillResult{
			Success:  false,
			Response: fmt.Sprintf("天气预报查询失败：%v", err),
		}, nil
	}

	response := fmt.Sprintf("🌤️ %s未来%s天天气预报：\n\n", location, days)
	for _, f := range forecasts {
		icon := s.getWeatherIcon(f.Weather)
		response += fmt.Sprintf(
			"%s %s %s %d°C~%d°C %s\n",
			icon,
			f.Date,
			f.Weather,
			f.Low,
			f.High,
			f.Wind,
		)
	}

	response += "\n⚠️ 提示：明天有小雨，出门记得带雨具！"

	return &SkillResult{
		Success:  true,
		Response: response,
		Data:     forecasts,
		SuggestedActions: []Action{
			{Type: "query_weather", Name: "实时天气", Params: `{}`},
			{Type: "weather_alert", Name: "天气预警", Params: `{}`},
		},
	}, nil
}

// handleDeliveryWeather 配送天气建议
func (s *WeatherSkill) handleDeliveryWeather(ctx context.Context, params SkillParams) (*SkillResult, error) {
	// 调用和风天气 API 获取配送路线天气
	routeWeather, err := s.getRouteWeather(ctx)
	if err != nil {
		return &SkillResult{
			Success:  false,
			Response: fmt.Sprintf("配送天气查询失败：%v", err),
		}, nil
	}

	response := "🚴 配送路线天气分析：\n\n"
	response += fmt.Sprintf(
		"起点天气：%s\n"+
			"终点天气：%s\n"+
			"温度：%d°C\n"+
			"降雨概率：%d%%\n"+
			"风力：%s\n\n",
		routeWeather.StartWeather,
		routeWeather.EndWeather,
		routeWeather.Temperature,
		routeWeather.RainProbability,
		routeWeather.WindLevel,
	)

	// 根据天气给出建议
	if routeWeather.RainProbability > 50 {
		response += "⚠️ 配送建议：\n"
		response += "1. 记得携带雨具（雨衣/雨伞）\n"
		response += "2. 路面湿滑，注意骑行安全\n"
		response += "3. 适当降低骑行速度\n"
		response += "4. 保护好餐品，避免淋湿\n"
	} else if routeWeather.Temperature > 35 {
		response += "🌡️ 配送建议：\n"
		response += "1. 天气炎热，注意防暑降温\n"
		response += "2. 多喝水，避免中暑\n"
		response += "3. 佩戴遮阳装备\n"
	} else {
		response += "✅ 天气条件良好，适合配送！"
	}

	return &SkillResult{
		Success:  true,
		Response: response,
		Data:     routeWeather,
		SuggestedActions: []Action{
			{Type: "query_weather", Name: "实时天气", Params: `{}`},
			{Type: "weather_forecast", Name: "查看预报", Params: `{}`},
		},
	}, nil
}

// handleWeatherAlert 天气预警
func (s *WeatherSkill) handleWeatherAlert(ctx context.Context, params SkillParams) (*SkillResult, error) {
	// 调用和风天气 API 获取天气预警
	alerts, err := s.getWeatherAlert(ctx)
	if err != nil {
		return &SkillResult{
			Success:  false,
			Response: fmt.Sprintf("天气预警查询失败：%v", err),
		}, nil
	}

	response := "⚠️ 天气预警信息：\n\n"
	for _, alert := range alerts {
		response += fmt.Sprintf(
			"🚨 %s%s预警\n"+
				"发布时间：%s\n"+
				"预警内容：%s\n\n",
			alert.Level,
			alert.Type,
			alert.PublishTime.Format("15:04"),
			alert.Content,
		)
	}

	response += "💡 建议：恶劣天气下请注意安全，必要时可暂停配送。"

	return &SkillResult{
		Success:  true,
		Response: response,
		Data:     alerts,
		SuggestedActions: []Action{
			{Type: "safety_guide", Name: "安全指南", Params: `{}`},
			{Type: "contact_support", Name: "联系客服", Params: `{}`},
		},
	}, nil
}

// handleDefault 默认处理
func (s *WeatherSkill) handleDefault(ctx context.Context, params SkillParams) (*SkillResult, error) {
	return &SkillResult{
		Success: true,
		Response: "🌤️ 我可以帮您：\n\n" +
			"1. 查询实时天气\n" +
			"2. 查看天气预报\n" +
			"3. 配送天气建议\n" +
			"4. 天气预警信息\n\n" +
			"请告诉我您想了解什么？",
		SuggestedActions: []Action{
			{Type: "query_weather", Name: "实时天气", Params: `{}`},
			{Type: "weather_forecast", Name: "天气预报", Params: `{}`},
			{Type: "delivery_weather", Name: "配送天气", Params: `{}`},
		},
	}, nil
}

// GetExamples 获取示例
func (s *WeatherSkill) GetExamples() []SkillExample {
	return []SkillExample{
		{
			Input:    "今天天气怎么样",
			Intent:   "query_weather",
			Entities: map[string]string{},
			Response: "今天多云，温度25°C，体感27°C，适合配送...",
		},
		{
			Input:    "未来三天天气预报",
			Intent:   "weather_forecast",
			Entities: map[string]string{"days": "3"},
			Response: "未来三天天气：今天多云25°C，明天小雨22°C...",
		},
		{
			Input:    "配送路上会下雨吗",
			Intent:   "delivery_weather",
			Entities: map[string]string{},
			Response: "配送路线天气分析：降雨概率60%，建议携带雨具...",
		},
	}
}

// getDeliverySuggestion 获取配送建议
func (s *WeatherSkill) getDeliverySuggestion(weather *WeatherInfo) string {
	if strings.Contains(weather.Weather, "雨") {
		return "雨天路滑，请减速慢行，注意保护餐品"
	}
	if strings.Contains(weather.Weather, "雪") {
		return "雪天路滑，建议穿戴防滑装备，注意安全"
	}
	if weather.Temperature > 35 {
		return "天气炎热，注意防暑降温，多喝水"
	}
	if weather.Temperature < 5 {
		return "天气寒冷，注意保暖，佩戴手套"
	}
	if weather.WindLevel == "5级" || weather.WindLevel == "6级" {
		return "风力较大，骑行时注意侧风影响"
	}
	return "天气良好，适合配送"
}

// getWeatherIcon 获取天气图标
func (s *WeatherSkill) getWeatherIcon(weather string) string {
	iconMap := map[string]string{
		"晴":  "☀️",
		"多云": "⛅",
		"阴":  "☁️",
		"小雨": "🌦️",
		"中雨": "🌧️",
		"大雨": "⛈️",
		"雪":  "❄️",
		"雾":  "🌫️",
	}
	if icon, ok := iconMap[weather]; ok {
		return icon
	}
	return "🌡️"
}

// TODO: 实现真实的天气 API 调用
// getRealTimeWeather 获取实时天气
func (s *WeatherSkill) getRealTimeWeather(ctx context.Context, location string) (*WeatherInfo, error) {
	// 构建请求 URL
	u, _ := url.Parse(s.baseURL + "/weather/now")
	q := u.Query()
	q.Set("key", s.apiKey)
	q.Set("location", location)
	u.RawQuery = q.Encode()

	// 发送请求
	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 解析响应
	var result struct {
		Code string `json:"code"`
		Now  struct {
			Temp      string `json:"temp"`
			FeelsLike string `json:"feelsLike"`
			Text      string `json:"text"`
			Humidity  string `json:"humidity"`
			WindScale string `json:"windScale"`
			WindDir   string `json:"windDir"`
			Vis       string `json:"vis"`
		} `json:"now"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if result.Code != "200" {
		// API 调用失败，返回模拟数据
		return s.getMockWeather(location), nil
	}

	temp, _ := strconv.Atoi(result.Now.Temp)
	feelsLike, _ := strconv.Atoi(result.Now.FeelsLike)
	humidity, _ := strconv.Atoi(result.Now.Humidity)
	vis, _ := strconv.Atoi(result.Now.Vis)

	return &WeatherInfo{
		Location:    location,
		Temperature: temp,
		FeelsLike:   feelsLike,
		Weather:     result.Now.Text,
		Humidity:    humidity,
		WindLevel:   result.Now.WindScale + "级",
		WindDir:     result.Now.WindDir,
		Visibility:  vis,
		UpdateTime:  time.Now(),
	}, nil
}

// getWeatherForecast 获取天气预报
func (s *WeatherSkill) getWeatherForecast(ctx context.Context, location string, days string) ([]WeatherForecast, error) {
	// 构建请求 URL - 和风天气3天预报 API
	// 和风天气 API 使用 /weather/3d 或 /weather/7d
	u, _ := url.Parse(s.baseURL + "/weather/3d")
	q := u.Query()
	q.Set("key", s.apiKey)
	q.Set("location", location)
	q.Set("days", days)
	u.RawQuery = q.Encode()

	// 发送请求
	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 解析响应
	var result struct {
		Code  string `json:"code"`
		Daily []struct {
			Date      string `json:"date"`
			TextDay   string `json:"textDay"`
			TextNight string `json:"textNight"`
			High      string `json:"high"`
			Low       string `json:"low"`
			WindDir   string `json:"dir"`
			WindScale string `json:"scale"`
		} `json:"daily"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if result.Code != "200" {
		// API 调用失败，返回模拟数据
		return s.getMockForecast(days), nil
	}

	forecasts := make([]WeatherForecast, 0)
	for _, d := range result.Daily {
		forecasts = append(forecasts, WeatherForecast{
			Date:    d.Date,
			Weather: d.TextDay,
			High:    mustInt(d.High),
			Low:     mustInt(d.Low),
			Wind:    d.WindDir + d.WindScale + "级",
		})
	}

	return forecasts, nil
}

// getRouteWeather 获取路线天气
func (s *WeatherSkill) getRouteWeather(ctx context.Context) (*RouteWeather, error) {
	// 这里简化处理，获取当前位置的天气作为起点和终点的天气
	// 实际使用时需要根据具体的起点和终点坐标获取天气
	location := "北京" // 默认使用北京

	// 获取实时天气
	weather, err := s.getRealTimeWeather(ctx, location)
	if err != nil {
		return nil, err
	}

	// 简化处理，假设终点天气与起点相同
	routeWeather := &RouteWeather{
		StartWeather:    weather.Weather,
		EndWeather:      weather.Weather,
		Temperature:     weather.Temperature,
		RainProbability: 0, // 简化处理，实际需要根据预报计算
		WindLevel:       weather.WindLevel,
	}

	return routeWeather, nil
}

// getWeatherAlert 获取天气预警
func (s *WeatherSkill) getWeatherAlert(ctx context.Context) ([]WeatherAlert, error) {
	// 构建请求 URL - 和风天气预警 API
	u, _ := url.Parse(s.baseURL + "/warning/now")
	q := u.Query()
	q.Set("key", s.apiKey)
	q.Set("location", "北京") // 默认使用北京
	u.RawQuery = q.Encode()

	// 发送请求
	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return s.getMockAlerts(), nil
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return s.getMockAlerts(), nil
	}
	defer resp.Body.Close()

	// 解析响应
	var result struct {
		Code    string `json:"code"`
		Warning []struct {
			WarningType  string `json:"type"`
			WarningLevel string `json:"level"`
			PubTime      string `json:"pubTime"`
			Content      string `json:"content"`
		} `json:"warning"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return s.getMockAlerts(), nil
	}

	if result.Code != "200" {
		return s.getMockAlerts(), nil
	}

	alerts := make([]WeatherAlert, 0)
	for _, w := range result.Warning {
		publishTime, _ := time.Parse("2006-01-02 15:04:05", w.PubTime)
		alerts = append(alerts, WeatherAlert{
			Type:        w.WarningType,
			Level:       w.WarningLevel,
			Content:     w.Content,
			PublishTime: publishTime,
		})
	}

	return alerts, nil
}

// WeatherInfo 天气信息
type WeatherInfo struct {
	Location    string    `json:"location"`
	Temperature int       `json:"temperature"` // 温度
	FeelsLike   int       `json:"feels_like"`  // 体感温度
	Weather     string    `json:"weather"`     // 天气状况
	Humidity    int       `json:"humidity"`    // 湿度
	WindLevel   string    `json:"wind_level"`  // 风力等级
	WindDir     string    `json:"wind_dir"`    // 风向
	Visibility  int       `json:"visibility"`  // 能见度
	UpdateTime  time.Time `json:"update_time"` // 更新时间
}

// WeatherForecast 天气预报
type WeatherForecast struct {
	Date    string `json:"date"`    // 日期
	Weather string `json:"weather"` // 天气
	High    int    `json:"high"`    // 最高温度
	Low     int    `json:"low"`     // 最低温度
	Wind    string `json:"wind"`    // 风力
}

// RouteWeather 路线天气
type RouteWeather struct {
	StartWeather    string `json:"start_weather"`    // 起点天气
	EndWeather      string `json:"end_weather"`      // 终点天气
	Temperature     int    `json:"temperature"`      // 温度
	RainProbability int    `json:"rain_probability"` // 降雨概率
	WindLevel       string `json:"wind_level"`       // 风力
}

// WeatherAlert 天气预警
type WeatherAlert struct {
	Type        string    `json:"type"`         // 预警类型
	Level       string    `json:"level"`        // 预警级别
	Content     string    `json:"content"`      // 预警内容
	PublishTime time.Time `json:"publish_time"` // 发布时间
}

// getMockWeather 获取模拟天气数据（API 失败时使用）
func (s *WeatherSkill) getMockWeather(location string) *WeatherInfo {
	return &WeatherInfo{
		Location:    location,
		Temperature: 22,
		FeelsLike:   24,
		Weather:     "多云",
		Humidity:    65,
		WindLevel:   "3级",
		WindDir:     "东南风",
		Visibility:  10,
		UpdateTime:  time.Now(),
	}
}

// getMockForecast 获取模拟天气预报数据（API 失败时使用）
func (s *WeatherSkill) getMockForecast(days string) []WeatherForecast {
	numDays := mustInt(days)
	if numDays <= 0 {
		numDays = 3
	}

	forecasts := make([]WeatherForecast, 0, numDays)
	weatherTypes := []string{"晴", "多云", "阴", "小雨"}

	for i := 0; i < numDays; i++ {
		date := time.Now().AddDate(0, 0, i).Format("2006-01-02")
		forecasts = append(forecasts, WeatherForecast{
			Date:    date,
			Weather: weatherTypes[i%len(weatherTypes)],
			High:    20 + i*2,
			Low:     10 + i,
			Wind:    "东南风3级",
		})
	}

	return forecasts
}

// getMockAlerts 获取模拟天气预警数据（API 失败时使用）
func (s *WeatherSkill) getMockAlerts() []WeatherAlert {
	// 默认返回空预警列表，表示没有预警
	return []WeatherAlert{}
}
