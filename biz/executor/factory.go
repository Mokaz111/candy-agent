package executor

import (
	"fmt"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"sync"

	"github.mokaz111.com/candy-agent/conf"
)

// ExecutorFactory 执行器工厂实现
type ExecutorFactory struct {
	executors map[string]Executor
	mu        sync.RWMutex
}

// NewExecutorFactory 创建执行器工厂
func NewExecutorFactory() *ExecutorFactory {
	return &ExecutorFactory{
		executors: make(map[string]Executor),
	}
}

// Register 注册执行器
func (f *ExecutorFactory) Register(executor Executor) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.executors[executor.Name()] = executor
	hlog.Infof("Registered executor: %s", executor.Name())
}

// Create 创建执行器
func (f *ExecutorFactory) Create(executorType string) (Executor, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	executor, ok := f.executors[executorType]
	if !ok {
		return nil, fmt.Errorf("executor not found: %s", executorType)
	}

	return executor, nil
}

var (
	factory *ExecutorFactory
	once    sync.Once
)

// GetExecutorFactory 获取执行器工厂单例
func GetExecutorFactory() *ExecutorFactory {
	once.Do(func() {
		factory = NewExecutorFactory()
	})

	return factory
}

// 为了兼容性，保留旧的函数名
func GetFactory() *ExecutorFactory {
	return GetExecutorFactory()
}

// InitExecutors 初始化所有执行器
func InitExecutors(cfg *conf.Config) error {
	f := GetExecutorFactory()

	// 根据配置创建并注册执行器
	// 注册 Prometheus 执行器
	if cfg.Executors.Prometheus.URL != "" {
		prometheusExecutor, err := NewPrometheusExecutor(cfg.Executors.Prometheus)
		if err != nil {
			return err
		}
		f.Register(prometheusExecutor)
	}

	// 注册 VM 执行器
	if cfg.Executors.VM.URL != "" {
		vmExecutor, err := NewVMExecutor(cfg.Executors.VM)
		if err != nil {
			return err
		}
		f.Register(vmExecutor)
	}

	// 注册 SSH 执行器
	sshExecutor := NewSSHExecutor(cfg.Executors.SSH)
	f.Register(sshExecutor)

	return nil
}
