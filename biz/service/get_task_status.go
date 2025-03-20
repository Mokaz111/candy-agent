package service

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	candyAgent "github.mokaz111.com/candy-agent/hertz_gen/candyAgent"
)

type GetTaskStatusService struct {
	RequestContext *app.RequestContext
	Context        context.Context
}

func NewGetTaskStatusService(Context context.Context, RequestContext *app.RequestContext) *GetTaskStatusService {
	return &GetTaskStatusService{RequestContext: RequestContext, Context: Context}
}

func (h *GetTaskStatusService) Run(req *candyAgent.TaskStatusRequest) (resp *candyAgent.TaskStatusResponse, err error) {
	defer func() {
		hlog.CtxInfof(h.Context, "req = %+v", req)
		hlog.CtxInfof(h.Context, "resp = %+v", resp)
	}()

	if req.TaskId == "" {
		return nil, fmt.Errorf("任务ID不能为空")
	}

	// 从任务管理器获取任务状态
	taskManager := GetTaskManager()
	task, err := taskManager.GetTask(req.TaskId)
	if err != nil {
		return nil, fmt.Errorf("获取任务状态失败: %v", err)
	}

	// 如果找不到任务
	if task == nil {
		return nil, fmt.Errorf("任务不存在: %s", req.TaskId)
	}

	// 将任务状态转换为响应
	resp = &candyAgent.TaskStatusResponse{
		TaskId:    task.ID,
		StartTime: task.StartTime.Format(time.RFC3339),
		Message:   task.Error,
	}

	// 将任务状态转换为枚举值
	switch task.Status {
	case "pending":
		resp.Status = candyAgent.TaskStatus_TASK_STATUS_PENDING
	case "running":
		resp.Status = candyAgent.TaskStatus_TASK_STATUS_RUNNING
	case "completed":
		resp.Status = candyAgent.TaskStatus_TASK_STATUS_COMPLETED
	case "failed":
		resp.Status = candyAgent.TaskStatus_TASK_STATUS_FAILED
	case "canceled":
		resp.Status = candyAgent.TaskStatus_TASK_STATUS_CANCELED
	default:
		resp.Status = candyAgent.TaskStatus_TASK_STATUS_UNKNOWN
	}

	return resp, nil
}
