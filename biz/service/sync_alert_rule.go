package service

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.mokaz111.com/candy-agent/biz/dal/alertrule"
	"github.mokaz111.com/candy-agent/biz/model"
	candyAgent "github.mokaz111.com/candy-agent/hertz_gen/candyAgent"
)

type SyncService struct {
	RequestContext *app.RequestContext
	Context        context.Context
}

func NewSyncAlertRuleService(Context context.Context, RequestContext *app.RequestContext) *SyncService {
	return &SyncService{
		RequestContext: RequestContext,
		Context:        Context,
	}
}

func (h *SyncService) Run(req *candyAgent.AlertRuleRequest) (resp *candyAgent.AlertRuleResponse, err error) {
	hlog.CtxInfof(h.Context, "Starting to sync all alert rules from Kubernetes cluster")

	resp = &candyAgent.AlertRuleResponse{}

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

	hlog.CtxInfof(h.Context, "Using label selector: %s", labelSelector)

	// 定义要查询的规则类型
	ruleTypes := []string{"vm", "prometheus"}

	// 初始化规则列表
	allRules := make([]*candyAgent.AlertRule, 0)

	// 从Kubernetes获取各类型的规则列表
	for _, ruleType := range ruleTypes {
		hlog.CtxInfof(h.Context, "Syncing alert rules of type: %s", ruleType)

		// 确定命名空间
		namespace := req.Namespace
		if namespace == "" {
			namespace = "monitoring" // 默认命名空间
		}

		// 获取当前类型的规则列表
		rules, err := k8sManager.ListRules(h.Context, ruleType, namespace, labelSelector)
		if err != nil {
			hlog.CtxWarnf(h.Context, "Failed to get rules of type %s: %v", ruleType, err)
			// 不要因为一种类型的规则获取失败而终止整个同步过程
			continue
		}

		hlog.CtxInfof(h.Context, "Found %d rules of type %s", len(rules), ruleType)

		// 将获取到的规则转换为响应格式并添加到总列表中
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
			allRules = append(allRules, alertRule)
		}
	}

	// 构建响应
	resp.Success = true
	resp.Message = fmt.Sprintf("同步告警规则成功: Found %d rules in total", len(allRules))
	resp.Rules = allRules

	hlog.CtxInfof(h.Context, "Alert rule sync completed: found %d rules in total", len(allRules))

	return resp, nil
}

// convertToProtoRule 将模型规则转换为proto规则
func convertToProtoRule(rule *model.AlertRule) *candyAgent.AlertRule {
	return &candyAgent.AlertRule{
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
}
