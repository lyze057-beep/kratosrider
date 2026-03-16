import { describe, it, expect, vi, beforeEach } from 'vitest';
import { orderApi } from './order';

vi.mock('./client', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
    interceptors: {
      request: { use: vi.fn() },
      response: { use: vi.fn() },
    },
  },
}));

import apiClient from './client';
const mockedApiClient = apiClient as any;

describe('orderApi', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('getOrderList', () => {
    it('should call get order list API with correct parameters', async () => {
      const mockResponse = {
        data: {
          code: 0,
          message: 'success',
          data: {
            orders: [],
            total: 0,
          },
        },
      };
      mockedApiClient.get.mockResolvedValueOnce(mockResponse);

      const result = await orderApi.getOrderList(-1, 1, 10, 1);

      expect(mockedApiClient.get).toHaveBeenCalledWith('/rider/v1/orders', {
        params: { status: -1, page: 1, page_size: 10, rider_id: 1 },
      });
      expect(result).toEqual(mockResponse.data);
    });
  });

  describe('getOrderDetail', () => {
    it('should call get order detail API', async () => {
      const mockResponse = {
        data: {
          code: 0,
          message: 'success',
          data: {
            id: 1,
            orderNo: '123456',
          },
        },
      };
      mockedApiClient.get.mockResolvedValueOnce(mockResponse);

      const result = await orderApi.getOrderDetail(1);

      expect(mockedApiClient.get).toHaveBeenCalledWith('/rider/v1/order/1');
      expect(result).toEqual(mockResponse.data);
    });
  });

  describe('acceptOrder', () => {
    it('should call accept order API', async () => {
      const mockResponse = { data: { code: 0, message: 'success' } };
      mockedApiClient.post.mockResolvedValueOnce(mockResponse);

      const result = await orderApi.acceptOrder(1, 1);

      expect(mockedApiClient.post).toHaveBeenCalledWith('/rider/v1/order/accept', {
        order_id: 1,
        rider_id: 1,
      });
      expect(result).toEqual(mockResponse.data);
    });
  });

  describe('updateOrderStatus', () => {
    it('should call update order status API', async () => {
      const mockResponse = { data: { code: 0, message: 'success' } };
      mockedApiClient.post.mockResolvedValueOnce(mockResponse);

      const result = await orderApi.updateOrderStatus(1, 2);

      expect(mockedApiClient.post).toHaveBeenCalledWith('/rider/v1/order/status', {
        order_id: 1,
        status: 2,
      });
      expect(result).toEqual(mockResponse.data);
    });
  });
});
