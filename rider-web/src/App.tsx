import React, { useEffect } from 'react';
import { createBrowserRouter, RouterProvider, Navigate } from 'react-router-dom';
import { ConfigProvider } from 'antd';
import zhCN from 'antd/locale/zh_CN';
import LoginPage from './pages/LoginPage';
import RegisterPage from './pages/RegisterPage';
import MainLayout from './pages/MainLayout';
import OrderPage from './pages/OrderPage';
import IncomePage from './pages/IncomePage';
import AIAgentPage from './pages/AIAgentPage';
import MessagePage from './pages/MessagePage';
import QualificationPage from './pages/QualificationPage';
import ProfilePage from './pages/ProfilePage';
import TicketPage from './pages/TicketPage';
import ReferralPage from './pages/ReferralPage';
import RatingPage from './pages/RatingPage';
import DeactivationPage from './pages/DeactivationPage';
import { useAuthStore } from './store/authStore';

const ProtectedRoute: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { isLoggedIn } = useAuthStore();
  return isLoggedIn ? <>{children}</> : <Navigate to="/login" replace />;
};

const App: React.FC = () => {
  const { loadStoredAuth } = useAuthStore();

  useEffect(() => {
    loadStoredAuth();
  }, []);

  const router = createBrowserRouter([
    {
      path: '/login',
      element: <LoginPage />,
    },
    {
      path: '/register',
      element: <RegisterPage />,
    },
    {
      path: '/',
      element: (
        <ProtectedRoute>
          <MainLayout />
        </ProtectedRoute>
      ),
      children: [
        {
          index: true,
          element: <OrderPage />,
        },
        {
          path: 'income',
          element: <IncomePage />,
        },
        {
          path: 'ai',
          element: <AIAgentPage />,
        },
        {
          path: 'messages',
          element: <MessagePage />,
        },
        {
          path: 'qualification',
          element: <QualificationPage />,
        },
        {
          path: 'profile',
          element: <ProfilePage />,
        },
        {
          path: 'tickets',
          element: <TicketPage />,
        },
        {
          path: 'referral',
          element: <ReferralPage />,
        },
        {
          path: 'rating',
          element: <RatingPage />,
        },
        {
          path: 'deactivation',
          element: <DeactivationPage />,
        },
      ],
    },
  ]);

  return (
    <ConfigProvider locale={zhCN} theme={{
      token: {
        colorPrimary: '#FFD700',
        borderRadius: 8,
      },
    }}>
      <RouterProvider router={router} />
    </ConfigProvider>
  );
};

export default App;
