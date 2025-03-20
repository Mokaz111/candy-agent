# Candy Agent

一个基于 Hertz 框架开发的集群巡检系统Agent组件，部署在各个Kubernetes集群中，负责执行巡检任务并上报结果。

## 项目结构

```
candy-agent/
├── biz/                    # 业务逻辑代码
│   ├── handler/           # HTTP 处理器
│   │   └── candyAgent/    # Agent相关处理器
│   ├── service/           # 业务服务层
│   ├── executor/          # 执行器实现
│   ├── model/             # 数据模型定义
│   ├── dal/               # 数据访问层
│   ├── router/            # 路由定义
│   └── utils/             # 工具函数
├── conf/                  # 配置文件目录
│   └── config.yaml        # 应用配置文件
├── idl/                   # 接口定义
├── hertz_gen/             # Hertz 生成的代码
├── script/                # 脚本文件
├── .hz                    # Hertz 配置
├── main.go                # 主程序入口
├── Dockerfile             # Docker 构建文件
├── docker-compose.yaml    # Docker Compose 配置
├── build.sh               # 构建脚本
├── go.mod                 # Go 模块文件
├── go.sum                 # Go 依赖校验文件
└── README.md              # 项目说明文档
```

## 功能特性

### 已完成功能

#### 心跳管理
- [x] 向服务器发送心跳 (`POST /api/v1/heartbeat/:agent_id`)
- [x] 上报集群基本信息
- [x] 上报 Agent 状态信息
- [x] 上报 Agent 访问地址

#### 巡检执行
- [x] 接收巡检任务 (`POST /api/v1/task`)
- [x] 执行巡检任务
  - [x] Prometheus 查询执行
  - [x] VictoriaMetrics 查询执行
  - [x] Bash 命令执行
- [x] 返回巡检结果

#### 执行引擎
- [x] 执行器适配层
- [x] Prometheus 执行器
- [x] VictoriaMetrics 执行器
- [x] SSH 执行器
- [x] 执行器工厂模式

#### 告警规则管理
- [x] 创建告警规则 (`POST /api/alert-rules`)
- [x] 更新告警规则 (`PUT /api/alert-rules`)
- [x] 删除告警规则 (`DELETE /api/alert-rules`)
- [x] 获取告警规则 (`GET /api/alert-rules/:name`)
- [x] 获取告警规则列表 (`GET /api/alert-rules`)
- [x] 同步告警规则 (`POST /api/alert-rules/sync`)
- [x] 支持 PrometheusRule 和 VMRule 资源

#### Kubernetes集成
- [x] 客户端自动初始化
- [x] 支持集群内外运行
- [x] 使用 client-go 操作资源
- [x] 使用 Dynamic Client 处理自定义资源

### 待完成功能

#### 任务执行
- [x] 异步任务执行机制
- [x] 任务队列实现
- [x] 任务状态管理
- [x] 任务回调机制

#### 系统功能
- [ ] 自动升级
- [ ] 健康检查
- [ ] 资源使用限制
- [ ] 执行日志

#### 安全性
- [ ] 加密通信
- [ ] 访问控制
- [ ] 凭证管理

## 系统架构

### Agent组件设计

```
+------------------+                  +------------------+
|                  |                  |                  |
|   Candy-Server   |<---------------->|   Candy-Agent    |
|                  |                  |                  |
+------------------+                  +--------+---------+
                                               |
                                               | 本地执行
                                               v
                                      +------------------+
                                      |                  |
                                      |  执行器适配层     |
                                      |                  |
                                      +--------+---------+
                                               |
                       +---------------------+---------------------+
                       |                     |                     |                       
                +------v------+      +-------v------+     +--------v------+         
                |             |      |              |     |               |          
                | Prometheus  |      |  VictoriaMetrics | |  SSH执行器     |           
                | 执行器      |      |  执行器        |    |               |          
                |             |      |              |     |               |            
                +-------------+      +--------------+     +---------------+
                       |                     |                     |
                       |                     |                     |
                +------v------+      +-------v------+     +--------v------+
                |             |      |              |     |               |
                | Prometheus  |      |  VictoriaMetrics |     |  目标主机     |
                |             |      |              |     |               |
                +-------------+      +--------------+     +---------------+
```

### 告警规则管理架构

```
+------------------+                  +------------------+
|                  |                  |                  |
|   Candy-Server   |<---------------->|   Candy-Agent    |
|                  |                  |                  |
+------------------+                  +--------+---------+
                                               |
                                               | client-go
                                               v
                                      +------------------+
                                      |                  |
                                      | Kubernetes API   |
                                      |                  |
                                      +--------+---------+
                                               |
                                               v
                                      +------------------+
                                      |                  |
                                      | VMRule/          |
                                      | PrometheusRule   |
                                      |                  |
                                      +------------------+
```

## 通信协议

### 1. 巡检任务下发 (Server -> Agent)

```json
{
  "task_id": "12345",
  "items": [
    {
      "id": 1,
      "name": "CPU使用率检查",
      "type": "prometheus",
      "params": {
        "query": "100 - (avg by(instance) (irate(node_cpu_seconds_total{mode=\"idle\"}[5m])) * 100)",
        "threshold": "80"
      }
    },
    {
      "id": 2,
      "name": "内存使用率检查",
      "type": "victoriaMetrics",
      "params": {
        "query": "100 * (1 - (node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes))",
        "threshold": "90"
      }
    }
  ],
  "timeout": 300
}
```

### 2. 巡检结果上报 (Agent -> Server)

```json
{
  "task_id": "12345",
  "status": "completed",
  "results": [
    {
      "item_id": 1,
      "status": "warning",
      "value": "45.2",
      "message": "有 3 个节点超过阈值 80.00，最高值为 node-01 的 95.78",
      "details": "Query: 100 - (avg by(instance) (irate(node_cpu_seconds_total{mode=\"idle\"}[5m])) * 100)\nResult Type: vector\n\n超过阈值的节点：\nnode-01 (95.78)\nnode-03 (85.21)\nnode-07 (82.64)",
      "duration": 120
    },
    {
      "item_id": 2,
      "status": "warning",
      "value": "92.5",
      "message": "节点 node-1 的值 92.50 超过阈值 90.00",
      "details": "Query: 100 * (1 - (node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes))\nResult Type: vector\n\n超过阈值的节点：\nnode-1 (92.50)",
      "duration": 150
    }
  ],
  "start_time": "2023-06-01T10:00:00Z",
  "end_time": "2023-06-01T10:05:00Z",
  "message": "Task completed with some warnings"
}
```

### 3. 心跳检测 (Agent -> Server)

```json
{
  "agent_id": "agent-001",
  "status": "running",
  "version": "1.0.0",
  "cluster_info": {
    "name": "production-cluster",
    "nodes": 5,
    "kubernetes_version": "1.25.0",
    "cpu_usage": 45.2,
    "memory_usage": 68.7,
    "disk_usage": 72.3
  },
  "timestamp": "2023-06-01T10:00:00Z",
  "agent_endpoint": "http://agent-address:8888",
  "api_key": "api-key-for-authentication"
}
```

### 4. 任务状态响应 (Agent -> Client)

```json
{
  "task_id": "12345",
  "status": "running", 
  "message": "",
  "start_time": "2023-06-01T10:00:00Z"
}
```

### 5. 获取任务结果响应 (Agent -> Client)

```json
{
  "task_id": "12345",
  "status": "completed",
  "results": [
    {
      "item_id": 1,
      "status": "warning",
      "value": "45.2",
      "message": "有3个节点超过阈值80.00，最高值为node-01的95.78",
      "details": "Query: 100 - (avg by(instance) (irate(node_cpu_seconds_total{mode=\"idle\"}[5m])) * 100)\nResult Type: vector\n\n超过阈值的节点：\nnode-01 (95.78)\nnode-03 (85.21)\nnode-07 (82.64)",
      "duration": 120
    }
  ],
  "start_time": "2023-06-01T10:00:00Z",
  "end_time": "2023-06-01T10:05:00Z",
  "message": ""
}
```

### 6. 任务执行回调 (Agent -> Server)

```json
{
  "task_id": "12345",
  "status": "completed",
  "results": [
    {
      "item_id": 1,
      "status": "warning",
      "value": "45.2",
      "message": "有3个节点超过阈值80.00，最高值为node-01的95.78",
      "details": "Query: 100 - (avg by(instance) (irate(node_cpu_seconds_total{mode=\"idle\"}[5m])) * 100)\nResult Type: vector\n\n超过阈值的节点：\nnode-01 (95.78)\nnode-03 (85.21)\nnode-07 (82.64)",
      "duration": 120
    }
  ],
  "start_time": "2023-06-01T10:00:00Z",
  "end_time": "2023-06-01T10:05:00Z",
}
```

## 最近更新

### 2023-03-25 异步任务执行系统

1. **异步任务执行机制**
   - 实现任务内存缓存，支持任务状态和结果的查询
   - 任务异步执行，立即返回任务ID
   - 实现任务取消功能和超时控制
   - 实现并发任务控制机制

2. **回调通知机制**
   - 任务完成后自动回调Server端，传递完整结果
   - 支持多种任务状态（等待中、运行中、已完成、失败、已取消）
   - 统一错误处理和日志记录

3. **API接口扩展**
   - 添加任务状态查询接口 `/api/v1/tasks/:task_id`
   - 添加任务取消接口 `/api/v1/tasks/:task_id`
   - 添加任务结果查询接口 `/api/v1/results/:task_id`
   - 完善API文档和通信协议规范

### 2023-03-20 功能优化

1. **告警规则管理重构**
   - 使用 client-go 直接操作 Kubernetes 自定义资源
   - 支持 PrometheusRule 和 VMRule 两种资源类型
   - 优化资源操作流程，提高可靠性

2. **执行引擎优化**
   - 改进执行器适配层，增强扩展性
   - 完善错误处理逻辑，提高任务执行稳定性
   - 优化日志记录，便于问题排查

3. **API 接口优化**
   - 统一返回格式，改善错误提示
   - 增强参数验证，提高接口安全性
   - 优化性能，减少不必要的数据处理

4. **心跳机制完善**
   - 心跳信息添加 Agent 访问地址和 API 密钥
   - 优化集群信息收集逻辑
   - 完善状态报告机制

### 2023-03-17 基础架构优化

1. **Kubernetes集成**
   - 实现 Kubernetes 客户端自动初始化
   - 支持集群内和集群外两种运行模式
   - 封装 Dynamic Client 操作，简化自定义资源处理

2. **无状态设计**
   - 采用无状态设计，提高系统稳定性
   - 移除本地存储依赖，降低维护成本
   - 使用 Kubernetes 原生资源存储，确保数据一致性

3. **安全性增强**
   - 实现 API 密钥认证机制
   - 添加请求验证流程
   - 优化错误处理和异常情况

## 技术特性

### 无状态设计

Candy-Agent 采用无状态设计，具有以下优势：

1. **任务执行**
   - Agent 不保存任务状态，每个任务执行完成后立即通过回调上报结果
   - 任务执行过程中的临时数据存储在内存缓存中，过期后自动清理
   - 任务状态变更实时同步到Server

2. **配置管理**
   - Agent 配置由 Server 统一管理，Agent 本地只保存必要的连接信息
   - 配置更新通过 Server 下发，Agent 应用后不需要保存历史版本

3. **心跳机制**
   - Agent 定期向 Server 发送心跳，Server 维护 Agent 状态
   - 心跳中包含基本的集群信息，无需 Agent 保存历史数据

### 客户端自动初始化

Candy-Agent在启动时自动初始化Kubernetes客户端，初始化过程如下：

1. **配置获取**
   - 首先尝试从环境变量`KUBECONFIG`获取kubeconfig路径
   - 如果环境变量未设置，则使用默认路径`~/.kube/config`
   - 检测是否在Kubernetes集群内运行，如果是则使用集群内配置

2. **客户端创建**
   - 根据运行环境选择合适的配置方式
   - 在集群内运行时使用`rest.InClusterConfig()`
   - 在集群外运行时使用`clientcmd.BuildConfigFromFlags("", kubeconfig)`
   - 创建动态客户端`dynamic.NewForConfig(clientConfig)`

## K8s容器开发最佳实践

1. **单一职责**
   - 每个容器只运行一个进程
   - 避免在容器中运行多个服务

2. **无状态设计**
   - 不依赖本地存储
   - 配置通过环境变量或配置文件注入

3. **健康检查**
   - 提供 /health 和 /ready 接口
   - 支持 K8s 的 liveness 和 readiness 探针

4. **优雅关闭**
   - 捕获 SIGTERM 信号
   - 处理完当前任务后再退出

5. **日志处理**
   - 日志输出到标准输出和标准错误
   - 不写入本地文件

6. **资源限制**
   - 设置合理的 CPU 和内存限制
   - 避免资源过度使用

7. **安全性**
   - 以非 root 用户运行
   - 使用只读文件系统
   - 设置适当的安全上下文

## 快速开始

### 环境要求

- Go 1.20+
- Kubernetes 集群

### 构建与部署

1. 构建二进制文件
```bash
sh build.sh
```

2. 构建 Docker 镜像
```bash
docker build -t candy-agent:latest .
```

3. 部署到 Kubernetes
```bash
kubectl apply -f deployments/kubernetes/candy-agent.yaml
```

### 配置

通过环境变量配置 Candy-Agent：

- `CANDY_SERVER_URL`: Candy-Server 的 URL
- `AGENT_ID`: Agent 的唯一标识
- `HEARTBEAT_INTERVAL`: 心跳间隔（秒）
- `LOG_LEVEL`: 日志级别
- `KUBECONFIG`: Kubernetes 配置文件路径

## 待实现功能

### 异步任务执行（V2版本）

当前版本的Candy-Agent使用同步任务执行模式，适用于执行时间较短的简单任务。在V2版本中，我们计划实现异步任务执行机制，以支持更复杂、耗时更长的巡检任务。

#### 异步执行设计

```
+------------------+                  +------------------+                  +------------------+
|                  |                  |                  |                  |                  |
|   Candy-Server   |<---------------->|   Candy-Agent    |<---------------->|    存储系统      |
|                  |                  |                  |                  |                  |
+------------------+                  +--------+---------+                  +------------------+
        ^                                     |                                      ^
        |                                     | 异步执行                              |
        |                                     v                                      |
        |                            +------------------+                            |
        |                            |                  |                            |
        |                            |    执行引擎       |                            |
        |                            |                  |                            |
        |                            +------------------+                            |
        |                                                                            |
        +----------------------------------------------------------------------------+
                                    任务完成回调
```

#### 主要组件

1. **任务队列**
   - 接收来自API服务器的任务请求
   - 将任务放入队列，立即返回任务接收确认
   - 支持任务优先级和调度控制

2. **执行引擎**
   - 从队列中获取任务并异步执行
   - 管理执行线程池，避免资源过度消耗
   - 实现任务超时控制和取消机制

3. **状态管理**
   - 使用临时存储记录任务执行状态
   - 支持查询任务进度和中间结果
   - 自动清理过期任务数据

4. **回调机制**
   - 任务完成后通过HTTP回调通知Server
   - 提供重试机制确保回调可靠送达
   - 支持自定义回调参数和格式

## 开发指南

### 添加新的执行器

1. 实现 `executor.Executor` 接口
2. 在 `executor/factory.go` 中注册执行器
3. 更新文档

### 运行测试

```bash
go test ./...
```

## API 文档

主要API分类如下：

### 心跳管理

| 接口 | 方法 | 路径 | 描述 |
| --- | --- | --- | --- |
| 发送心跳 | POST | /api/v1/heartbeat/:agent_id | 向Server发送心跳 |

### 巡检任务

| 接口 | 方法 | 路径 | 描述 |
| --- | --- | --- | --- |
| 接收任务 | POST | /api/v1/task | 接收Server下发的巡检任务 |
| 获取任务状态 | GET | /api/v1/tasks/:task_id | 获取指定任务的执行状态 |
| 取消任务 | DELETE | /api/v1/tasks/:task_id | 取消正在执行的任务 |
| 获取任务结果 | GET | /api/v1/results/:task_id | 获取指定任务的执行结果 |

### 告警规则管理

| 接口 | 方法 | 路径 | 描述 |
| --- | --- | --- | --- |
| 创建告警规则 | POST | /api/alert-rules | 创建新告警规则 |
| 更新告警规则 | PUT | /api/alert-rules | 更新现有告警规则 |
| 删除告警规则 | DELETE | /api/alert-rules | 删除告警规则 |
| 获取告警规则 | GET | /api/alert-rules/:name | 获取指定告警规则 |
| 获取告警规则列表 | GET | /api/alert-rules | 获取所有告警规则 |
| 同步告警规则 | POST | /api/alert-rules/sync | 将规则同步到Kubernetes |

### 健康检查

| 接口 | 方法 | 路径 | 描述 |
| --- | --- | --- | --- |
| 健康检查 | GET | /health | 获取Agent健康状态 |
| 就绪检查 | GET | /ready | 检查Agent是否就绪 |

### 请求参数示例

#### 巡检任务请求

```json
POST /api/v1/task
{
  "task_id": "12345",
  "items": [
    {
      "id": 1,
      "name": "CPU使用率检查",
      "type": "prometheus",
      "params": {
        "query": "100 - (avg by(instance) (irate(node_cpu_seconds_total{mode=\"idle\"}[5m])) * 100)",
        "threshold": "80"
      }
    }
  ],
  "timeout": 300
}
```

#### 告警规则创建请求

```json
POST /api/alert-rules
{
  "name": "HighCPUUsage",
  "type": "prometheus",
  "content": "100 - (avg by(instance) (irate(node_cpu_seconds_total{mode=\"idle\"}[5m])) * 100) > 80",
  "namespace": "monitoring",
  "config_map": "prometheus-rules",
  "key": "cpu-alerts.yaml"
}
```

## 贡献指南

欢迎提交 Issue 和 Pull Request

## 开源协议

MIT License 