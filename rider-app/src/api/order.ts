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

export interface GetOrderListRequest {
  status: number;
  page: number;
  page_size: number;
}

export interface GetOrderListReply {
  orders: OrderInfo[];
  total: number;
  page: number;
  page_size: number;
}

export interface CreateOrderRequest {
  origin: string;
  destination: string;
  origin_lat: number;
  origin_lng: number;
  dest_lat: number;
  dest_lng: number;
  amount: number;
}

export interface CreateOrderReply {
  success: boolean;
  message: string;
  order: OrderInfo;
}

export interface AcceptOrderRequest {
  order_id: number;
  rider_id: number;
}

export interface AcceptOrderReply {
  success: boolean;
  message: string;
}

export interface UpdateOrderStatusRequest {
  order_id: number;
  status: number;
}

export interface UpdateOrderStatusReply {
  success: boolean;
  message: string;
}

export const orderApi = {
  createOrder: async (data: CreateOrderRequest): Promise<CreateOrderReply> => {
    const response = await apiClient.post('/rider/v1/order/create', data);
    return response.data;
  },

  getOrderList: async (params: GetOrderListRequest): Promise<GetOrderListReply> => {
    const response = await apiClient.get('/rider/v1/orders', { params });
    return response.data;
  },

  getOrderDetail: async (orderId: number): Promise<{ order: OrderInfo }> => {
    const response = await apiClient.get(`/rider/v1/order/${orderId}`);
    return response.data;
  },

  acceptOrder: async (data: AcceptOrderRequest): Promise<AcceptOrderReply> => {
    const response = await apiClient.post('/rider/v1/order/accept', data);
    return response.data;
  },

  updateOrderStatus: async (data: UpdateOrderStatusRequest): Promise<UpdateOrderStatusReply> => {
    const response = await apiClient.post('/rider/v1/order/status', data);
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
