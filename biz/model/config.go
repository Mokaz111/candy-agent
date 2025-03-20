package model

// ConfigType 配置类型
type ConfigType string

const (
	// ConfigTypeAlertRules 告警规则配置
	ConfigTypeAlertRules ConfigType = "alert_rules"
	// ConfigTypeAgent Agent 配置
	ConfigTypeAgent ConfigType = "agent"
)

// ConfigUpdate 配置更新
type ConfigUpdate struct {
	ConfigType ConfigType        `json:"config_type"`
	Config     map[string]string `json:"config"`
}

// AlertRulesConfig 告警规则配置
type AlertRulesConfig struct {
	PrometheusRules string `json:"prometheus_rules"`
	VMRules         string `json:"vm_rules"`
}

// AgentConfig Agent 配置
type AgentConfig struct {
	HeartbeatInterval int    `json:"heartbeat_interval"`
	LogLevel          string `json:"log_level"`
}
