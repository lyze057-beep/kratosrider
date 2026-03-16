# Kratos 骑手端压测脚本使用指南

## 📋 目录
- [简介](#简介)
- [项目结构](#项目结构)
- [快速开始](#快速开始)
- [压测脚本说明](#压测脚本说明)
- [依赖安装](#依赖安装)
- [运行示例](#运行示例)
- [性能分析](#性能分析)
- [常见问题](#常见问题)

## 简介

本目录包含针对美团外卖骑手端 Kratos 项目的完整压测脚本，涵盖以下场景：

1. **骑手注册接口** - 验证接口限流/防刷能力
2. **骑手位置上报** - 高频低延迟场景
3. **骑手消息通知** - WebSocket长连接稳定性
4. **智能抢单推送** - 模拟高并发抢单场景（gRPC + WebSocket）
5. **多订单处理** - 模拟订单状态更新场景（gRPC + HTTP）
6. **性能分析** - 使用 pprof 进行性能瓶颈分析

## 项目结构

```
scripts/benchmark/
├── http_register.sh              # 骑手注册接口压测 (wrk)
├── http_location.sh              # 骑手位置上报压测 (wrk)
├── ws_message.py                 # 骑手消息通知压测 (WebSocket)
├── grpc_accept_order.sh          # gRPC 抢单压测脚本
├── grpc_update_order_status.sh   # gRPC 订单状态更新压测脚本
├── http_update_order_status.sh   # HTTP 订单状态更新压测脚本 (wrk)
├── locust_order_push.py          # Locust 混合场景压测脚本
├── PPROF_ANALYSIS_GUIDE.md       # pprof 性能分析指南
└── README.md                      # 本文件
```

## 快速开始

### 1. 启动服务

```bash
# 启动 Kratos 服务
go run cmd/hellokratos/main.go -conf configs/config.yaml
```

### 2. 运行压测

```bash
# 进入压测脚本目录
cd scripts/benchmark

# 骑手注册接口压测
./http_register.sh

# 骑手位置上报压测
./http_location.sh

# 骑手消息通知压测
python ws_message.py

# 智能抢单压测
./grpc_accept_order.sh

# 订单状态更新压测
./grpc_update_order_status.sh

# HTTP压测
./http_update_order_status.sh
```

## 压测脚本说明

### 1. 骑手手机号注册接口 (HTTP)

**脚本**: `http_register.sh`

**场景**: 验证接口限流/防刷能力，1000并发，持续3分钟

**协议**: HTTP/JSON

**参数**:
- `TARGET_URL`: 服务地址 (默认: `http://localhost:8000`)
- `CONCURRENT_CLIENTS`: 并发客户端数 (默认: `1000`)
- `DURATION`: 持续时间 (默认: `180s`)
- `THREADS`: 线程数 (默认: `8`)

**运行命令**:
```bash
# 使用默认参数
./http_register.sh

# 自定义参数
TARGET_URL=http://localhost:8000 \
CONCURRENT_CLIENTS=1000 \
DURATION=180s \
./http_register.sh
```

**技术特点**:
- 使用 wrk 进行 HTTP 压测
- 预生成手机号池和验证码池
- 模拟真实注册场景
- 验证接口限流/防刷能力

### 2. 骑手位置上报接口 (HTTP)

**脚本**: `http_location.sh`

**场景**: 模拟1000名骑手每秒上报1次，持续5分钟

**协议**: HTTP/JSON

**参数**:
- `TARGET_URL`: 服务地址 (默认: `http://localhost:8000`)
- `CONCURRENT_CLIENTS`: 并发客户端数 (默认: `1000`)
- `DURATION`: 持续时间 (默认: `300s`)
- `THREADS`: 线程数 (默认: `8`)

**运行命令**:
```bash
# 使用默认参数
./http_location.sh

# 自定义参数
TARGET_URL=http://localhost:8000 \
CONCURRENT_CLIENTS=1000 \
DURATION=300s \
./http_location.sh
```

**技术特点**:
- 使用 wrk 进行 HTTP 压测
- 预生成骑手ID池
- 模拟北京周边随机经纬度
- 高频低延迟场景
- 验证LBS定位更新性能

### 3. 骑手消息通知 (WebSocket)

**脚本**: `ws_message.py`

**场景**: 模拟1000名骑手在线收发消息，验证长连接稳定性和消息吞吐量

**协议**: WebSocket

**参数**:
- `NUM_RIDERS`: 骑手数量 (默认: `1000`)
- `DURATION_SECONDS`: 持续时间 (默认: `300`)
- `SERVER_URL`: WebSocket服务地址 (默认: `ws://localhost:8000/ws/rider/message`)

**运行命令**:
```bash
# 安装依赖
pip install websockets

# 运行压测
python ws_message.py
```

**技术特点**:
- 使用 asyncio 异步 WebSocket 客户端
- 模拟多种消息类型（文本、系统、订单、促销）
- 记录延迟统计（P50/P95/P99）
- 验证长连接稳定性和消息吞吐量
- 保存结果到JSON文件

### 4. 智能抢单推送 (gRPC)

**脚本**: `grpc_accept_order.sh`

**场景**: 模拟 1000 名骑手同时抢单，1000 并发，总请求 10 万次

**协议**: gRPC

**参数**:
- `TARGET_HOST`: 服务地址 (默认: `localhost:9000`)
- `CONCURRENT_CLIENTS`: 并发客户端数 (默认: `1000`)
- `TOTAL_REQUESTS`: 总请求数 (默认: `100000`)
- `DURATION`: 持续时间 (默认: `60s`)

**运行命令**:
```bash
# 使用默认参数
./grpc_accept_order.sh

# 自定义参数
TARGET_HOST=localhost:9000 \
CONCURRENT_CLIENTS=1000 \
TOTAL_REQUESTS=100000 \
./grpc_accept_order.sh
```

**技术特点**:
- 使用 ghz 进行 gRPC 压测
- 生成随机骑手 ID、经纬度和订单 ID
- 模拟 Redis 分布式锁场景
- 支持 Kafka 消息队列集成

### 2. 多订单处理 (gRPC)

**脚本**: `grpc_update_order_status.sh`

**场景**: 模拟 RabbitMQ 削峰后的订单状态更新，1000 并发，持续 5 分钟

**协议**: gRPC

**参数**:
- `TARGET_HOST`: 服务地址 (默认: `localhost:9000`)
- `CONCURRENT_CLIENTS`: 并发客户端数 (默认: `1000`)
- `DURATION`: 持续时间 (默认: `300s`)

**运行命令**:
```bash
# 使用默认参数
./grpc_update_order_status.sh

# 自定义参数
TARGET_HOST=localhost:9000 \
CONCURRENT_CLIENTS=1000 \
DURATION=300s \
./grpc_update_order_status.sh
```

**技术特点**:
- 使用 ghz 进行 gRPC 压测
- 模拟订单状态更新（待接单、已接单、配送中、已完成、已取消）
- 支持 RabbitMQ 消息队列

### 3. 多订单处理 (HTTP)

**脚本**: `http_update_order_status.sh`

**场景**: 模拟 HTTP 协议的订单状态更新，使用 wrk 进行压测

**协议**: HTTP/JSON

**参数**:
- `TARGET_URL`: 服务地址 (默认: `http://localhost:8000`)
- `CONCURRENT_CONNECTIONS`: 并发连接数 (默认: `1000`)
- `DURATION`: 持续时间 (默认: `300s`)
- `THREADS`: 线程数 (默认: `4`)

**运行命令**:
```bash
# 使用默认参数
./http_update_order_status.sh

# 自定义参数
TARGET_URL=http://localhost:8000 \
CONCURRENT_CONNECTIONS=1000 \
DURATION=300s \
./http_update_order_status.sh
```

**技术特点**:
- 使用 wrk 进行 HTTP 压测
- Lua 脚本生成随机请求
- 支持高并发连接

### 4. 智能抢单推送 (Locust)

**脚本**: `locust_order_push.py`

**场景**: 模拟 WebSocket + gRPC 混合场景，1000 并发用户，持续 5 分钟

**协议**: WebSocket (模拟) + gRPC (模拟)

**参数**:
- `NUM_RIDERS`: 骑手数量 (默认: `1000`)
- `TOTAL_REQUESTS`: 总请求数 (默认: `100000`)
- `DURATION_SECONDS`: 持续时间 (默认: `300`)

**运行命令**:
```bash
# 安装依赖
pip install locust

# 运行压测
python locust_order_push.py
```

**技术特点**:
- 使用 Locust 进行混合场景压测
- 模拟 WebSocket 订阅订单推送
- 模拟 LBS 定位更新
- 支持分布式部署

## 依赖安装

### 1. wrk (HTTP 压测工具)

```bash
# Mac
brew install wrk

# Linux
git clone https://github.com/wg/wrk.git
cd wrk
make
sudo cp wrk /usr/local/bin/

# Windows
# 下载预编译版本: https://github.com/wg/wrk/releases
```

### 2. Locust (Python)

```bash
pip install locust
```

### 3. websockets (Python WebSocket)

```bash
pip install websockets
```

### 4. graphviz (火焰图生成)

```bash
# Mac
brew install graphviz

# Ubuntu/Debian
sudo apt-get install graphviz

# CentOS/RHEL
sudo yum install graphviz
```

## 运行示例

### 示例 1: 智能抢单压测

```bash
cd scripts/benchmark

# 使用默认参数
./grpc_accept_order.sh

# 自定义参数
TARGET_HOST=localhost:9000 \
CONCURRENT_CLIENTS=1000 \
TOTAL_REQUESTS=100000 \
DURATION=60s \
./grpc_accept_order.sh
```

**预期输出**:
```
==========================================
智能抢单推送压测 - gRPC协议
==========================================
目标服务: localhost:9000
并发客户端: 1000
总请求数: 100000
持续时间: 60s
==========================================
...
Summary:
  Total:        100000
  Slowest:      0.5234s
  Fastest:      0.0123s
  Average:      0.0456s
  Requests/sec: 1666.67
  ...
```

### 示例 2: 订单状态更新压测

```bash
cd scripts/benchmark

# 使用默认参数
./grpc_update_order_status.sh

# 自定义参数
TARGET_HOST=localhost:9000 \
CONCURRENT_CLIENTS=1000 \
DURATION=300s \
./grpc_update_order_status.sh
```

### 示例 3: HTTP 压测

```bash
cd scripts/benchmark

# 使用默认参数
./http_update_order_status.sh

# 自定义参数
TARGET_URL=http://localhost:8000 \
CONCURRENT_CONNECTIONS=1000 \
DURATION=300s \
THREADS=8 \
./http_update_order_status.sh
```

### 示例 4: Locust 混合场景压测

```bash
cd scripts/benchmark

# 安装依赖
pip install locust

# 运行压测
python locust_order_push.py

# 访问 Web 界面
# http://localhost:8089
```

## 性能分析

### 使用 pprof 进行性能分析

#### 1. 启用 pprof

在 `cmd/hellokratos/main.go` 中添加：

```go
import (
    _ "net/http/pprof"
    "net/http"
)

go func() {
    http.ListenAndServe("0.0.0.0:6060", nil)
}()
```

#### 2. 采集 CPU 性能数据

```bash
# 采集 30 秒的 CPU 数据
go tool pprof -http=:8080 http://localhost:6060/debug/pprof/profile?seconds=30
```

#### 3. 采集内存性能数据

```bash
# 采集内存数据
go tool pprof -http=:8080 http://localhost:6060/debug/pprof/heap
```

#### 4. 生成火焰图

```bash
# 生成 PNG 火焰图
go tool pprof -png -output=cpu_profile.png \
    http://localhost:6060/debug/pprof/profile?seconds=30

# 生成 SVG 火焰图
go tool pprof -svg -output=cpu_profile.svg \
    http://localhost:6060/debug/pprof/profile?seconds=30
```

详细使用方法请参考 [PPROF_ANALYSIS_GUIDE.md](./PPROF_ANALYSIS_GUIDE.md)

## 常见问题

### Q1: 如何修改压测参数？

**A**: 所有脚本都支持通过环境变量自定义参数，例如：

```bash
TARGET_HOST=localhost:9000 \
CONCURRENT_CLIENTS=500 \
TOTAL_REQUESTS=50000 \
./grpc_accept_order.sh
```

### Q2: 压测时服务响应慢怎么办？

**A**: 可能的原因：
- 数据库查询慢：添加索引、优化查询
- Redis 操作慢：使用 Pipeline、添加缓存
- 内存不足：增加内存、优化内存使用
- CPU 使用率高：优化算法、减少计算

使用 pprof 分析性能瓶颈。

### Q3: 如何分析压测结果？

**A**: ghz 和 wrk 都会输出详细的统计信息：
- Total: 总请求数
- Slowest/Fastest/Average: 最慢/最快/平均响应时间
- Requests/sec: 每秒请求数
- P50/P95/P99: 响应时间百分位

### Q4: 如何进行分布式压测？

**A**: Locust 支持分布式部署：

```bash
# 主节点
locust -f locust_order_push.py --master

# 工作节点
locust -f locust_order_push.py --worker --master-host=192.168.1.100
```

### Q5: 如何监控压测过程中的服务状态？

**A**: 可以使用以下命令监控：

```bash
# 查看 CPU 使用率
top -p $(pgrep hellokratos)

# 查看内存使用率
ps aux | grep hellokratos

# 查看 goroutine 数量
curl http://localhost:6060/debug/pprof/goroutine?debug=1
```

## 性能基准

### 期望指标

| 场景 | P95 | P99 | 说明 |
|------|-----|-----|------|
| 订单接单 | < 50ms | < 100ms | 高并发抢单场景 |
| 订单查询 | < 30ms | < 60ms | 订单列表查询 |
| LBS定位更新 | < 20ms | < 40ms | 实时定位更新 |

### 监控指标

| 指标 | 阈值 | 说明 |
|------|------|------|
| CPU 使用率 | < 70% | 避免 CPU 瓶颈 |
| 内存使用率 | < 80% | 避免内存泄漏 |
| Goroutine 数量 | < 10000 | 避免协程泄漏 |
| 数据库连接数 | < 100 | 避免连接池耗尽 |

## 贡献

如有问题或建议，请提交 Issue 或 PR。

## 许可证

MIT License
