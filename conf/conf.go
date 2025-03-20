package conf

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/kr/pretty"
	"gopkg.in/validator.v2"
	"gopkg.in/yaml.v2"
)

var (
	conf *Config
	once sync.Once
)

type Config struct {
	Env string

	Hertz       Hertz             `yaml:"hertz"`
	CandyServer CandyServerConfig `yaml:"candy_server"`
	Agent       AgentConfig       `yaml:"agent"`
	Executors   ExecutorsConfig   `yaml:"executors"`
	Kubernetes  KubernetesConfig  `yaml:"kubernetes"`
	Server      ServerConfig      `yaml:"server"`
	TaskManager TaskManagerConfig `yaml:"task_manager"`
}

type CandyServerConfig struct {
	URL    string `yaml:"url"`
	APIKey string `yaml:"api_key"`
}

// AgentConfig Agent 配置
type AgentConfig struct {
	ID                string `yaml:"id"`
	HeartbeatInterval int    `yaml:"heartbeat_interval"`
	ClusterName       string `yaml:"cluster_name"`
}

// ExecutorsConfig 执行器配置
type ExecutorsConfig struct {
	Prometheus PrometheusConfig `yaml:"prometheus"`
	VM         VMConfig         `yaml:"vm"`
	SSH        SSHConfig        `yaml:"ssh"`
	Kubernetes KubernetesConfig `yaml:"kubernetes"`
}

// PrometheusConfig Prometheus 执行器配置
type PrometheusConfig struct {
	URL     string `yaml:"url"`
	Timeout int    `yaml:"timeout"`
}

// VMConfig VictoriaMetrics 执行器配置
type VMConfig struct {
	URL     string `yaml:"url"`
	Token   string `yaml:"token"`
	Timeout int    `yaml:"timeout"`
}

// SSHConfig SSH 执行器配置
type SSHConfig struct {
	Timeout           int `yaml:"timeout"`
	ConnectionTimeout int `yaml:"connection_timeout"`
}

// KubernetesConfig Kubernetes 执行器配置
type KubernetesConfig struct {
	KubeConfig     string `yaml:"kube_config"`
	KubeconfigPath string `yaml:"kubeconfig_path"`
	InCluster      bool   `yaml:"in_cluster"`
	Timeout        int    `yaml:"timeout"`
}

// ServerConfig 服务端配置
type ServerConfig struct {
	URL         string `yaml:"url"`          // 服务端地址
	CallbackURL string `yaml:"callback_url"` // 回调URL
	APIKey      string `yaml:"api_key"`      // API密钥
}

// TaskManagerConfig 任务管理器配置
type TaskManagerConfig struct {
	MaxWorkers     int   `yaml:"max_workers"`     // 最大工作线程数
	TaskExpiration int64 `yaml:"task_expiration"` // 任务过期时间(小时)
}

// HertzConfig Hertz配置
type Hertz struct {
	Address         string `yaml:"address"`
	Host            string `yaml:"host"`
	Port            int    `yaml:"port"`
	EnablePprof     bool   `yaml:"enable_pprof"`
	EnableGzip      bool   `yaml:"enable_gzip"`
	EnableAccessLog bool   `yaml:"enable_access_log"`
	LogLevel        string `yaml:"log_level"`
	LogFileName     string `yaml:"log_file_name"`
	LogMaxSize      int    `yaml:"log_max_size"`
	LogMaxBackups   int    `yaml:"log_max_backups"`
	LogMaxAge       int    `yaml:"log_max_age"`
	ReadTimeout     int    `yaml:"read_timeout"`
	WriteTimeout    int    `yaml:"write_timeout"`
	MaxHeaderBytes  int    `yaml:"max_header_bytes"`
}

// GetConf gets configuration instance
func GetConf() *Config {
	once.Do(initConf)
	return conf
}

func initConf() {
	prefix := "conf"
	confFileRelPath := filepath.Join(prefix, filepath.Join(GetEnv(), "conf.yaml"))
	content, err := ioutil.ReadFile(confFileRelPath)
	if err != nil {
		panic(err)
	}

	conf = new(Config)
	err = yaml.Unmarshal(content, conf)
	if err != nil {
		hlog.Error("parse yaml error - %v", err)
		panic(err)
	}
	if err := validator.Validate(conf); err != nil {
		hlog.Error("validate config error - %v", err)
		panic(err)
	}

	conf.Env = GetEnv()
	overrideFromEnv(conf)
	pretty.Printf("%+v\n", conf)
}

func GetEnv() string {
	e := os.Getenv("GO_ENV")
	if len(e) == 0 {
		return "test"
	}
	return e
}

func LogLevel() hlog.Level {
	level := GetConf().Hertz.LogLevel
	switch level {
	case "trace":
		return hlog.LevelTrace
	case "debug":
		return hlog.LevelDebug
	case "info":
		return hlog.LevelInfo
	case "notice":
		return hlog.LevelNotice
	case "warn":
		return hlog.LevelWarn
	case "error":
		return hlog.LevelError
	case "fatal":
		return hlog.LevelFatal
	default:
		return hlog.LevelInfo
	}
}

func overrideFromEnv(config *Config) {
	// 优先使用环境变量中的 Agent ID
	if agentID := os.Getenv("CANDY_AGENT_ID"); agentID != "" {
		config.Agent.ID = agentID
	}

	// 如果没有设置 Agent ID，则使用主机名
	if config.Agent.ID == "" {
		hostname, err := os.Hostname()
		if err == nil {
			config.Agent.ID = hostname
		} else {
			config.Agent.ID = "unknown"
		}
	}

	// 优先使用环境变量中的 Candy-Server URL
	if serverURL := os.Getenv("CANDY_SERVER_URL"); serverURL != "" {
		config.CandyServer.URL = serverURL
	}

	// 优先使用环境变量中的 API Key
	if apiKey := os.Getenv("CANDY_API_KEY"); apiKey != "" {
		config.CandyServer.APIKey = apiKey
	}

	// 优先使用环境变量中的集群名称
	if clusterName := os.Getenv("CANDY_CLUSTER_NAME"); clusterName != "" {
		config.Agent.ClusterName = clusterName
	}
}
