package candyAgent

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.mokaz111.com/candy-agent/biz/service"
	"github.mokaz111.com/candy-agent/biz/utils"
	candyAgent "github.mokaz111.com/candy-agent/hertz_gen/candyAgent"
)

// CreateAlertRule .
// @router /api/alert-rule/create [POST]
func CreateAlertRule(ctx context.Context, c *app.RequestContext) {
	var err error
	var req candyAgent.AlertRuleRequest
	err = c.BindAndValidate(&req)
	if err != nil {
		hlog.Errorf("Failed to bind alert rule creation request: %v", err)
		utils.SendErrResponse(ctx, c, consts.StatusOK, err)
		return
	}
	hlog.Infof("Received alert rule creation request: %s", req.RuleId)

	resp, err := service.NewCreateAlertRuleService(ctx, c).Run(&req)

	if err != nil {
		hlog.Errorf("Failed to create alert rule: %v", err)
		utils.SendErrResponse(ctx, c, consts.StatusOK, err)
		return
	}
	hlog.Infof("Successfully created alert rule with ID: %s", req.RuleId)
	utils.SendSuccessResponse(ctx, c, consts.StatusOK, resp)
}

// UpdateAlertRule .
// @router /api/alert-rule/update [POST]
func UpdateAlertRule(ctx context.Context, c *app.RequestContext) {
	var err error
	var req candyAgent.AlertRuleRequest
	err = c.BindAndValidate(&req)
	if err != nil {
		hlog.Errorf("Failed to bind alert rule update request: %v", err)
		utils.SendErrResponse(ctx, c, consts.StatusOK, err)
		return
	}
	hlog.Infof("Received alert rule update request for rule ID: %s", req.RuleId)

	resp, err := service.NewUpdateAlertRuleService(ctx, c).Run(&req)

	if err != nil {
		hlog.Errorf("Failed to update alert rule: %v", err)
		utils.SendErrResponse(ctx, c, consts.StatusOK, err)
		return
	}
	hlog.Infof("Successfully updated alert rule with ID: %s", req.RuleId)
	utils.SendSuccessResponse(ctx, c, consts.StatusOK, resp)
}

// DeleteAlertRule .
// @router /api/alert-rule/delete [POST]
func DeleteAlertRule(ctx context.Context, c *app.RequestContext) {
	var err error
	var req candyAgent.AlertRuleRequest
	err = c.BindAndValidate(&req)
	if err != nil {
		hlog.Errorf("Failed to bind alert rule deletion request: %v", err)
		utils.SendErrResponse(ctx, c, consts.StatusOK, err)
		return
	}
	hlog.Infof("Received alert rule deletion request for rule ID: %s", req.RuleId)

	resp, err := service.NewDeleteAlertRuleService(ctx, c).Run(&req)

	if err != nil {
		hlog.Errorf("Failed to delete alert rule: %v", err)
		utils.SendErrResponse(ctx, c, consts.StatusOK, err)
		return
	}
	hlog.Infof("Successfully deleted alert rule with ID: %s", req.RuleId)
	utils.SendSuccessResponse(ctx, c, consts.StatusOK, resp)
}

// GetAlertRule .
// @router /api/alert-rule/get [POST]
func GetAlertRule(ctx context.Context, c *app.RequestContext) {
	var err error
	var req candyAgent.AlertRuleRequest
	err = c.BindAndValidate(&req)
	if err != nil {
		hlog.Errorf("Failed to bind alert rule get request: %v", err)
		utils.SendErrResponse(ctx, c, consts.StatusOK, err)
		return
	}
	hlog.Infof("Received alert rule get request for rule ID: %s", req.RuleId)

	resp, err := service.NewGetAlertRuleService(ctx, c).Run(&req)

	if err != nil {
		hlog.Errorf("Failed to get alert rule: %v", err)
		utils.SendErrResponse(ctx, c, consts.StatusOK, err)
		return
	}
	hlog.Infof("Successfully retrieved alert rule with ID: %s", req.RuleId)
	utils.SendSuccessResponse(ctx, c, consts.StatusOK, resp)
}

// ListAlertRule .
// @router /api/alert-rule/list [POST]
func ListAlertRule(ctx context.Context, c *app.RequestContext) {
	var err error
	var req candyAgent.AlertRuleRequest
	err = c.BindAndValidate(&req)
	if err != nil {
		hlog.Errorf("Failed to bind alert rule list request: %v", err)
		utils.SendErrResponse(ctx, c, consts.StatusOK, err)
		return
	}
	hlog.Infof("Received alert rule list request")

	resp, err := service.NewListAlertRuleService(ctx, c).Run(&req)

	if err != nil {
		hlog.Errorf("Failed to list alert rules: %v", err)
		utils.SendErrResponse(ctx, c, consts.StatusOK, err)
		return
	}
	hlog.Infof("Successfully retrieved alert rule list")
	utils.SendSuccessResponse(ctx, c, consts.StatusOK, resp)
}

// SyncAlertRule .
// @router /api/alert-rule/sync [POST]
func SyncAlertRule(ctx context.Context, c *app.RequestContext) {
	var err error
	var req candyAgent.AlertRuleRequest
	err = c.BindAndValidate(&req)
	if err != nil {
		hlog.Errorf("Failed to bind alert rule sync request: %v", err)
		utils.SendErrResponse(ctx, c, consts.StatusOK, err)
		return
	}
	hlog.Infof("Received alert rule sync request")

	resp, err := service.NewSyncAlertRuleService(ctx, c).Run(&req)

	if err != nil {
		hlog.Errorf("Failed to sync alert rules: %v", err)
		utils.SendErrResponse(ctx, c, consts.StatusOK, err)
		return
	}
	hlog.Infof("Successfully synced alert rules")
	utils.SendSuccessResponse(ctx, c, consts.StatusOK, resp)
}
