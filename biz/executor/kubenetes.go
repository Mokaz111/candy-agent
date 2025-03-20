package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"github.mokaz111.com/candy-agent/biz/model"
	"github.mokaz111.com/candy-agent/conf"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// KubernetesExecutor Kubernetes 执行器
type KubernetesExecutor struct {
	clientset *kubernetes.Clientset
	config    conf.KubernetesConfig
}

// NewKubernetesExecutor 创建 Kubernetes 执行器
func NewKubernetesExecutor(config conf.KubernetesConfig) (*KubernetesExecutor, error) {
	var (
		clientConfig *rest.Config
		err          error
	)

	// 根据配置决定使用集群内配置还是外部配置
	if config.InCluster {
		// 使用集群内配置
		clientConfig, err = rest.InClusterConfig()
		if err != nil {
			hlog.Errorf("Failed to create in-cluster config: %v", err)
			return nil, err
		}
	} else {
		// 使用外部配置
		clientConfig, err = clientcmd.BuildConfigFromFlags("", config.KubeConfig)
		if err != nil {
			hlog.Errorf("Failed to build config from kubeconfig: %v", err)
			return nil, err
		}
	}

	// 创建 Kubernetes 客户端
	clientset, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		hlog.Errorf("Failed to create Kubernetes client: %v", err)
		return nil, err
	}

	return &KubernetesExecutor{
		clientset: clientset,
		config:    config,
	}, nil
}

// Name 执行器名称
func (e *KubernetesExecutor) Name() string {
	return "kubernetes"
}

// Execute 执行巡检项
func (e *KubernetesExecutor) Execute(ctx context.Context, item model.TaskItem) (model.TaskResult, error) {
	startTime := time.Now()
	result := model.TaskResult{
		ItemID: item.ID,
		Status: model.ResultStatusNormal,
	}

	// 获取操作类型
	operation, ok := item.Params["operation"].(string)
	if !ok {
		result.Status = model.ResultStatusFailed
		result.Message = "Missing operation parameter"
		result.Duration = time.Since(startTime).Milliseconds()
		return result, fmt.Errorf("missing operation parameter")
	}

	// 设置超时上下文
	timeout := e.config.Timeout
	if timeout <= 0 {
		timeout = 30
	}
	execCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	// 根据操作类型执行不同的操作
	switch strings.ToLower(operation) {
	case "get_nodes":
		return e.getNodes(execCtx, item, startTime)
	case "get_pods":
		return e.getPods(execCtx, item, startTime)
	case "check_deployments":
		return e.checkDeployments(execCtx, item, startTime)
	case "check_services":
		return e.checkServices(execCtx, item, startTime)
	default:
		result.Status = model.ResultStatusFailed
		result.Message = fmt.Sprintf("Unsupported operation: %s", operation)
		result.Duration = time.Since(startTime).Milliseconds()
		return result, fmt.Errorf("unsupported operation: %s", operation)
	}
}

// getNodes 获取节点信息
func (e *KubernetesExecutor) getNodes(ctx context.Context, item model.TaskItem, startTime time.Time) (model.TaskResult, error) {
	result := model.TaskResult{
		ItemID: item.ID,
		Status: model.ResultStatusNormal,
	}

	// 获取所有节点
	nodes, err := e.clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		result.Status = model.ResultStatusFailed
		result.Message = fmt.Sprintf("Failed to get nodes: %v", err)
		result.Duration = time.Since(startTime).Milliseconds()
		return result, err
	}

	// 统计节点状态
	var readyNodes, notReadyNodes int
	var nodeDetails []string

	for _, node := range nodes.Items {
		isReady := false
		for _, condition := range node.Status.Conditions {
			if condition.Type == "Ready" {
				isReady = condition.Status == "True"
				break
			}
		}

		if isReady {
			readyNodes++
		} else {
			notReadyNodes++
			nodeDetails = append(nodeDetails, fmt.Sprintf("Node %s is not ready", node.Name))
		}
	}

	// 设置结果
	totalNodes := len(nodes.Items)
	result.Value = fmt.Sprintf("%d/%d", readyNodes, totalNodes)

	if notReadyNodes > 0 {
		result.Status = model.ResultStatusWarning
		result.Message = fmt.Sprintf("%d out of %d nodes are not ready", notReadyNodes, totalNodes)
		result.Details = strings.Join(nodeDetails, "\n")
	} else {
		result.Message = fmt.Sprintf("All %d nodes are ready", totalNodes)
		result.Details = fmt.Sprintf("Total nodes: %d\nReady nodes: %d", totalNodes, readyNodes)
	}

	result.Duration = time.Since(startTime).Milliseconds()
	return result, nil
}

// getPods 获取Pod信息
func (e *KubernetesExecutor) getPods(ctx context.Context, item model.TaskItem, startTime time.Time) (model.TaskResult, error) {
	result := model.TaskResult{
		ItemID: item.ID,
		Status: model.ResultStatusNormal,
	}

	// 获取命名空间参数
	namespace, _ := item.Params["namespace"].(string)
	if namespace == "" {
		namespace = "default"
	}

	// 获取标签选择器
	labelSelector, _ := item.Params["label_selector"].(string)

	// 获取Pod列表
	pods, err := e.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		result.Status = model.ResultStatusFailed
		result.Message = fmt.Sprintf("Failed to get pods: %v", err)
		result.Duration = time.Since(startTime).Milliseconds()
		return result, err
	}

	// 统计Pod状态
	var runningPods, pendingPods, failedPods, otherPods int
	var problemPods []string

	for _, pod := range pods.Items {
		switch pod.Status.Phase {
		case "Running":
			runningPods++
		case "Pending":
			pendingPods++
			problemPods = append(problemPods, fmt.Sprintf("Pod %s is pending", pod.Name))
		case "Failed":
			failedPods++
			problemPods = append(problemPods, fmt.Sprintf("Pod %s failed", pod.Name))
		default:
			otherPods++
			problemPods = append(problemPods, fmt.Sprintf("Pod %s is in %s state", pod.Name, pod.Status.Phase))
		}
	}

	// 设置结果
	totalPods := len(pods.Items)
	result.Value = fmt.Sprintf("%d/%d", runningPods, totalPods)

	if pendingPods > 0 || failedPods > 0 {
		if failedPods > 0 {
			result.Status = model.ResultStatusCritical
		} else {
			result.Status = model.ResultStatusWarning
		}
		result.Message = fmt.Sprintf("Issues found: %d pending, %d failed pods", pendingPods, failedPods)
		result.Details = strings.Join(problemPods, "\n")
	} else {
		result.Message = fmt.Sprintf("All %d pods are running", totalPods)
		result.Details = fmt.Sprintf("Total pods: %d\nRunning: %d\nPending: %d\nFailed: %d\nOther: %d",
			totalPods, runningPods, pendingPods, failedPods, otherPods)
	}

	result.Duration = time.Since(startTime).Milliseconds()
	return result, nil
}

// checkDeployments 检查部署状态
func (e *KubernetesExecutor) checkDeployments(ctx context.Context, item model.TaskItem, startTime time.Time) (model.TaskResult, error) {
	result := model.TaskResult{
		ItemID: item.ID,
		Status: model.ResultStatusNormal,
	}

	// 获取命名空间参数
	namespace, _ := item.Params["namespace"].(string)
	if namespace == "" {
		namespace = "default"
	}

	// 获取部署名称
	deploymentName, _ := item.Params["deployment"].(string)

	var deployments []string
	var details []string
	var problemDeployments []string

	// 如果指定了部署名称，则只检查该部署
	if deploymentName != "" {
		deployment, err := e.clientset.AppsV1().Deployments(namespace).Get(ctx, deploymentName, metav1.GetOptions{})
		if err != nil {
			if errors.IsNotFound(err) {
				result.Status = model.ResultStatusCritical
				result.Message = fmt.Sprintf("Deployment %s not found", deploymentName)
				result.Duration = time.Since(startTime).Milliseconds()
				return result, err
			}
			result.Status = model.ResultStatusFailed
			result.Message = fmt.Sprintf("Failed to get deployment: %v", err)
			result.Duration = time.Since(startTime).Milliseconds()
			return result, err
		}
		deployments = append(deployments, deploymentName)

		// 检查部署状态
		if deployment.Status.ReadyReplicas < deployment.Status.Replicas {
			problemDeployments = append(problemDeployments, deploymentName)
			details = append(details, fmt.Sprintf("Deployment %s: %d/%d replicas ready",
				deploymentName, deployment.Status.ReadyReplicas, deployment.Status.Replicas))
		} else {
			details = append(details, fmt.Sprintf("Deployment %s: %d/%d replicas ready",
				deploymentName, deployment.Status.ReadyReplicas, deployment.Status.Replicas))
		}
	} else {
		// 获取所有部署
		deploymentList, err := e.clientset.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			result.Status = model.ResultStatusFailed
			result.Message = fmt.Sprintf("Failed to list deployments: %v", err)
			result.Duration = time.Since(startTime).Milliseconds()
			return result, err
		}

		// 检查每个部署的状态
		for _, deployment := range deploymentList.Items {
			deployments = append(deployments, deployment.Name)
			if deployment.Status.ReadyReplicas < deployment.Status.Replicas {
				problemDeployments = append(problemDeployments, deployment.Name)
				details = append(details, fmt.Sprintf("Deployment %s: %d/%d replicas ready",
					deployment.Name, deployment.Status.ReadyReplicas, deployment.Status.Replicas))
			} else {
				details = append(details, fmt.Sprintf("Deployment %s: %d/%d replicas ready",
					deployment.Name, deployment.Status.ReadyReplicas, deployment.Status.Replicas))
			}
		}
	}

	// 设置结果
	totalDeployments := len(deployments)
	problemCount := len(problemDeployments)
	result.Value = fmt.Sprintf("%d/%d", totalDeployments-problemCount, totalDeployments)

	if problemCount > 0 {
		result.Status = model.ResultStatusWarning
		result.Message = fmt.Sprintf("%d out of %d deployments have issues", problemCount, totalDeployments)
		result.Details = strings.Join(details, "\n")
	} else {
		result.Message = fmt.Sprintf("All %d deployments are healthy", totalDeployments)
		result.Details = strings.Join(details, "\n")
	}

	result.Duration = time.Since(startTime).Milliseconds()
	return result, nil
}

// checkServices 检查服务状态
func (e *KubernetesExecutor) checkServices(ctx context.Context, item model.TaskItem, startTime time.Time) (model.TaskResult, error) {
	result := model.TaskResult{
		ItemID: item.ID,
		Status: model.ResultStatusNormal,
	}

	// 获取命名空间参数
	namespace, _ := item.Params["namespace"].(string)
	if namespace == "" {
		namespace = "default"
	}

	// 获取服务名称
	serviceName, _ := item.Params["service"].(string)

	var services []map[string]interface{}
	var problemServices []string

	// 如果指定了服务名称，则只检查该服务
	if serviceName != "" {
		service, err := e.clientset.CoreV1().Services(namespace).Get(ctx, serviceName, metav1.GetOptions{})
		if err != nil {
			if errors.IsNotFound(err) {
				result.Status = model.ResultStatusCritical
				result.Message = fmt.Sprintf("Service %s not found", serviceName)
				result.Duration = time.Since(startTime).Milliseconds()
				return result, err
			}
			result.Status = model.ResultStatusFailed
			result.Message = fmt.Sprintf("Failed to get service: %v", err)
			result.Duration = time.Since(startTime).Milliseconds()
			return result, err
		}

		// 检查服务是否有端点
		endpoints, err := e.clientset.CoreV1().Endpoints(namespace).Get(ctx, serviceName, metav1.GetOptions{})
		if err != nil && !errors.IsNotFound(err) {
			result.Status = model.ResultStatusFailed
			result.Message = fmt.Sprintf("Failed to get endpoints: %v", err)
			result.Duration = time.Since(startTime).Milliseconds()
			return result, err
		}

		serviceInfo := map[string]interface{}{
			"name":       service.Name,
			"type":       service.Spec.Type,
			"cluster_ip": service.Spec.ClusterIP,
		}

		// 检查端点
		hasEndpoints := false
		if endpoints != nil {
			for _, subset := range endpoints.Subsets {
				if len(subset.Addresses) > 0 {
					hasEndpoints = true
					break
				}
			}
		}

		serviceInfo["has_endpoints"] = hasEndpoints
		services = append(services, serviceInfo)

		if !hasEndpoints {
			problemServices = append(problemServices, service.Name)
		}
	} else {
		// 获取所有服务
		serviceList, err := e.clientset.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			result.Status = model.ResultStatusFailed
			result.Message = fmt.Sprintf("Failed to list services: %v", err)
			result.Duration = time.Since(startTime).Milliseconds()
			return result, err
		}

		// 检查每个服务
		for _, service := range serviceList.Items {
			// 跳过 Kubernetes 服务
			if service.Name == "kubernetes" && namespace == "default" {
				continue
			}

			// 检查服务是否有端点
			endpoints, err := e.clientset.CoreV1().Endpoints(namespace).Get(ctx, service.Name, metav1.GetOptions{})
			if err != nil && !errors.IsNotFound(err) {
				hlog.Warnf("Failed to get endpoints for service %s: %v", service.Name, err)
				continue
			}

			serviceInfo := map[string]interface{}{
				"name":       service.Name,
				"type":       service.Spec.Type,
				"cluster_ip": service.Spec.ClusterIP,
			}

			// 检查端点
			hasEndpoints := false
			if endpoints != nil {
				for _, subset := range endpoints.Subsets {
					if len(subset.Addresses) > 0 {
						hasEndpoints = true
						break
					}
				}
			}

			serviceInfo["has_endpoints"] = hasEndpoints
			services = append(services, serviceInfo)

			if !hasEndpoints {
				problemServices = append(problemServices, service.Name)
			}
		}
	}

	// 设置结果
	totalServices := len(services)
	problemCount := len(problemServices)
	result.Value = fmt.Sprintf("%d/%d", totalServices-problemCount, totalServices)

	// 将服务信息转换为JSON
	servicesJSON, err := json.Marshal(services)
	if err != nil {
		hlog.Warnf("Failed to marshal services to JSON: %v", err)
	}

	if problemCount > 0 {
		result.Status = model.ResultStatusWarning
		result.Message = fmt.Sprintf("%d out of %d services have no endpoints", problemCount, totalServices)
		result.Details = fmt.Sprintf("Services without endpoints: %s\n\nAll services: %s",
			strings.Join(problemServices, ", "), string(servicesJSON))
	} else {
		result.Message = fmt.Sprintf("All %d services have endpoints", totalServices)
		result.Details = fmt.Sprintf("All services: %s", string(servicesJSON))
	}

	result.Duration = time.Since(startTime).Milliseconds()
	return result, nil
}
