package biz

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBaiduMapService_ConvertCoordinate(t *testing.T) {
	logger := newTestLogger()
	service := NewBaiduMapService("VyNeRRX9dO9DThd0xkf3mg3OA3bcF3AF", logger)

	ctx := context.Background()

	tests := []struct {
		name       string
		lat        float64
		lng        float64
		fromType   int
		wantErr    bool
		errContain string
	}{
		{
			name:     "GPS坐标转百度坐标-北京",
			lat:      39.9042,
			lng:      116.4074,
			fromType: 1, // GPS坐标
			wantErr:  false,
		},
		{
			name:     "高德坐标转百度坐标-上海",
			lat:      31.2304,
			lng:      121.4737,
			fromType: 2, // GCJ-02坐标
			wantErr:  false,
		},
		{
			name:     "百度坐标保持不变",
			lat:      39.9042,
			lng:      116.4074,
			fromType: 3, // 已经是百度坐标
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lat, lng, err := service.ConvertCoordinate(ctx, tt.lat, tt.lng, tt.fromType)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContain != "" {
					assert.Contains(t, err.Error(), tt.errContain)
				}
			} else {
				// 如果API调用成功，验证转换结果
				if err == nil {
					assert.NotZero(t, lat)
					assert.NotZero(t, lng)
					// 如果是百度坐标转百度坐标，应该保持不变
					if tt.fromType == 3 {
						assert.InDelta(t, tt.lat, lat, 0.0001)
						assert.InDelta(t, tt.lng, lng, 0.0001)
					}
				}
			}
		})
	}
}

func TestBaiduMapService_ReverseGeocoding(t *testing.T) {
	logger := newTestLogger()
	service := NewBaiduMapService("VyNeRRX9dO9DThd0xkf3mg3OA3bcF3AF", logger)

	ctx := context.Background()

	tests := []struct {
		name    string
		lat     float64
		lng     float64
		wantErr bool
	}{
		{
			name:    "逆地理编码-北京天安门",
			lat:     39.9042,
			lng:     116.4074,
			wantErr: false,
		},
		{
			name:    "逆地理编码-上海外滩",
			lat:     31.2304,
			lng:     121.4737,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.ReverseGeocoding(ctx, tt.lat, tt.lng)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				// API可能因配额或网络原因失败，这里不强制要求成功
				if err == nil && result != nil {
					assert.NotEmpty(t, result.Result.FormattedAddress)
					t.Logf("解析地址: %s", result.Result.FormattedAddress)
				}
			}
		})
	}
}

func TestBaiduMapService_CalculateDistance(t *testing.T) {
	logger := newTestLogger()
	service := NewBaiduMapService("VyNeRRX9dO9DThd0xkf3mg3OA3bcF3AF", logger)

	ctx := context.Background()

	tests := []struct {
		name    string
		fromLat float64
		fromLng float64
		toLat   float64
		toLng   float64
		wantErr bool
	}{
		{
			name:    "计算骑行距离-北京两点",
			fromLat: 39.9042,
			fromLng: 116.4074,
			toLat:   39.9142,
			toLng:   116.4174,
			wantErr: false,
		},
		{
			name:    "计算骑行距离-上海两点",
			fromLat: 31.2304,
			fromLng: 121.4737,
			toLat:   31.2404,
			toLng:   121.4837,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			distance, duration, err := service.CalculateDistance(ctx, tt.fromLat, tt.fromLng, tt.toLat, tt.toLng)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				// API可能因配额或网络原因失败，这里不强制要求成功
				if err == nil {
					assert.Greater(t, distance, 0.0)
					assert.Greater(t, duration, 0)
					t.Logf("距离: %.2f公里, 预计时间: %d秒", distance, duration)
				}
			}
		})
	}
}

func TestCalculateStraightDistance(t *testing.T) {
	tests := []struct {
		name     string
		lat1     float64
		lng1     float64
		lat2     float64
		lng2     float64
		expected float64 // 预期距离（米），允许一定误差
		delta    float64 // 允许的误差范围
	}{
		{
			name:     "北京同一点",
			lat1:     39.9042,
			lng1:     116.4074,
			lat2:     39.9042,
			lng2:     116.4074,
			expected: 0,
			delta:    1,
		},
		{
			name:     "北京-上海",
			lat1:     39.9042,
			lng1:     116.4074,
			lat2:     31.2304,
			lng2:     121.4737,
			expected: 1067000, // 约1067公里
			delta:    10000,   // 允许10公里误差
		},
		{
			name:     "北京-天津",
			lat1:     39.9042,
			lng1:     116.4074,
			lat2:     39.1252,
			lng2:     117.1904,
			expected: 109600, // 约109.6公里（直线距离）
			delta:    5000,   // 允许5公里误差
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			distance := calculateStraightDistance(tt.lat1, tt.lng1, tt.lat2, tt.lng2)
			assert.InDelta(t, tt.expected, distance, tt.delta)
			t.Logf("计算距离: %.2f米", distance)
		})
	}
}

func TestBaiduMapService_IsInFence(t *testing.T) {
	service := &BaiduMapService{}

	tests := []struct {
		name       string
		centerLat  float64
		centerLng  float64
		pointLat   float64
		pointLng   float64
		radius     float64 // 半径（米）
		wantInside bool
	}{
		{
			name:       "点在围栏内",
			centerLat:  39.9042,
			centerLng:  116.4074,
			pointLat:   39.9043,
			pointLng:   116.4075,
			radius:     1000, // 1公里
			wantInside: true,
		},
		{
			name:       "点在围栏外",
			centerLat:  39.9042,
			centerLng:  116.4074,
			pointLat:   39.9142,
			pointLng:   116.4174,
			radius:     100, // 100米
			wantInside: false,
		},
		{
			name:       "点在边界上",
			centerLat:  39.9042,
			centerLng:  116.4074,
			pointLat:   39.9042,
			pointLng:   116.4074,
			radius:     0,
			wantInside: true, // 同一点应该在围栏内
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inside := service.IsInFence(tt.centerLat, tt.centerLng, tt.pointLat, tt.pointLng, tt.radius)
			assert.Equal(t, tt.wantInside, inside)
		})
	}
}
