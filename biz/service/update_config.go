package service

import (
	"context"
	"fmt"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	candyAgent "github.mokaz111.com/candy-agent/hertz_gen/candyAgent"
)

type UpdateConfigService struct {
	RequestContext *app.RequestContext
	Context        context.Context
}

func NewUpdateConfigService(Context context.Context, RequestContext *app.RequestContext) *UpdateConfigService {
	return &UpdateConfigService{RequestContext: RequestContext, Context: Context}
}

func (h *UpdateConfigService) Run(req *candyAgent.ConfigUpdateRequest) (resp *candyAgent.ConfigUpdateResponse, err error) {
	defer func() {
		hlog.CtxInfof(h.Context, "Config update requested, type: %v", req.ConfigType)
	}()

	resp = &candyAgent.ConfigUpdateResponse{}

	switch req.ConfigType {
	case candyAgent.ConfigType_CONFIG_TYPE_AGENT:
		if req.GetAgentConfig() != nil {
			err = h.updateAgentConfig(req.GetAgentConfig())
			if err != nil {
				resp.Message = fmt.Sprintf("Failed to update agent config: %v", err)
				return resp, err
			}
			resp.Message = "Agent config updated successfully"
		} else {
			resp.Message = "No agent config provided"
			err = fmt.Errorf("no agent config provided")
			return resp, err
		}

	default:
		resp.Message = fmt.Sprintf("Unknown config type: %v", req.ConfigType)
		err = fmt.Errorf("unknown config type: %v", req.ConfigType)
		return resp, err
	}

	hlog.Infof("Config updated: %s", resp.Message)
	return resp, nil
}

// 更新告警规则
func (h *UpdateConfigService) updateAlertRules(config *candyAgent.AlertRulesConfig) error {
	// 这里应该实现将告警规则应用到Kubernetes集群的逻辑
	// 例如创建或更新PrometheusRule和VMRule资源

	if config.PrometheusRules != "" {
		hlog.Infof("Updating Prometheus rules: %d bytes", len(config.PrometheusRules))
		// TODO: 实现更新Prometheus规则的逻辑
	}

	if config.VmRules != "" {
		hlog.Infof("Updating VM rules: %d bytes", len(config.VmRules))
		// TODO: 实现更新VictoriaMetrics规则的逻辑
	}

	return nil
}

// 更新Agent配置
func (h *UpdateConfigService) updateAgentConfig(config *candyAgent.AgentConfig) error {
	// 更新心跳间隔
	if config.HeartbeatInterval > 0 {
		hlog.Infof("Updating heartbeat interval to %d seconds", config.HeartbeatInterval)
		// TODO: 实现更新心跳间隔的逻辑
	}

	// 更新日志级别
	if config.LogLevel != "" {
		hlog.Infof("Updating log level to %s", config.LogLevel)
		// TODO: 实现更新日志级别的逻辑
	}

	return nil
}
