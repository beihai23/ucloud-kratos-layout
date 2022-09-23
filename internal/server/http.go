package server

import (
	"git.ucloudadmin.com/framework/micro-mini/middleware/tracing"
	uhttp "git.ucloudadmin.com/framework/micro-mini/transport/http"

	kprom "github.com/go-kratos/kratos/contrib/metrics/prometheus/v2"
	"github.com/go-kratos/kratos/v2/log"
	kmetrics "github.com/go-kratos/kratos/v2/metrics"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/validate"
	"github.com/go-kratos/kratos/v2/transport/http"
	v1 "github.com/nobugtodebug/ucloud-kratos-layout/api/helloworld/v1"
	"github.com/nobugtodebug/ucloud-kratos-layout/internal/conf"
	"github.com/nobugtodebug/ucloud-kratos-layout/internal/service"
	prom "github.com/prometheus/client_golang/prometheus"
)

var (
	retCodeCounter kmetrics.Counter
)

func init() {
	counter := prom.NewCounterVec(
		prom.CounterOpts{
			Namespace:   "uge",
			Subsystem:   "uresource",
			Name:        "action_ret_code",
			ConstLabels: prom.Labels{"app": "go-uresource-api"},
		}, []string{"action", "retcode"})
	prom.MustRegister(counter)

	retCodeCounter = kprom.NewCounter(counter)
}

// NewHTTPServer new a HTTP server.
func NewHTTPServer(c *conf.Server, greeter *service.GreeterService, logger log.Logger) *http.Server {
	var opts = []http.ServerOption{
		http.Middleware(
			metrics.Server(
				metrics.WithSeconds(kprom.NewHistogram(_metricSeconds)),
				metrics.WithRequests(kprom.NewCounter(_metricRequests)),
			),
			recovery.Recovery(),
			tracing.Server(),
			logging.Server(logger),
			validate.Validator(),
		),
		http.ErrorEncoder(uhttp.WithUCloudErrorEncoder(v1.ErrorReason_value, v1.UCloudErrorReasonTab, retCodeCounter)),
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
	srv.Handle("/metrics", promhttp.Handler())
	srv.HandlePrefix("/q/", openapiv2.NewHandler())

	v1.RegisterGreeterHTTPServer(srv, greeter)
	return srv
}
