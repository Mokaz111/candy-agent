FROM golang:1.20-alpine AS builder

WORKDIR /app

# 复制go.mod和go.sum
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 编译
RUN CGO_ENABLED=0 GOOS=linux go build -o candy-agent

# 使用轻量级的alpine镜像
FROM alpine:3.17

WORKDIR /app

# 安装必要的工具
RUN apk --no-cache add ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai

# 从builder阶段复制编译好的二进制文件
COPY --from=builder /app/candy-agent /app/
COPY --from=builder /app/conf /app/conf

# 创建日志目录
RUN mkdir -p /app/log

# 设置环境变量
ENV GO_ENV=prod
ENV CANDY_AGENT_ID=""
ENV CANDY_SERVER_URL=""
ENV CANDY_API_KEY=""
ENV CANDY_CLUSTER_NAME=""

# 暴露端口
EXPOSE 8080

# 以非root用户运行
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
RUN chown -R appuser:appgroup /app
USER appuser

# 启动命令
CMD ["/app/candy-agent"] 