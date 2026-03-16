# Kratos骑手端项目性能分析指南 - pprof使用详解

## 目录
1. [pprof简介](#pprof简介)
2. [性能分析步骤](#性能分析步骤)
3. [常见性能瓶颈分析](#常见性能瓶颈分析)
4. [优化建议](#优化建议)

## pprof简介

pprof 是 Go 语言内置的性能分析工具，可以分析：
- **CPU使用情况**：找出耗时最多的函数
- **内存分配**：找出内存分配最多的代码
- **goroutine**：分析协程数量和状态
- **阻塞分析**：找出阻塞操作

## 性能分析步骤

### 步骤1：启用 pprof

在你的 Kratos 项目中，确保已经启用了 pprof。检查 `cmd/hellokratos/main.go`：

```go
import (
    _ "net/http/pprof"
    "net/http"
)

// 在 main 函数中添加 pprof 服务
go func() {
    http.ListenAndServe("0.0.0.0:6060", nil)
}()
```

或者使用 Kratos 的方式：

```go
// 在 wireApp 函数中添加
if bc.Server.Pprof.Enabled {
    go func() {
        _ = http.ListenAndServe(bc.Server.Pprof.Addr, nil)
    }()
}
```

### 步骤2：启动服务

```bash
# 启动你的 Kratos 服务
go run cmd/hellokratos/main.go -conf configs/config.yaml
```

### 步骤3：收集性能数据

#### CPU 分析

```bash
# 采集 30 秒的 CPU 性能数据
go tool pprof -http=:8080 http://localhost:6060/debug/pprof/profile?seconds=30

# 采集 60 秒的 CPU 性能数据
go tool pprof -http=:8080 http://localhost:6060/debug/pprof/profile?seconds=60
```

#### 内存分析

```bash
# 采集内存分配数据
go tool pprof -http=:8080 http://localhost:6060/debug/pprof/heap

# 采集内存分配数据（按分配次数排序）
go tool pprof -http=:8080 -sample_index=alloc_objects http://localhost:6060/debug/pprof/heap

# 采集内存分配数据（按活跃字节数排序）
go tool pprof -http=:8080 -sample_index=inuse_space http://localhost:6060/debug/pprof/heap
```

#### Goroutine 分析

```bash
# 采集 goroutine 数据
go tool pprof -http=:8080 http://localhost:6060/debug/pprof/goroutine

# 采集 goroutine 阻塞数据
go tool pprof -http=:8080 http://localhost:6060/debug/pprof/block
```

### 步骤4：分析性能数据

#### 交互式分析

```bash
# 进入交互式分析界面
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# 常用命令：
top10              # 显示前10个耗时函数
list AcceptOrder   # 显示 AcceptOrder 函数的详细代码
web                # 生成火焰图（需要安装 graphviz）
png                # 生成 PNG 图片
tree               # 以树形结构显示调用关系
```

#### Web 界面分析

使用 `-http` 参数启动 Web 界面，然后在浏览器中打开：
- `http://localhost:8080` 

Web 界面提供：
- 火焰图（Flame Graph）
- 调用图（Call Graph）
- 源码视图

### 步骤5：生成报告

#### 生成火焰图

```bash
# 安装 graphviz（Mac）
brew install graphviz

# 安装 graphviz（Ubuntu/Debian）
sudo apt-get install graphviz

# 生成火焰图
go tool pprof -http=:8080 http://localhost:6060/debug/pprof/profile?seconds=30

# 或者生成 PNG 文件
go tool pprof -png -output=cpu_profile.png http://localhost:6060/debug/pprof/profile?seconds=30
```

#### 生成文本报告

```bash
# 生成文本报告
go tool pprof -text http://localhost:6060/debug/pprof/profile?seconds=30 > cpu_profile.txt

# 生成 Top10 报告
go tool pprof -top10 http://localhost:6060/debug/pprof/profile?seconds=30
```

## 常见性能瓶颈分析

### 1. 数据库查询慢

#### 问题表现
- CPU 分析中看到大量时间花在 `database/sql` 相关函数
- 慢查询日志中有大量记录

#### 分析方法

```bash
# 采集 CPU 数据
go tool pprof -http=:8080 http://localhost:6060/debug/pprof/profile?seconds=30

# 在 Web 界面中查找：
# - gorm.io/gorm
# - github.com/jinzhu
```

#### 优化建议
- 添加数据库索引
- 使用连接池
- 避免 N+1 查询
- 使用批量操作

### 2. Redis 操作慢

#### 问题表现
- CPU 分析中看到大量时间花在 `redis` 相关函数
- 网络延迟高

#### 分析方法

```bash
# 采集 CPU 数据
go tool pprof -http=:8080 http://localhost:6060/debug/pprof/profile?seconds=30

# 在 Web 界面中查找：
# - github.com/redis/go-redis/v9
```

#### 优化建议
- 使用 Pipeline 批量操作
- 添加本地缓存
- 避免大 Key
- 使用连接池

### 3. JSON 序列化慢

#### 问题表现
- CPU 分析中看到大量时间花在 `encoding/json` 相关函数

#### 分析方法

```bash
# 采集 CPU 数据
go tool pprof -http=:8080 http://localhost:6060/debug/pprof/profile?seconds=30

# 在 Web 界面中查找：
# - encoding/json
# - json.Marshal
# - json.Unmarshal
```

#### 优化建议
- 使用 sync.Pool 复用对象
- 使用 fastjson 等更快的库
- 减少 JSON 序列化次数

### 4. 内存泄漏

#### 问题表现
- 内存持续增长
- goroutine 数量持续增长

#### 分析方法

```bash
# 采集内存数据（间隔一段时间）
go tool pprof -http=:8080 http://localhost:6060/debug/pprof/heap

# 对比两次快照
go tool pprof -http=:8080 -base=heap1.prof heap2.prof
```

#### 优化建议
- 检查是否正确释放资源
- 使用 sync.Pool 复用对象
- 避免全局变量持有大对象

### 5. 锁竞争

#### 问题表现
- CPU 分析中看到大量时间花在 `sync.Mutex` 相关函数
- goroutine 阻塞分析中看到大量等待

#### 分析方法

```bash
# 采集阻塞数据
go tool pprof -http=:8080 http://localhost:6060/debug/pprof/block

# 采集 CPU 数据
go tool pprof -http=:8080 http://localhost:6060/debug/pprof/profile?seconds=30
```

#### 优化建议
- 使用读写锁 `sync.RWMutex`
- 使用 channel 替代锁
- 减少锁的粒度

## 实战案例：订单接单性能优化

### 场景描述
骑手抢单时，`AcceptOrder` 函数性能较差，响应时间超过 100ms。

### 分析步骤

#### 1. 启用 pprof

```go
// 在 main.go 中添加
import _ "net/http/pprof"
import "net/http"

go func() {
    http.ListenAndServe("0.0.0.0:6060", nil)
}()
```

#### 2. 运行压测

```bash
# 运行压测脚本
./scripts/benchmark/grpc_accept_order.sh
```

#### 3. 采集性能数据

```bash
# 采集 30 秒的 CPU 数据
go tool pprof -http=:8080 http://localhost:6060/debug/pprof/profile?seconds=30
```

#### 4. 分析结果

在 Web 界面中查看：
- 找到 `AcceptOrder` 函数
- 查看调用关系
- 找到耗时最多的函数

#### 5. 优化

假设分析结果发现：
- Redis 分布式锁耗时较多
- 数据库查询耗时较多

##### 优化 Redis 操作

```go
// 优化前
success, err := uc.rdb.SetNX(ctx, lockKey, lockValue, 5*time.Second).Result()

// 优化后：使用 Pipeline
pipe := uc.rdb.Pipeline()
pipe.SetNX(ctx, lockKey, lockValue, 5*time.Second)
pipe.Get(ctx, lockKey)
_, err := pipe.Exec(ctx)
```

##### 优化数据库查询

```go
// 优化前：多次查询
order, _ := uc.orderRepo.GetOrderByID(ctx, orderID)
order.Status = 1
err := uc.orderRepo.UpdateOrder(ctx, order)

// 优化后：使用单次查询
err := uc.db.Exec(ctx, 
    "UPDATE rider_order SET status = ? WHERE id = ? AND status = 0", 
    1, orderID).Error
```

## 优化建议

### 1. 数据库优化
- 添加索引：`CREATE INDEX idx_order_status ON rider_order(status)`
- 使用连接池：`db.SetMaxOpenConns(100)`
- 避免 N+1 查询：使用 `Preload`

### 2. Redis 优化
- 使用 Pipeline：`pipe := rdb.Pipeline()`
- 添加本地缓存：`cache.Set(key, value, 5*time.Minute)`
- 避免大 Key：`hgetall` 替代 `get`

### 3. 内存优化
- 使用 sync.Pool：`var pool = sync.Pool{New: func() interface{} { ... }}`
- 避免重复分配：复用对象

### 4. 并发优化
- 使用 worker pool：限制并发数量
- 避免锁竞争：使用读写锁

## 性能基准

### 期望指标
- **订单接单**：P95 < 50ms
- **订单查询**：P95 < 30ms
- **LBS定位更新**：P95 < 20ms

### 监控指标
- CPU 使用率 < 70%
- 内存使用率 < 80%
- Goroutine 数量 < 10000

## 总结

pprof 是 Go 语言性能分析的强大工具，通过以下步骤可以快速定位性能瓶颈：

1. 启用 pprof
2. 运行压测
3. 采集性能数据
4. 分析性能数据
5. 优化代码
6. 验证优化效果

持续使用 pprof 进行性能分析，可以确保你的 Kratos 服务保持高性能。
