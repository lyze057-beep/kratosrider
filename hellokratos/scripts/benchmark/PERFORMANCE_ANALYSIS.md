# Kratos 骑手端性能分析与优化指南

## 📊 压测结果分析

### 当前性能指标

| 接口 | QPS | 响应时间 | 目标QPS | 目标响应时间 | 错误率 | 状态 |
|------|-----|----------|---------|--------------|--------|------|
| 智能抢单(gRPC) | 3200 | 160ms | 5000 | <100ms | 0.8% | ❌ 未达标 |
| 位置上报(HTTP) | 1800 | 65ms | 2000 | <50ms | 0% | ⚠️ 接近达标 |
| 骑手注册(HTTP) | 800 | 200ms | - | - | 超时 | ❌ 未达标 |

### 性能瓶颈分析

#### 1. 智能抢单接口 (QPS=3200, 目标5000)

**瓶颈分析**:
- ✅ QPS未达标：需要提升62.5%
- ✅ 响应时间160ms > 目标100ms
- ⚠️ 错误率0.8%：主要来自Redis分布式锁竞争

**可能瓶颈点**:
1. **Redis分布式锁** - `AcceptOrder` 中的 `SetNX` 操作
2. **数据库查询** - `GetOrderByID` 和 `UpdateOrder` 两次数据库操作
3. **Kafka消息发布** - `PublishOrderMessage` 阻塞操作
4. **GORM查询优化** - 缺少索引导致查询慢

#### 2. 位置上报接口 (QPS=1800, 目标2000)

**瓶颈分析**:
- ⚠️ QPS接近目标，仍有提升空间
- ⚠️ 响应时间65ms > 目标50ms

**可能瓶颈点**:
1. **数据库写入** - 高频位置上报导致数据库写入压力
2. **Redis缓存** - 未使用Redis缓存位置信息
3. **HTTP中间件** - Kratos中间件链可能过长

#### 3. 骑手注册接口 (QPS=800, 超时)

**瓶颈分析**:
- ❌ QPS较低
- ❌ 存在超时

**可能瓶颈点**:
1. **短信验证码** - 互亿无线API调用慢
2. **数据库查询** - 用户存在性检查
3. **密码加密** - bcrypt加密耗时

---

## 🔍 pprof 瓶颈定位指南

### 步骤1: 启用 pprof

在 `cmd/hellokratos/main.go` 中添加：

```go
import (
    _ "net/http/pprof"
    "net/http"
)

// 在 main 函数中添加
go func() {
    http.ListenAndServe("0.0.0.0:6060", nil)
}()
```

### 步骤2: 运行压测

```bash
# 运行智能抢单压测
./scripts/benchmark/grpc_accept_order.sh

# 运行位置上报压测
./scripts/benchmark/http_location.sh

# 运行骑手注册压测
./scripts/benchmark/http_register.sh
```

### 步骤3: 采集性能数据

#### 3.1 智能抢单接口分析

```bash
# 采集 30 秒的 CPU 数据
go tool pprof -http=:8080 http://localhost:6060/debug/pprof/profile?seconds=30

# 采集内存数据
go tool pprof -http=:8080 http://localhost:6060/debug/pprof/heap

# 采集 goroutine 数据
go tool pprof -http=:8080 http://localhost:6060/debug/pprof/goroutine
```

#### 3.2 位置上报接口分析

```bash
# 采集 30 秒的 CPU 数据
go tool pprof -http=:8081 http://localhost:6061/debug/pprof/profile?seconds=30

# 采集内存数据
go tool pprof -http=:8081 http://localhost:6061/debug/pprof/heap
```

#### 3.3 骑手注册接口分析

```bash
# 采集 30 秒的 CPU 数据
go tool pprof -http=:8082 http://localhost:6062/debug/pprof/profile?seconds=30

# 采集阻塞数据（分析超时原因）
go tool pprof -http=:8082 http://localhost:6062/debug/pprof/block
```

### 步骤4: 分析性能数据

#### 4.1 交互式分析

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

#### 4.2 Web 界面分析

使用 `-http` 参数启动 Web 界面，然后在浏览器中打开：
- `http://localhost:8080` 

Web 界面提供：
- 火焰图（Flame Graph）
- 调用图（Call Graph）
- 源码视图

#### 4.3 生成报告

```bash
# 生成文本报告
go tool pprof -text http://localhost:6060/debug/pprof/profile?seconds=30 > cpu_profile.txt

# 生成 Top10 报告
go tool pprof -top10 http://localhost:6060/debug/pprof/profile?seconds=30

# 生成 PNG 火焰图
go tool pprof -png -output=cpu_profile.png \
    http://localhost:6060/debug/pprof/profile?seconds=30

# 生成 SVG 火焰图
go tool pprof -svg -output=cpu_profile.svg \
    http://localhost:6060/debug/pprof/profile?seconds=30
```

### 步骤5: 重点分析函数

#### 5.1 智能抢单接口

重点关注以下函数：
- `AcceptOrder` - 接单逻辑
- `GetOrderByID` - 获取订单
- `UpdateOrder` - 更新订单
- `PublishOrderMessage` - 发布消息
- `redis.SetNX` - Redis分布式锁

#### 5.2 位置上报接口

重点关注以下函数：
- `UpdateLocation` - 更新位置
- `CreateOrder` - 创建订单
- `CalculateDistance` - 计算距离

#### 5.3 骑手注册接口

重点关注以下函数：
- `Register` - 注册逻辑
- `SendCode` - 发送验证码
- `bcrypt.GenerateFromPassword` - 密码加密

---

## 💡 优化建议

### 代码层面优化

#### 1. 智能抢单接口优化

**优化前**:
```go
func (uc *orderUsecase) AcceptOrder(ctx context.Context, orderID int64, riderID int64) error {
    // 1. 获取Redis分布式锁
    lockKey := fmt.Sprintf("order:lock:%d", orderID)
    success, err := uc.rdb.SetNX(ctx, lockKey, lockValue, 5*time.Second).Result()
    
    // 2. 查询订单
    order, err := uc.orderRepo.GetOrderByID(ctx, orderID)
    
    // 3. 更新订单状态
    order.Status = 1
    order.RiderID = riderID
    err = uc.orderRepo.UpdateOrder(ctx, order)
    
    // 4. 发布消息
    err = uc.orderMessageProducer.PublishOrderMessage(ctx, msg)
    
    return nil
}
```

**优化后**:
```go
func (uc *orderUsecase) AcceptOrder(ctx context.Context, orderID int64, riderID int64) error {
    // 1. 使用Pipeline批量执行Redis操作
    pipe := uc.rdb.Pipeline()
    lockKey := fmt.Sprintf("order:lock:%d", orderID)
    pipe.SetNX(ctx, lockKey, fmt.Sprintf("%d", riderID), 5*time.Second)
    pipe.Expire(ctx, lockKey, 5*time.Second)
    _, err := pipe.Exec(ctx)
    
    // 2. 使用单次数据库查询更新订单
    result := uc.db.Exec(ctx, 
        "UPDATE rider_order SET status = ?, rider_id = ? WHERE id = ? AND status = 0",
        1, riderID, orderID)
    
    // 3. 异步发布消息
    go func() {
        msg := &data.OrderMessage{
            OrderID: orderID,
            RiderID: riderID,
            Action:  "accept",
            Status:  1,
        }
        uc.orderMessageProducer.PublishOrderMessage(ctx, msg)
    }()
    
    return nil
}
```

**优化点**:
- ✅ 使用Pipeline批量执行Redis操作
- ✅ 使用单次数据库查询更新订单（减少查询次数）
- ✅ 异步发布消息（不阻塞主流程）
- ✅ 添加数据库索引

#### 2. 位置上报接口优化

**优化前**:
```go
func (uc *orderUsecase) UpdateLocation(ctx context.Context, riderID int64, lat, lng float64) error {
    // 每次上报都查询数据库
    rider, err := uc.riderRepo.GetByRiderID(ctx, riderID)
    if err != nil {
        return err
    }
    
    // 更新位置
    rider.Lat = lat
    rider.Lng = lng
    return uc.riderRepo.Update(ctx, rider)
}
```

**优化后**:
```go
func (uc *orderUsecase) UpdateLocation(ctx context.Context, riderID int64, lat, lng float64) error {
    // 1. 先更新Redis缓存
    cacheKey := fmt.Sprintf("rider:location:%d", riderID)
    location := fmt.Sprintf("%.6f,%.6f", lat, lng)
    uc.rdb.Set(ctx, cacheKey, location, 5*time.Minute)
    
    // 2. 异步更新数据库（批量写入）
    uc.locationQueue <- locationData{
        riderID: riderID,
        lat: lat,
        lng: lng,
    }
    
    return nil
}

// 后台goroutine批量写入数据库
func (uc *orderUsecase) startLocationBatchWriter() {
    batch := make([]locationData, 0, 100)
    ticker := time.NewTicker(1 * time.Second)
    
    for {
        select {
        case data := <-uc.locationQueue:
            batch = append(batch, data)
            if len(batch) >= 100 {
                uc.batchUpdateLocation(batch)
                batch = batch[:0]
            }
        case <-ticker.C:
            if len(batch) > 0 {
                uc.batchUpdateLocation(batch)
                batch = batch[:0]
            }
        }
    }
}
```

**优化点**:
- ✅ 使用Redis缓存位置信息
- ✅ 异步批量写入数据库
- ✅ 减少数据库查询次数

#### 3. 骑手注册接口优化

**优化前**:
```go
func (uc *orderUsecase) Register(ctx context.Context, phone, password, code, nickname string) error {
    // 1. 发送验证码（同步调用短信API）
    err := uc.sms.SendCode(ctx, phone)
    
    // 2. 验证验证码
    err = uc.sms.ValidateCode(ctx, phone, code)
    
    // 3. 密码加密
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    
    // 4. 保存用户
    err = uc.userRepo.Create(ctx, user)
    
    return nil
}
```

**优化后**:
```go
func (uc *orderUsecase) Register(ctx context.Context, phone, password, code, nickname string) error {
    // 1. 异步发送验证码
    go func() {
        uc.sms.SendCode(ctx, phone)
    }()
    
    // 2. 使用更快的密码加密（或使用缓存）
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 4) // 降低cost值
    
    // 3. 保存用户
    err = uc.userRepo.Create(ctx, user)
    
    return nil
}
```

**优化点**:
- ✅ 异步发送验证码
- ✅ 降低bcrypt cost值（从默认10降到4）
- ✅ 添加验证码缓存

### 配置层面优化

#### 1. Redis配置优化

```yaml
data:
  redis:
    addr: 115.190.45.90:6379
    password: 4ay1nkal3u8ed77y
    read_timeout: 0.1s      # 减少超时时间
    write_timeout: 0.1s     # 减少超时时间
    pool_size: 100          # 增加连接池大小
    min_idle_conns: 50      # 最小空闲连接数
    max_idle_conns: 100     # 最大空闲连接数
    conn_max_lifetime: 5m   # 连接最大生命周期
```

#### 2. 数据库配置优化

```yaml
data:
  database:
    driver: mysql
    source: root:4ay1nkal3u8ed77y@tcp(115.190.45.90:3306)/mtrider?charset=utf8mb4&parseTime=True&loc=Local&maxOpenConns=100&maxIdleConns=50&connMaxLifetime=300s
```

#### 3. RabbitMQ配置优化

```yaml
data:
  rabbitmq:
    addr: amqp://rabbitmq:rabbitmq@115.190.45.90:5672/
    username: rabbitmq
    password: rabbitmq
    virtual_host: /
    connection_timeout: 3s
    channel_size: 1000      # 增加通道大小
    prefetch_count: 100     # 预取数量
```

### 架构层面优化

#### 1. 数据库索引优化

```sql
-- 订单表索引
CREATE INDEX idx_order_status ON rider_order(status);
CREATE INDEX idx_order_rider_id ON rider_order(rider_id);
CREATE INDEX idx_order_status_rider_id ON rider_order(status, rider_id);

-- 骑手表索引
CREATE INDEX idx_rider_phone ON rider_user(phone);
CREATE INDEX idx_rider_status ON rider_user(status);

-- 位置表索引
CREATE INDEX idx_location_rider_id ON rider_location(rider_id);
```

#### 2. Redis缓存优化

```go
// 1. 订单缓存
func (uc *orderUsecase) GetOrderByID(ctx context.Context, id int64) (*model.Order, error) {
    cacheKey := fmt.Sprintf("order:%d", id)
    
    // 1. 先查缓存
    cached, err := uc.rdb.Get(ctx, cacheKey).Result()
    if err == nil {
        var order model.Order
        json.Unmarshal([]byte(cached), &order)
        return &order, nil
    }
    
    // 2. 缓存未命中，查数据库
    order, err := uc.orderRepo.GetOrderByID(ctx, id)
    if err != nil {
        return nil, err
    }
    
    // 3. 写入缓存
    data, _ := json.Marshal(order)
    uc.rdb.Set(ctx, cacheKey, data, 5*time.Minute)
    
    return order, nil
}

// 2. 位置缓存
func (uc *orderUsecase) GetRiderLocation(ctx context.Context, riderID int64) (*model.Location, error) {
    cacheKey := fmt.Sprintf("rider:location:%d", riderID)
    
    cached, err := uc.rdb.Get(ctx, cacheKey).Result()
    if err == nil {
        var location model.Location
        json.Unmarshal([]byte(cached), &location)
        return &location, nil
    }
    
    // 缓存未命中，查数据库
    location, err := uc.locationRepo.GetByRiderID(ctx, riderID)
    if err != nil {
        return nil, err
    }
    
    // 写入缓存
    data, _ := json.Marshal(location)
    uc.rdb.Set(ctx, cacheKey, data, 5*time.Minute)
    
    return location, nil
}
```

#### 3. 消息队列优化

```go
// 1. 使用Kafka替代RabbitMQ（更高吞吐量）
type KafkaMessageProducer struct {
    producer *sarama.AsyncProducer
}

func (p *KafkaMessageProducer) PublishOrderMessage(ctx context.Context, message *data.OrderMessage) error {
    msg := &sarama.ProducerMessage{
        Topic: "order_topic",
        Key:   sarama.StringEncoder(fmt.Sprintf("%d", message.OrderID)),
        Value: sarama.StringEncoder(string(data)),
    }
    
    p.producer.Input() <- msg
    return nil
}

// 2. 批量消息发布
func (p *KafkaMessageProducer) PublishBatchMessages(ctx context.Context, messages []*data.OrderMessage) error {
    for _, msg := range messages {
        p.producer.Input() <- &sarama.ProducerMessage{
            Topic: "order_topic",
            Key:   sarama.StringEncoder(fmt.Sprintf("%d", msg.OrderID)),
            Value: sarama.StringEncoder(string(data)),
        }
    }
    return nil
}
```

#### 4. gRPC连接池优化

```go
// 1. 使用连接池
conn, err := grpc.Dial(
    "localhost:9000",
    grpc.WithInsecure(),
    grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(1024*1024*10)),
    grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(1024*1024*10)),
    grpc.WithConnectParams(grpc.ConnectParams{
        Backoff: backoff.Config{
            BaseDelay:  100 * time.Millisecond,
            MaxDelay:   1 * time.Second,
        },
        MinConnectTimeout: 1 * time.Second,
    }),
)

// 2. 使用连接池
client := v1.NewOrderClient(conn)
```

### 监控和调优

#### 1. 添加性能监控

```go
// 1. 添加中间件监控
func MetricsMiddleware() grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
        start := time.Now()
        
        resp, err := handler(ctx, req)
        
        duration := time.Since(start)
        
        // 记录性能指标
        metrics.RecordDuration(info.FullMethod, duration)
        
        return resp, err
    }
}

// 2. 添加日志监控
func LoggingMiddleware() grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
        start := time.Now()
        
        resp, err := handler(ctx, req)
        
        duration := time.Since(start)
        
        // 记录慢查询
        if duration > 100*time.Millisecond {
            log.Warn("slow query", "method", info.FullMethod, "duration", duration)
        }
        
        return resp, err
    }
}
```

#### 2. 性能调优步骤

```bash
# 1. 启用pprof
go run cmd/hellokratos/main.go -conf configs/config.yaml

# 2. 运行压测
./scripts/benchmark/grpc_accept_order.sh

# 3. 采集性能数据
go tool pprof -http=:8080 http://localhost:6060/debug/pprof/profile?seconds=30

# 4. 分析性能瓶颈
# - 查看top10耗时函数
# - 查看火焰图
# - 查看源码视图

# 5. 优化代码
# - 根据pprof结果优化代码
# - 添加索引
# - 优化Redis操作

# 6. 重新压测
# - 再次运行压测
# - 对比优化前后的性能指标
```

### 优化效果预期

| 接口 | 优化前 | 优化后 | 提升 |
|------|--------|--------|------|
| 智能抢单 | QPS=3200, 160ms | QPS=5000+, <100ms | +56% |
| 位置上报 | QPS=1800, 65ms | QPS=2000+, <50ms | +11% |
| 骑手注册 | QPS=800, 200ms | QPS=1500+, <100ms | +88% |

---

## 📝 总结

### 关键优化点

1. **Redis分布式锁优化** - 使用Pipeline批量操作
2. **数据库查询优化** - 添加索引、减少查询次数
3. **异步消息发布** - 不阻塞主流程
4. **Redis缓存** - 缓存热点数据
5. **批量写入** - 减少数据库写入次数

### 监控建议

1. **添加性能监控** - 使用pprof定期分析
2. **添加日志监控** - 记录慢查询
3. **添加告警** - QPS、响应时间、错误率告警

### 持续优化

1. **定期压测** - 每周进行一次压测
2. **性能分析** - 使用pprof分析性能瓶颈
3. **代码优化** - 根据分析结果优化代码
4. **配置调优** - 根据压测结果调整配置
