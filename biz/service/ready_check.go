package service

import (
	"context"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	dal "github.mokaz111.com/candy-agent/biz/dal/executor"
	"github.mokaz111.com/candy-agent/biz/executor"
	"github.mokaz111.com/candy-agent/conf"
	candyAgent "github.mokaz111.com/candy-agent/hertz_gen/candyAgent"
)

type ReadyCheckService struct {
	RequestContext *app.RequestContext
	Context        context.Context
}

func NewReadyCheckService(Context context.Context, RequestContext *app.RequestContext) *ReadyCheckService {
	return &ReadyCheckService{RequestContext: RequestContext, Context: Context}
}

func (h *ReadyCheckService) Run(req *candyAgent.Empty) (resp *candyAgent.HealthResponse, err error) {
	defer func() {
		hlog.CtxInfof(h.Context, "req = %+v", req)
		hlog.CtxInfof(h.Context, "resp = %+v", resp)
	}()

	// 检查关键组件是否准备就绪
	_ = GetTaskManager()              // 仅检查是否能正常初始化
	_ = executor.GetExecutorFactory() // 仅检查是否能正常初始化
	_ = dal.ExecutorFactory           // 仅检查是否能正常初始化

	// 检查配置是否加载成功
	config := conf.GetConf()
	if config == nil {
		return &candyAgent.HealthResponse{
			Status:  "not ready",
			Version: "unknown",
			Uptime:  "0",
		}, nil
	}

	// 创建响应
	resp = &candyAgent.HealthResponse{
		Status:  "ready",
		Version: "1.0.0", // TODO: 从版本文件或环境变量中获取
		Uptime:  time.Since(readyCheckStartTime).String(),
	}

	return resp, nil
}

// 记录服务启动时间
var readyCheckStartTime = time.Now()
