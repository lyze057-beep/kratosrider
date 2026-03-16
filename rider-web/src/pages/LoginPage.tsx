import React, { useState } from 'react';
import { Form, Input, Button, Card, message, Tabs } from 'antd';
import { UserOutlined, LockOutlined, PhoneOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { authApi } from '../api/auth';
import { useAuthStore } from '../store/authStore';

const LoginPage: React.FC = () => {
  const navigate = useNavigate();
  const { setAuth } = useAuthStore();
  const [loading, setLoading] = useState(false);
  const [countdown, setCountdown] = useState(0);
  const [passwordForm] = Form.useForm();
  const [phoneForm] = Form.useForm();

  const handleSendCode = async () => {
    const phone = phoneForm.getFieldValue('phone');
    if (!phone || phone.length !== 11) {
      message.error('请输入正确的手机号');
      return;
    }

    try {
      const result = await authApi.sendCode(phone, 'login');
      message.success('验证码已发送，请查看后端日志获取验证码');
      console.log('验证码:', result.code);
      setCountdown(60);
      const timer = setInterval(() => {
        setCountdown((prev) => {
          if (prev <= 1) {
            clearInterval(timer);
            return 0;
          }
          return prev - 1;
        });
      }, 1000);
    } catch (error: any) {
      message.error(error.response?.data?.message || '发送验证码失败');
    }
  };

  const handlePasswordLogin = async (values: { phone: string; password: string }) => {
    setLoading(true);
    try {
      const result = await authApi.loginByPassword(values.phone, values.password);
      setAuth(result.token, result.refreshToken, result.userInfo);
      message.success('登录成功');
      navigate('/');
    } catch (error: any) {
      message.error(error.response?.data?.message || '登录失败');
    } finally {
      setLoading(false);
    }
  };

  const handlePhoneLogin = async (values: { phone: string; code: string }) => {
    setLoading(true);
    try {
      const result = await authApi.loginByPhone(values.phone, values.code);
      setAuth(result.token, result.refreshToken, result.userInfo);
      message.success('登录成功');
      navigate('/');
    } catch (error: any) {
      message.error(error.response?.data?.message || '登录失败');
    } finally {
      setLoading(false);
    }
  };

  const tabItems = [
    {
      key: 'password',
      label: '密码登录',
      children: (
        <Form form={passwordForm} onFinish={handlePasswordLogin} autoComplete="off">
          <Form.Item name="phone" label="手机号" rules={[{ required: true, message: '请输入手机号' }]}>
            <Input prefix={<PhoneOutlined />} placeholder="请输入手机号" maxLength={11} />
          </Form.Item>
          <Form.Item name="password" label="密码" rules={[{ required: true, message: '请输入密码' }]}>
            <Input.Password prefix={<LockOutlined />} placeholder="请输入密码" />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" loading={loading} block size="large">
              登录
            </Button>
          </Form.Item>
        </Form>
      )
    },
    {
      key: 'phone',
      label: '验证码登录',
      children: (
        <Form form={phoneForm} onFinish={handlePhoneLogin} autoComplete="off">
          <Form.Item name="phone" label="手机号" rules={[{ required: true, message: '请输入手机号' }]}>
            <Input prefix={<PhoneOutlined />} placeholder="请输入手机号" maxLength={11} />
          </Form.Item>
          <Form.Item name="code" label="验证码" rules={[{ required: true, message: '请输入验证码' }]}>
            <Input.Search
              prefix={<UserOutlined />}
              placeholder="请输入验证码"
              maxLength={6}
              enterButton={
                <Button type="primary" disabled={countdown > 0}>
                  {countdown > 0 ? `${countdown}s` : '获取验证码'}
                </Button>
              }
              onSearch={handleSendCode}
            />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" loading={loading} block size="large">
              登录
            </Button>
          </Form.Item>
        </Form>
      )
    }
  ];

  return (
    <div style={{ 
      display: 'flex', 
      justifyContent: 'center', 
      alignItems: 'center', 
      minHeight: '100vh',
      background: 'linear-gradient(135deg, #FFD700 0%, #FFA500 100%)'
    }}>
      <Card 
        title={<div style={{ textAlign: 'center', fontSize: '24px' }}>🚴 美团骑手</div>}
        style={{ width: 400, borderRadius: 12 }}
      >
        <Tabs items={tabItems} centered />
        <div style={{ textAlign: 'center', marginTop: 16 }}>
          还没有账号？
          <Button type="link" onClick={() => navigate('/register')}>
            立即注册
          </Button>
        </div>
      </Card>
    </div>
  );
};

export default LoginPage;
