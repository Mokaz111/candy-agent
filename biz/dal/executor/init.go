package executor

import (
	"os"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.mokaz111.com/candy-agent/biz/executor"
	"github.mokaz111.com/candy-agent/conf"
)

var ExecutorFactory *executor.ExecutorFactory

func Init() {
	ExecutorFactory = executor.GetExecutorFactory()
	if err := executor.InitExecutors(conf.GetConf()); err != nil {
		hlog.Errorf("Failed to initialize executors: %v", err)
		os.Exit(1)
	}
}
