package biz

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

// BaiduMapService 百度地图服务
type BaiduMapService struct {
	apiKey  string
	baseURL string
	client  *http.Client
	log     *log.Helper
}

// NewBaiduMapService 创建百度地图服务
func NewBaiduMapService(apiKey string, logger log.Logger) *BaiduMapService {
	return &BaiduMapService{
		apiKey:  apiKey,
		baseURL: "https://api.map.baidu.com",
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
		log: log.NewHelper(logger),
	}
}

// CoordinateConvertRequest 坐标转换请求
type CoordinateConvertRequest struct {
	Lat      float64 // 纬度
	Lng      float64 // 经度
	FromType int     // 源坐标类型: 1-GPS, 2-谷歌/高德(GCJ-02), 3-百度(BD09)
	ToType   int     // 目标坐标类型: 3-百度(BD09), 4-火星(GCJ-02)
}

// CoordinateConvertResponse 坐标转换响应
type CoordinateConvertResponse struct {
	Status int `json:"status"`
	Result []struct {
		X float64 `json:"x"` // 经度
		Y float64 `json:"y"` // 纬度
	} `json:"result"`
}

// ConvertCoordinate 坐标转换
// 将GPS/高德坐标转换为百度坐标
func (s *BaiduMapService) ConvertCoordinate(ctx context.Context, lat, lng float64, fromType int) (float64, float64, error) {
	// 如果已经是百度坐标，直接返回
	if fromType == 3 {
		return lat, lng, nil
	}

	apiURL := fmt.Sprintf("%s/geoconv/v1/", s.baseURL)
	params := url.Values{
		"coords": {fmt.Sprintf("%f,%f", lng, lat)}, // 注意：百度API是先经度后纬度
		"from":   {fmt.Sprintf("%d", fromType)},
		"to":     {"3"}, // 转换为百度坐标
		"ak":     {s.apiKey},
		"output": {"json"},
	}

	fullURL := fmt.Sprintf("%s?%s", apiURL, params.Encode())

	resp, err := s.client.Get(fullURL)
	if err != nil {
		s.log.Error("百度坐标转换请求失败", "err", err)
		return lat, lng, err
	}
	defer resp.Body.Close()

	var result CoordinateConvertResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		s.log.Error("百度坐标转换解析失败", "err", err)
		return lat, lng, err
	}

	if result.Status != 0 {
		s.log.Error("百度坐标转换失败", "status", result.Status)
		return lat, lng, fmt.Errorf("坐标转换失败，status: %d", result.Status)
	}

	if len(result.Result) == 0 {
		return lat, lng, fmt.Errorf("坐标转换结果为空")
	}

	// 返回转换后的坐标（百度坐标）
	return result.Result[0].Y, result.Result[0].X, nil
}

// GeoCoderResponse 地理编码响应
type GeoCoderResponse struct {
	Status int `json:"status"`
	Result struct {
		AddressComponent struct {
			Country   string `json:"country"`
			Province  string `json:"province"`
			City      string `json:"city"`
			District  string `json:"district"`
			Street    string `json:"street"`
			StreetNum string `json:"street_number"`
		} `json:"addressComponent"`
		FormattedAddress string `json:"formatted_address"`
		Business         string `json:"business"`
	} `json:"result"`
}

// ReverseGeocoding 逆地理编码
// 经纬度 -> 地址信息
func (s *BaiduMapService) ReverseGeocoding(ctx context.Context, lat, lng float64) (*GeoCoderResponse, error) {
	apiURL := fmt.Sprintf("%s/reverse_geocoding/v3/", s.baseURL)
	params := url.Values{
		"location":   {fmt.Sprintf("%f,%f", lat, lng)},
		"ak":         {s.apiKey},
		"output":     {"json"},
		"extensions": {"poi"},
	}

	fullURL := fmt.Sprintf("%s?%s", apiURL, params.Encode())

	resp, err := s.client.Get(fullURL)
	if err != nil {
		s.log.Error("百度逆地理编码请求失败", "err", err)
		return nil, err
	}
	defer resp.Body.Close()

	var result GeoCoderResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		s.log.Error("百度逆地理编码解析失败", "err", err)
		return nil, err
	}

	if result.Status != 0 {
		s.log.Error("百度逆地理编码失败", "status", result.Status)
		return nil, fmt.Errorf("逆地理编码失败，status: %d", result.Status)
	}

	return &result, nil
}

// CalculateDistance 计算两点间距离（骑行距离）
func (s *BaiduMapService) CalculateDistance(ctx context.Context, fromLat, fromLng, toLat, toLng float64) (float64, int, error) {
	apiURL := fmt.Sprintf("%s/directionlite/v1/riding", s.baseURL)
	params := url.Values{
		"origin":      {fmt.Sprintf("%f,%f", fromLat, fromLng)},
		"destination": {fmt.Sprintf("%f,%f", toLat, toLng)},
		"ak":          {s.apiKey},
	}

	fullURL := fmt.Sprintf("%s?%s", apiURL, params.Encode())

	resp, err := s.client.Get(fullURL)
	if err != nil {
		s.log.Error("百度骑行路线规划请求失败", "err", err)
		return 0, 0, err
	}
	defer resp.Body.Close()

	var result struct {
		Status int `json:"status"`
		Result struct {
			Routes []struct {
				Distance int `json:"distance"` // 距离（米）
				Duration int `json:"duration"` // 时间（秒）
			} `json:"routes"`
		} `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		s.log.Error("百度骑行路线规划解析失败", "err", err)
		return 0, 0, err
	}

	if result.Status != 0 {
		s.log.Error("百度骑行路线规划失败", "status", result.Status)
		return 0, 0, fmt.Errorf("路线规划失败，status: %d", result.Status)
	}

	if len(result.Result.Routes) == 0 {
		return 0, 0, fmt.Errorf("无路线结果")
	}

	route := result.Result.Routes[0]
	return float64(route.Distance) / 1000.0, route.Duration, nil // 返回公里数和秒数
}

// IsInFence 判断点是否在围栏内（圆形围栏）
func (s *BaiduMapService) IsInFence(centerLat, centerLng, pointLat, pointLng float64, radius float64) bool {
	// 使用百度地图的距离计算API或本地计算
	// 简化版本：使用直线距离计算
	distance := calculateStraightDistance(centerLat, centerLng, pointLat, pointLng)
	return distance <= radius
}

// calculateStraightDistance 计算直线距离（使用Haversine公式）
func calculateStraightDistance(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 6371000 // 地球半径（米）
	phi1 := lat1 * math.Pi / 180
	phi2 := lat2 * math.Pi / 180
	deltaPhi := (lat2 - lat1) * math.Pi / 180
	deltaLambda := (lng2 - lng1) * math.Pi / 180

	a := math.Sin(deltaPhi/2)*math.Sin(deltaPhi/2) +
		math.Cos(phi1)*math.Cos(phi2)*
			math.Sin(deltaLambda/2)*math.Sin(deltaLambda/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c // 距离（米）
}
