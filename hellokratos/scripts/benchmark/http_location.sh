#!/bin/bash
# 骑手位置上报接口压测脚本 - HTTP POST (高频低延迟)
# 用于模拟1000名骑手每秒上报1次，持续5分钟

# 配置参数
TARGET_URL="${TARGET_URL:-http://localhost:8001}"
CONCURRENT_CLIENTS="${CONCURRENT_CLIENTS:-1000}"
DURATION="${DURATION:-300s}"  # 5分钟
THREADS="${THREADS:-8}"

echo "=========================================="
echo "骑手位置上报接口压测 - HTTP POST (高频低延迟)"
echo "=========================================="
echo "目标URL: $TARGET_URL"
echo "并发客户端: $CONCURRENT_CLIENTS"
echo "线程数: $THREADS"
echo "持续时间: $DURATION"
echo "=========================================="

# 创建Lua脚本用于生成随机请求
cat > /tmp/wrk_location.lua << 'EOF'
wrk.method = "POST"
wrk.headers["Content-Type"] = "application/json"
wrk.headers["Accept"] = "application/json"

-- 预生成骑手ID池
local rider_ids = {}
for i = 1, 10000 do
    rider_ids[i] = string.format("rider_%d", 10000 + i)
end

-- 北京周边随机经纬度
local base_lat = 39.9042
local base_lng = 116.4074

function request()
    local index = math.random(1, #rider_ids)
    local rider_id = rider_ids[index]
    
    -- 生成随机经纬度（约1公里范围内）
    local lat = base_lat + math.random(-1000, 1000) / 10000
    local lng = base_lng + math.random(-1000, 1000) / 10000
    local timestamp = os.time()
    
    local body = string.format(
        '{"rider_id":"%s","latitude":%.6f,"longitude":%.6f,"timestamp":%d}',
        rider_id,
        lat,
        lng,
        timestamp
    )
    
    wrk.body = body
    return wrk.format(nil, "/api/v1/rider/location")
end
EOF

echo "开始压测..."
echo ""

# 使用 wrk 进行 HTTP 压测
wrk -t$THREADS -c$CONCURRENT_CLIENTS -d$DURATION \
    --script /tmp/wrk_location.lua \
    --latency \
    "$TARGET_URL"

# 清理临时文件
rm -f /tmp/wrk_location.lua

echo ""
echo "=========================================="
echo "压测完成！"
echo "=========================================="
