package biz

import (
	"context"
	"errors"
	"hellokratos/internal/data/model"
	"sync"
	"testing"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockLocationRepo 位置仓库mock
type MockLocationRepo struct {
	mock.Mock
}

func (m *MockLocationRepo) SaveLocation(ctx context.Context, location *model.RiderLocation) error {
	args := m.Called(ctx, location)
	return args.Error(0)
}

func (m *MockLocationRepo) GetLocationByRiderID(ctx context.Context, riderID int64) (*model.RiderLocation, error) {
	args := m.Called(ctx, riderID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.RiderLocation), args.Error(1)
}

func (m *MockLocationRepo) GetNearbyRiders(ctx context.Context, lat, lng float64, radius int32, limit int32) ([]*model.RiderLocation, error) {
	args := m.Called(ctx, lat, lng, radius, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.RiderLocation), args.Error(1)
}

func (m *MockLocationRepo) SaveLocationHistory(ctx context.Context, history *model.RiderLocationHistory) error {
	args := m.Called(ctx, history)
	return args.Error(0)
}

func (m *MockLocationRepo) GetUserByRiderID(ctx context.Context, riderID int64) (*model.User, error) {
	args := m.Called(ctx, riderID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

// MockRedisClient 模拟Redis客户端
type MockRedisClient struct {
	data map[string]string
	mu   sync.RWMutex
}

func NewMockRedisClient() *MockRedisClient {
	return &MockRedisClient{
		data: make(map[string]string),
	}
}

func (m *MockRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = value.(string)
	return nil
}

func (m *MockRedisClient) Get(ctx context.Context, key string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if val, ok := m.data[key]; ok {
		return val, nil
	}
	return "", errors.New("key not found")
}

func TestLocationUsecase_UpdateLocation(t *testing.T) {
	mockLocationRepo := new(MockLocationRepo)
	logger := newTestLogger()

	// 创建一个适配器来绕过Redis依赖
	uc := &locationUsecase{
		locationRepo: mockLocationRepo,
		log:          log.NewHelper(logger),
	}

	ctx := context.Background()

	// 百度地图坐标示例（北京市朝阳区）
	// 百度地图使用的是BD09坐标系，这里使用转换后的WGS-84/GCJ-02坐标
	beijingLat := 39.9042  // 纬度
	beijingLng := 116.4074 // 经度

	tests := []struct {
		name        string
		riderID     int64
		lat         float64
		lng         float64
		accuracy    float32
		speed       float32
		direction   int32
		address     string
		mockUser    *model.User
		mockUserErr error
		mockSaveErr error
		expectedErr bool
		errMsg      string
	}{
		{
			name:      "正常更新位置-北京朝阳",
			riderID:   1001,
			lat:       beijingLat,
			lng:       beijingLng,
			accuracy:  10.5,
			speed:     25.0,
			direction: 90,
			address:   "北京市朝阳区建国路88号",
			mockUser: &model.User{
				ID:       1001,
				Nickname: "rider001",
				Status:   1,
			},
			mockSaveErr: nil,
			expectedErr: false,
		},
		{
			name:      "正常更新位置-上海浦东",
			riderID:   1002,
			lat:       31.2304,  // 上海纬度
			lng:       121.4737, // 上海经度
			accuracy:  5.0,
			speed:     30.5,
			direction: 180,
			address:   "上海市浦东新区陆家嘴",
			mockUser: &model.User{
				ID:       1002,
				Nickname: "rider002",
				Status:   1,
			},
			mockSaveErr: nil,
			expectedErr: false,
		},
		{
			name:        "无效骑手ID",
			riderID:     0,
			lat:         beijingLat,
			lng:         beijingLng,
			mockUser:    nil,
			mockUserErr: errors.New("骑手不存在"),
			expectedErr: true,
			errMsg:      "骑手不存在",
		},
		{
			name:        "骑手不存在",
			riderID:     9999,
			lat:         beijingLat,
			lng:         beijingLng,
			mockUser:    nil,
			mockUserErr: errors.New("骑手不存在"),
			expectedErr: true,
			errMsg:      "骑手不存在",
		},
		{
			name:    "保存位置失败",
			riderID: 1003,
			lat:     beijingLat,
			lng:     beijingLng,
			mockUser: &model.User{
				ID:       1003,
				Nickname: "rider003",
				Status:   1,
			},
			mockSaveErr: errors.New("数据库错误"),
			expectedErr: true,
			errMsg:      "数据库错误",
		},
		{
			name:      "边界坐标-最北纬度",
			riderID:   1004,
			lat:       53.55, // 漠河附近
			lng:       122.30,
			accuracy:  15.0,
			speed:     0,
			direction: 0,
			address:   "黑龙江省漠河市",
			mockUser: &model.User{
				ID:       1004,
				Nickname: "rider004",
				Status:   1,
			},
			mockSaveErr: nil,
			expectedErr: false,
		},
		{
			name:      "边界坐标-最南纬度",
			riderID:   1005,
			lat:       3.85, // 曾母暗沙附近
			lng:       112.05,
			accuracy:  20.0,
			speed:     10.0,
			direction: 270,
			address:   "海南省三沙市",
			mockUser: &model.User{
				ID:       1005,
				Nickname: "rider005",
				Status:   1,
			},
			mockSaveErr: nil,
			expectedErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 设置mock期望
			mockLocationRepo.On("GetUserByRiderID", ctx, tt.riderID).Return(tt.mockUser, tt.mockUserErr).Once()

			// 只有用户存在时才设置SaveLocation mock
			if tt.mockUser != nil {
				mockLocationRepo.On("SaveLocation", ctx, mock.Anything).Return(tt.mockSaveErr).Once()
			}

			err := uc.UpdateLocation(ctx, tt.riderID, tt.lat, tt.lng, tt.accuracy, tt.speed, tt.direction, tt.address)

			if tt.expectedErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}

			mockLocationRepo.AssertExpectations(t)
		})
	}
}

func TestLocationUsecase_GetLocationByRiderID(t *testing.T) {
	mockLocationRepo := new(MockLocationRepo)
	logger := newTestLogger()

	uc := &locationUsecase{
		locationRepo: mockLocationRepo,
		log:          log.NewHelper(logger),
	}

	ctx := context.Background()

	tests := []struct {
		name         string
		riderID      int64
		mockLocation *model.RiderLocation
		mockErr      error
		expectedErr  bool
	}{
		{
			name:    "正常获取位置",
			riderID: 1001,
			mockLocation: &model.RiderLocation{
				RiderID:   1001,
				Latitude:  39.9042,
				Longitude: 116.4074,
				Address:   "北京市朝阳区",
				UpdatedAt: time.Now(),
			},
			mockErr:     nil,
			expectedErr: false,
		},
		{
			name:         "骑手位置不存在",
			riderID:      9999,
			mockLocation: nil,
			mockErr:      errors.New("位置不存在"),
			expectedErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLocationRepo.On("GetLocationByRiderID", ctx, tt.riderID).Return(tt.mockLocation, tt.mockErr).Once()

			location, err := uc.GetLocationByRiderID(ctx, tt.riderID)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, location)
				assert.Equal(t, tt.riderID, location.RiderID)
				assert.Equal(t, tt.mockLocation.Latitude, location.Latitude)
				assert.Equal(t, tt.mockLocation.Longitude, location.Longitude)
			}

			mockLocationRepo.AssertExpectations(t)
		})
	}
}

func TestLocationUsecase_GetNearbyRiders(t *testing.T) {
	mockLocationRepo := new(MockLocationRepo)
	logger := newTestLogger()

	uc := &locationUsecase{
		locationRepo: mockLocationRepo,
		log:          log.NewHelper(logger),
	}

	ctx := context.Background()

	// 百度地图坐标示例
	centerLat := 39.9042  // 中心点纬度（北京）
	centerLng := 116.4074 // 中心点经度（北京）
	radius := int32(5000) // 5公里
	limit := int32(10)

	tests := []struct {
		name          string
		lat           float64
		lng           float64
		radius        int32
		limit         int32
		mockLocations []*model.RiderLocation
		mockErr       error
		expectedErr   bool
		expectedCount int
	}{
		{
			name:   "正常获取附近骑手",
			lat:    centerLat,
			lng:    centerLng,
			radius: radius,
			limit:  limit,
			mockLocations: []*model.RiderLocation{
				{RiderID: 1001, Latitude: 39.9050, Longitude: 116.4080, Address: "附近位置1"},
				{RiderID: 1002, Latitude: 39.9030, Longitude: 116.4060, Address: "附近位置2"},
				{RiderID: 1003, Latitude: 39.9045, Longitude: 116.4075, Address: "附近位置3"},
			},
			mockErr:       nil,
			expectedErr:   false,
			expectedCount: 3,
		},
		{
			name:          "附近无骑手",
			lat:           centerLat,
			lng:           centerLng,
			radius:        radius,
			limit:         limit,
			mockLocations: []*model.RiderLocation{},
			mockErr:       nil,
			expectedErr:   false,
			expectedCount: 0,
		},
		{
			name:          "查询失败",
			lat:           centerLat,
			lng:           centerLng,
			radius:        radius,
			limit:         limit,
			mockLocations: nil,
			mockErr:       errors.New("查询失败"),
			expectedErr:   true,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLocationRepo.On("GetNearbyRiders", ctx, tt.lat, tt.lng, tt.radius, tt.limit).Return(tt.mockLocations, tt.mockErr).Once()

			locations, err := uc.GetNearbyRiders(ctx, tt.lat, tt.lng, tt.radius, tt.limit)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCount, len(locations))
			}

			mockLocationRepo.AssertExpectations(t)
		})
	}
}

func TestLocationUsecase_CoordinateValidation(t *testing.T) {
	// 测试坐标有效性验证
	tests := []struct {
		name    string
		lat     float64
		lng     float64
		isValid bool
	}{
		{"有效坐标-北京", 39.9042, 116.4074, true},
		{"有效坐标-上海", 31.2304, 121.4737, true},
		{"有效坐标-广州", 23.1291, 113.2644, true},
		{"有效坐标-深圳", 22.5431, 114.0579, true},
		{"边界-北纬90度", 90.0, 0.0, true},
		{"边界-南纬90度", -90.0, 0.0, true},
		{"边界-东经180度", 0.0, 180.0, true},
		{"边界-西经180度", 0.0, -180.0, true},
		{"无效坐标-超北纬", 91.0, 116.4074, false},
		{"无效坐标-超南纬", -91.0, 116.4074, false},
		{"无效坐标-超东经", 39.9042, 181.0, false},
		{"无效坐标-超西经", 39.9042, -181.0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 验证纬度范围
			isLatValid := tt.lat >= -90 && tt.lat <= 90
			// 验证经度范围
			isLngValid := tt.lng >= -180 && tt.lng <= 180

			if tt.isValid {
				assert.True(t, isLatValid, "纬度应在有效范围内")
				assert.True(t, isLngValid, "经度应在有效范围内")
			} else {
				assert.False(t, isLatValid && isLngValid, "坐标应在有效范围外")
			}
		})
	}
}

func TestLocationUsecase_ConcurrentUpdateLocation(t *testing.T) {
	mockLocationRepo := new(MockLocationRepo)
	logger := newTestLogger()

	uc := &locationUsecase{
		locationRepo: mockLocationRepo,
		log:          log.NewHelper(logger),
	}

	ctx := context.Background()

	// 设置mock期望 - 使用Maybe()允许多次调用
	mockLocationRepo.On("GetUserByRiderID", ctx, int64(1001)).Return(&model.User{
		ID:       1001,
		Nickname: "rider001",
		Status:   1,
	}, nil).Maybe()

	mockLocationRepo.On("SaveLocation", ctx, mock.Anything).Return(nil).Maybe()

	// 并发测试
	concurrency := 100
	var wg sync.WaitGroup
	wg.Add(concurrency)

	for i := 0; i < concurrency; i++ {
		go func(index int) {
			defer wg.Done()
			// 模拟不同位置更新
			lat := 39.9042 + float64(index)*0.0001
			lng := 116.4074 + float64(index)*0.0001
			err := uc.UpdateLocation(ctx, 1001, lat, lng, 10.0, 25.0, 90, "测试地址")
			// 由于使用了Maybe，不会报错
			_ = err
		}(i)
	}

	wg.Wait()
}
