#!/bin/bash
# 多订单处理压测脚本 - gRPC协议
# 用于模拟RabbitMQ削峰后的订单状态更新场景

# 配置参数
TARGET_HOST="${TARGET_HOST:-localhost:9000}"
CONCURRENT_CLIENTS="${CONCURRENT_CLIENTS:-1000}"
DURATION="${DURATION:-300s}"  # 5分钟

echo "=========================================="
echo "多订单处理压测 - gRPC协议"
echo "=========================================="
echo "目标服务: $TARGET_HOST"
echo "并发客户端: $CONCURRENT_CLIENTS"
echo "持续时间: $DURATION"
echo "=========================================="

# 创建临时数据文件
echo "生成测试数据..."
mkdir -p /tmp/kratos_benchmark
> /tmp/kratos_benchmark/order_status_data.json

# 生成订单ID和状态数据
for i in $(seq 1 10000); do
    order_id=$((20000 + RANDOM % 90000))
    status=$((0 + RANDOM % 5))  # 0-待接单, 1-已接单, 2-配送中, 3-已完成, 4-已取消
    
    echo "{\"order_id\": $order_id, \"status\": $status}" >> /tmp/kratos_benchmark/order_status_data.json
done

echo "测试数据生成完成，共 10000 条"

# 使用 ghz 进行 gRPC 压测
echo ""
echo "开始压测..."
echo ""

ghz \
    --proto ./api/rider/v1/order.proto \
    --call api.rider.v1.Order.UpdateOrderStatus \
    --json /tmp/kratos_benchmark/order_status_data.json \
    --concurrency $CONCURRENT_CLIENTS \
    --duration $DURATION \
    --timeout 30s \
    --insecure \
    $TARGET_HOST

# 清理临时文件
rm -rf /tmp/kratos_benchmark

echo ""
echo "=========================================="
echo "压测完成！"
echo "=========================================="
