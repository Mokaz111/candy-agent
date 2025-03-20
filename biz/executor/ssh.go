package executor

import (
	"bytes"
	"context"
	"fmt"
	"github.mokaz111.com/candy-agent/biz/model"
	config "github.mokaz111.com/candy-agent/conf"
	"net"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

// SSHExecutor SSH 执行器
type SSHExecutor struct {
	config config.SSHConfig
}

// NewSSHExecutor 创建 SSH 执行器
func NewSSHExecutor(config config.SSHConfig) *SSHExecutor {
	return &SSHExecutor{
		config: config,
	}
}

// Name 执行器名称
func (e *SSHExecutor) Name() string {
	return "ssh"
}

// Execute 执行巡检项
func (e *SSHExecutor) Execute(ctx context.Context, item model.TaskItem) (model.TaskResult, error) {
	startTime := time.Now()
	result := model.TaskResult{
		ItemID: item.ID,
		Status: model.ResultStatusNormal,
	}

	// 获取 SSH 连接参数
	host, ok := item.Params["host"].(string)
	if !ok {
		result.Status = model.ResultStatusFailed
		result.Message = "Missing host parameter"
		result.Duration = time.Since(startTime).Milliseconds()
		return result, fmt.Errorf("missing host parameter")
	}

	port := 22
	if portParam, ok := item.Params["port"].(float64); ok {
		port = int(portParam)
	}

	username, ok := item.Params["username"].(string)
	if !ok {
		result.Status = model.ResultStatusFailed
		result.Message = "Missing username parameter"
		result.Duration = time.Since(startTime).Milliseconds()
		return result, fmt.Errorf("missing username parameter")
	}

	password, _ := item.Params["password"].(string)
	privateKey, _ := item.Params["private_key"].(string)

	if password == "" && privateKey == "" {
		result.Status = model.ResultStatusFailed
		result.Message = "Missing authentication method (password or private_key)"
		result.Duration = time.Since(startTime).Milliseconds()
		return result, fmt.Errorf("missing authentication method")
	}

	command, ok := item.Params["command"].(string)
	if !ok {
		result.Status = model.ResultStatusFailed
		result.Message = "Missing command parameter"
		result.Duration = time.Since(startTime).Milliseconds()
		return result, fmt.Errorf("missing command parameter")
	}

	// 创建 SSH 客户端配置
	clientConfig := &ssh.ClientConfig{
		User:            username,
		Auth:            []ssh.AuthMethod{},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         time.Duration(e.config.ConnectionTimeout) * time.Second,
	}

	// 添加认证方法
	if password != "" {
		clientConfig.Auth = append(clientConfig.Auth, ssh.Password(password))
	}

	if privateKey != "" {
		signer, err := ssh.ParsePrivateKey([]byte(privateKey))
		if err != nil {
			result.Status = model.ResultStatusFailed
			result.Message = fmt.Sprintf("Failed to parse private key: %v", err)
			result.Duration = time.Since(startTime).Milliseconds()
			return result, err
		}
		clientConfig.Auth = append(clientConfig.Auth, ssh.PublicKeys(signer))
	}

	// 连接 SSH 服务器
	client, err := ssh.Dial("tcp", net.JoinHostPort(host, strconv.Itoa(port)), clientConfig)
	if err != nil {
		result.Status = model.ResultStatusFailed
		result.Message = fmt.Sprintf("Failed to connect to SSH server: %v", err)
		result.Duration = time.Since(startTime).Milliseconds()
		return result, err
	}
	defer client.Close()

	// 创建会话
	session, err := client.NewSession()
	if err != nil {
		result.Status = model.ResultStatusFailed
		result.Message = fmt.Sprintf("Failed to create SSH session: %v", err)
		result.Duration = time.Since(startTime).Milliseconds()
		return result, err
	}
	defer session.Close()

	// 设置标准输出和标准错误
	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	// 设置超时上下文
	timeout := e.config.Timeout
	if timeout <= 0 {
		timeout = 30
	}

	// 创建一个通道来接收命令完成信号
	done := make(chan error, 1)
	go func() {
		done <- session.Run(command)
	}()

	// 等待命令完成或超时
	select {
	case <-ctx.Done():
		// 上下文被取消
		result.Status = model.ResultStatusFailed
		result.Message = "Command execution cancelled"
		result.Duration = time.Since(startTime).Milliseconds()
		return result, ctx.Err()
	case <-time.After(time.Duration(timeout) * time.Second):
		// 命令执行超时
		result.Status = model.ResultStatusFailed
		result.Message = fmt.Sprintf("Command execution timed out after %d seconds", timeout)
		result.Duration = time.Since(startTime).Milliseconds()
		return result, fmt.Errorf("command execution timed out")
	case err := <-done:
		// 命令执行完成
		if err != nil {
			// 命令执行失败
			result.Status = model.ResultStatusFailed
			result.Message = fmt.Sprintf("Command execution failed: %v", err)
			result.Value = stderr.String()
		} else {
			// 命令执行成功
			result.Value = stdout.String()
			result.Message = "Command executed successfully"

			// 检查阈值
			if threshold, ok := item.Params["threshold"].(string); ok {
				// 根据阈值判断状态
				if strings.Contains(result.Value, threshold) {
					result.Status = model.ResultStatusWarning
					result.Message = fmt.Sprintf("Output contains threshold string: %s", threshold)
				}
			}
		}
	}

	// 设置详细信息
	var details strings.Builder
	details.WriteString(fmt.Sprintf("Command: %s\n", command))
	details.WriteString(fmt.Sprintf("Host: %s:%d\n", host, port))
	details.WriteString(fmt.Sprintf("Username: %s\n", username))
	details.WriteString(fmt.Sprintf("Stdout: %s\n", stdout.String()))
	if stderr.Len() > 0 {
		details.WriteString(fmt.Sprintf("Stderr: %s\n", stderr.String()))
	}
	result.Details = details.String()
	result.Duration = time.Since(startTime).Milliseconds()

	return result, nil
}
