import React, { useState, useEffect } from 'react';
import { Card, Avatar, Form, Input, Button, message, Upload, Tabs, Statistic, Row, Col } from 'antd';
import { UserOutlined, CameraOutlined, SafetyOutlined, BellOutlined } from '@ant-design/icons';
import { useAuthStore } from '../store/authStore';

const { TabPane } = Tabs;

const ProfilePage: React.FC = () => {
  const { userInfo, setUserInfo } = useAuthStore();
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [stats] = useState({
    totalOrders: 0,
    totalIncome: 0,
    rating: 5.0,
    completionRate: 100
  });

  useEffect(() => {
    if (userInfo) {
      form.setFieldsValue({
        nickname: userInfo.nickname,
        phone: userInfo.phone,
      });
    }
    loadStats();
  }, [userInfo]);

  const loadStats = async () => {
    // 这里可以加载骑手的统计数据
    // const result = await riderApi.getStats(userInfo?.userId);
    // setStats(result);
  };

  const handleUpdateProfile = async (values: any) => {
    setLoading(true);
    try {
      // await authApi.updateProfile(userInfo?.userId, values);
      message.success('资料更新成功');
      // 更新本地存储的用户信息
      if (userInfo) {
        setUserInfo({ ...userInfo, ...values });
      }
    } catch (error) {
      message.error('更新失败');
    } finally {
      setLoading(false);
    }
  };

  const handleAvatarChange = (info: any) => {
    if (info.file.status === 'done') {
      message.success('头像上传成功');
      // 更新用户信息
    }
  };

  return (
    <div style={{ padding: 24 }}>
      <h2>个人中心</h2>
      
      <Row gutter={24} style={{ marginBottom: 24 }}>
        <Col span={6}>
          <Card>
            <Statistic title="累计接单" value={stats.totalOrders} suffix="单" />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic title="累计收入" value={stats.totalIncome} prefix="¥" />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic title="评分" value={stats.rating} suffix="分" />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic title="准时率" value={stats.completionRate} suffix="%" />
          </Card>
        </Col>
      </Row>

      <Card>
        <Tabs defaultActiveKey="1">
          <TabPane tab="基本信息" key="1">
            <div style={{ display: 'flex', alignItems: 'center', marginBottom: 24 }}>
              <Avatar 
                size={100} 
                icon={<UserOutlined />} 
                src={userInfo?.avatar}
                style={{ marginRight: 24 }}
              />
              <Upload
                showUploadList={false}
                onChange={handleAvatarChange}
              >
                <Button icon={<CameraOutlined />}>更换头像</Button>
              </Upload>
            </div>

            <Form
              form={form}
              layout="vertical"
              onFinish={handleUpdateProfile}
              style={{ maxWidth: 500 }}
            >
              <Form.Item
                name="nickname"
                label="昵称"
                rules={[{ required: true, message: '请输入昵称' }]}
              >
                <Input placeholder="请输入昵称" />
              </Form.Item>
              <Form.Item
                name="phone"
                label="手机号"
                rules={[{ required: true, message: '请输入手机号' }]}
              >
                <Input disabled />
              </Form.Item>
              <Form.Item>
                <Button type="primary" htmlType="submit" loading={loading}>
                  保存修改
                </Button>
              </Form.Item>
            </Form>
          </TabPane>

          <TabPane tab="安全设置" key="2">
            <Card title="修改密码" style={{ maxWidth: 500 }}>
              <Form layout="vertical">
                <Form.Item
                  name="oldPassword"
                  label="原密码"
                  rules={[{ required: true, message: '请输入原密码' }]}
                >
                  <Input.Password placeholder="请输入原密码" />
                </Form.Item>
                <Form.Item
                  name="newPassword"
                  label="新密码"
                  rules={[{ required: true, message: '请输入新密码' }]}
                >
                  <Input.Password placeholder="请输入新密码" />
                </Form.Item>
                <Form.Item
                  name="confirmPassword"
                  label="确认新密码"
                  rules={[
                    { required: true, message: '请确认新密码' },
                    ({ getFieldValue }) => ({
                      validator(_, value) {
                        if (!value || getFieldValue('newPassword') === value) {
                          return Promise.resolve();
                        }
                        return Promise.reject(new Error('两次输入的密码不一致'));
                      },
                    }),
                  ]}
                >
                  <Input.Password placeholder="请确认新密码" />
                </Form.Item>
                <Form.Item>
                  <Button type="primary" icon={<SafetyOutlined />}>
                    修改密码
                  </Button>
                </Form.Item>
              </Form>
            </Card>
          </TabPane>

          <TabPane tab="通知设置" key="3">
            <Card title="消息通知" style={{ maxWidth: 500 }}>
              <Form layout="vertical">
                <Form.Item>
                  <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16 }}>
                    <span>订单推送通知</span>
                    <input type="checkbox" defaultChecked />
                  </div>
                  <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16 }}>
                    <span>收入到账通知</span>
                    <input type="checkbox" defaultChecked />
                  </div>
                  <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16 }}>
                    <span>系统公告通知</span>
                    <input type="checkbox" defaultChecked />
                  </div>
                  <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16 }}>
                    <span>活动优惠通知</span>
                    <input type="checkbox" />
                  </div>
                </Form.Item>
                <Form.Item>
                  <Button type="primary" icon={<BellOutlined />}>
                    保存设置
                  </Button>
                </Form.Item>
              </Form>
            </Card>
          </TabPane>
        </Tabs>
      </Card>
    </div>
  );
};

export default ProfilePage;
