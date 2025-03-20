package alertrule

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.mokaz111.com/candy-agent/biz/model"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// K8sRuleManager 使用client-go直接管理Kubernetes中的告警规则资源
type K8sRuleManager struct {
	client dynamic.Interface
}

// VMRule 资源定义
var vmRuleGVR = schema.GroupVersionResource{
	Group:    "operator.victoriametrics.com",
	Version:  "v1beta1",
	Resource: "vmrules",
}

// PrometheusRule 资源定义
var prometheusRuleGVR = schema.GroupVersionResource{
	Group:    "monitoring.coreos.com",
	Version:  "v1",
	Resource: "prometheusrules",
}

// NewK8sRuleManager 创建一个新的K8s规则管理器
func NewK8sRuleManager(kubeconfig string, inCluster bool) (*K8sRuleManager, error) {
	var (
		clientConfig *rest.Config
		err          error
	)

	// 根据配置决定使用集群内配置还是外部配置
	if inCluster {
		// 使用集群内配置
		clientConfig, err = rest.InClusterConfig()
		if err != nil {
			hlog.Errorf("Failed to create in-cluster config: %v", err)
			return nil, err
		}
	} else {
		// 使用外部配置
		clientConfig, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			hlog.Errorf("Failed to build config from kubeconfig: %v", err)
			return nil, err
		}
	}

	// 创建动态客户端
	dynamicClient, err := dynamic.NewForConfig(clientConfig)
	if err != nil {
		hlog.Errorf("Failed to create dynamic client: %v", err)
		return nil, err
	}

	return &K8sRuleManager{
		client: dynamicClient,
	}, nil
}

// getGVRByType 根据规则类型获取对应的GroupVersionResource
func (m *K8sRuleManager) getGVRByType(ruleType string) (schema.GroupVersionResource, error) {
	switch ruleType {
	case "victoriaMetrics":
		return vmRuleGVR, nil
	case "prometheus":
		return prometheusRuleGVR, nil
	default:
		return schema.GroupVersionResource{}, fmt.Errorf("unsupported rule type: %s", ruleType)
	}
}

// CreateRule 创建告警规则
func (m *K8sRuleManager) CreateRule(ctx context.Context, rule *model.AlertRule) error {
	// 获取对应的GVR
	gvr, err := m.getGVRByType(string(rule.Type))
	if err != nil {
		hlog.CtxErrorf(ctx, "Failed to get GVR for rule type %s: %v", rule.Type, err)
		return err
	}

	// 如果未指定命名空间，使用默认命名空间
	namespace := rule.Namespace
	if namespace == "" {
		namespace = "monitoring"
	}

	hlog.CtxInfof(ctx, "Creating rule %s with GVR %s/%s/%s in namespace %s",
		rule.Name, gvr.Group, gvr.Version, gvr.Resource, namespace)

	// 尝试解析规则的完整内容
	var specMap map[string]interface{}

	if rule.Content != "" {
		// 尝试从Content解析JSON
		err = json.Unmarshal([]byte(rule.Content), &specMap)
		if err != nil {
			hlog.CtxWarnf(ctx, "Failed to parse rule content as JSON, using simplified approach: %v", err)

			// 如果无法解析为JSON，则使用简化的方法创建规则结构
			specMap = map[string]interface{}{
				"groups": []map[string]interface{}{
					{
						"name": rule.Name,
						"rules": []map[string]interface{}{
							{
								"alert": rule.Name,
								"expr":  rule.Content,
								"for":   "5m",
								"labels": map[string]interface{}{
									"severity": "warning",
								},
								"annotations": map[string]interface{}{
									"summary":     fmt.Sprintf("%s alert", rule.Name),
									"description": fmt.Sprintf("Alert for %s", rule.Name),
								},
							},
						},
					},
				},
			}
		}
	} else {
		hlog.CtxWarnf(ctx, "Rule content is empty for %s", rule.Name)
		return fmt.Errorf("rule content cannot be empty")
	}

	// 创建标签
	labels := map[string]interface{}{
		"app":        "candy-agent",
		"managed-by": "candy-agent",
	}

	// 创建规则对象
	ruleObj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": gvr.Group + "/" + gvr.Version,
			"kind":       m.getKindByGVR(gvr),
			"metadata": map[string]interface{}{
				"name":      rule.Name,
				"namespace": namespace,
				"labels":    labels,
			},
			"spec": specMap,
		},
	}

	// 创建规则资源
	_, err = m.client.Resource(gvr).Namespace(namespace).Create(ctx, ruleObj, metav1.CreateOptions{})
	if err != nil {
		hlog.CtxErrorf(ctx, "Failed to create rule %s in namespace %s: %v", rule.Name, namespace, err)
		return err
	}

	hlog.CtxInfof(ctx, "Successfully created rule %s in namespace %s", rule.Name, namespace)
	return nil
}

// UpdateRule 更新告警规则
func (m *K8sRuleManager) UpdateRule(ctx context.Context, rule *model.AlertRule) error {
	// 获取对应的GVR
	gvr, err := m.getGVRByType(string(rule.Type))
	if err != nil {
		hlog.CtxErrorf(ctx, "Failed to get GVR for rule type %s: %v", rule.Type, err)
		return err
	}

	// 如果未指定命名空间，使用默认命名空间
	namespace := rule.Namespace
	if namespace == "" {
		namespace = "monitoring"
	}

	hlog.CtxInfof(ctx, "Updating rule %s with GVR %s/%s/%s in namespace %s",
		rule.Name, gvr.Group, gvr.Version, gvr.Resource, namespace)

	// 获取现有规则
	existing, err := m.client.Resource(gvr).Namespace(namespace).Get(ctx, rule.Name, metav1.GetOptions{})
	if err != nil {
		hlog.CtxErrorf(ctx, "Failed to get existing rule %s in namespace %s: %v", rule.Name, namespace, err)
		return err
	}

	// 尝试解析规则的完整内容
	var specMap map[string]interface{}

	if rule.Content != "" {
		// 尝试从Content解析JSON
		err = json.Unmarshal([]byte(rule.Content), &specMap)
		if err != nil {
			hlog.CtxWarnf(ctx, "Failed to parse rule content as JSON, using simplified approach: %v", err)

			// 如果无法解析为JSON，则使用简化的方法创建规则结构
			specMap = map[string]interface{}{
				"groups": []map[string]interface{}{
					{
						"name": rule.Name,
						"rules": []map[string]interface{}{
							{
								"alert": rule.Name,
								"expr":  rule.Content,
								"for":   "5m",
								"labels": map[string]interface{}{
									"severity": "warning",
								},
								"annotations": map[string]interface{}{
									"summary":     fmt.Sprintf("%s alert", rule.Name),
									"description": fmt.Sprintf("Alert for %s", rule.Name),
								},
							},
						},
					},
				},
			}
		}
	} else {
		hlog.CtxWarnf(ctx, "Rule content is empty for %s", rule.Name)
		return fmt.Errorf("rule content cannot be empty")
	}

	// 更新规则规范
	err = unstructured.SetNestedMap(existing.Object, specMap, "spec")
	if err != nil {
		hlog.CtxErrorf(ctx, "Failed to set spec in rule object: %v", err)
		return err
	}

	// 更新标签
	labels := existing.GetLabels()
	if labels == nil {
		labels = make(map[string]string)
	}
	labels["app"] = "candy-agent"
	labels["managed-by"] = "candy-agent"
	existing.SetLabels(labels)

	// 更新规则资源
	_, err = m.client.Resource(gvr).Namespace(namespace).Update(ctx, existing, metav1.UpdateOptions{})
	if err != nil {
		hlog.CtxErrorf(ctx, "Failed to update rule %s in namespace %s: %v", rule.Name, namespace, err)
		return err
	}

	hlog.CtxInfof(ctx, "Successfully updated rule %s in namespace %s", rule.Name, namespace)
	return nil
}

// DeleteRule 删除告警规则
func (m *K8sRuleManager) DeleteRule(ctx context.Context, ruleType, name, namespace string) error {
	// 获取对应的GVR
	gvr, err := m.getGVRByType(ruleType)
	if err != nil {
		return err
	}

	// 如果未指定命名空间，使用默认命名空间
	if namespace == "" {
		namespace = "monitoring"
	}

	// 删除规则资源
	return m.client.Resource(gvr).Namespace(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

// GetRule 获取单个告警规则
func (m *K8sRuleManager) GetRule(ctx context.Context, ruleType, name, namespace string) (*model.AlertRule, error) {
	// 获取对应的GVR
	gvr, err := m.getGVRByType(ruleType)
	if err != nil {
		hlog.CtxErrorf(ctx, "Failed to get GVR for rule type %s: %v", ruleType, err)
		return nil, err
	}

	// 如果未指定命名空间，使用默认命名空间
	if namespace == "" {
		namespace = "monitoring"
	}

	hlog.CtxInfof(ctx, "Getting rule %s with GVR %s/%s/%s in namespace %s",
		name, gvr.Group, gvr.Version, gvr.Resource, namespace)

	// 获取规则资源
	ruleObj, err := m.client.Resource(gvr).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		hlog.CtxErrorf(ctx, "Failed to get rule %s in namespace %s: %v", name, namespace, err)
		return nil, err
	}

	// 获取完整的规则规范
	specMap, ok := ruleObj.Object["spec"].(map[string]interface{})
	if !ok {
		hlog.CtxWarnf(ctx, "Rule %s does not have a valid spec", name)
		specMap = map[string]interface{}{}
	}

	// 将整个spec转换为JSON字符串
	fullContentBytes, err := json.Marshal(specMap)
	fullContent := string(fullContentBytes)
	if err != nil {
		hlog.CtxWarnf(ctx, "Failed to marshal rule spec to JSON: %v", err)
		fullContent = "{}" // 设置默认空对象
	}

	hlog.CtxInfof(ctx, "Successfully retrieved rule %s with full content", name)

	// 创建AlertRule对象，使用完整内容作为Content字段
	rule := &model.AlertRule{
		Name:        ruleObj.GetName(),
		Description: ruleObj.GetName(),
		Type:        model.AlertRuleType(ruleType),
		Content:     fullContent, // 使用完整的规则内容
		Namespace:   namespace,
		Status:      model.AlertRuleStatusEnabled,
		CreatedAt:   ruleObj.GetCreationTimestamp().Time,
		UpdatedAt:   time.Now(),
	}

	// 尝试提取标签信息
	labels := ruleObj.GetLabels()
	if appLabel, ok := labels["app"]; ok {
		rule.ConfigMap = appLabel
	}

	return rule, nil
}

// ListRules 获取告警规则列表
func (m *K8sRuleManager) ListRules(ctx context.Context, ruleType, namespace string, labelSelector string) ([]*model.AlertRule, error) {
	// 获取对应的GVR
	gvr, err := m.getGVRByType(ruleType)
	if err != nil {
		hlog.CtxErrorf(ctx, "Failed to get GVR for rule type %s: %v", ruleType, err)
		return nil, err
	}

	// 如果未指定命名空间，使用默认命名空间
	if namespace == "" {
		namespace = "monitoring"
	}

	hlog.CtxInfof(ctx, "Listing rules with GVR %s/%s/%s in namespace %s with labelSelector: %s",
		gvr.Group, gvr.Version, gvr.Resource, namespace, labelSelector)

	// 获取规则资源列表
	list, err := m.client.Resource(gvr).Namespace(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		hlog.CtxErrorf(ctx, "Failed to list resources for type %s in namespace %s: %v", ruleType, namespace, err)
		return nil, err
	}

	hlog.CtxInfof(ctx, "Found %d items for rule type %s in namespace %s", len(list.Items), ruleType, namespace)

	// 创建规则列表
	rules := make([]*model.AlertRule, 0, len(list.Items))
	for _, item := range list.Items {
		// 获取完整的规则规范
		specMap, ok := item.Object["spec"].(map[string]interface{})
		var content string = "{}"

		if ok {
			// 将spec转换为JSON字符串
			contentBytes, err := json.Marshal(specMap)
			if err == nil {
				content = string(contentBytes)
			} else {
				hlog.CtxWarnf(ctx, "Failed to marshal spec to JSON for rule %s: %v", item.GetName(), err)
			}
		} else {
			hlog.CtxWarnf(ctx, "Rule %s does not have a valid spec", item.GetName())
		}

		rule := &model.AlertRule{
			Name:        item.GetName(),
			Description: item.GetName(),
			Type:        model.AlertRuleType(ruleType),
			Content:     content, // 使用完整的规则内容
			Namespace:   namespace,
			Status:      model.AlertRuleStatusEnabled,
			CreatedAt:   item.GetCreationTimestamp().Time,
			UpdatedAt:   time.Now(),
		}

		// 提取标签信息
		labels := item.GetLabels()
		if appLabel, ok := labels["app"]; ok {
			rule.ConfigMap = appLabel
		}

		rules = append(rules, rule)
	}

	return rules, nil
}

// getKindByGVR 根据GVR获取资源Kind
func (m *K8sRuleManager) getKindByGVR(gvr schema.GroupVersionResource) string {
	if gvr == vmRuleGVR {
		return "VMRule"
	}
	if gvr == prometheusRuleGVR {
		return "PrometheusRule"
	}
	return ""
}

// RuleToJSON 将规则转换为JSON字符串
func (m *K8sRuleManager) RuleToJSON(rule *model.AlertRule) (string, error) {
	data, err := json.Marshal(rule)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// RulesToJSON 将规则列表转换为JSON字符串
func (m *K8sRuleManager) RulesToJSON(rules []*model.AlertRule) (string, error) {
	data, err := json.Marshal(rules)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// 全局K8s规则管理器实例
var (
	k8sRuleManager *K8sRuleManager
)

// GetK8sRuleManager 获取K8s规则管理器实例
func GetK8sRuleManager() *K8sRuleManager {
	return k8sRuleManager
}

// InitK8sRuleManager 初始化K8s规则管理器
func InitK8sRuleManager(kubeconfig string, inCluster bool) error {
	var err error
	k8sRuleManager, err = NewK8sRuleManager(kubeconfig, inCluster)
	return err
}
