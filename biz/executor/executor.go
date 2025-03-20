package executor

import (
	"context"
	"github.mokaz111.com/candy-agent/biz/model"
)

// Executor 执行器接口
type Executor interface {
	// Execute 执行巡检项
	Execute(ctx context.Context, item model.TaskItem) (model.TaskResult, error)
	// Name 执行器名称
	Name() string
}

// Factory 执行器工厂
type Factory interface {
	// Create 创建执行器
	Create(executorType string) (Executor, error)
}
