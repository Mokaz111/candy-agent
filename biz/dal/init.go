package dal

import (
	"github.mokaz111.com/candy-agent/biz/dal/executor"
	"github.mokaz111.com/candy-agent/biz/dal/k8s"
)

func Init() {
	//redis.Init()
	//mysql.Init()
	executor.Init()
	k8s.Init()
}
