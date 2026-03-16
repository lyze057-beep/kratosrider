import React from 'react';
import { Layout, Menu, Avatar, Dropdown } from 'antd';
import { Outlet, useNavigate, useLocation } from 'react-router-dom';
import { 
  FileTextOutlined, 
  DollarOutlined, 
  RobotOutlined, 
  UserOutlined,
  LogoutOutlined,
  SettingOutlined,
  MessageOutlined,
  SafetyCertificateOutlined,
  CustomerServiceOutlined,
  ShareAltOutlined,
  StarOutlined,
  DeleteOutlined
} from '@ant-design/icons';
import { useAuthStore } from '../store/authStore';
import { authApi } from '../api/auth';

const { Header, Sider } = Layout;

const MainLayout: React.FC = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const { userInfo, logout } = useAuthStore();

  const handleLogout = async () => {
    try {
      await authApi.logout(userInfo?.userId || '');
      logout();
      navigate('/login');
    } catch (error) {
      logout();
      navigate('/login');
    }
  };

  const menuItems = [
    {
      key: '/',
      icon: <FileTextOutlined />,
      label: '订单管理',
    },
    {
      key: '/income',
      icon: <DollarOutlined />,
      label: '收入管理',
    },
    {
      key: '/ai',
      icon: <RobotOutlined />,
      label: 'AI助手',
    },
    {
      key: '/messages',
      icon: <MessageOutlined />,
      label: '消息中心',
    },
    {
      key: '/qualification',
      icon: <SafetyCertificateOutlined />,
      label: '资质认证',
    },
    {
      key: '/profile',
      icon: <UserOutlined />,
      label: '个人中心',
    },
    {
      key: '/tickets',
      icon: <CustomerServiceOutlined />,
      label: '工单中心',
    },
    {
      key: '/referral',
      icon: <ShareAltOutlined />,
      label: '拉新推广',
    },
    {
      key: '/rating',
      icon: <StarOutlined />,
      label: '我的评价',
    },
    {
      key: '/deactivation',
      icon: <DeleteOutlined />,
      label: '账号注销',
    },
  ];

  const userMenuItems = [
    {
      key: 'profile',
      icon: <UserOutlined />,
      label: '个人资料',
      onClick: () => navigate('/profile'),
    },
    {
      key: 'settings',
      icon: <SettingOutlined />,
      label: '设置',
      onClick: () => navigate('/profile'),
    },
    {
      type: 'divider' as const,
    },
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: '退出登录',
      onClick: handleLogout,
    },
  ];

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider theme="light" width={200}>
        <div style={{ 
          height: 64, 
          display: 'flex', 
          alignItems: 'center', 
          justifyContent: 'center',
          background: 'linear-gradient(135deg, #FFD700 0%, #FFA500 100%)'
        }}>
          <span style={{ fontSize: 18, fontWeight: 'bold', color: '#333' }}>
            🚴 美团骑手
          </span>
        </div>
        <Menu
          mode="inline"
          selectedKeys={[location.pathname]}
          items={menuItems}
          onClick={({ key }) => navigate(key)}
        />
      </Sider>
      <Layout>
        <Header style={{ 
          background: '#fff', 
          padding: '0 24px',
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          boxShadow: '0 2px 4px rgba(0,0,0,0.05)'
        }}>
          <div />
          <Dropdown menu={{ items: userMenuItems }} placement="bottomRight">
            <div style={{ display: 'flex', alignItems: 'center', cursor: 'pointer' }}>
              <Avatar 
                style={{ backgroundColor: '#FFD700', marginRight: 12 }}
                icon={<UserOutlined />}
              >
                {userInfo?.nickname?.charAt(0)}
              </Avatar>
              <span>{userInfo?.nickname}</span>
            </div>
          </Dropdown>
        </Header>
        <Outlet />
      </Layout>
    </Layout>
  );
};

export default MainLayout;
