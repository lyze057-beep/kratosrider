#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
智能抢单推送压测 - Locust (WebSocket + gRPC混合场景)
模拟1000名骑手同时抢单，包含LBS定位和订单推送
"""

import random
import time
import json
from locust import HttpUser, task, between, events
from locust.env import Environment
from locust.stats import stats_history, stats_printer
import threading
import queue

# 配置参数
NUM_RIDERS = 1000  # 骑手数量
TOTAL_REQUESTS = 100000  # 总请求数
DURATION_SECONDS = 300  # 持续时间（秒）

# 生成随机骑手数据
def generate_rider_data():
    """生成随机骑手ID、经纬度和订单ID"""
    rider_id = random.randint(10000, 99999)
    # 北京周边随机经纬度（约1公里范围内）
    base_lat = 39.9042
    base_lng = 116.4074
    lat = base_lat + random.uniform(-0.01, 0.01)
    lng = base_lng + random.uniform(-0.01, 0.01)
    order_id = random.randint(20000, 99999)
    return rider_id, lat, lng, order_id

# 订单状态枚举
ORDER_STATUS = {
    "PENDING": 0,    # 待接单
    "ACCEPTED": 1,   # 已接单
    "DELIVERING": 2, # 配送中
    "COMPLETED": 3,  # 已完成
    "CANCELLED": 4   # 已取消
}

class RiderClient:
    """模拟骑手客户端"""
    
    def __init__(self, host):
        self.host = host
        self.rider_id = None
        self.token = None
        self.orders = []
    
    def subscribe_orders(self, lat, lng):
        """订阅订单推送（WebSocket模拟）"""
        # 模拟WebSocket连接
        return {
            "action": "subscribe",
            "rider_id": self.rider_id,
            "lat": lat,
            "lng": lng,
            "timestamp": time.time()
        }
    
    def accept_order(self, order_id):
        """接单（gRPC模拟）"""
        return {
            "action": "accept",
            "order_id": order_id,
            "rider_id": self.rider_id,
            "timestamp": time.time()
        }
    
    def update_location(self, lat, lng):
        """更新LBS定位"""
        return {
            "action": "update_location",
            "rider_id": self.rider_id,
            "lat": lat,
            "lng": lng,
            "timestamp": time.time()
        }

class OrderPushUser(HttpUser):
    """订单推送压测用户"""
    
    # 模拟骑手客户端
    client_instance = None
    
    # 请求队列
    request_queue = queue.Queue()
    
    @task(3)
    def subscribe_and_accept_orders(self):
        """订阅订单并尝试接单（高频任务）"""
        rider_id, lat, lng, order_id = generate_rider_data()
        
        # 模拟订阅订单推送
        subscribe_payload = {
            "rider_id": rider_id,
            "lat": lat,
            "lng": lng
        }
        
        # 发送订阅请求
        self.client.post(
            "/rider/v1/order/subscribe",
            json=subscribe_payload,
            name="subscribe_orders"
        )
        
        # 模拟接单
        accept_payload = {
            "order_id": order_id,
            "rider_id": rider_id
        }
        
        self.client.post(
            "/rider/v1/order/accept",
            json=accept_payload,
            name="accept_order"
        )
        
        # 模拟LBS定位更新
        location_payload = {
            "rider_id": rider_id,
            "lat": lat,
            "lng": lng
        }
        
        self.client.post(
            "/rider/v1/order/location",
            json=location_payload,
            name="update_location"
        )
    
    @task(1)
    def update_order_status(self):
        """更新订单状态（低频任务）"""
        rider_id, lat, lng, order_id = generate_rider_data()
        
        status = random.choice([1, 2, 3])  # 已接单、配送中、已完成
        
        status_payload = {
            "order_id": order_id,
            "status": status
        }
        
        self.client.post(
            "/rider/v1/order/status",
            json=status_payload,
            name="update_order_status"
        )
    
    def on_start(self):
        """用户启动时初始化"""
        self.client_instance = RiderClient(self.host)
        self.rider_id, self.lat, self.lng, _ = generate_rider_data()
        self.client_instance.rider_id = self.rider_id

def run_benchmark():
    """运行压测"""
    
    print("=" * 60)
    print("智能抢单推送压测 - Locust")
    print("=" * 60)
    print(f"目标服务: http://localhost:8000")
    print(f"并发用户数: {NUM_RIDERS}")
    print(f"总请求数: {TOTAL_REQUESTS}")
    print(f"持续时间: {DURATION_SECONDS}秒")
    print("=" * 60)
    
    # 创建环境
    env = Environment(user_classes=[OrderPushUser])
    
    # 配置用户数
    env.runner.user_count = NUM_RIDERS
    env.runner.start(user_count=NUM_RIDERS, spawn_rate=100)  # 每秒启动100个用户
    
    # 启动统计报告
    stats_history_thread = threading.Thread(
        target=stats_history,
        args=(env,),
        daemon=True
    )
    stats_history_thread.start()
    
    # 启动统计打印
    stats_printer_thread = threading.Thread(
        target=stats_printer,
        args=(env.stats,),
        daemon=True
    )
    stats_printer_thread.start()
    
    # 运行指定时间
    print(f"\n开始压测，预计持续 {DURATION_SECONDS} 秒...")
    time.sleep(DURATION_SECONDS)
    
    # 停止压测
    env.runner.quit()
    
    # 打印最终统计
    print("\n" + "=" * 60)
    print("压测完成！")
    print("=" * 60)
    print(f"总请求数: {env.stats.total.num_requests}")
    print(f"失败请求数: {env.stats.total.num_failures}")
    print(f"请求/秒: {env.stats.total.current_rps:.2f}")
    print(f"平均响应时间: {env.stats.total.avg_response_time:.2f}ms")
    print(f"95%响应时间: {env.stats.total.get_response_time_percentile(0.95):.2f}ms")
    print(f"99%响应时间: {env.stats.total.get_response_time_percentile(0.99):.2f}ms")
    print("=" * 60)
    
    # 保存结果
    env.stats.save_csv("/tmp/locust_benchmark_results.csv")
    print("\n结果已保存到: /tmp/locust_benchmark_results.csv")

if __name__ == "__main__":
    run_benchmark()
