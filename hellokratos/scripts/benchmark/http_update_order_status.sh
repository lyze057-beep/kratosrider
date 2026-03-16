#!/bin/bash
# 多订单处理压测脚本 - HTTP协议 (wrk)
# 用于模拟RabbitMQ削峰后的订单状态更新场景

# 配置参数
TARGET_URL="${TARGET_URL:-http://localhost:8000}"
CONCURRENT_CONNECTIONS="${CONCURRENT_CONNECTIONS:-1000}"
DURATION="${DURATION:-300s}"  # 5分钟
THREADS="${THREADS:-4}"

echo "=========================================="
echo "多订单处理压测 - HTTP协议 (wrk)"
echo "=========================================="
echo "目标URL: $TARGET_URL"
echo "并发连接数: $CONCURRENT_CONNECTIONS"
echo "线程数: $THREADS"
echo "持续时间: $DURATION"
echo "=========================================="

# 创建Lua脚本用于生成随机请求
cat > /tmp/wrk_order_status.lua << 'EOF'
wrk.method = "POST"
wrk.headers["Content-Type"] = "application/json"

function request()
    -- 生成随机订单ID和状态
    local order_id = 20000 + math.random(0, 79999)
    local status = math.random(0, 4)  -- 0-待接单, 1-已接单, 2-配送中, 3-已完成, 4-已取消
    
    local body = string.format(
        '{"order_id": %d, "status": %d}',
        order_id,
        status
    )
    
    wrk.body = body
    return wrk.format(nil, "/rider/v1/order/status")
end
EOF

echo "开始压测..."
echo ""

# 使用 wrk 进行 HTTP 压测
wrk -t$THREADS -c$CONCURRENT_CONNECTIONS -d$DURATION \
    --script /tmp/wrk_order_status.lua \
    --latency \
    "$TARGET_URL"

# 清理临时文件
rm -f /tmp/wrk_order_status.lua

echo ""
echo "=========================================="
echo "压测完成！"
echo "=========================================="
