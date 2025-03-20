package candyAgent

import (
	"context"
	"fmt"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	dal "github.mokaz111.com/candy-agent/biz/dal/executor"
	"github.mokaz111.com/candy-agent/biz/service"
	"github.mokaz111.com/candy-agent/biz/utils"
	candyAgent "github.mokaz111.com/candy-agent/hertz_gen/candyAgent"
)

// ExecuteTask .
// @router /api/task [POST]
func ExecuteTask(ctx context.Context, c *app.RequestContext) {
	var err error
	var req candyAgent.TaskRequest
	err = c.BindAndValidate(&req)
	if err != nil {
		hlog.Errorf("Failed to bind task request: %v", err)
		utils.SendErrResponse(ctx, c, consts.StatusOK, err)
		return
	}
	hlog.Infof("Received task request: %+v", req.TaskId)

	// 验证请求
	if req.TaskId == "" {
		hlog.Errorf("Missing task_id in request")
		utils.SendErrResponse(ctx, c, consts.StatusOK, fmt.Errorf("task_id is required"))
		return
	}

	if len(req.Items) == 0 {
		hlog.Errorf("No task items in request")
		utils.SendErrResponse(ctx, c, consts.StatusOK, fmt.Errorf("at least one task item is required"))
		return
	}

	// 设置默认超时
	if req.Timeout <= 0 {
		req.Timeout = 30 // 默认30秒
	}

	e := dal.ExecutorFactory
	// 同步执行任务
	resp, err := service.NewExecuteTaskService(ctx, c, e).Run(&req)
	if err != nil {
		hlog.Errorf("Failed to execute task: %v", err)
		utils.SendErrResponse(ctx, c, consts.StatusOK, err)
		return
	}
	hlog.Infof("Successfully executed task %s with %d results", req.TaskId, len(resp.Results))
	utils.SendSuccessResponse(ctx, c, consts.StatusOK, resp)
}

// SendHeartbeat .
// @router /api/heartbeat [POST]
func SendHeartbeat(ctx context.Context, c *app.RequestContext) {
	var err error
	var req candyAgent.HeartbeatRequest
	err = c.BindAndValidate(&req)
	if err != nil {
		hlog.Errorf("Failed to bind heartbeat request: %v", err)
		utils.SendErrResponse(ctx, c, consts.StatusOK, err)
		return
	}
	hlog.Infof("Received heartbeat request from agent: %s", req.AgentId)

	resp, err := service.NewSendHeartbeatService(ctx, c).Run(&req)
	if err != nil {
		hlog.Errorf("Failed to process heartbeat: %v", err)
		utils.SendErrResponse(ctx, c, consts.StatusOK, err)
		return
	}
	hlog.Infof("Successfully processed heartbeat from agent: %s", req.AgentId)
	utils.SendSuccessResponse(ctx, c, consts.StatusOK, resp)
}

// UpdateConfig .
// @router /api/config [POST]
func UpdateConfig(ctx context.Context, c *app.RequestContext) {
	var err error
	var req candyAgent.ConfigUpdateRequest
	err = c.BindAndValidate(&req)
	if err != nil {
		hlog.Errorf("Failed to bind config update request: %v", err)
		utils.SendErrResponse(ctx, c, consts.StatusOK, err)
		return
	}
	hlog.Infof("Received config update request for agent: %s")

	resp, err := service.NewUpdateConfigService(ctx, c).Run(&req)
	if err != nil {
		hlog.Errorf("Failed to update config: %v", err)
		utils.SendErrResponse(ctx, c, consts.StatusOK, err)
		return
	}
	hlog.Infof("Successfully updated config for agent: %s")
	utils.SendSuccessResponse(ctx, c, consts.StatusOK, resp)
}

// HealthCheck .
// @router /health [GET]
func HealthCheck(ctx context.Context, c *app.RequestContext) {
	var err error
	var req candyAgent.Empty
	err = c.BindAndValidate(&req)
	if err != nil {
		hlog.Errorf("Failed to bind health check request: %v", err)
		utils.SendErrResponse(ctx, c, consts.StatusOK, err)
		return
	}
	hlog.Infof("Received health check request")

	resp, err := service.NewHealthCheckService(ctx, c).Run(&req)
	if err != nil {
		hlog.Errorf("Failed to process health check: %v", err)
		utils.SendErrResponse(ctx, c, consts.StatusOK, err)
		return
	}
	hlog.Infof("Health check completed successfully")
	utils.SendSuccessResponse(ctx, c, consts.StatusOK, resp)
}

// GetTaskStatus .
// @router /api/v1/tasks/:task_id [GET]
func GetTaskStatus(ctx context.Context, c *app.RequestContext) {
	var err error
	var req candyAgent.TaskStatusRequest
	err = c.BindAndValidate(&req)
	if err != nil {
		utils.SendErrResponse(ctx, c, consts.StatusOK, err)
		return
	}

	resp, err := service.NewGetTaskStatusService(ctx, c).Run(&req)

	if err != nil {
		utils.SendErrResponse(ctx, c, consts.StatusOK, err)
		return
	}
	utils.SendSuccessResponse(ctx, c, consts.StatusOK, resp)
}

// CancelTask .
// @router /api/v1/tasks/:task_id [DELETE]
func CancelTask(ctx context.Context, c *app.RequestContext) {
	var err error
	var req candyAgent.TaskCancelRequest
	err = c.BindAndValidate(&req)
	if err != nil {
		utils.SendErrResponse(ctx, c, consts.StatusOK, err)
		return
	}

	resp, err := service.NewCancelTaskService(ctx, c).Run(&req)

	if err != nil {
		utils.SendErrResponse(ctx, c, consts.StatusOK, err)
		return
	}
	utils.SendSuccessResponse(ctx, c, consts.StatusOK, resp)
}

// ReadyCheck .
// @router /ready [GET]
func ReadyCheck(ctx context.Context, c *app.RequestContext) {
	var err error
	var req candyAgent.Empty
	err = c.BindAndValidate(&req)
	if err != nil {
		utils.SendErrResponse(ctx, c, consts.StatusOK, err)
		return
	}

	resp, err := service.NewReadyCheckService(ctx, c).Run(&req)

	if err != nil {
		utils.SendErrResponse(ctx, c, consts.StatusOK, err)
		return
	}
	utils.SendSuccessResponse(ctx, c, consts.StatusOK, resp)
}
