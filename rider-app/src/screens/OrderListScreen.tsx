import React, { useState, useEffect } from 'react';
import {
  View,
  Text,
  FlatList,
  TouchableOpacity,
  StyleSheet,
  RefreshControl,
  Alert,
} from 'react-native';
import { orderApi, OrderInfo, ORDER_STATUS_TEXT } from '../api/order';
import { useAuthStore } from '../store/authStore';

export const OrderListScreen: React.FC = () => {
  const { userInfo } = useAuthStore();
  const [orders, setOrders] = useState<OrderInfo[]>([]);
  const [loading, setLoading] = useState(false);
  const [refreshing, setRefreshing] = useState(false);
  const [activeTab, setActiveTab] = useState(-1);

  const tabs = [
    { status: -1, label: '全部' },
    { status: 0, label: '待接单' },
    { status: 1, label: '已接单' },
    { status: 3, label: '配送中' },
    { status: 4, label: '已完成' },
  ];

  useEffect(() => {
    loadOrders();
  }, [activeTab]);

  const loadOrders = async () => {
    setLoading(true);
    try {
      const result = await orderApi.getOrderList({
        status: activeTab,
        page: 1,
        page_size: 20,
      });
      setOrders(result.orders || []);
    } catch (error) {
      console.error('Load orders error:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleRefresh = async () => {
    setRefreshing(true);
    await loadOrders();
    setRefreshing(false);
  };

  const handleAcceptOrder = async (orderId: number) => {
    try {
      const result = await orderApi.acceptOrder({
        order_id: orderId,
        rider_id: Number(userInfo?.user_id || 0),
      });
      if (result.success) {
        Alert.alert('成功', '接单成功');
        loadOrders();
      } else {
        Alert.alert('提示', result.message);
      }
    } catch (error: any) {
      Alert.alert('错误', error.response?.data?.message || '接单失败');
    }
  };

  const handleUpdateStatus = async (orderId: number, status: number) => {
    try {
      const result = await orderApi.updateOrderStatus({
        order_id: orderId,
        status,
      });
      if (result.success) {
        Alert.alert('成功', '状态更新成功');
        loadOrders();
      } else {
        Alert.alert('提示', result.message);
      }
    } catch (error: any) {
      Alert.alert('错误', error.response?.data?.message || '更新失败');
    }
  };

  const renderOrderItem = ({ item }: { item: OrderInfo }) => (
    <View style={styles.orderCard}>
      <View style={styles.orderHeader}>
        <Text style={styles.orderNo}>#{item.order_no}</Text>
        <Text style={[styles.orderStatus, getStatusStyle(item.status)]}>
          {ORDER_STATUS_TEXT[item.status] || '未知'}
        </Text>
      </View>

      <View style={styles.orderBody}>
        <View style={styles.locationRow}>
          <Text style={styles.locationIcon}>📍</Text>
          <Text style={styles.locationText} numberOfLines={1}>
            {item.origin}
          </Text>
        </View>
        <View style={styles.locationRow}>
          <Text style={styles.locationIcon}>🎯</Text>
          <Text style={styles.locationText} numberOfLines={1}>
            {item.destination}
          </Text>
        </View>
      </View>

      <View style={styles.orderFooter}>
        <View style={styles.orderInfo}>
          <Text style={styles.distance}>{item.distance.toFixed(1)}km</Text>
          <Text style={styles.amount}>¥{item.amount.toFixed(2)}</Text>
        </View>

        {renderActionButtons(item)}
      </View>
    </View>
  );

  const renderActionButtons = (order: OrderInfo) => {
    switch (order.status) {
      case 0:
        return (
          <TouchableOpacity
            style={styles.actionBtn}
            onPress={() => handleAcceptOrder(order.id)}
          >
            <Text style={styles.actionBtnText}>接单</Text>
          </TouchableOpacity>
        );
      case 1:
        return (
          <TouchableOpacity
            style={styles.actionBtn}
            onPress={() => handleUpdateStatus(order.id, 2)}
          >
            <Text style={styles.actionBtnText}>已取货</Text>
          </TouchableOpacity>
        );
      case 2:
        return (
          <TouchableOpacity
            style={styles.actionBtn}
            onPress={() => handleUpdateStatus(order.id, 3)}
          >
            <Text style={styles.actionBtnText}>开始配送</Text>
          </TouchableOpacity>
        );
      case 3:
        return (
          <TouchableOpacity
            style={styles.actionBtn}
            onPress={() => handleUpdateStatus(order.id, 4)}
          >
            <Text style={styles.actionBtnText}>完成配送</Text>
          </TouchableOpacity>
        );
      default:
        return null;
    }
  };

  const getStatusStyle = (status: number) => {
    switch (status) {
      case 0:
        return styles.statusPending;
      case 1:
      case 2:
        return styles.statusAccepted;
      case 3:
        return styles.statusDelivering;
      case 4:
        return styles.statusCompleted;
      default:
        return styles.statusCancelled;
    }
  };

  return (
    <View style={styles.container}>
      <View style={styles.tabs}>
        {tabs.map((tab) => (
          <TouchableOpacity
            key={tab.status}
            style={[styles.tab, activeTab === tab.status && styles.tabActive]}
            onPress={() => setActiveTab(tab.status)}
          >
            <Text
              style={[
                styles.tabText,
                activeTab === tab.status && styles.tabTextActive,
              ]}
            >
              {tab.label}
            </Text>
          </TouchableOpacity>
        ))}
      </View>

      <FlatList
        data={orders}
        keyExtractor={(item) => item.id.toString()}
        renderItem={renderOrderItem}
        contentContainerStyle={styles.listContent}
        refreshControl={
          <RefreshControl refreshing={refreshing} onRefresh={handleRefresh} />
        }
        ListEmptyComponent={
          <View style={styles.emptyContainer}>
            <Text style={styles.emptyText}>暂无订单</Text>
          </View>
        }
      />
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#f5f5f5',
  },
  tabs: {
    flexDirection: 'row',
    backgroundColor: '#fff',
    paddingHorizontal: 8,
    paddingVertical: 8,
    borderBottomWidth: 1,
    borderBottomColor: '#eee',
  },
  tab: {
    flex: 1,
    paddingVertical: 8,
    alignItems: 'center',
  },
  tabActive: {
    borderBottomWidth: 2,
    borderBottomColor: '#FFD700',
  },
  tabText: {
    fontSize: 14,
    color: '#666',
  },
  tabTextActive: {
    color: '#FFD700',
    fontWeight: 'bold',
  },
  listContent: {
    padding: 16,
  },
  orderCard: {
    backgroundColor: '#fff',
    borderRadius: 12,
    padding: 16,
    marginBottom: 12,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.05,
    shadowRadius: 4,
    elevation: 2,
  },
  orderHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 12,
  },
  orderNo: {
    fontSize: 14,
    color: '#666',
  },
  orderStatus: {
    fontSize: 14,
    fontWeight: 'bold',
    paddingHorizontal: 8,
    paddingVertical: 4,
    borderRadius: 4,
  },
  statusPending: {
    backgroundColor: '#FFF3E0',
    color: '#FF9800',
  },
  statusAccepted: {
    backgroundColor: '#E3F2FD',
    color: '#2196F3',
  },
  statusDelivering: {
    backgroundColor: '#F3E5F5',
    color: '#9C27B0',
  },
  statusCompleted: {
    backgroundColor: '#E8F5E9',
    color: '#4CAF50',
  },
  statusCancelled: {
    backgroundColor: '#FFEBEE',
    color: '#F44336',
  },
  orderBody: {
    marginBottom: 12,
  },
  locationRow: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 8,
  },
  locationIcon: {
    fontSize: 16,
    marginRight: 8,
  },
  locationText: {
    flex: 1,
    fontSize: 14,
    color: '#333',
  },
  orderFooter: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    borderTopWidth: 1,
    borderTopColor: '#eee',
    paddingTop: 12,
  },
  orderInfo: {
    flexDirection: 'row',
    alignItems: 'center',
  },
  distance: {
    fontSize: 14,
    color: '#666',
    marginRight: 16,
  },
  amount: {
    fontSize: 18,
    fontWeight: 'bold',
    color: '#FF6B00',
  },
  actionBtn: {
    backgroundColor: '#FFD700',
    paddingHorizontal: 20,
    paddingVertical: 8,
    borderRadius: 20,
  },
  actionBtnText: {
    fontSize: 14,
    fontWeight: 'bold',
    color: '#333',
  },
  emptyContainer: {
    alignItems: 'center',
    paddingVertical: 40,
  },
  emptyText: {
    fontSize: 16,
    color: '#999',
  },
});
