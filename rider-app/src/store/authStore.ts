import { create } from 'zustand';
import AsyncStorage from '@react-native-async-storage/async-storage';
import { UserInfo } from '../api/auth';

interface AuthState {
  isLoggedIn: boolean;
  token: string | null;
  refreshToken: string | null;
  userInfo: UserInfo | null;
  setAuth: (token: string, refreshToken: string, userInfo: UserInfo) => Promise<void>;
  logout: () => Promise<void>;
  loadStoredAuth: () => Promise<void>;
}

export const useAuthStore = create<AuthState>((set) => ({
  isLoggedIn: false,
  token: null,
  refreshToken: null,
  userInfo: null,

  setAuth: async (token, refreshToken, userInfo) => {
    await AsyncStorage.setItem('token', token);
    await AsyncStorage.setItem('refreshToken', refreshToken);
    await AsyncStorage.setItem('userInfo', JSON.stringify(userInfo));
    set({
      isLoggedIn: true,
      token,
      refreshToken,
      userInfo,
    });
  },

  logout: async () => {
    await AsyncStorage.removeItem('token');
    await AsyncStorage.removeItem('refreshToken');
    await AsyncStorage.removeItem('userInfo');
    set({
      isLoggedIn: false,
      token: null,
      refreshToken: null,
      userInfo: null,
    });
  },

  loadStoredAuth: async () => {
    try {
      const token = await AsyncStorage.getItem('token');
      const refreshToken = await AsyncStorage.getItem('refreshToken');
      const userInfoStr = await AsyncStorage.getItem('userInfo');
      
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
