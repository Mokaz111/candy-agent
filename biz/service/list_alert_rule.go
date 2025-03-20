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

type ListService struct {
	RequestContext *app.RequestContext
	Context        context.Context
}

func NewListAlertRuleService(Context context.Context, RequestContext *app.RequestContext) *ListService {
	return &ListService{
		RequestContext: RequestContext,
		Context:        Context,
	}
}

func (h *ListService) Run(req *candyAgent.AlertRuleRequest) (resp *candyAgent.AlertRuleResponse, err error) {
	defer func() {
		hlog.CtxInfof(h.Context, "List alert rules requested: type=%s, namespace=%s", req.Type, req.Namespace)
	}()

	resp = &candyAgent.AlertRuleResponse{}

	// 验证请求
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

	// 构建标签选择器
	labelSelector := "app=candy-agent,managed-by=candy-agent"
	if req.ConfigMap != "" {
		labelSelector = fmt.Sprintf("%s,config-map=%s", labelSelector, req.ConfigMap)
	}

	hlog.CtxInfof(h.Context, "Getting alert rules from Kubernetes with type=%s, namespace=%s, labelSelector=%s",
		req.Type, req.Namespace, labelSelector)

	// 从Kubernetes获取规则列表
	rules, err := k8sManager.ListRules(h.Context, req.Type, req.Namespace, labelSelector)
	if err != nil {
		resp.Success = false
		resp.Message = fmt.Sprintf("获取规则列表失败: %v", err)
		resp.Error = err.Error()
		return resp, err
	}

	// 构建响应
	resp.Success = true
	resp.Message = fmt.Sprintf("获取规则列表成功: Found %d %ss in namespace %s",
		len(rules), getKindByType(req.Type), req.Namespace)
	resp.Rules = make([]*candyAgent.AlertRule, 0, len(rules))

	// 将规则列表转换为响应格式
	for _, rule := range rules {
		alertRule := &candyAgent.AlertRule{
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
		resp.Rules = append(resp.Rules, alertRule)
	}

	return resp, nil
}
