package service

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.mokaz111.com/candy-agent/biz/dal/alertrule"
	candyAgent "github.mokaz111.com/candy-agent/hertz_gen/candyAgent"
)

type GetService struct {
	RequestContext *app.RequestContext
	Context        context.Context
}

func NewGetAlertRuleService(Context context.Context, RequestContext *app.RequestContext) *GetService {
	return &GetService{
		RequestContext: RequestContext,
		Context:        Context,
	}
}

func (h *GetService) Run(req *candyAgent.AlertRuleRequest) (resp *candyAgent.AlertRuleResponse, err error) {
	defer func() {
		hlog.CtxInfof(h.Context, "Get alert rule requested: name=%s, type=%s", req.Name, req.Type)
	}()

	resp = &candyAgent.AlertRuleResponse{}

	// 验证请求
	if req.Name == "" {
		resp.Success = false
		resp.Message = "规则名称不能为空"
		resp.Error = "name is required"
		return resp, fmt.Errorf("name is required")
	}

	if req.Type == "" {
		resp.Success = false
		resp.Message = "规则类型不能为空"
		resp.Error = "type is required"
		return resp, fmt.Errorf("type is required")
	}

	// 获取K8s规则管理器
	k8sManager := alertrule.GetK8sRuleManager()
	if k8sManager == nil {
		resp.Success = false
		resp.Message = "K8s规则管理器未初始化"
		resp.Error = "k8s rule manager not initialized"
		return resp, fmt.Errorf("k8s rule manager not initialized")
	}

	// 从Kubernetes获取规则
	rule, err := k8sManager.GetRule(h.Context, req.Type, req.Name, req.Namespace)
	if err != nil {
		resp.Success = false
		resp.Message = fmt.Sprintf("获取规则失败: %v", err)
		resp.Error = err.Error()
		return resp, err
	}

	// 构建响应
	resp.Success = true
	resp.Message = fmt.Sprintf("获取规则成功: Successfully retrieved %s %s in namespace %s",
		getKindByType(req.Type), req.Name, rule.Namespace)

	// 构建规则信息返回
	resp.Rule = &candyAgent.AlertRule{
		Name:        rule.Name,
		Description: rule.Description,
		ClusterId:   uint32(rule.ClusterID),
		Type:        string(rule.Type),
		Content:     rule.Content,
		ConfigMap:   rule.ConfigMap,
		Namespace:   rule.Namespace,
		Key:         rule.Key,
		Status:      int32(rule.Status),
		CreatedBy:   uint32(rule.CreatedBy),
		Kind:        rule.Kind,
		CreatedAt:   rule.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   rule.UpdatedAt.Format(time.RFC3339),
	}

	return resp, nil
}

// getKindByType 根据规则类型获取资源Kind
func getKindByType(ruleType string) string {
	switch ruleType {
	case "vm":
		return "VMRule"
	case "prometheus":
		return "PrometheusRule"
	default:
		return "Rule"
	}
}
