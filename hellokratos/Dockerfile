# 1. 构建阶段：强制编译Linux amd64架构可执行文件，彻底解决格式错误
FROM golang:1.26-alpine AS builder
# 安装必要工具
RUN apk add --no-cache git
# 设置工作目录
WORKDIR /src
# 复制项目依赖文件
COPY go.mod go.sum ./
# 下载依赖，提前缓存
RUN GOPROXY=https://goproxy.cn go mod download
# 复制全量项目代码
COPY . .
# ✅ 核心：直接写死编译命令，不用makefile中转，100%确保参数生效
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/server ./cmd/hellokratos/

# 2. 运行阶段：极简配置，彻底解决权限+路径问题
FROM debian:stable-slim
# 安装系统依赖，避免缺失
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    netbase \
    && rm -rf /var/lib/apt/lists/ \
    && apt-get autoremove -y && apt-get autoclean -y
# 工作目录
WORKDIR /app
# 复制编译好的可执行文件
COPY --from=builder /src/bin/server /app/
# 复制配置文件
COPY --from=builder /src/configs /app/configs
# 复制数据库文件
COPY --from=builder /src/hellokratos.db /app/
# ✅ 终极权限配置，直接给最高权限，杜绝所有权限问题
RUN chmod +x /app/server && chmod -R 777 /app
# 暴露端口（和你项目配置完全一致）
EXPOSE 8001 9002
# ✅ 绝对路径启动，指定正确的配置目录，不会找不到文件
CMD ["/app/server", "-conf", "/app/configs"]