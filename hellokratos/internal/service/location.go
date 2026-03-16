package service

import (
	"context"
	v1 "hellokratos/api/rider/v1"
	"hellokratos/internal/biz"
	"hellokratos/internal/data/model"

	"github.com/go-kratos/kratos/v2/log"
)

// LocationService 位置服务
type LocationService struct {
	v1.UnimplementedLocationServer
	locationUsecase biz.LocationUsecase
	log             *log.Helper
}

// NewLocationService 创建位置服务实例
func NewLocationService(locationUsecase biz.LocationUsecase, logger log.Logger) *LocationService {
	return &LocationService{
		locationUsecase: locationUsecase,
		log:             log.NewHelper(logger),
	}
}

// UpdateLocation 更新骑手位置
func (s *LocationService) UpdateLocation(ctx context.Context, req *v1.UpdateLocationRequest) (*v1.UpdateLocationReply, error) {
	// 1. 验证骑手ID
	if req.RiderId <= 0 {
		s.log.Error("invalid rider_id", "rider_id", req.RiderId)
		return &v1.UpdateLocationReply{
			Success: false,
			Message: "无效的骑手ID",
		}, nil
	}

	// 2. 更新位置
	err := s.locationUsecase.UpdateLocation(
		ctx,
		req.RiderId,
		req.Latitude,
		req.Longitude,
		req.Accuracy,
		req.Speed,
		req.Direction,
		req.Address,
	)
	if err != nil {
		s.log.Error("update location failed", "err", err, "rider_id", req.RiderId)
		return &v1.UpdateLocationReply{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	// 3. 推送位置更新（可选）
	// TODO: 推送位置更新给相关方

	s.log.Info("rider location updated", "rider_id", req.RiderId, "lat", req.Latitude, "lng", req.Longitude)

	return &v1.UpdateLocationReply{
		Success: true,
		Message: "位置上报成功",
	}, nil
}

// GetLocation 获取骑手位置
func (s *LocationService) GetLocation(ctx context.Context, req *v1.GetLocationRequest) (*v1.GetLocationReply, error) {
	location, err := s.locationUsecase.GetLocationByRiderID(ctx, req.RiderId)
	if err != nil {
		s.log.Error("get location failed", "err", err, "rider_id", req.RiderId)
		return &v1.GetLocationReply{
			Location: nil,
		}, nil
	}

	locationInfo := &v1.LocationInfo{
		RiderId:   location.RiderID,
		Latitude:  location.Latitude,
		Longitude: location.Longitude,
		Accuracy:  float32(location.Accuracy),
		Speed:     float32(location.Speed),
		Direction: location.Direction,
		Address:   location.Address,
		City:      location.City,
		Province:  location.Province,
		Country:   location.Country,
		UpdatedAt: location.UpdatedAt.Unix(),
	}

	return &v1.GetLocationReply{
		Location: locationInfo,
	}, nil
}

// GetNearbyRiders 获取附近骑手位置
func (s *LocationService) GetNearbyRiders(ctx context.Context, req *v1.GetNearbyRidersRequest) (*v1.GetNearbyRidersReply, error) {
	locations, err := s.locationUsecase.GetNearbyRiders(ctx, req.Latitude, req.Longitude, req.Radius, req.Limit)
	if err != nil {
		s.log.Error("get nearby riders failed", "err", err)
		return &v1.GetNearbyRidersReply{
			Riders: nil,
		}, nil
	}

	riders := make([]*v1.LocationInfo, len(locations))
	for i, location := range locations {
		riders[i] = &v1.LocationInfo{
			RiderId:   location.RiderID,
			Latitude:  location.Latitude,
			Longitude: location.Longitude,
			Accuracy:  float32(location.Accuracy),
			Speed:     float32(location.Speed),
			Direction: location.Direction,
			Address:   location.Address,
			City:      location.City,
			Province:  location.Province,
			Country:   location.Country,
			UpdatedAt: location.UpdatedAt.Unix(),
		}
	}

	return &v1.GetNearbyRidersReply{
		Riders: riders,
	}, nil
}

// convertLocationToInfo 将位置模型转换为API响应格式
func (s *LocationService) convertLocationToInfo(location *model.RiderLocation) *v1.LocationInfo {
	return &v1.LocationInfo{
		RiderId:   location.RiderID,
		Latitude:  location.Latitude,
		Longitude: location.Longitude,
		Accuracy:  float32(location.Accuracy),
		Speed:     float32(location.Speed),
		Direction: location.Direction,
		Address:   location.Address,
		City:      location.City,
		Province:  location.Province,
		Country:   location.Country,
		UpdatedAt: location.UpdatedAt.Unix(),
	}
}
