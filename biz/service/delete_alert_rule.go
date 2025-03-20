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

type DeleteService struct {
	RequestContext *app.RequestContext
	Context        context.Context
}

func NewDeleteAlertRuleService(Context context.Context, RequestContext *app.RequestContext) *DeleteService {
	hlog.Infof("Creating new delete alert rule service")
	return &DeleteService{
		RequestContext: RequestContext,
		Context:        Context,
	}
}

func (h *DeleteService) Run(req *candyAgent.AlertRuleRequest) (resp *candyAgent.AlertRuleResponse, err error) {
	defer func() {
		hlog.CtxInfof(h.Context, "Delete alert rule completed: name=%s, type=%s", req.Name, req.Type)
	}()

	hlog.CtxInfof(h.Context, "Starting alert rule deletion: name=%s, type=%s", req.Name, req.Type)
	resp = &candyAgent.AlertRuleResponse{}

	// 验证请求
	if req.Name == "" {
		hlog.CtxErrorf(h.Context, "Alert rule deletion failed: name is required")
		resp.Success = false
		resp.Message = "规则名称不能为空"
		resp.Error = "name is required"
		return resp, fmt.Errorf("name is required")
	}

	if req.Type == "" {
		hlog.CtxErrorf(h.Context, "Alert rule deletion failed: type is required")
		resp.Success = false
		resp.Message = "规则类型不能为空"
		resp.Error = "type is required"
		return resp, fmt.Errorf("type is required")
	}

	// 获取K8s规则管理器
	k8sManager := alertrule.GetK8sRuleManager()
	if k8sManager == nil {
		hlog.CtxErrorf(h.Context, "Alert rule deletion failed: k8s rule manager not initialized")
		resp.Success = false
		resp.Message = "K8s规则管理器未初始化"
		resp.Error = "k8s rule manager not initialized"
		return resp, fmt.Errorf("k8s rule manager not initialized")
	}

	// 获取现有规则
	existingRule, err := k8sManager.GetRule(h.Context, req.Type, req.Name, req.Namespace)
	if err != nil {
		hlog.CtxErrorf(h.Context, "Failed to get existing rule: %v", err)
		resp.Success = false
		resp.Message = "获取现有规则失败"
		resp.Error = err.Error()
		return resp, err
	}

	hlog.CtxInfof(h.Context, "Found existing rule to delete: ID=%d, name=%s", existingRule.ID, existingRule.Name)

	// 删除规则
	err = k8sManager.DeleteRule(h.Context, req.Type, req.Name, req.Namespace)
	if err != nil {
		hlog.CtxErrorf(h.Context, "Failed to delete alert rule: %v", err)
		resp.Success = false
		resp.Message = "删除规则失败"
		resp.Error = err.Error()
		return resp, err
	}

	hlog.CtxInfof(h.Context, "Successfully deleted alert rule: ID=%d, name=%s", existingRule.ID, existingRule.Name)

	// 构建响应
	resp.Success = true
	resp.Message = "删除规则成功"
	resp.Rule = &candyAgent.AlertRule{
		Name:        existingRule.Name,
		Description: existingRule.Description,
		ClusterId:   uint32(existingRule.ClusterID),
		Type:        string(existingRule.Type),
		Content:     existingRule.Content,
		ConfigMap:   existingRule.ConfigMap,
		Namespace:   existingRule.Namespace,
		Key:         existingRule.Key,
		Status:      int32(existingRule.Status),
		CreatedBy:   uint32(existingRule.CreatedBy),
		Kind:        existingRule.Kind,
		CreatedAt:   existingRule.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   existingRule.UpdatedAt.Format(time.RFC3339),
	}
	return resp, nil
}
