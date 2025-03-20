package model

import "time"

// AgentStatus Agent 状态
type AgentStatus string

const (
	// AgentStatusRunning 运行中
	AgentStatusRunning AgentStatus = "running"
	// AgentStatusBusy 繁忙
	AgentStatusBusy AgentStatus = "busy"
	// AgentStatusError 错误
	AgentStatusError AgentStatus = "error"
)

// Heartbeat 心跳信息
type Heartbeat struct {
	AgentID     string      `json:"agent_id"`
	Status      AgentStatus `json:"status"`
	Version     string      `json:"version"`
	ClusterInfo ClusterInfo `json:"cluster_info"`
	Timestamp   time.Time   `json:"timestamp"`
}

// ClusterInfo 集群信息
type ClusterInfo struct {
	Name              string  `json:"name"`
	Nodes             int     `json:"nodes"`
	KubernetesVersion string  `json:"kubernetes_version,omitempty"`
	CPUUsage          float64 `json:"cpu_usage,omitempty"`
	MemoryUsage       float64 `json:"memory_usage,omitempty"`
	DiskUsage         float64 `json:"disk_usage,omitempty"`
}
