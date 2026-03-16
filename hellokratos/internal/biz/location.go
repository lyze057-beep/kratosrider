package biz

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"hellokratos/internal/data/model"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"
)

// LocationUsecase 位置相关的业务逻辑接口
type LocationUsecase interface {
	// UpdateLocation 更新骑手位置
	UpdateLocation(ctx context.Context, riderID int64, lat float64, lng float64, accuracy float32, speed float32, direction int32, address string) error
	// GetLocationByRiderID 根据骑手ID获取位置
	GetLocationByRiderID(ctx context.Context, riderID int64) (*model.RiderLocation, error)
	// GetNearbyRiders 获取附近骑手位置
	GetNearbyRiders(ctx context.Context, lat, lng float64, radius int32, limit int32) ([]*model.RiderLocation, error)
	// CalculateDistance 计算骑手到目的地的距离和预计时间
	CalculateDistance(ctx context.Context, riderID int64, destLat, destLng float64) (distance float64, duration int, err error)
}

// locationUsecase 位置相关的业务逻辑实现
type locationUsecase struct {
	locationRepo LocationRepo     // 位置数据访问接口
	rdb          *redis.Client    // Redis客户端
	baiduMap     *BaiduMapService // 百度地图服务
	log          *log.Helper      // 日志记录器
}

// LocationRepo 位置数据访问接口
type LocationRepo interface {
	SaveLocation(ctx context.Context, location *model.RiderLocation) error
	GetLocationByRiderID(ctx context.Context, riderID int64) (*model.RiderLocation, error)
	GetNearbyRiders(ctx context.Context, lat, lng float64, radius int32, limit int32) ([]*model.RiderLocation, error)
	SaveLocationHistory(ctx context.Context, history *model.RiderLocationHistory) error
}

// NewLocationUsecase 创建位置业务逻辑实例
func NewLocationUsecase(locationRepo LocationRepo, rdb *redis.Client, logger log.Logger) LocationUsecase {
	return &locationUsecase{
		locationRepo: locationRepo,
		rdb:          rdb,
		baiduMap:     NewBaiduMapService("VyNeRRX9dO9DThd0xkf3mg3OA3bcF3AF", logger),
		log:          log.NewHelper(logger),
	}
}

// UpdateLocation 更新骑手位置
func (uc *locationUsecase) UpdateLocation(ctx context.Context, riderID int64, lat float64, lng float64, accuracy float32, speed float32, direction int32, address string) error {
	// 1. 验证骑手是否存在
	authRepo, ok := uc.locationRepo.(interface {
		GetUserByRiderID(ctx context.Context, riderID int64) (*model.User, error)
	})
	if !ok {
		uc.log.Error("location repo does not support GetUserByRiderID")
		return errors.New("位置服务未初始化")
	}

	user, err := authRepo.GetUserByRiderID(ctx, riderID)
	if err != nil {
		uc.log.Error("failed to get rider", "err", err)
		return err
	}

	if user == nil {
		uc.log.Error("rider not found", "rider_id", riderID)
		return errors.New("骑手不存在")
	}

	// 2. 坐标转换（GPS坐标 -> 百度坐标）
	// 骑手APP上报的是GPS坐标(WGS-84)，需要转换为百度坐标(BD09)以便使用百度地图服务
	baiduLat, baiduLng := lat, lng
	if uc.baiduMap != nil {
		convertedLat, convertedLng, err := uc.baiduMap.ConvertCoordinate(ctx, lat, lng, 1) // 1 = GPS坐标
		if err == nil {
			baiduLat, baiduLng = convertedLat, convertedLng
			uc.log.Info("coordinate converted", "from_lat", lat, "from_lng", lng, "to_lat", baiduLat, "to_lng", baiduLng)
		} else {
			uc.log.Warn("coordinate conversion failed, using original", "err", err)
		}
	}

	// 3. 逆地理编码（经纬度 -> 地址）
	// 如果前端没有传地址，调用百度地图API解析
	if address == "" && uc.baiduMap != nil {
		geoResult, err := uc.baiduMap.ReverseGeocoding(ctx, baiduLat, baiduLng)
		if err == nil && geoResult != nil {
			address = geoResult.Result.FormattedAddress
			uc.log.Info("address resolved", "address", address)
		}
	}

	// 4. 保存位置数据
	location := &model.RiderLocation{
		RiderID:      riderID,
		Latitude:     baiduLat, // 保存百度坐标
		Longitude:    baiduLng,
		Accuracy:     float64(accuracy),
		Speed:        float64(speed),
		Direction:    direction,
		Address:      address,
		City:         "北京", // TODO: 从逆地理编码结果中获取
		Province:     "北京",
		Country:      "中国",
		LocationType: "gps",
	}

	err = uc.locationRepo.SaveLocation(ctx, location)
	if err != nil {
		uc.log.Error("failed to save location", "err", err)
		return err
	}

	// 5. 更新Redis缓存（5分钟过期）
	cacheKey := fmt.Sprintf("rider:location:%d", riderID)
	data, _ := json.Marshal(map[string]interface{}{
		"rider_id":   riderID,
		"latitude":   baiduLat,
		"longitude":  baiduLng,
		"accuracy":   accuracy,
		"speed":      speed,
		"direction":  direction,
		"address":    address,
		"updated_at": time.Now().Unix(),
	})
	if uc.rdb != nil {
		uc.rdb.Set(ctx, cacheKey, data, 5*time.Minute)
	}

	// 6. 推送位置给订单系统（用于智能派单）
	// TODO: 发布位置消息

	uc.log.Info("rider location updated", "rider_id", riderID, "lat", baiduLat, "lng", baiduLng, "address", address)
	return nil
}

// GetLocationByRiderID 根据骑手ID获取位置
func (uc *locationUsecase) GetLocationByRiderID(ctx context.Context, riderID int64) (*model.RiderLocation, error) {
	// 1. 先查缓存
	if uc.rdb != nil {
		cacheKey := fmt.Sprintf("rider:location:%d", riderID)
		cached, err := uc.rdb.Get(ctx, cacheKey).Result()
		if err == nil {
			var data map[string]interface{}
			if err := json.Unmarshal([]byte(cached), &data); err == nil {
				uc.log.Info("location from cache", "rider_id", riderID)
				return &model.RiderLocation{
					RiderID:   riderID,
					Latitude:  data["latitude"].(float64),
					Longitude: data["longitude"].(float64),
					Address:   data["address"].(string),
					UpdatedAt: time.Unix(int64(data["updated_at"].(float64)), 0),
				}, nil
			}
		}
	}

	// 2. 缓存未命中，查数据库
	location, err := uc.locationRepo.GetLocationByRiderID(ctx, riderID)
	if err != nil {
		return nil, err
	}

	// 3. 更新缓存
	if uc.rdb != nil {
		cacheKey := fmt.Sprintf("rider:location:%d", riderID)
		data, _ := json.Marshal(location)
		uc.rdb.Set(ctx, cacheKey, data, 5*time.Minute)
	}

	return location, nil
}

// GetNearbyRiders 获取附近骑手位置
func (uc *locationUsecase) GetNearbyRiders(ctx context.Context, lat, lng float64, radius int32, limit int32) ([]*model.RiderLocation, error) {
	// 1. 先查缓存
	if uc.rdb != nil {
		cacheKey := fmt.Sprintf("rider:nearby:%.6f:%.6f:%d:%d", lat, lng, radius, limit)
		cached, err := uc.rdb.Get(ctx, cacheKey).Result()
		if err == nil {
			var locations []*model.RiderLocation
			if err := json.Unmarshal([]byte(cached), &locations); err == nil {
				uc.log.Info("nearby riders from cache", "lat", lat, "lng", lng)
				return locations, nil
			}
		}
	}

	// 2. 缓存未命中，查数据库
	locations, err := uc.locationRepo.GetNearbyRiders(ctx, lat, lng, radius, limit)
	if err != nil {
		return nil, err
	}

	// 3. 更新缓存（30秒过期）
	if uc.rdb != nil {
		cacheKey := fmt.Sprintf("rider:nearby:%.6f:%.6f:%d:%d", lat, lng, radius, limit)
		data, _ := json.Marshal(locations)
		uc.rdb.Set(ctx, cacheKey, data, 30*time.Second)
	}

	return locations, nil
}

// CalculateDistance 计算骑手到目的地的距离和预计时间
func (uc *locationUsecase) CalculateDistance(ctx context.Context, riderID int64, destLat, destLng float64) (float64, int, error) {
	// 1. 获取骑手当前位置
	riderLocation, err := uc.GetLocationByRiderID(ctx, riderID)
	if err != nil {
		uc.log.Error("failed to get rider location", "err", err)
		return 0, 0, err
	}

	// 2. 调用百度地图API计算骑行距离
	if uc.baiduMap != nil {
		distance, duration, err := uc.baiduMap.CalculateDistance(ctx, riderLocation.Latitude, riderLocation.Longitude, destLat, destLng)
		if err == nil {
			uc.log.Info("distance calculated", "rider_id", riderID, "distance", distance, "duration", duration)
			return distance, duration, nil
		}
		uc.log.Warn("baidu distance calculation failed, using straight line", "err", err)
	}

	// 3. 如果百度地图API失败，使用直线距离估算
	straightDistance := calculateStraightDistance(riderLocation.Latitude, riderLocation.Longitude, destLat, destLng)
	// 估算骑行时间（假设平均速度15km/h）
	estimatedDuration := int(straightDistance * 3600 / 15) // 秒

	return straightDistance, estimatedDuration, nil
}
