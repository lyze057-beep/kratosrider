# Kratos Project Template

## Install Kratos
```
go install github.com/go-kratos/kratos/cmd/kratos/v2@latest
```
## Create a service
```
# Create a template project
kratos new server

cd server
# Add a proto template
kratos proto add api/server/server.proto
# Generate the proto code
kratos proto client api/server/server.proto
# Generate the source code of service by proto file
kratos proto server api/server/server.proto -t internal/service

go generate ./...
go build -o ./bin/ ./...
./bin/server -conf ./configs
```
## Generate other auxiliary files by Makefile
```
# Download and update dependencies
make init
# Generate API files (include: pb.go, http, grpc, validate, swagger) by proto file
make api
# Generate all files
make all
```
## Automated Initialization (wire)
```
# install wire
go get github.com/google/wire/cmd/wire

# generate wire
cd cmd/server
wire
```

## Docker
```bash
# build
docker build -t <your-docker-image-name> .

# run
docker run --rm -p 8000:8000 -p 9000:9000 -v </path/to/your/configs>:/data/conf <your-docker-image-name>
```

## Performance Testing

### 压测报告
详细的压测报告请参考：[压测报告.md](./scripts/benchmark/压测报告.md)

### 压测结果对比

| 接口 | 优化前 | 优化后 | 提升幅度 |
|------|--------|--------|----------|
| 智能抢单(gRPC) | QPS=3200, 160ms | **QPS=5200, 85ms** | **+62.5%** |
| 骑手注册(HTTP) | QPS=800, 200ms | **QPS=1500, 95ms** | **+87.5%** |
| 位置上报(HTTP) | QPS=1800, 65ms | **QPS=2100, 45ms** | **+16.7%** |

### 压测脚本

#### 1. 智能抢单接口压测 (gRPC)
```bash
cd scripts/benchmark
./grpc_accept_order.sh

# 自定义参数
TARGET_HOST=localhost:9002 CONCURRENT_CLIENTS=1000 TOTAL_REQUESTS=100000 DURATION=60s ./grpc_accept_order.sh
```

#### 2. 骑手注册接口压测 (HTTP)
```bash
cd scripts/benchmark
./http_register.sh

# 自定义参数
TARGET_URL=http://localhost:8001 CONCURRENT_CLIENTS=1000 DURATION=180s ./http_register.sh
```

#### 3. 位置上报接口压测 (HTTP)
```bash
cd scripts/benchmark
./http_location.sh

# 自定义参数
TARGET_URL=http://localhost:8001 CONCURRENT_CLIENTS=1000 DURATION=300s ./http_location.sh
```

### 优化措施

#### 代码优化
- ✅ 使用Redis Pipeline批量操作
- ✅ 添加Redis缓存
- ✅ 异步消息发布
- ✅ 单次数据库查询更新
- ✅ 降低bcrypt cost值
- ✅ 异步短信发送
- ✅ 添加数据库索引

#### 配置优化
- ✅ 优化数据库连接池
- ✅ 优化Redis连接池
- ✅ 优化RabbitMQ配置

### 性能瓶颈分析

#### 当前瓶颈
1. **数据库查询**: 虽然已添加索引，但在高并发场景下仍有瓶颈
2. **Redis连接**: 高并发时连接池可能成为瓶颈
3. **消息队列**: Kafka/RabbitMQ吞吐量可能限制整体性能

#### 优化建议
1. **数据库读写分离**: 主库写，从库读
2. **Redis集群**: 分布式缓存，提升吞吐量
3. **消息队列优化**: 使用Kafka替代RabbitMQ，提升吞吐量
4. **服务拆分**: 将订单、用户、消息等模块拆分为独立服务
5. **缓存预热**: 启动时预热热点数据

## 项目结构

### API层
- `api/rider/v1/` - API定义
  - `order.proto` - 订单服务
  - `auth.proto` - 认证服务
  - `message.proto` - 消息服务
  - `income.proto` - 收入服务

### 业务层
- `internal/biz/` - 业务逻辑
  - `order.go` - 订单业务
  - `auth.go` - 认证业务
  - `message.go` - 消息业务
  - `income.go` - 收入业务

### 数据层
- `internal/data/` - 数据访问
  - `order.go` - 订单数据访问
  - `auth.go` - 认证数据访问
  - `message.go` - 消息数据访问
  - `data.go` - 数据层初始化

### 服务层
- `internal/server/` - 服务启动
  - `http.go` - HTTP服务
  - `grpc.go` - gRPC服务
  - `ws.go` - WebSocket服务

### 配置
- `configs/` - 配置文件
  - `config.yaml` - 应用配置

### 脚本
- `scripts/` - 脚本文件
  - `benchmark/` - 压测脚本
    - `grpc_accept_order.sh` - 智能抢单压测
    - `http_register.sh` - 注册压测
    - `http_location.sh` - 位置上报压测
    - `PERFORMANCE_ANALYSIS.md` - 性能分析指南
    - `压测报告.md` - 压测报告

## Docker

## License
MIT
