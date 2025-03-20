package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.mokaz111.com/candy-agent/conf"
	candyAgent "github.mokaz111.com/candy-agent/hertz_gen/candyAgent"
)

// HeartbeatService 心跳服务
type HeartbeatService struct {
	client      *http.Client
	serverURL   string
	apiKey      string
	agentID     string
	clusterName string
	interval    time.Duration
	version     string
	stopChan    chan struct{}
	kubeVersion string
	nodeCount   int32
}

// NewHeartbeatService 创建心跳服务
func NewHeartbeatService() *HeartbeatService {
	config := conf.GetConf()

	// 设置默认值
	interval := time.Duration(config.Agent.HeartbeatInterval) * time.Second
	if interval == 0 {
		interval = 30 * time.Second
	}

	hlog.Infof("Creating new heartbeat service with agent ID: %s, cluster: %s, interval: %v",
		config.Agent.ID, config.Agent.ClusterName, interval)

	return &HeartbeatService{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		serverURL:   config.CandyServer.URL,
		apiKey:      config.CandyServer.APIKey,
		agentID:     config.Agent.ID,
		clusterName: config.Agent.ClusterName,
		interval:    interval,
		version:     "1.0.0", // 与健康检查服务保持一致
		stopChan:    make(chan struct{}),
		kubeVersion: "unknown",
		nodeCount:   0,
	}
}

// Start 启动心跳服务
func (s *HeartbeatService) Start() {
	hlog.Infof("Starting heartbeat service for agent: %s", s.agentID)

	// 启动时立即发送一次心跳
	s.sendHeartbeat()

	// 启动定时发送心跳的goroutine
	go func() {
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		for {
			select {
			case <-s.stopChan:
				hlog.Infof("Heartbeat service stopped for agent: %s", s.agentID)
				return
			case <-ticker.C:
				s.sendHeartbeat()
			}
		}
	}()

	// 启动定时更新集群信息的goroutine
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-s.stopChan:
				hlog.Infof("Cluster info update stopped for agent: %s", s.agentID)
				return
			case <-ticker.C:
				s.updateClusterInfo()
			}
		}
	}()

	hlog.Infof("Heartbeat service started successfully for agent: %s", s.agentID)
}

// Stop 停止心跳服务
func (s *HeartbeatService) Stop() {
	hlog.Infof("Stopping heartbeat service for agent: %s", s.agentID)
	close(s.stopChan)
}

// sendHeartbeat 发送心跳
func (s *HeartbeatService) sendHeartbeat() {
	// 更新集群信息
	s.updateClusterInfo()

	// 构建心跳请求
	req := &candyAgent.HeartbeatRequest{
		AgentId:   s.agentID,
		Status:    "running",
		Version:   s.version,
		Timestamp: time.Now().Format(time.RFC3339),
		ClusterInfo: &candyAgent.ClusterInfo{
			Name:              s.clusterName,
			Nodes:             s.nodeCount,
			KubernetesVersion: s.kubeVersion,
		},
	}

	// 序列化请求
	jsonData, err := json.Marshal(req)
	if err != nil {
		hlog.Errorf("Failed to marshal heartbeat request: %v", err)
		return
	}

	// 发送请求
	url := fmt.Sprintf("%s/api/heartbeat", s.serverURL)
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		hlog.Errorf("Failed to create heartbeat request: %v", err)
		return
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-API-Key", s.apiKey)

	resp, err := s.client.Do(httpReq)
	if err != nil {
		hlog.Errorf("Failed to send heartbeat request: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		hlog.Errorf("Heartbeat request failed with status code: %d", resp.StatusCode)
		return
	}

	hlog.Infof("Successfully sent heartbeat for agent: %s", s.agentID)
}

// updateClusterInfo 更新集群信息
func (s *HeartbeatService) updateClusterInfo() {
	// 获取Kubernetes版本
	kubeVersion, err := s.getKubernetesVersion()
	if err != nil {
		hlog.Errorf("Failed to get Kubernetes version: %v", err)
		return
	}
	s.kubeVersion = kubeVersion

	// 获取节点数量
	nodeCount, err := s.getNodeCount()
	if err != nil {
		hlog.Errorf("Failed to get node count: %v", err)
		return
	}
	s.nodeCount = nodeCount

	hlog.Infof("Updated cluster info - KubeVersion: %s, NodeCount: %d", s.kubeVersion, s.nodeCount)
}

// getKubernetesVersion 获取Kubernetes版本
func (s *HeartbeatService) getKubernetesVersion() (string, error) {
	// TODO: 实现获取Kubernetes版本的逻辑
	return "v1.20.0", nil
}

// getNodeCount 获取节点数量
func (s *HeartbeatService) getNodeCount() (int32, error) {
	// TODO: 实现获取节点数量的逻辑
	return 3, nil
}
