import React, { useState } from 'react';
import { Form, Input, Button, Card, message } from 'antd';
import { UserOutlined, LockOutlined, PhoneOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { authApi } from '../api/auth';
import { useAuthStore } from '../store/authStore';

const RegisterPage: React.FC = () => {
  const navigate = useNavigate();
  const { setAuth } = useAuthStore();
  const [loading, setLoading] = useState(false);
  const [countdown, setCountdown] = useState(0);
  const [form] = Form.useForm();

  const handleSendCode = async () => {
    const phone = form.getFieldValue('phone');
    console.log('获取验证码，手机号:', phone);

    if (!phone) {
      message.error('请先输入手机号');
      return;
    }

    if (phone.length !== 11) {
      message.error('请输入11位手机号');
      return;
    }

    try {
      const result = await authApi.sendCode(phone, 'register');
      message.success('验证码已发送！请查看浏览器控制台或后端日志');
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
      console.error('发送验证码失败:', error);
      message.error(error.response?.data?.message || '发送验证码失败');
    }
  };

  const handleRegister = async (values: any) => {
    console.log('注册表单数据:', values);

    if (values.password !== values.confirmPassword) {
      message.error('两次密码不一致');
      return;
    }

    if (values.password.length < 6) {
      message.error('密码至少6位');
      return;
    }

    setLoading(true);
    try {
      const result = await authApi.register({
        phone: values.phone,
        password: values.password,
        code: values.code,
        nickname: values.nickname
      });
      setAuth(result.token, result.refresh_token, result.user_info);
      message.success('注册成功');
      navigate('/');
    } catch (error: any) {
      console.error('注册失败:', error);
      message.error(error.response?.data?.message || '注册失败');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={{
      display: 'flex',
      justifyContent: 'center',
      alignItems: 'center',
      minHeight: '100vh',
      background: 'linear-gradient(135deg, #FFD700 0%, #FFA500 100%)'
    }}>
      <Card
        title={<div style={{ textAlign: 'center', fontSize: '24px' }}>🚴 注册账号</div>}
        style={{ width: 400, borderRadius: 12 }}
      >
        <Form form={form} onFinish={handleRegister} autoComplete="off">
          <Form.Item
            name="phone"
            rules={[
              { required: true, message: '请输入手机号' },
              { pattern: /^1[3-9]\d{9}$/, message: '请输入正确的手机号' }
            ]}
          >
            <Input prefix={<PhoneOutlined />} placeholder="请输入手机号" maxLength={11} />
          </Form.Item>

          <Form.Item
            name="code"
            rules={[{ required: true, message: '请输入验证码' }]}
          >
            <Input.Search
              prefix={<UserOutlined />}
              placeholder="请输入验证码"
              maxLength={6}
              enterButton={countdown > 0 ? `${countdown}s` : '获取验证码'}
              disabled={countdown > 0}
              onSearch={handleSendCode}
            />
          </Form.Item>

          <Form.Item name="nickname" rules={[{ required: true, message: '请输入昵称' }]}>
            <Input prefix={<UserOutlined />} placeholder="请输入昵称" />
          </Form.Item>

          <Form.Item
            name="password"
            rules={[
              { required: true, message: '请输入密码' },
              { min: 6, message: '密码至少6位' }
            ]}
          >
            <Input.Password prefix={<LockOutlined />} placeholder="请输入密码（至少6位）" />
          </Form.Item>

          <Form.Item name="confirmPassword" rules={[{ required: true, message: '请确认密码' }]}>
            <Input.Password prefix={<LockOutlined />} placeholder="请确认密码" />
          </Form.Item>

          <Form.Item>
            <Button type="primary" htmlType="submit" loading={loading} block size="large">
              注册
            </Button>
          </Form.Item>
        </Form>

        <div style={{ textAlign: 'center' }}>
          已有账号？
          <Button type="link" onClick={() => navigate('/login')}>
            立即登录
          </Button>
        </div>
      </Card>
    </div>
  );
};

export default RegisterPage;
