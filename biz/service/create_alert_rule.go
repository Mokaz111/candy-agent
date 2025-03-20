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

type CreateService struct {
	RequestContext *app.RequestContext
	Context        context.Context
}

func NewCreateAlertRuleService(Context context.Context, RequestContext *app.RequestContext) *CreateService {
	hlog.Infof("Creating new alert rule service")
	return &CreateService{
		RequestContext: RequestContext,
		Context:        Context,
	}
}

func (h *CreateService) Run(req *candyAgent.AlertRuleRequest) (resp *candyAgent.AlertRuleResponse, err error) {
	defer func() {
		hlog.CtxInfof(h.Context, "Create alert rule completed: name=%s, type=%s", req.Name, req.Type)
	}()

	hlog.CtxInfof(h.Context, "Starting alert rule creation: name=%s, type=%s", req.Name, req.Type)
	resp = &candyAgent.AlertRuleResponse{}

	// 验证请求
	if req.Name == "" {
		hlog.CtxErrorf(h.Context, "Alert rule creation failed: name is required")
		resp.Success = false
		resp.Message = "规则名称不能为空"
		resp.Error = "name is required"
		return resp, fmt.Errorf("name is required")
	}

	if req.Type == "" {
		hlog.CtxErrorf(h.Context, "Alert rule creation failed: type is required")
		resp.Success = false
		resp.Message = "规则类型不能为空"
		resp.Error = "type is required"
		return resp, fmt.Errorf("type is required")
	}

	if req.Content == "" {
		hlog.CtxErrorf(h.Context, "Alert rule creation failed: content is required")
		resp.Success = false
		resp.Message = "规则内容不能为空"
		resp.Error = "content is required"
		return resp, fmt.Errorf("content is required")
	}

	// 创建规则
	rule := &model.AlertRule{
		Name:        req.Name,
		Description: req.Name, // 使用名称作为描述
		ClusterID:   1,        // 默认集群ID
		Type:        model.AlertRuleType(req.Type),
		Content:     req.Content,
		ConfigMap:   req.ConfigMap,
		Namespace:   req.Namespace,
		Key:         req.Key,
		Status:      model.AlertRuleStatusEnabled,
		CreatedBy:   1, // 默认用户ID
		Kind:        "AlertRule",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 保存规则
	err = alertrule.GetK8sRuleManager().CreateRule(h.Context, rule)
	if err != nil {
		hlog.CtxErrorf(h.Context, "Failed to create alert rule: %v", err)
		resp.Success = false
		resp.Message = "创建规则失败"
		resp.Error = err.Error()
		return resp, err
	}

	hlog.CtxInfof(h.Context, "Successfully created alert rule: ID=%d, name=%s", rule.ID, rule.Name)

	// 构建响应
	resp.Success = true
	resp.Message = "创建规则成功"
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
