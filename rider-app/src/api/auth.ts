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

export interface RegisterRequest {
  phone: string;
  password: string;
  code: string;
  nickname: string;
}

export interface SendCodeRequest {
  phone: string;
  type: string;
}

export interface SendCodeReply {
  request_id: string;
  expire_seconds: number;
}

export interface LoginByPhoneRequest {
  phone: string;
  code: string;
}

export interface LoginByPasswordRequest {
  phone: string;
  password: string;
}

export const authApi = {
  sendCode: async (data: SendCodeRequest): Promise<SendCodeReply> => {
    const response = await apiClient.post('/rider/v1/code/send', data);
    return response.data;
  },

  register: async (data: RegisterRequest): Promise<LoginReply> => {
    const response = await apiClient.post('/rider/v1/register', data);
    return response.data;
  },

  loginByPhone: async (data: LoginByPhoneRequest): Promise<LoginReply> => {
    const response = await apiClient.post('/rider/v1/login/phone', data);
    return response.data;
  },

  loginByPassword: async (data: LoginByPasswordRequest): Promise<LoginReply> => {
    const response = await apiClient.post('/rider/v1/login/password', data);
    return response.data;
  },

  loginByThirdParty: async (platform: string, code: string, state?: string): Promise<LoginReply> => {
    const response = await apiClient.post('/rider/v1/login/third', {
      platform,
      code,
      state,
    });
    return response.data;
  },

  logout: async (userId: string): Promise<{ success: boolean }> => {
    const response = await apiClient.post('/rider/v1/logout', {
      user_id: userId,
    });
    return response.data;
  },

  refreshToken: async (refreshToken: string): Promise<LoginReply> => {
    const response = await apiClient.post('/rider/v1/token/refresh', {
      refresh_token: refreshToken,
    });
    return response.data;
  },
};
