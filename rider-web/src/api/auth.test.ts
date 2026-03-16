import { describe, it, expect, vi, beforeEach } from 'vitest';
import { authApi } from './auth';

vi.mock('./client', () => ({
  default: {
    post: vi.fn(),
    get: vi.fn(),
    interceptors: {
      request: { use: vi.fn() },
      response: { use: vi.fn() },
    },
  },
}));

import apiClient from './client';
const mockedApiClient = apiClient as any;

describe('authApi', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('sendCode', () => {
    it('should call send code API with correct parameters', async () => {
      const mockResponse = {
        data: { code: 0, message: 'success' },
      };
      mockedApiClient.post.mockResolvedValueOnce(mockResponse);

      const result = await authApi.sendCode('13800138000', 'login');

      expect(mockedApiClient.post).toHaveBeenCalledWith('/rider/v1/code/send', {
        phone: '13800138000',
        type: 'login',
      });
      expect(result).toEqual(mockResponse.data);
    });
  });

  describe('register', () => {
    it('should call register API with correct parameters', async () => {
      const mockResponse = {
        data: { code: 0, message: 'success' },
      };
      mockedApiClient.post.mockResolvedValueOnce(mockResponse);

      const result = await authApi.register({
        phone: '13800138000',
        password: '123456',
        code: '123456',
        nickname: 'Test User',
      });

      expect(mockedApiClient.post).toHaveBeenCalledWith('/rider/v1/register', {
        phone: '13800138000',
        password: '123456',
        code: '123456',
        nickname: 'Test User',
      });
      expect(result).toEqual(mockResponse.data);
    });
  });

  describe('loginByPhone', () => {
    it('should call phone login API with correct parameters', async () => {
      const mockResponse = {
        data: { code: 0, message: 'success' },
      };
      mockedApiClient.post.mockResolvedValueOnce(mockResponse);

      const result = await authApi.loginByPhone('13800138000', '123456');

      expect(mockedApiClient.post).toHaveBeenCalledWith('/rider/v1/login/phone', {
        phone: '13800138000',
        code: '123456',
      });
      expect(result).toEqual(mockResponse.data);
    });
  });

  describe('loginByPassword', () => {
    it('should call password login API with correct parameters', async () => {
      const mockResponse = {
        data: { code: 0, message: 'success' },
      };
      mockedApiClient.post.mockResolvedValueOnce(mockResponse);

      const result = await authApi.loginByPassword('13800138000', '123456');

      expect(mockedApiClient.post).toHaveBeenCalledWith('/rider/v1/login/password', {
        phone: '13800138000',
        password: '123456',
      });
      expect(result).toEqual(mockResponse.data);
    });
  });

  describe('logout', () => {
    it('should call logout API with correct parameters', async () => {
      const mockResponse = {
        data: { code: 0, message: 'success' },
      };
      mockedApiClient.post.mockResolvedValueOnce(mockResponse);

      const result = await authApi.logout('1');

      expect(mockedApiClient.post).toHaveBeenCalledWith('/rider/v1/logout', {
        user_id: '1',
      });
      expect(result).toEqual(mockResponse.data);
    });
  });
});
