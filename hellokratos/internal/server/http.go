package server

import (
	v1 "hellokratos/api/helloworld/v1"
	riderV1 "hellokratos/api/rider/v1"
	"hellokratos/internal/conf"
	"hellokratos/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/http"
)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(c *conf.Server, greeter *service.GreeterService, authService *service.AuthService, orderService *service.OrderService, messageService *service.MessageService, incomeService *service.IncomeService, aiAgentService *service.AIAgentService, qualificationService *service.QualificationService, referralService *service.ReferralService, logger log.Logger) *http.Server {
	var opts = []http.ServerOption{
		http.Middleware(
			recovery.Recovery(),
		),
	}
	if c.Http.Network != "" {
		opts = append(opts, http.Network(c.Http.Network))
	}
	if c.Http.Addr != "" {
		opts = append(opts, http.Address(c.Http.Addr))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, http.Timeout(c.Http.Timeout.AsDuration()))
	}
	srv := http.NewServer(opts...)
	v1.RegisterGreeterHTTPServer(srv, greeter)
	riderV1.RegisterAuthHTTPServer(srv, authService)
	riderV1.RegisterOrderHTTPServer(srv, orderService)
	riderV1.RegisterMessageHTTPServer(srv, messageService)
	riderV1.RegisterIncomeHTTPServer(srv, incomeService)
	// 新增：AI智能体客服和资质验证服务
	riderV1.RegisterAIAgentHTTPServer(srv, aiAgentService)
	riderV1.RegisterQualificationHTTPServer(srv, qualificationService)
	// 新增：骑手拉新服务
	riderV1.RegisterReferralHTTPServer(srv, referralService)
	return srv
}
