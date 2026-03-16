import { create } from 'zustand';
import { UserInfo } from '../api/auth';

interface AuthState {
  isLoggedIn: boolean;
  token: string | null;
  refreshToken: string | null;
  userInfo: UserInfo | null;
  setAuth: (token: string, refreshToken: string, userInfo: UserInfo) => void;
  setUserInfo: (userInfo: UserInfo) => void;
  logout: () => void;
  loadStoredAuth: () => void;
}

export const useAuthStore = create<AuthState>((set) => ({
  isLoggedIn: false,
  token: null,
  refreshToken: null,
  userInfo: null,

  setAuth: (token, refreshToken, userInfo) => {
    localStorage.setItem('token', token);
    localStorage.setItem('refreshToken', refreshToken);
    localStorage.setItem('userInfo', JSON.stringify(userInfo));
    set({
      isLoggedIn: true,
      token,
      refreshToken,
      userInfo,
    });
  },

  setUserInfo: (userInfo) => {
    localStorage.setItem('userInfo', JSON.stringify(userInfo));
    set({ userInfo });
  },

  logout: () => {
    localStorage.removeItem('token');
    localStorage.removeItem('refreshToken');
    localStorage.removeItem('userInfo');
    set({
      isLoggedIn: false,
      token: null,
      refreshToken: null,
      userInfo: null,
    });
  },

  loadStoredAuth: () => {
    try {
      const token = localStorage.getItem('token');
      const refreshToken = localStorage.getItem('refreshToken');
      const userInfoStr = localStorage.getItem('userInfo');
      
      if (token && refreshToken && userInfoStr) {
        const userInfo = JSON.parse(userInfoStr);
        set({
          isLoggedIn: true,
          token,
          refreshToken,
          userInfo,
        });
      }
    } catch (error) {
      console.error('Load stored auth error:', error);
    }
  },
}));
