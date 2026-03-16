package server

import (
	"net/http"

	v1 "hellokratos/api/helloworld/v1"
	riderV1 "hellokratos/api/rider/v1"
	"hellokratos/internal/conf"
	"hellokratos/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	kratoshttp "github.com/go-kratos/kratos/v2/transport/http"
)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(c *conf.Server, greeter *service.GreeterService, authService *service.AuthService, orderService *service.OrderService, messageService *service.MessageService, incomeService *service.IncomeService, aiAgentService *service.AIAgentService, qualificationService *service.QualificationService, referralService *service.ReferralService, locationService *service.LocationService, appealService *service.AppealService, safetyService *service.SafetyService, logger log.Logger) *kratoshttp.Server {
	var opts = []kratoshttp.ServerOption{
		kratoshttp.Middleware(
			recovery.Recovery(),
		),
	}
	if c.Http.Network != "" {
		opts = append(opts, kratoshttp.Network(c.Http.Network))
	}
	if c.Http.Addr != "" {
		opts = append(opts, kratoshttp.Address(c.Http.Addr))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, kratoshttp.Timeout(c.Http.Timeout.AsDuration()))
	}
	srv := kratoshttp.NewServer(opts...)

	// 根路径处理 - 解决访问域名返回404问题
	srv.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello from Kratos Rider Service on CloudBase!"))
	})

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
	// 新增：骑手位置服务
	riderV1.RegisterLocationHTTPServer(srv, locationService)
	// 新增：申诉服务
	riderV1.RegisterAppealHTTPServer(srv, appealService)
	// 新增：安全服务
	riderV1.RegisterSafetyHTTPServer(srv, safetyService)
	return srv
}
