#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
骑手消息通知压测脚本 - WebSocket
用于模拟1000名骑手在线收发消息，验证长连接稳定性和消息吞吐量
"""

import asyncio
import json
import random
import time
import statistics
from datetime import datetime
from websockets import connect
from websockets.exceptions import ConnectionClosed
import sys

# 配置参数
NUM_RIDERS = 1000  # 骑手数量
DURATION_SECONDS = 300  # 持续时间（秒）
SERVER_URL = "ws://localhost:8000/ws/rider/message"  # WebSocket服务地址

# 消息类型
MESSAGE_TYPES = {
    "text": "文本消息",
    "system": "系统通知",
    "order": "订单通知",
    "promotion": "促销活动"
}

# 消息内容模板
MESSAGE_TEMPLATES = {
    "text": [
        "你好，我在哪里接单？",
        "订单状态更新了吗？",
        "需要帮助，请联系客服",
        "今天天气怎么样？",
        "最近有什么优惠活动？"
    ],
    "system": [
        "系统维护通知：今晚22:00-23:00进行系统升级",
        "重要提醒：请确保每天完成至少10单",
        "新功能上线：智能派单系统已开启",
        "安全提示：请佩戴头盔，注意交通安全"
    ],
    "order": [
        "您有新的订单，请及时处理",
        "订单已取消，请查看取消原因",
        "订单已完成，感谢您的服务",
        "订单状态更新：配送中"
    ],
    "promotion": [
        "🎉 新人特惠：首单立减10元",
        "🔥 热门活动：满50减15，限时抢购",
        "💰 骑手奖励：今日完成20单额外奖励50元",
        "🎁 福利放送：连续7天签到领好礼"
    ]
}

class RiderClient:
    """模拟骑手WebSocket客户端"""
    
    def __init__(self, rider_id, websocket):
        self.rider_id = rider_id
        self.websocket = websocket
        self.messages_sent = 0
        self.messages_received = 0
        self.latencies = []
        self.start_time = time.time()
    
    async def send_message(self, message_type, content):
        """发送消息"""
        message = {
            "rider_id": self.rider_id,
            "message_type": message_type,
            "content": content,
            "timestamp": int(time.time()),
            "message_id": f"{self.rider_id}_{self.messages_sent}"
        }
        
        start_time = time.time()
        await self.websocket.send(json.dumps(message))
        self.messages_sent += 1
        
        return start_time
    
    async def receive_message(self):
        """接收消息"""
        try:
            response = await self.websocket.recv()
            end_time = time.time()
            
            # 计算延迟
            latency = (end_time - self.start_time) * 1000  # 转换为毫秒
            self.latencies.append(latency)
            self.messages_received += 1
            
            return response
        except ConnectionClosed:
            return None

async def rider_task(rider_id, messages_queue, stats):
    """骑手任务：发送和接收消息"""
    try:
        async with connect(SERVER_URL) as websocket:
            client = RiderClient(rider_id, websocket)
            
            # 持续发送和接收消息
            end_time = time.time() + DURATION_SECONDS
            
            while time.time() < end_time:
                # 随机选择消息类型
                message_type = random.choice(list(MESSAGE_TYPES.keys()))
                content = random.choice(MESSAGE_TEMPLATES[message_type])
                
                # 发送消息
                await client.send_message(message_type, content)
                
                # 接收消息
                response = await client.receive_message()
                
                # 短暂延迟，模拟真实场景
                await asyncio.sleep(random.uniform(0.1, 0.5))
            
            # 收集统计信息
            stats['clients'].append({
                'rider_id': rider_id,
                'messages_sent': client.messages_sent,
                'messages_received': client.messages_received,
                'latencies': client.latencies
            })
            
    except Exception as e:
        print(f"骑手 {rider_id} 连接错误: {e}")
        stats['errors'].append(str(e))

async def run_benchmark():
    """运行压测"""
    
    print("=" * 70)
    print("骑手消息通知压测 - WebSocket")
    print("=" * 70)
    print(f"目标服务: {SERVER_URL}")
    print(f"并发骑手数: {NUM_RIDERS}")
    print(f"持续时间: {DURATION_SECONDS}秒 ({DURATION_SECONDS/60}分钟)")
    print("=" * 70)
    
    # 统计信息
    stats = {
        'clients': [],
        'errors': [],
        'start_time': time.time()
    }
    
    # 创建骑手任务
    tasks = []
    for i in range(NUM_RIDERS):
        rider_id = f"rider_{10000 + i}"
        tasks.append(rider_task(rider_id, None, stats))
    
    print(f"\n开始压测，预计持续 {DURATION_SECONDS} 秒...")
    print(f"预计总消息数: {NUM_RIDERS * (DURATION_SECONDS // 0.3)} 条")
    print("-" * 70)
    
    # 运行所有任务
    await asyncio.gather(*tasks)
    
    # 计算统计结果
    end_time = time.time()
    duration = end_time - stats['start_time']
    
    # 合并所有延迟数据
    all_latencies = []
    total_sent = 0
    total_received = 0
    
    for client in stats['clients']:
        all_latencies.extend(client['latencies'])
        total_sent += client['messages_sent']
        total_received += client['messages_received']
    
    # 计算延迟统计
    if all_latencies:
        avg_latency = statistics.mean(all_latencies)
        min_latency = min(all_latencies)
        max_latency = max(all_latencies)
        p50_latency = statistics.quantiles(all_latencies, n=100)[49]
        p95_latency = statistics.quantiles(all_latencies, n=100)[94]
        p99_latency = statistics.quantiles(all_latencies, n=100)[98]
    else:
        avg_latency = min_latency = max_latency = p50_latency = p95_latency = p99_latency = 0
    
    # 计算吞吐量
    messages_per_second = (total_sent + total_received) / duration
    
    # 打印结果
    print("\n" + "=" * 70)
    print("压测完成！")
    print("=" * 70)
    print(f"总耗时: {duration:.2f}秒")
    print(f"总发送消息数: {total_sent}")
    print(f"总接收消息数: {total_received}")
    print(f"消息吞吐量: {messages_per_second:.2f} 条/秒")
    print("-" * 70)
    print("延迟统计 (毫秒):")
    print(f"  最小: {min_latency:.2f}ms")
    print(f"  最大: {max_latency:.2f}ms")
    print(f"  平均: {avg_latency:.2f}ms")
    print(f"  P50:  {p50_latency:.2f}ms")
    print(f"  P95:  {p95_latency:.2f}ms")
    print(f"  P99:  {p99_latency:.2f}ms")
    print("-" * 70)
    print(f"连接错误数: {len(stats['errors'])}")
    
    if stats['errors']:
        print("\n错误详情:")
        for error in stats['errors'][:10]:  # 只显示前10个错误
            print(f"  - {error}")
    
    print("=" * 70)
    
    # 保存结果
    result = {
        'duration': duration,
        'total_sent': total_sent,
        'total_received': total_received,
        'messages_per_second': messages_per_second,
        'latency': {
            'min': min_latency,
            'max': max_latency,
            'avg': avg_latency,
            'p50': p50_latency,
            'p95': p95_latency,
            'p99': p99_latency
        },
        'errors': len(stats['errors'])
    }
    
    # 保存到JSON文件
    with open('/tmp/websocket_benchmark_results.json', 'w', encoding='utf-8') as f:
        json.dump(result, f, indent=2, ensure_ascii=False)
    
    print("\n结果已保存到: /tmp/websocket_benchmark_results.json")
    
    return result

def main():
    """主函数"""
    try:
        asyncio.run(run_benchmark())
    except KeyboardInterrupt:
        print("\n\n压测被用户中断")
        sys.exit(0)
    except Exception as e:
        print(f"\n压测失败: {e}")
        sys.exit(1)

if __name__ == "__main__":
    main()
