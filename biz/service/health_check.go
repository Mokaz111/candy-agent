package service

import (
	"context"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	candyAgent "github.mokaz111.com/candy-agent/hertz_gen/candyAgent"
)

// 应用启动时间
var startTime = time.Now()

// 版本信息
const version = "1.0.0"

type HealthCheckService struct {
	RequestContext *app.RequestContext
	Context        context.Context
}

func NewHealthCheckService(Context context.Context, RequestContext *app.RequestContext) *HealthCheckService {
	return &HealthCheckService{RequestContext: RequestContext, Context: Context}
}

func (h *HealthCheckService) Run(req *candyAgent.Empty) (resp *candyAgent.HealthResponse, err error) {
	defer func() {
		hlog.CtxInfof(h.Context, "Health check requested")
	}()

	// 计算运行时间
	uptime := time.Since(startTime).String()

	resp = &candyAgent.HealthResponse{
		Status:  "running",
		Version: version,
		Uptime:  uptime,
	}

	hlog.Infof("Health check: status=%s, version=%s, uptime=%s", resp.Status, resp.Version, resp.Uptime)
	return resp, nil
}
