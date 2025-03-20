package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.mokaz111.com/candy-agent/biz/executor"
	"github.mokaz111.com/candy-agent/biz/model"

	"github.com/cloudwego/hertz/pkg/app"
	candyAgent "github.mokaz111.com/candy-agent/hertz_gen/candyAgent"
)

type ExecuteTaskService struct {
	RequestContext  *app.RequestContext
	Context         context.Context
	executorFactory *executor.ExecutorFactory
	tasks           map[string]*model.Task
	mutex           sync.RWMutex
}

func NewExecuteTaskService(Context context.Context, RequestContext *app.RequestContext, executorFactory *executor.ExecutorFactory) *ExecuteTaskService {
	hlog.Infof("Creating new ExecuteTaskService")
	return &ExecuteTaskService{RequestContext: RequestContext, Context: Context, executorFactory: executorFactory, tasks: make(map[string]*model.Task)}
}

func (s *ExecuteTaskService) Run(req *candyAgent.TaskRequest) (resp *candyAgent.TaskResponse, err error) {
	startTime := time.Now()
	hlog.CtxInfof(s.Context, "Starting task execution: %s with %d items", req.TaskId, len(req.Items))

	// 创建任务响应对象
	resp = &candyAgent.TaskResponse{
		TaskId:    req.TaskId,
		Status:    candyAgent.TaskStatus_TASK_STATUS_COMPLETED,
		Results:   make([]*candyAgent.TaskResult, 0, len(req.Items)),
		StartTime: startTime.Format(time.RFC3339),
	}

	// 创建上下文，支持超时控制
	timeout := time.Duration(req.Timeout) * time.Second
	if timeout <= 0 {
		timeout = 30 * time.Second // 默认30秒超时
	}
	ctx, cancel := context.WithTimeout(s.Context, timeout)
	defer cancel()

	// 执行每个任务项
	for _, item := range req.Items {
		hlog.CtxInfof(ctx, "Processing task item: %s (ID: %d) for task: %s", item.Name, item.Id, req.TaskId)

		// 创建任务项模型
		modelItem := &model.TaskItem{
			ID:     uint(item.Id),
			Name:   item.Name,
			Type:   item.Type,
			Params: make(map[string]interface{}),
		}

		// 转换参数
		for k, v := range item.Params {
			modelItem.Params[k] = v
		}

		// 检查上下文是否已取消
		if ctx.Err() != nil {
			hlog.CtxWarnf(ctx, "Task execution interrupted: %v", ctx.Err())
			resp.Status = candyAgent.TaskStatus_TASK_STATUS_FAILED
			resp.Message = fmt.Sprintf("Task interrupted: %v", ctx.Err())
			break
		}

		// 获取执行器
		executor, err := s.executorFactory.Create(item.Type)
		if err != nil {
			hlog.CtxErrorf(ctx, "Failed to get executor for item %d: %v", item.Id, err)
			resp.Results = append(resp.Results, &candyAgent.TaskResult{
				ItemId:   item.Id,
				Status:   candyAgent.ResultStatus_RESULT_STATUS_FAILED,
				Message:  fmt.Sprintf("Failed to get executor: %v", err),
				Details:  fmt.Sprintf("Error creating executor of type %s: %v", item.Type, err),
				Duration: time.Since(startTime).Milliseconds(),
			})
			continue
		}

		// 执行任务项
		itemStartTime := time.Now()
		result, err := executor.Execute(ctx, *modelItem)
		executionDuration := time.Since(itemStartTime).Milliseconds()

		if err != nil {
			hlog.CtxErrorf(ctx, "Failed to execute item %d: %v", item.Id, err)
			resp.Results = append(resp.Results, &candyAgent.TaskResult{
				ItemId:   item.Id,
				Status:   candyAgent.ResultStatus_RESULT_STATUS_FAILED,
				Message:  fmt.Sprintf("Execution failed: %v", err),
				Details:  fmt.Sprintf("Error executing task item: %v", err),
				Duration: executionDuration,
			})
			continue
		}

		// 将结果添加到响应中
		hlog.CtxInfof(ctx, "Successfully executed item %d for task %s: %s", item.Id, req.TaskId, result.Value)
		resp.Results = append(resp.Results, &candyAgent.TaskResult{
			ItemId:   item.Id,
			Status:   s.convertResultStatus(result.Status),
			Message:  result.Message,
			Details:  result.Details,
			Duration: result.Duration,
		})
	}

	// 设置结束时间
	endTime := time.Now()
	resp.EndTime = endTime.Format(time.RFC3339)
	resp.Message = fmt.Sprintf("Task completed in %d ms", time.Since(startTime).Milliseconds())

	// 根据结果调整任务状态
	if resp.Status != candyAgent.TaskStatus_TASK_STATUS_FAILED {
		// 检查是否有任何任务项失败
		hasFailures := false
		for _, result := range resp.Results {
			if result.Status == candyAgent.ResultStatus_RESULT_STATUS_FAILED {
				hasFailures = true
				break
			}
		}

		if hasFailures {
			// 使用WARNING状态代替COMPLETED_WITH_ERRORS
			resp.Message = "Task completed with some failures"
		}
	}

	hlog.CtxInfof(s.Context, "Task %s completed with status %s in %d ms",
		req.TaskId, resp.Status, time.Since(startTime).Milliseconds())
	return resp, nil
}

// 转换结果状态
func (s *ExecuteTaskService) convertResultStatus(status model.ResultStatus) candyAgent.ResultStatus {
	switch status {
	case model.ResultStatusNormal:
		return candyAgent.ResultStatus_RESULT_STATUS_NORMAL
	case model.ResultStatusWarning:
		return candyAgent.ResultStatus_RESULT_STATUS_WARNING
	case model.ResultStatusCritical:
		return candyAgent.ResultStatus_RESULT_STATUS_FAILED // 将CRITICAL映射到FAILED
	case model.ResultStatusFailed:
		return candyAgent.ResultStatus_RESULT_STATUS_FAILED
	default:
		return candyAgent.ResultStatus_RESULT_STATUS_UNKNOWN
	}
}

// GetTask 获取任务
func (s *ExecuteTaskService) GetTask(taskID string) (*model.Task, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	task, ok := s.tasks[taskID]
	if !ok {
		return nil, fmt.Errorf("task not found: %s", taskID)
	}

	return task, nil
}

// CleanupTasks 清理过期任务
func (s *ExecuteTaskService) CleanupTasks(maxAge time.Duration) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	now := time.Now()
	for id, task := range s.tasks {
		// 如果任务已完成且超过最大保留时间，则删除
		if (task.Status == model.TaskStatusCompleted || task.Status == model.TaskStatusFailed) &&
			task.EndTime.Add(maxAge).Before(now) {
			delete(s.tasks, id)
			hlog.Infof("Cleaned up task %s", id)
		}
	}
}
