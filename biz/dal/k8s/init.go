package k8s

import (
	"os"
	"path/filepath"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.mokaz111.com/candy-agent/biz/dal/alertrule"
)

// InitK8sClient 初始化Kubernetes客户端
func InitK8sClient() error {
	// 尝试获取kubeconfig路径
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		// 如果环境变量中没有设置，尝试使用默认路径
		homeDir, err := os.UserHomeDir()
		if err == nil {
			kubeconfig = filepath.Join(homeDir, ".kube", "config")
		}
	}

	// 检查是否在集群内运行
	inCluster := false
	if _, err := os.Stat("/var/run/secrets/kubernetes.io/serviceaccount/token"); err == nil {
		inCluster = true
		hlog.Info("Running in Kubernetes cluster, using in-cluster config")
	} else {
		hlog.Infof("Running outside Kubernetes cluster, using kubeconfig: %s", kubeconfig)
	}

	// 初始化K8s规则管理器
	err := alertrule.InitK8sRuleManager(kubeconfig, inCluster)
	if err != nil {
		hlog.Errorf("Failed to initialize K8s rule manager: %v", err)
		return err
	}

	hlog.Info("K8s rule manager initialized successfully")
	return nil
}
func Init() {
	// 初始化K8s客户端
	if err := InitK8sClient(); err != nil {
		hlog.Fatalf("Failed to initialize K8s client: %v", err)
	}
}
