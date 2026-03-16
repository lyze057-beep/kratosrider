import React, { useState, useEffect } from 'react';
import { Table, Button, Tag, message, Card, Tabs } from 'antd';
import { CheckOutlined, CarOutlined, CheckCircleOutlined } from '@ant-design/icons';
import { orderApi, OrderInfo, ORDER_STATUS, ORDER_STATUS_TEXT } from '../api/order';
import { useAuthStore } from '../store/authStore';

const OrderPage: React.FC = () => {
  const { userInfo } = useAuthStore();
  const [orders, setOrders] = useState<OrderInfo[]>([]);
  const [loading, setLoading] = useState(false);
  const [activeTab, setActiveTab] = useState('-1');

  const loadOrders = async (status: number) => {
    setLoading(true);
    try {
      const riderId = userInfo?.userId ? Number(userInfo.userId) : undefined;
      const result = await orderApi.getOrderList(status, 1, 20, riderId);
      setOrders(result.orders || []);
    } catch (error) {
      message.error('加载订单失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadOrders(parseInt(activeTab));
  }, [activeTab]);

  const handleAccept = async (orderId: number) => {
    try {
      const result = await orderApi.acceptOrder(orderId, Number(userInfo?.userId || 0));
      if (result.success) {
        message.success('接单成功');
        loadOrders(parseInt(activeTab));
      } else {
        message.warning(result.message);
      }
    } catch (error: any) {
      message.error(error.response?.data?.message || '接单失败');
    }
  };

  const handleUpdateStatus = async (orderId: number, status: number) => {
    try {
      const result = await orderApi.updateOrderStatus(orderId, status);
      if (result.success) {
        message.success('状态更新成功');
        loadOrders(parseInt(activeTab));
      } else {
        message.warning(result.message);
      }
    } catch (error: any) {
      message.error(error.response?.data?.message || '更新失败');
    }
  };

  const getStatusColor = (status: number) => {
    switch (status) {
      case ORDER_STATUS.PENDING:
        return 'orange';
      case ORDER_STATUS.ACCEPTED:
      case ORDER_STATUS.PICKED_UP:
        return 'blue';
      case ORDER_STATUS.DELIVERING:
        return 'purple';
      case ORDER_STATUS.COMPLETED:
        return 'green';
      default:
        return 'default';
    }
  };

  const getActionButtons = (order: OrderInfo) => {
    switch (order.status) {
      case ORDER_STATUS.PENDING:
        return (
          <Button type="primary" icon={<CheckOutlined />} onClick={() => handleAccept(order.id)}>
            接单
          </Button>
        );
      case ORDER_STATUS.ACCEPTED:
        return (
          <Button type="primary" icon={<CarOutlined />} onClick={() => handleUpdateStatus(order.id, ORDER_STATUS.PICKED_UP)}>
            已取货
          </Button>
        );
      case ORDER_STATUS.PICKED_UP:
        return (
          <Button type="primary" icon={<CarOutlined />} onClick={() => handleUpdateStatus(order.id, ORDER_STATUS.DELIVERING)}>
            开始配送
          </Button>
        );
      case ORDER_STATUS.DELIVERING:
        return (
          <Button type="primary" icon={<CheckCircleOutlined />} onClick={() => handleUpdateStatus(order.id, ORDER_STATUS.COMPLETED)}>
            完成配送
          </Button>
        );
      default:
        return null;
    }
  };

  const columns = [
    {
      title: '订单号',
      dataIndex: 'order_no',
      key: 'order_no',
      width: 150,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: number) => (
        <Tag color={getStatusColor(status)}>{ORDER_STATUS_TEXT[status]}</Tag>
      ),
    },
    {
      title: '起点',
      dataIndex: 'origin',
      key: 'origin',
      ellipsis: true,
    },
    {
      title: '终点',
      dataIndex: 'destination',
      key: 'destination',
      ellipsis: true,
    },
    {
      title: '距离',
      dataIndex: 'distance',
      key: 'distance',
      width: 100,
      render: (distance: number) => `${distance.toFixed(1)}km`,
    },
    {
      title: '金额',
      dataIndex: 'amount',
      key: 'amount',
      width: 100,
      render: (amount: number) => <span style={{ color: '#FF6B00', fontWeight: 'bold' }}>¥{amount.toFixed(2)}</span>,
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 180,
    },
    {
      title: '操作',
      key: 'action',
      width: 150,
      render: (_: any, record: OrderInfo) => getActionButtons(record),
    },
  ];

  const tabItems = [
    { key: '-1', label: '全部' },
    { key: '0', label: '待接单' },
    { key: '1', label: '已接单' },
    { key: '3', label: '配送中' },
    { key: '4', label: '已完成' },
  ];

  return (
    <div style={{ padding: 24 }}>
      <Card>
        <Tabs 
          activeKey={activeTab} 
          onChange={setActiveTab} 
          items={tabItems}
          style={{ marginBottom: 16 }}
        />
        <Table
          columns={columns}
          dataSource={orders}
          rowKey="id"
          loading={loading}
          pagination={{ pageSize: 10 }}
        />
      </Card>
    </div>
  );
};

export default OrderPage;
