#!/bin/bash
# 智能抢单推送压测脚本 - gRPC协议
# 用于模拟1000名骑手同时抢单的高并发场景

# 配置参数
TARGET_HOST="${TARGET_HOST:-localhost:9002}"
CONCURRENT_CLIENTS="${CONCURRENT_CLIENTS:-1000}"
TOTAL_REQUESTS="${TOTAL_REQUESTS:-100000}"
DURATION="${DURATION:-60s}"

echo "=========================================="
echo "智能抢单推送压测 - gRPC协议"
echo "=========================================="
echo "目标服务: $TARGET_HOST"
echo "并发客户端: $CONCURRENT_CLIENTS"
echo "总请求数: $TOTAL_REQUESTS"
echo "持续时间: $DURATION"
echo "=========================================="

# 生成骑手ID和经纬度随机数据
generate_rider_data() {
    local rider_id=$((10000 + RANDOM % 90000))
    local lat=$(echo "scale=6; 39.9 + ($RANDOM % 1000) / 10000" | bc)
    local lng=$(echo "scale=6; 116.3 + ($RANDOM % 1000) / 10000" | bc)
    local order_id=$((20000 + RANDOM % 90000))
    echo "$rider_id $lat $lng $order_id"
}

# 创建临时数据文件
echo "生成测试数据..."
mkdir -p /tmp/kratos_benchmark
> /tmp/kratos_benchmark/rider_data.json

for i in $(seq 1 $TOTAL_REQUESTS); do
    data=$(generate_rider_data)
    rider_id=$(echo $data | cut -d' ' -f1)
    lat=$(echo $data | cut -d' ' -f2)
    lng=$(echo $data | cut -d' ' -f3)
    order_id=$(echo $data | cut -d' ' -f4)
    
    echo "{\"rider_id\": $rider_id, \"lat\": $lat, \"lng\": $lng, \"order_id\": $order_id}" >> /tmp/kratos_benchmark/rider_data.json
done

echo "测试数据生成完成，共 $TOTAL_REQUESTS 条"

# 使用 ghz 进行 gRPC 压测
# ghz 是一个高性能的 gRPC 压测工具
echo ""
echo "开始压测..."
echo ""

ghz \
    --proto ./api/rider/v1/order.proto \
    --call api.rider.v1.Order.AcceptOrder \
    --json /tmp/kratos_benchmark/rider_data.json \
    --concurrency $CONCURRENT_CLIENTS \
    --total $TOTAL_REQUESTS \
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
