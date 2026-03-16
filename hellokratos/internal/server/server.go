package server

import (
	"hellokratos/internal/service"

	"github.com/google/wire"
)

// ProviderSet is server providers.
var ProviderSet = wire.NewSet(NewGRPCServer, NewHTTPServer, NewLocationService)

// NewLocationService 创建位置服务实例
func NewLocationService(locationUsecase *service.LocationService) *service.LocationService {
	return locationUsecase
}
