package model

import (
	"time"
)

// AlertRuleType 告警规则类型
type AlertRuleType string

const (
	// AlertRuleTypePrometheus Prometheus告警规则类型
	AlertRuleTypePrometheus AlertRuleType = "prometheus"
	// AlertRuleTypeVM VictoriaMetrics告警规则类型
	AlertRuleTypeVM AlertRuleType = "vm"
)

// AlertRuleStatus 告警规则状态
type AlertRuleStatus int

const (
	// AlertRuleStatusDisabled 禁用状态
	AlertRuleStatusDisabled AlertRuleStatus = 0
	// AlertRuleStatusEnabled 启用状态
	AlertRuleStatusEnabled AlertRuleStatus = 1
)

// AlertRule 告警规则模型
type AlertRule struct {
	ID          uint            `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	ClusterID   uint            `json:"cluster_id"`
	Type        AlertRuleType   `json:"type"`
	Content     string          `json:"content"` // 包含完整的规则规范JSON
	ConfigMap   string          `json:"config_map"`
	Namespace   string          `json:"namespace"`
	Key         string          `json:"key"`
	Status      AlertRuleStatus `json:"status"`
	CreatedBy   uint            `json:"created_by"`
	Kind        string          `json:"kind"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// NewAlertRule 创建一个新的告警规则
func NewAlertRule(name, description string, ruleType AlertRuleType, content, namespace string) *AlertRule {
	now := time.Now()
	return &AlertRule{
		Name:        name,
		Description: description,
		ClusterID:   1, // 默认集群ID
		Type:        ruleType,
		Content:     content,
		Namespace:   namespace,
		Status:      AlertRuleStatusEnabled,
		CreatedBy:   1, // 默认用户ID
		Kind:        "AlertRule",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}
