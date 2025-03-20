package model

// import "time"

// // AlertRuleType 告警规则类型
// type AlertRuleType string

// const (
// 	// AlertRuleTypePrometheus Prometheus 告警规则
// 	AlertRuleTypePrometheus AlertRuleType = "prometheus"
// 	// AlertRuleTypeVM VictoriaMetrics 告警规则
// 	AlertRuleTypeVM AlertRuleType = "vm"
// )

// // AlertRuleStatus 告警规则状态
// type AlertRuleStatus int

// const (
// 	// AlertRuleStatusDisabled 禁用
// 	AlertRuleStatusDisabled AlertRuleStatus = 0
// 	// AlertRuleStatusEnabled 启用
// 	AlertRuleStatusEnabled AlertRuleStatus = 1
// )

// // AlertRule 告警规则
// type AlertRule struct {
// 	ID          uint            `json:"id"`
// 	Name        string          `json:"name"`
// 	Description string          `json:"description"`
// 	ClusterID   uint            `json:"cluster_id"`
// 	Type        AlertRuleType   `json:"type"`
// 	Content     string          `json:"content"`
// 	ConfigMap   string          `json:"config_map"`
// 	Namespace   string          `json:"namespace"`
// 	Key         string          `json:"key"`
// 	Status      AlertRuleStatus `json:"status"`
// 	CreatedBy   uint            `json:"created_by"`
// 	Kind        string          `json:"kind"`
// 	CreatedAt   time.Time       `json:"created_at"`
// 	UpdatedAt   time.Time       `json:"updated_at"`
// }

// // AlertRuleStore 告警规则存储接口
// type AlertRuleStore interface {
// 	// Create 创建告警规则
// 	Create(rule *AlertRule) error
// 	// Update 更新告警规则
// 	Update(rule *AlertRule) error
// 	// Delete 删除告警规则
// 	Delete(id uint) error
// 	// Get 获取告警规则
// 	Get(id uint) (*AlertRule, error)
// 	// List 获取告警规则列表
// 	List(ruleType string, namespace string) ([]*AlertRule, error)
// }

// // AlertRuleRequest 告警规则请求
// type AlertRuleRequest struct {
// 	Action    string `json:"action"`
// 	RuleID    uint   `json:"rule_id"`
// 	Name      string `json:"name"`
// 	Type      string `json:"type"`
// 	Content   string `json:"content"`
// 	ConfigMap string `json:"config_map"`
// 	Namespace string `json:"namespace"`
// 	Key       string `json:"key"`
// }

// // AlertRuleResponse 告警规则响应
// type AlertRuleResponse struct {
// 	Success bool         `json:"success"`
// 	Message string       `json:"message"`
// 	Rule    *AlertRule   `json:"rule,omitempty"`
// 	Rules   []*AlertRule `json:"rules,omitempty"`
// 	Error   string       `json:"error,omitempty"`
// }
