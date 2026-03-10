import React, { useState, useEffect } from 'react';
import { Table, Button, Card, Tabs, Statistic, Modal, Form, Input, message, Select, Space } from 'antd';
import { DollarOutlined, ArrowUpOutlined, ArrowDownOutlined } from '@ant-design/icons';
import { incomeApi, IncomeInfo, WithdrawalInfo } from '../api/income';
import { useAuthStore } from '../store/authStore';

const IncomePage: React.FC = () => {
  const { userInfo } = useAuthStore();
  const [totalIncome, setTotalIncome] = useState(0);
  const [incomes, setIncomes] = useState<IncomeInfo[]>([]);
  const [withdrawals, setWithdrawals] = useState<WithdrawalInfo[]>([]);
  const [loading, setLoading] = useState(false);
  const [activeTab, setActiveTab] = useState('income');
  const [withdrawModalVisible, setWithdrawModalVisible] = useState(false);
  const [withdrawForm] = Form.useForm();

  const loadData = async () => {
    setLoading(true);
    try {
      const riderId = Number(userInfo?.user_id || 0);
      const [totalRes, incomeRes, withdrawalRes] = await Promise.all([
        incomeApi.getTotalIncome(riderId),
        incomeApi.getIncomeList(riderId, 20),
        incomeApi.getWithdrawalList(riderId, 20)
      ]);
      
      setTotalIncome(totalRes.total);
      setIncomes(incomeRes.incomes || []);
      setWithdrawals(withdrawalRes.withdrawals || []);
    } catch (error) {
      message.error('加载数据失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, []);

  const handleWithdraw = async (values: any) => {
    try {
      const riderId = Number(userInfo?.user_id || 0);
      const result = await incomeApi.applyWithdrawal(
        riderId,
        parseFloat(values.amount),
        values.platform,
        values.account
      );
      
      if (result.success) {
        message.success('提现申请已提交');
        setWithdrawModalVisible(false);
        withdrawForm.resetFields();
        loadData();
      } else {
        message.warning(result.message);
      }
    } catch (error: any) {
      message.error(error.response?.data?.message || '提现失败');
    }
  };

  const incomeColumns = [
    {
      title: '类型',
      dataIndex: 'income_type',
      key: 'income_type',
      width: 120,
      render: (type: number) => type === 0 ? '订单收入' : '其他收入',
    },
    {
      title: '订单ID',
      dataIndex: 'order_id',
      key: 'order_id',
      width: 120,
    },
    {
      title: '金额',
      dataIndex: 'amount',
      key: 'amount',
      width: 120,
      render: (amount: number) => (
        <span style={{ color: '#52c41a', fontWeight: 'bold' }}>
          <ArrowUpOutlined /> ¥{amount.toFixed(2)}
        </span>
      ),
    },
    {
      title: '状态',
      dataIndex: 'income_status',
      key: 'income_status',
      width: 120,
      render: (status: number) => {
        const statusMap = ['待结算', '已结算', '已提现'];
        return statusMap[status] || '未知';
      },
    },
    {
      title: '时间',
      dataIndex: 'created_at',
      key: 'created_at',
    },
  ];

  const withdrawalColumns = [
    {
      title: '提现方式',
      dataIndex: 'platform',
      key: 'platform',
      width: 120,
      render: (platform: string) => {
        const map: Record<string, string> = { alipay: '支付宝', wechat: '微信', bank: '银行卡' };
        return map[platform] || platform;
      },
    },
    {
      title: '账号',
      dataIndex: 'account',
      key: 'account',
    },
    {
      title: '金额',
      dataIndex: 'amount',
      key: 'amount',
      width: 120,
      render: (amount: number) => (
        <span style={{ color: '#ff4d4f', fontWeight: 'bold' }}>
          <ArrowDownOutlined /> ¥{amount.toFixed(2)}
        </span>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 120,
      render: (status: number) => {
        const statusMap = ['待处理', '处理中', '成功', '失败'];
        const colorMap = ['orange', 'blue', 'green', 'red'];
        return <span style={{ color: colorMap[status] }}>{statusMap[status]}</span>;
      },
    },
    {
      title: '时间',
      dataIndex: 'created_at',
      key: 'created_at',
    },
  ];

  const tabItems = [
    { key: 'income', label: '收入明细' },
    { key: 'withdrawal', label: '提现记录' },
  ];

  return (
    <div style={{ padding: 24 }}>
      <Card style={{ marginBottom: 24 }}>
        <Space size="large">
          <Statistic
            title="总收入"
            value={totalIncome}
            precision={2}
            valueStyle={{ color: '#3f8600' }}
            prefix={<DollarOutlined />}
          />
          <Button 
            type="primary" 
            size="large"
            onClick={() => setWithdrawModalVisible(true)}
            disabled={totalIncome <= 0}
          >
            申请提现
          </Button>
        </Space>
      </Card>

      <Card>
        <Tabs 
          activeKey={activeTab} 
          onChange={setActiveTab} 
          items={tabItems}
          style={{ marginBottom: 16 }}
        />
        <Table
          columns={activeTab === 'income' ? incomeColumns : withdrawalColumns}
          dataSource={activeTab === 'income' ? incomes : withdrawals}
          rowKey="id"
          loading={loading}
          pagination={{ pageSize: 10 }}
        />
      </Card>

      <Modal
        title="申请提现"
        open={withdrawModalVisible}
        onCancel={() => setWithdrawModalVisible(false)}
        footer={null}
      >
        <Form form={withdrawForm} onFinish={handleWithdraw} layout="vertical">
          <Form.Item
            name="amount"
            label="提现金额"
            rules={[{ required: true, message: '请输入提现金额' }]}
          >
            <Input placeholder={`最多可提现 ¥${totalIncome.toFixed(2)}`} />
          </Form.Item>
          <Form.Item
            name="platform"
            label="提现方式"
            rules={[{ required: true, message: '请选择提现方式' }]}
            initialValue="alipay"
          >
            <Select>
              <Select.Option value="alipay">支付宝</Select.Option>
              <Select.Option value="wechat">微信</Select.Option>
              <Select.Option value="bank">银行卡</Select.Option>
            </Select>
          </Form.Item>
          <Form.Item
            name="account"
            label="提现账号"
            rules={[{ required: true, message: '请输入提现账号' }]}
          >
            <Input placeholder="请输入提现账号" />
          </Form.Item>
          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit">确认提现</Button>
              <Button onClick={() => setWithdrawModalVisible(false)}>取消</Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default IncomePage;
