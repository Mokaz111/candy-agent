package service

import (
	"context"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	candyAgent "github.mokaz111.com/candy-agent/hertz_gen/candyAgent"
)

type SendHeartbeatService struct {
	RequestContext *app.RequestContext
	Context        context.Context
}

func NewSendHeartbeatService(Context context.Context, RequestContext *app.RequestContext) *SendHeartbeatService {
	return &SendHeartbeatService{RequestContext: RequestContext, Context: Context}
}

func (h *SendHeartbeatService) Run(req *candyAgent.HeartbeatRequest) (resp *candyAgent.HeartbeatResponse, err error) {
	defer func() {
		hlog.CtxInfof(h.Context, "Heartbeat received from agent %s", req.AgentId)
	}()

	// 获取当前服务器时间
	serverTime := time.Now().Format(time.RFC3339)

	// 构建响应
	resp = &candyAgent.HeartbeatResponse{
		ServerTime: serverTime,
		Message:    "Heartbeat acknowledged",
	}

	hlog.Infof("Heartbeat from agent %s: status=%s, version=%s, cluster=%s",
		req.AgentId, req.Status, req.Version, req.ClusterInfo.Name)

	return resp, nil
}
