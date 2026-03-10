import apiClient from './client';

export interface UserInfo {
  user_id: string;
  phone: string;
  nickname: string;
  avatar: string;
  status: number;
}

export interface LoginReply {
  user_id: string;
  token: string;
  refresh_token: string;
  expires_in: number;
  user_info: UserInfo;
}

export const authApi = {
  sendCode: async (phone: string, type: string) => {
    const response = await apiClient.post('/rider/v1/code/send', { phone, type });
    return response.data;
  },

  register: async (data: { phone: string; password: string; code: string; nickname: string }) => {
    const response = await apiClient.post('/rider/v1/register', data);
    return response.data;
  },

  loginByPhone: async (phone: string, code: string) => {
    const response = await apiClient.post('/rider/v1/login/phone', { phone, code });
    return response.data;
  },

  loginByPassword: async (phone: string, password: string) => {
    const response = await apiClient.post('/rider/v1/login/password', { phone, password });
    return response.data;
  },

  logout: async (userId: string) => {
    const response = await apiClient.post('/rider/v1/logout', { user_id: userId });
    return response.data;
  },
};
