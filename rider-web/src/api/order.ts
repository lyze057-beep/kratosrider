import apiClient from './client';

export interface OrderInfo {
  id: number;
  order_no: string;
  status: number;
  origin: string;
  destination: string;
  origin_lat: number;
  origin_lng: number;
  dest_lat: number;
  dest_lng: number;
  distance: number;
  amount: number;
  rider_id: number;
  created_at: string;
  updated_at: string;
}

export const orderApi = {
  getOrderList: async (status: number, page: number, pageSize: number) => {
    const response = await apiClient.get('/rider/v1/orders', {
      params: { status, page, page_size: pageSize }
    });
    return response.data;
  },

  getOrderDetail: async (orderId: number) => {
    const response = await apiClient.get(`/rider/v1/order/${orderId}`);
    return response.data;
  },

  acceptOrder: async (orderId: number, riderId: number) => {
    const response = await apiClient.post('/rider/v1/order/accept', {
      order_id: orderId,
      rider_id: riderId
    });
    return response.data;
  },

  updateOrderStatus: async (orderId: number, status: number) => {
    const response = await apiClient.post('/rider/v1/order/status', {
      order_id: orderId,
      status
    });
    return response.data;
  },
};

export const ORDER_STATUS = {
  PENDING: 0,
  ACCEPTED: 1,
  PICKED_UP: 2,
  DELIVERING: 3,
  COMPLETED: 4,
  CANCELLED: 5,
};

export const ORDER_STATUS_TEXT: Record<number, string> = {
  [ORDER_STATUS.PENDING]: '待接单',
  [ORDER_STATUS.ACCEPTED]: '已接单',
  [ORDER_STATUS.PICKED_UP]: '已取货',
  [ORDER_STATUS.DELIVERING]: '配送中',
  [ORDER_STATUS.COMPLETED]: '已完成',
  [ORDER_STATUS.CANCELLED]: '已取消',
};
