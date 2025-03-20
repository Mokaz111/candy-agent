package model

import (
	"sync"
	"time"
)

// TaskStatus 任务状态
type TaskStatus string

const (
	// TaskStatusPending 等待执行
	TaskStatusPending TaskStatus = "pending"
	// TaskStatusRunning 执行中
	TaskStatusRunning TaskStatus = "running"
	// TaskStatusCompleted 已完成
	TaskStatusCompleted TaskStatus = "completed"
	// TaskStatusFailed 执行失败
	TaskStatusFailed TaskStatus = "failed"
	// TaskStatusCanceled 已取消
	TaskStatusCanceled TaskStatus = "canceled"
)

// ResultStatus 结果状态
type ResultStatus string

const (
	// ResultStatusNormal 正常
	ResultStatusNormal ResultStatus = "normal"
	// ResultStatusWarning 警告
	ResultStatusWarning ResultStatus = "warning"
	// ResultStatusCritical 严重
	ResultStatusCritical ResultStatus = "critical"
	// ResultStatusFailed 失败
	ResultStatusFailed ResultStatus = "failed"
)

// TaskItem 任务项
type TaskItem struct {
	ID     uint                   `json:"id"`
	Name   string                 `json:"name"`
	Type   string                 `json:"type"`
	Params map[string]interface{} `json:"params"`
}

// TaskResult 任务结果
type TaskResult struct {
	ItemID   uint         `json:"item_id"`
	Status   ResultStatus `json:"status"`
	Value    string       `json:"value"`
	Message  string       `json:"message"`
	Details  string       `json:"details"`
	Duration int64        `json:"duration"`
}

// Task 任务对象，包含任务信息和执行结果
type Task struct {
	ID        string        `json:"task_id"`
	Status    TaskStatus    `json:"status"`
	Items     []TaskItem    `json:"items"`
	Results   []TaskResult  `json:"results"`
	Error     string        `json:"error,omitempty"`
	StartTime time.Time     `json:"start_time"`
	EndTime   time.Time     `json:"end_time,omitempty"`
	Timeout   int           `json:"timeout"`
	Cancel    chan struct{} `json:"-"` // 用于发送取消信号
}

// TaskCallback 任务回调信息
type TaskCallback struct {
	TaskID    string       `json:"task_id"`
	Status    TaskStatus   `json:"status"`
	Results   []TaskResult `json:"results"`
	StartTime time.Time    `json:"start_time"`
	EndTime   time.Time    `json:"end_time"`
	Error     string       `json:"error,omitempty"`
}

// TaskCache 任务缓存，保存所有任务状态和结果
type TaskCache struct {
	tasks map[string]*Task
	mutex sync.RWMutex
	// 清理配置
	maxAge      time.Duration // 最大保留时间
	cleanTicker *time.Ticker  // 清理定时器
}

// NewTaskCache 创建新的任务缓存
func NewTaskCache() *TaskCache {
	cache := &TaskCache{
		tasks:       make(map[string]*Task),
		maxAge:      24 * time.Hour, // 默认保留24小时
		cleanTicker: time.NewTicker(1 * time.Hour),
	}

	// 启动定时清理
	go cache.startCleaner()

	return cache
}

// startCleaner 启动缓存清理器
func (c *TaskCache) startCleaner() {
	for range c.cleanTicker.C {
		c.cleanExpiredTasks()
	}
}

// cleanExpiredTasks 清理过期任务
func (c *TaskCache) cleanExpiredTasks() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	now := time.Now()
	for id, task := range c.tasks {
		// 清理已完成且过期的任务
		if (task.Status == TaskStatusCompleted ||
			task.Status == TaskStatusFailed ||
			task.Status == TaskStatusCanceled) &&
			task.EndTime.Add(c.maxAge).Before(now) {
			delete(c.tasks, id)
		}
	}
}

// AddTask 添加任务
func (c *TaskCache) AddTask(task *Task) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.tasks[task.ID] = task
}

// GetTask 获取任务
func (c *TaskCache) GetTask(taskID string) (*Task, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	task, exists := c.tasks[taskID]
	return task, exists
}

// UpdateTaskStatus 更新任务状态
func (c *TaskCache) UpdateTaskStatus(taskID string, status TaskStatus) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	task, exists := c.tasks[taskID]
	if !exists {
		return false
	}

	task.Status = status
	if status == TaskStatusCompleted || status == TaskStatusFailed || status == TaskStatusCanceled {
		task.EndTime = time.Now()
	}

	return true
}

// UpdateTaskResult 更新任务结果
func (c *TaskCache) UpdateTaskResult(taskID string, results []TaskResult) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	task, exists := c.tasks[taskID]
	if !exists {
		return false
	}

	task.Results = results
	return true
}

// CancelTask 取消任务
func (c *TaskCache) CancelTask(taskID string) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	task, exists := c.tasks[taskID]
	if !exists || task.Status != TaskStatusRunning {
		return false
	}

	// 发送取消信号
	close(task.Cancel)
	task.Status = TaskStatusCanceled
	task.EndTime = time.Now()

	return true
}
