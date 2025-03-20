package service

import (
	"context"
	"fmt"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	candyAgent "github.mokaz111.com/candy-agent/hertz_gen/candyAgent"
)

type CancelTaskService struct {
	RequestContext *app.RequestContext
	Context        context.Context
}

func NewCancelTaskService(Context context.Context, RequestContext *app.RequestContext) *CancelTaskService {
	return &CancelTaskService{RequestContext: RequestContext, Context: Context}
}

func (h *CancelTaskService) Run(req *candyAgent.TaskCancelRequest) (resp *candyAgent.TaskCancelResponse, err error) {
	defer func() {
		hlog.CtxInfof(h.Context, "req = %+v", req)
		hlog.CtxInfof(h.Context, "resp = %+v", resp)
	}()

	if req.TaskId == "" {
		return &candyAgent.TaskCancelResponse{
			TaskId:  req.TaskId,
			Success: false,
			Message: "任务ID不能为空",
		}, nil
	}

	// 从任务管理器取消任务
	taskManager := GetTaskManager()
	err = taskManager.CancelTask(req.TaskId)
	if err != nil {
		return &candyAgent.TaskCancelResponse{
			TaskId:  req.TaskId,
			Success: false,
			Message: fmt.Sprintf("取消任务失败: %v", err),
		}, nil
	}

	// 返回成功响应
	resp = &candyAgent.TaskCancelResponse{
		TaskId:  req.TaskId,
		Success: true,
		Message: "任务已成功取消",
	}

	return resp, nil
}
