#!/bin/bash
# 骑手手机号注册接口压测脚本 - HTTP POST
# 用于验证接口限流/防刷能力

# 配置参数
TARGET_URL="${TARGET_URL:-http://localhost:8001}"
CONCURRENT_CLIENTS="${CONCURRENT_CLIENTS:-1000}"
DURATION="${DURATION:-180s}"  # 3分钟
THREADS="${THREADS:-8}"

echo "=========================================="
echo "骑手手机号注册接口压测 - HTTP POST"
echo "=========================================="
echo "目标URL: $TARGET_URL"
echo "并发客户端: $CONCURRENT_CLIENTS"
echo "线程数: $THREADS"
echo "持续时间: $DURATION"
echo "=========================================="

# 创建Lua脚本用于生成随机请求
cat > /tmp/wrk_register.lua << 'EOF'
wrk.method = "POST"
wrk.headers["Content-Type"] = "application/json"
wrk.headers["Accept"] = "application/json"

-- 预生成手机号池
local phones = {}
for i = 1, 10000 do
    local prefix = {"138", "139", "158", "159", "188", "189"}
    local p = prefix[math.random(1, #prefix)]
    local suffix = string.format("%08d", math.random(0, 99999999))
    phones[i] = p .. suffix
end

-- 预生成验证码池
local codes = {}
for i = 1, 10000 do
    codes[i] = string.format("%06d", math.random(0, 999999))
end

function request()
    local index = math.random(1, #phones)
    local phone = phones[index]
    local code = codes[index]
    
    local body = string.format(
        '{"phone":"%s","code":"%s"}',
        phone,
        code
    )
    
    wrk.body = body
    return wrk.format(nil, "/api/v1/rider/register")
end
EOF

echo "开始压测..."
echo ""

# 使用 wrk 进行 HTTP 压测
wrk -t$THREADS -c$CONCURRENT_CLIENTS -d$DURATION \
    --script /tmp/wrk_register.lua \
    --latency \
    "$TARGET_URL"

# 清理临时文件
rm -f /tmp/wrk_register.lua

echo ""
echo "=========================================="
echo "压测完成！"
echo "=========================================="
