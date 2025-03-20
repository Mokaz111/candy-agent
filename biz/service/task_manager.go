package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.mokaz111.com/candy-agent/biz/executor"
	"github.mokaz111.com/candy-agent/biz/model"
	"github.mokaz111.com/candy-agent/conf"
)

// TaskManager 任务管理器，负责任务的异步执行和管理
type TaskManager struct {
	cache           *model.TaskCache
	executorFactory *executor.ExecutorFactory
	workerPool      chan struct{} // 控制并发执行的任务数
	callbackURL     string
	mutex           sync.RWMutex
}

var (
	taskManagerInstance *TaskManager
	taskManagerOnce     sync.Once
)

// GetTaskManager 获取任务管理器单例
func GetTaskManager() *TaskManager {
	taskManagerOnce.Do(func() {
		maxWorkers := conf.GetConf().TaskManager.MaxWorkers
		if maxWorkers <= 0 {
			maxWorkers = 10 // 默认最大10个并发任务
		}

		taskManagerInstance = &TaskManager{
			cache:           model.NewTaskCache(),
			executorFactory: executor.GetExecutorFactory(),
			workerPool:      make(chan struct{}, maxWorkers),
			callbackURL:     conf.GetConf().Server.CallbackURL,
		}
	})
	return taskManagerInstance
}

// CreateTask 创建并开始执行任务
func (tm *TaskManager) CreateTask(taskID string, items []model.TaskItem, timeout int) (*model.Task, error) {
	// 检查任务是否已存在
	if existingTask, exists := tm.cache.GetTask(taskID); exists {
		return existingTask, fmt.Errorf("任务已存在: %s", taskID)
	}

	// 创建任务对象
	task := &model.Task{
		ID:        taskID,
		Status:    model.TaskStatusPending,
		Items:     items,
		Results:   make([]model.TaskResult, 0),
		StartTime: time.Now(),
		Timeout:   timeout,
		Cancel:    make(chan struct{}),
	}

	// 保存到缓存
	tm.cache.AddTask(task)

	// 异步执行任务
	go tm.executeTask(task)

	return task, nil
}

// GetTask 获取任务状态和结果
func (tm *TaskManager) GetTask(taskID string) (*model.Task, error) {
	task, exists := tm.cache.GetTask(taskID)
	if !exists {
		return nil, fmt.Errorf("任务不存在: %s", taskID)
	}
	return task, nil
}

// CancelTask 取消任务
func (tm *TaskManager) CancelTask(taskID string) error {
	if success := tm.cache.CancelTask(taskID); !success {
		return fmt.Errorf("取消任务失败: %s，任务可能不存在或已完成", taskID)
	}
	return nil
}

// executeTask 异步执行任务
func (tm *TaskManager) executeTask(task *model.Task) {
	// 获取工作池令牌
	tm.workerPool <- struct{}{}
	defer func() {
		// 释放工作池令牌
		<-tm.workerPool
	}()

	// 更新任务状态为执行中
	tm.cache.UpdateTaskStatus(task.ID, model.TaskStatusRunning)

	// 创建带超时的上下文
	timeout := time.Duration(task.Timeout) * time.Second
	if timeout <= 0 {
		timeout = 30 * time.Second // 默认30秒
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// 创建取消监听goroutine
	go func() {
		select {
		case <-task.Cancel:
			// 任务被手动取消
			cancel()
		case <-ctx.Done():
			// 上下文超时或已取消，不需要额外操作
		}
	}()

	results := make([]model.TaskResult, 0, len(task.Items))

	for _, item := range task.Items {
		// 检查上下文是否已取消
		if ctx.Err() != nil {
			break
		}

		// 执行单个任务项
		result, err := tm.executeTaskItem(ctx, item)
		if err != nil {
			hlog.Errorf("执行任务项失败: %v", err)
			results = append(results, model.TaskResult{
				ItemID:   item.ID,
				Status:   model.ResultStatusFailed,
				Message:  fmt.Sprintf("执行失败: %v", err),
				Details:  fmt.Sprintf("任务项执行错误: %v", err),
				Duration: 0,
			})
			continue
		}

		results = append(results, result)
	}

	// 更新任务结果
	tm.cache.UpdateTaskResult(task.ID, results)

	// 根据上下文判断任务状态
	if ctx.Err() == context.Canceled {
		// 如果是被取消，状态已更新为Canceled
		if task.Status != model.TaskStatusCanceled {
			tm.cache.UpdateTaskStatus(task.ID, model.TaskStatusCanceled)
		}
	} else if ctx.Err() == context.DeadlineExceeded {
		// 任务超时
		tm.cache.UpdateTaskStatus(task.ID, model.TaskStatusFailed)
		task.Error = "任务执行超时"
	} else {
		// 任务正常完成
		tm.cache.UpdateTaskStatus(task.ID, model.TaskStatusCompleted)
	}

	// 发送回调
	tm.sendCallback(task)
}

// executeTaskItem 执行单个任务项
func (tm *TaskManager) executeTaskItem(ctx context.Context, item model.TaskItem) (model.TaskResult, error) {
	startTime := time.Now()

	// 获取执行器
	executor, err := tm.executorFactory.Create(item.Type)
	if err != nil {
		return model.TaskResult{}, fmt.Errorf("创建执行器失败: %v", err)
	}

	// 执行任务
	executionResult, err := executor.Execute(ctx, item)
	if err != nil {
		return model.TaskResult{}, err
	}

	// 计算执行时间
	duration := time.Since(startTime).Milliseconds()

	// 构建结果
	result := model.TaskResult{
		ItemID:   item.ID,
		Status:   executionResult.Status,
		Value:    executionResult.Value,
		Message:  executionResult.Message,
		Details:  executionResult.Details,
		Duration: duration,
	}

	return result, nil
}

// sendCallback 发送回调通知Server任务完成
func (tm *TaskManager) sendCallback(task *model.Task) {
	if tm.callbackURL == "" {
		hlog.Warn("回调URL未配置，跳过回调")
		return
	}

	// 准备回调数据
	callback := model.TaskCallback{
		TaskID:    task.ID,
		Status:    task.Status,
		Results:   task.Results,
		StartTime: task.StartTime,
		EndTime:   task.EndTime,
		Error:     task.Error,
	}

	// 序列化数据
	data, err := json.Marshal(callback)
	if err != nil {
		hlog.Errorf("序列化回调数据失败: %v", err)
		return
	}

	// 发送HTTP请求
	req, err := http.NewRequest(http.MethodPost, tm.callbackURL, bytes.NewBuffer(data))
	if err != nil {
		hlog.Errorf("创建回调请求失败: %v", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	// 设置API Key (如果配置了)
	if apiKey := conf.GetConf().Server.APIKey; apiKey != "" {
		req.Header.Set("X-API-Key", apiKey)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		hlog.Errorf("发送回调请求失败: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		hlog.Errorf("回调请求返回非200状态码: %d", resp.StatusCode)
		return
	}

	hlog.Infof("成功发送任务回调: %s", task.ID)
}
