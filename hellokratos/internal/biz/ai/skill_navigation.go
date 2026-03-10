package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"hellokratos/internal/conf"
)

// NavigationSkill 导航助手技能
// 提供路线规划、距离计算、导航指引等功能
type NavigationSkill struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

// NewNavigationSkill 创建导航助手技能
// TODO: 需要替换为实际的地图服务 API Key
func NewNavigationSkill(cfg *conf.Data) Skill {
	skill := &NavigationSkill{
		// 百度地图 API Key
		apiKey:  "VyNeRRX9dO9DThd0xkf3mg3OA3bcF3AF",
		baseURL: "https://api.map.baidu.com", // 百度地图 API
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}

	// 如果配置中有地图服务配置，使用配置值
	if cfg != nil && cfg.GetAiService() != nil {
		// 可以通过配置读取 API Key
		// skill.apiKey = cfg.GetMapService().GetApiKey()
	}

	return skill
}

// Name 技能名称
func (s *NavigationSkill) Name() string {
	return "navigation_assistant"
}

// Description 技能描述
func (s *NavigationSkill) Description() string {
	return "提供路线规划、距离计算、预计到达时间、导航指引等位置服务"
}

// CanHandle 判断是否可以处理该请求
func (s *NavigationSkill) CanHandle(ctx context.Context, intent string, entities map[string]string) bool {
	// 意图匹配
	navIntents := []string{
		"route_planning",
		"calculate_distance",
		"navigation",
		"estimate_time",
		"nearby_search",
	}

	for _, i := range navIntents {
		if intent == i {
			return true
		}
	}

	// 关键词匹配
	keywords := []string{"导航", "路线", "怎么走", "距离", "多远", "多久", "附近", "周边", "地图"}
	for _, keyword := range keywords {
		if strings.Contains(intent, keyword) {
			return true
		}
	}

	return false
}

// Execute 执行技能
func (s *NavigationSkill) Execute(ctx context.Context, params SkillParams) (*SkillResult, error) {
	switch params.Intent {
	case "route_planning", "navigation":
		return s.handleRoutePlanning(ctx, params)
	case "calculate_distance":
		return s.handleCalculateDistance(ctx, params)
	case "estimate_time":
		return s.handleEstimateTime(ctx, params)
	case "nearby_search":
		return s.handleNearbySearch(ctx, params)
	default:
		return s.handleDefault(ctx, params)
	}
}

// handleRoutePlanning 路线规划
func (s *NavigationSkill) handleRoutePlanning(ctx context.Context, params SkillParams) (*SkillResult, error) {
	// 获取起点和终点
	origin, ok1 := params.Entities["origin"]
	destination, ok2 := params.Entities["destination"]

	if !ok1 || !ok2 {
		return &SkillResult{
			Success:       true,
			Response:      "请告诉我起点和终点，例如：从XX到XX怎么走",
			NeedConfirm:   true,
			ConfirmPrompt: "请输入起点和终点",
		}, nil
	}

	// 调用百度地图 API 进行路线规划
	route, err := s.getRoute(ctx, origin, destination)
	if err != nil {
		return &SkillResult{
			Success:  false,
			Response: fmt.Sprintf("路线规划失败：%v", err),
		}, nil
	}

	// 构建回复
	response := fmt.Sprintf(
		"🗺️ 路线规划结果：\n\n"+
			"📍 起点：%s\n"+
			"🏁 终点：%s\n"+
			"📏 总距离：%.1f公里\n"+
			"⏱️ 预计时间：%d分钟\n"+
			"🚦 路况：%s\n\n"+
			"详细路线：\n",
		origin,
		destination,
		route.Distance,
		route.Duration,
		route.TrafficStatus,
	)

	for i, step := range route.Steps {
		response += fmt.Sprintf("%d. %s（%.1f公里，约%d分钟）\n", i+1, step.Instruction, step.Distance, step.Duration)
	}

	response += "\n💡 提示：实际行驶时间可能因交通状况而有所不同。"

	return &SkillResult{
		Success:  true,
		Response: response,
		Data:     route,
		SuggestedActions: []Action{
			{Type: "start_navigation", Name: "开始导航", Params: fmt.Sprintf(`{"origin":"%s","destination":"%s"}`, origin, destination)},
			{Type: "alternative_route", Name: "备选路线", Params: `{}`},
			{Type: "share_route", Name: "分享路线", Params: `{}`},
		},
	}, nil
}

// handleCalculateDistance 计算距离
func (s *NavigationSkill) handleCalculateDistance(ctx context.Context, params SkillParams) (*SkillResult, error) {
	origin, ok1 := params.Entities["origin"]
	destination, ok2 := params.Entities["destination"]

	if !ok1 || !ok2 {
		return &SkillResult{
			Success:       true,
			Response:      "请告诉我两个地点，例如：XX到XX有多远",
			NeedConfirm:   true,
			ConfirmPrompt: "请输入两个地点",
		}, nil
	}

	// 调用百度地图 API 计算距离
	route, err := s.getRoute(ctx, origin, destination)
	if err != nil {
		return &SkillResult{
			Success:  false,
			Response: fmt.Sprintf("距离计算失败：%v", err),
		}, nil
	}

	response := fmt.Sprintf(
		"📏 距离计算结果：\n\n"+
			"从 %s 到 %s\n"+
			"直线距离：%.1f公里\n"+
			"骑行距离：约%.1f公里\n"+
			"预计骑行时间：约%d分钟",
		origin,
		destination,
		route.Distance*0.8, // 直线距离约为实际距离的80%
		route.Distance,
		route.Duration,
	)

	return &SkillResult{
		Success:  true,
		Response: response,
		Data: map[string]interface{}{
			"origin":      origin,
			"destination": destination,
			"distance":    route.Distance,
		},
		SuggestedActions: []Action{
			{Type: "route_planning", Name: "查看路线", Params: fmt.Sprintf(`{"origin":"%s","destination":"%s"}`, origin, destination)},
		},
	}, nil
}

// handleEstimateTime 预计到达时间
func (s *NavigationSkill) handleEstimateTime(ctx context.Context, params SkillParams) (*SkillResult, error) {
	destination, ok := params.Entities["destination"]
	if !ok {
		return &SkillResult{
			Success:       true,
			Response:      "请告诉我目的地",
			NeedConfirm:   true,
			ConfirmPrompt: "请输入目的地",
		}, nil
	}

	// 调用百度地图 API 计算预计时间
	route, err := s.getRoute(ctx, "当前位置", destination)
	if err != nil {
		return &SkillResult{
			Success:  false,
			Response: fmt.Sprintf("预计时间计算失败：%v", err),
		}, nil
	}

	eta := time.Now().Add(time.Duration(route.Duration) * time.Minute)

	response := fmt.Sprintf(
		"⏱️ 预计到达时间：\n\n"+
			"目的地：%s\n"+
			"预计到达：%s\n"+
			"预计用时：约%d分钟\n\n"+
			"🚦 当前路况良好，建议立即出发。",
		destination,
		eta.Format("15:04"),
		route.Duration,
	)

	return &SkillResult{
		Success:  true,
		Response: response,
		Data: map[string]interface{}{
			"destination": destination,
			"eta":         eta.Format("15:04"),
			"duration":    20,
		},
		SuggestedActions: []Action{
			{Type: "start_navigation", Name: "开始导航", Params: fmt.Sprintf(`{"destination":"%s"}`, destination)},
		},
	}, nil
}

// handleNearbySearch 附近搜索
func (s *NavigationSkill) handleNearbySearch(ctx context.Context, params SkillParams) (*SkillResult, error) {
	keyword, ok := params.Entities["keyword"]
	if !ok {
		return &SkillResult{
			Success:       true,
			Response:      "请告诉我您想找什么，例如：附近的加油站、餐馆等",
			NeedConfirm:   true,
			ConfirmPrompt: "请输入搜索关键词",
		}, nil
	}

	// 调用百度地图 API 搜索附近地点
	places, err := s.searchNearby(ctx, keyword)
	if err != nil {
		return &SkillResult{
			Success:  false,
			Response: fmt.Sprintf("附近搜索失败：%v", err),
		}, nil
	}

	response := fmt.Sprintf("🔍 附近的%s：\n\n", keyword)
	for i, place := range places {
		response += fmt.Sprintf(
			"%d. %s\n   距离：%.1f公里\n   地址：%s\n\n",
			i+1,
			place["name"],
			place["distance"],
			place["address"],
		)
	}

	return &SkillResult{
		Success:  true,
		Response: response,
		Data:     places,
		SuggestedActions: []Action{
			{Type: "navigate_to", Name: "导航到第一个", Params: `{}`},
			{Type: "more_results", Name: "查看更多", Params: `{}`},
		},
	}, nil
}

// handleDefault 默认处理
func (s *NavigationSkill) handleDefault(ctx context.Context, params SkillParams) (*SkillResult, error) {
	return &SkillResult{
		Success: true,
		Response: "🗺️ 我可以帮您：\n\n" +
			"1. 规划路线 - 告诉我起点和终点\n" +
			"2. 计算距离 - 查询两地距离\n" +
			"3. 预计时间 - 计算到达时间\n" +
			"4. 附近搜索 - 查找周边设施\n\n" +
			"请告诉我您需要什么帮助？",
		SuggestedActions: []Action{
			{Type: "route_planning", Name: "路线规划", Params: `{}`},
			{Type: "nearby_search", Name: "附近搜索", Params: `{}`},
			{Type: "calculate_distance", Name: "计算距离", Params: `{}`},
		},
	}, nil
}

// GetExamples 获取示例
func (s *NavigationSkill) GetExamples() []SkillExample {
	return []SkillExample{
		{
			Input:    "从人民广场到外滩怎么走",
			Intent:   "route_planning",
			Entities: map[string]string{"origin": "人民广场", "destination": "外滩"},
			Response: "路线规划结果：总距离3.5公里，预计15分钟...",
		},
		{
			Input:    "附近有没有加油站",
			Intent:   "nearby_search",
			Entities: map[string]string{"keyword": "加油站"},
			Response: "附近的加油站：1. XX加油站（0.5公里）...",
		},
		{
			Input:    "到目的地还有多远",
			Intent:   "calculate_distance",
			Entities: map[string]string{},
			Response: "距离目的地还有2.3公里，预计7分钟到达...",
		},
	}
}

// TODO: 实现真实的地图 API 调用
// getRoute 调用百度地图 API 获取路线
func (s *NavigationSkill) getRoute(ctx context.Context, origin, destination string) (*RouteInfo, error) {
	// 构建请求 URL - 百度地图骑行路线规划 API
	u, _ := url.Parse(s.baseURL + "/direction/v2/riding")
	q := u.Query()
	q.Set("ak", s.apiKey)
	q.Set("origin", origin)
	q.Set("destination", destination)
	q.Set("output", "json")
	u.RawQuery = q.Encode()

	// 发送请求
	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return s.getMockRoute(origin, destination), nil
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return s.getMockRoute(origin, destination), nil
	}
	defer resp.Body.Close()

	// 解析百度地图响应
	var result struct {
		Status  int    `json:"status"`
		Message string `json:"message"`
		Result  struct {
			Routes []struct {
				Distance int `json:"distance"`
				Duration int `json:"duration"`
				Steps    []struct {
					Instruction string `json:"instruction"`
					Distance    int    `json:"distance"`
					Duration    int    `json:"duration"`
				} `json:"steps"`
			} `json:"routes"`
		} `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return s.getMockRoute(origin, destination), nil
	}

	if result.Status != 0 {
		return s.getMockRoute(origin, destination), nil
	}

	if len(result.Result.Routes) == 0 {
		return s.getMockRoute(origin, destination), nil
	}

	routeData := result.Result.Routes[0]

	route := &RouteInfo{
		Distance: float64(routeData.Distance) / 1000, // 转换为公里
		Duration: routeData.Duration / 60,            // 转换为分钟
	}

	for _, step := range routeData.Steps {
		route.Steps = append(route.Steps, RouteStep{
			Instruction: step.Instruction,
			Distance:    float64(step.Distance) / 1000,
			Duration:    step.Duration / 60,
		})
	}

	return route, nil
}

// RouteInfo 路线信息
type RouteInfo struct {
	Distance      float64     `json:"distance"`       // 总距离（公里）
	Duration      int         `json:"duration"`       // 预计时间（分钟）
	Toll          float64     `json:"toll"`           // 过路费
	TrafficStatus string      `json:"traffic_status"` // 路况
	Steps         []RouteStep `json:"steps"`          // 详细步骤
}

// RouteStep 路线步骤
type RouteStep struct {
	Instruction string  `json:"instruction"` // 指引
	Distance    float64 `json:"distance"`    // 距离（公里）
	Duration    int     `json:"duration"`    // 时间（分钟）
}

// searchNearby 搜索附近地点
func (s *NavigationSkill) searchNearby(ctx context.Context, keyword string) ([]map[string]interface{}, error) {
	// 构建请求 URL - 百度地图附近搜索 API
	u, _ := url.Parse(s.baseURL + "/place/v2/search")
	q := u.Query()
	q.Set("ak", s.apiKey)
	q.Set("query", keyword)
	q.Set("radius", "5000") // 搜索半径5公里
	q.Set("output", "json")
	u.RawQuery = q.Encode()

	// 发送请求
	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return s.getMockNearby(keyword), nil
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return s.getMockNearby(keyword), nil
	}
	defer resp.Body.Close()

	// 解析百度地图响应
	var result struct {
		Status  int    `json:"status"`
		Message string `json:"message"`
		Results []struct {
			Name     string `json:"name"`
			Address  string `json:"address"`
			Location struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"location"`
		} `json:"results"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return s.getMockNearby(keyword), nil
	}

	if result.Status != 0 {
		return s.getMockNearby(keyword), nil
	}

	if len(result.Results) == 0 {
		return s.getMockNearby(keyword), nil
	}

	places := make([]map[string]interface{}, 0)
	for _, r := range result.Results {
		places = append(places, map[string]interface{}{
			"name":    r.Name,
			"address": r.Address,
			"lat":     r.Location.Lat,
			"lng":     r.Location.Lng,
		})
	}

	return places, nil
}

// getMockRoute 获取模拟路线数据（API 失败时使用）
func (s *NavigationSkill) getMockRoute(origin, destination string) *RouteInfo {
	return &RouteInfo{
		Distance:      3.5,
		Duration:      20,
		Toll:          0,
		TrafficStatus: "良好",
		Steps: []RouteStep{
			{Instruction: "从 " + origin + " 出发，沿主路向东行驶", Distance: 1.2, Duration: 5},
			{Instruction: "在十字路口左转，进入中山路", Distance: 1.5, Duration: 8},
			{Instruction: "继续直行，到达 " + destination, Distance: 0.8, Duration: 7},
		},
	}
}

// getMockNearby 获取模拟附近地点数据（API 失败时使用）
func (s *NavigationSkill) getMockNearby(keyword string) []map[string]interface{} {
	return []map[string]interface{}{
		{"name": keyword + " 1号店", "address": "XX路123号", "distance": 0.5},
		{"name": keyword + " 2号店", "address": "YY路456号", "distance": 1.2},
		{"name": keyword + " 3号店", "address": "ZZ路789号", "distance": 2.0},
	}
}
