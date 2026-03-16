import React, { useState, useEffect } from 'react';
import { Card, Table, Button, Tag, Modal, Form, Input, Select, message, Tabs, Badge } from 'antd';
import { PlusOutlined, EyeOutlined, MessageOutlined } from '@ant-design/icons';
import { ticketApi, TicketInfo, TICKET_STATUS, TICKET_STATUS_TEXT, TICKET_TYPE_TEXT } from '../api/ticket';
import { useAuthStore } from '../store/authStore';

const { TextArea } = Input;
const { TabPane } = Tabs;

const TicketPage: React.FC = () => {
  useAuthStore();
  const [tickets, setTickets] = useState<TicketInfo[]>([]);
  const [loading, setLoading] = useState(false);
  const [createModalVisible, setCreateModalVisible] = useState(false);
  const [detailModalVisible, setDetailModalVisible] = useState(false);
  const [selectedTicket, setSelectedTicket] = useState<TicketInfo | null>(null);
  const [activeTab, setActiveTab] = useState('all');
  const [form] = Form.useForm();
  const [replyForm] = Form.useForm();
  const [stats, setStats] = useState({
    total: 0,
    pending: 0,
    processing: 0,
    resolved: 0
  });

  useEffect(() => {
    loadTickets();
    loadStats();
  }, [activeTab]);

  const loadTickets = async () => {
    setLoading(true);
    try {
      const params: any = { page: 1, page_size: 20 };
      if (activeTab !== 'all') {
        params.status = activeTab;
      }
      const result = await ticketApi.getTicketList(params);
      setTickets(result.tickets || []);
    } catch (error) {
      message.error('加载工单失败');
    } finally {
      setLoading(false);
    }
  };

  const loadStats = async () => {
    try {
      const result = await ticketApi.getTicketStatistics();
      setStats(result.statistics || { total: 0, pending: 0, processing: 0, resolved: 0 });
    } catch (error) {
      console.error('Load stats error:', error);
    }
  };

  const handleCreateTicket = async (values: any) => {
    try {
      await ticketApi.createTicket(values);
      message.success('工单创建成功');
      setCreateModalVisible(false);
      form.resetFields();
      loadTickets();
      loadStats();
    } catch (error) {
      message.error('创建失败');
    }
  };

  const handleViewDetail = (ticket: TicketInfo) => {
    setSelectedTicket(ticket);
    setDetailModalVisible(true);
  };

  const handleReply = async (values: any) => {
    if (!selectedTicket) return;
    try {
      await ticketApi.addTicketReply(selectedTicket.id, values.content);
      message.success('回复成功');
      replyForm.resetFields();
      loadTickets();
    } catch (error) {
      message.error('回复失败');
    }
  };

  const getStatusColor = (status: string) => {
    const colorMap: Record<string, string> = {
      [TICKET_STATUS.PENDING]: 'orange',
      [TICKET_STATUS.PROCESSING]: 'blue',
      [TICKET_STATUS.RESOLVED]: 'green',
      [TICKET_STATUS.CLOSED]: 'default'
    };
    return colorMap[status] || 'default';
  };

  const columns = [
    {
      title: '工单编号',
      dataIndex: 'id',
      key: 'id',
      width: 100
    },
    {
      title: '类型',
      dataIndex: 'ticket_type',
      key: 'ticket_type',
      width: 120,
      render: (type: string) => TICKET_TYPE_TEXT[type] || type
    },
    {
      title: '标题',
      dataIndex: 'title',
      key: 'title',
      ellipsis: true
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: string) => (
        <Tag color={getStatusColor(status)}>
          {TICKET_STATUS_TEXT[status] || status}
        </Tag>
      )
    },
    {
      title: '回复数',
      dataIndex: 'reply_count',
      key: 'reply_count',
      width: 80,
      render: (count: number) => count > 0 ? <Badge count={count} /> : '-'
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 180
    },
    {
      title: '操作',
      key: 'action',
      width: 120,
      render: (_: any, record: TicketInfo) => (
        <Button type="link" icon={<EyeOutlined />} onClick={() => handleViewDetail(record)}>
          查看
        </Button>
      )
    }
  ];

  return (
    <div style={{ padding: 24 }}>
      <Card style={{ marginBottom: 24 }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <div>
            <h2>工单中心</h2>
            <p style={{ color: '#666', margin: 0 }}>
              总工单: {stats.total} | 待处理: {stats.pending} | 处理中: {stats.processing} | 已解决: {stats.resolved}
            </p>
          </div>
          <Button type="primary" icon={<PlusOutlined />} onClick={() => setCreateModalVisible(true)}>
            创建工单
          </Button>
        </div>
      </Card>

      <Card>
        <Tabs activeKey={activeTab} onChange={setActiveTab}>
          <TabPane tab="全部" key="all" />
          <TabPane tab="待处理" key={TICKET_STATUS.PENDING} />
          <TabPane tab="处理中" key={TICKET_STATUS.PROCESSING} />
          <TabPane tab="已解决" key={TICKET_STATUS.RESOLVED} />
          <TabPane tab="已关闭" key={TICKET_STATUS.CLOSED} />
        </Tabs>
        <Table
          columns={columns}
          dataSource={tickets}
          rowKey="id"
          loading={loading}
          pagination={{ pageSize: 10 }}
        />
      </Card>

      {/* 创建工单弹窗 */}
      <Modal
        title="创建工单"
        open={createModalVisible}
        onCancel={() => setCreateModalVisible(false)}
        footer={null}
      >
        <Form form={form} onFinish={handleCreateTicket} layout="vertical">
          <Form.Item
            name="ticket_type"
            label="工单类型"
            rules={[{ required: true, message: '请选择工单类型' }]}
          >
            <Select placeholder="请选择工单类型">
              <Select.Option value="order_issue">订单问题</Select.Option>
              <Select.Option value="income_issue">收入问题</Select.Option>
              <Select.Option value="account_issue">账号问题</Select.Option>
              <Select.Option value="appeal">申诉</Select.Option>
              <Select.Option value="suggestion">建议反馈</Select.Option>
              <Select.Option value="other">其他</Select.Option>
            </Select>
          </Form.Item>
          <Form.Item
            name="title"
            label="标题"
            rules={[{ required: true, message: '请输入标题' }]}
          >
            <Input placeholder="请输入工单标题" />
          </Form.Item>
          <Form.Item
            name="description"
            label="详细描述"
            rules={[{ required: true, message: '请输入详细描述' }]}
          >
            <TextArea rows={4} placeholder="请详细描述您的问题" />
          </Form.Item>
          <Form.Item name="order_id" label="关联订单号">
            <Input placeholder="如有关联订单，请输入订单号" />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" block>
              提交工单
            </Button>
          </Form.Item>
        </Form>
      </Modal>

      {/* 工单详情弹窗 */}
      <Modal
        title="工单详情"
        open={detailModalVisible}
        onCancel={() => setDetailModalVisible(false)}
        footer={null}
        width={700}
      >
        {selectedTicket && (
          <div>
            <div style={{ marginBottom: 16, padding: 16, background: '#f5f5f5', borderRadius: 8 }}>
              <p><strong>工单编号:</strong> #{selectedTicket.id}</p>
              <p><strong>类型:</strong> {TICKET_TYPE_TEXT[selectedTicket.ticket_type]}</p>
              <p><strong>标题:</strong> {selectedTicket.title}</p>
              <p><strong>状态:</strong> <Tag color={getStatusColor(selectedTicket.status)}>
                {TICKET_STATUS_TEXT[selectedTicket.status]}
              </Tag></p>
              <p><strong>描述:</strong> {selectedTicket.description}</p>
            </div>

            {selectedTicket.status !== TICKET_STATUS.CLOSED && (
              <Form form={replyForm} onFinish={handleReply}>
                <Form.Item
                  name="content"
                  rules={[{ required: true, message: '请输入回复内容' }]}
                >
                  <TextArea rows={3} placeholder="输入回复内容..." />
                </Form.Item>
                <Form.Item>
                  <Button type="primary" htmlType="submit" icon={<MessageOutlined />}>
                    回复
                  </Button>
                </Form.Item>
              </Form>
            )}
          </div>
        )}
      </Modal>
    </div>
  );
};

export default TicketPage;
