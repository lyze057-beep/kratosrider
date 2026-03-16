import apiClient from './client';

export interface MessageInfo {
  id: number;
  sender_id: number;
  receiver_id: number;
  content: string;
  message_type: number;
  content_type: number;
  is_read: boolean;
  created_at: string;
}

export const messageApi = {
  // 获取消息列表
  getMessageList: async (params: {
    rider_id: number;
    page?: number;
    page_size?: number;
  }) => {
    const response = await apiClient.get('/rider/v1/messages', { params });
    return response.data;
  },

  // 发送消息
  sendMessage: async (data: {
    sender_id: number;
    receiver_id: number;
    content: string;
    message_type?: number;
  }) => {
    const response = await apiClient.post('/rider/v1/message/send', data);
    return response.data;
  },

  // 标记消息已读
  markAsRead: async (messageIds: number[]) => {
    const response = await apiClient.post('/rider/v1/message/read', {
      message_ids: messageIds
    });
    return response.data;
  },

  // 获取未读消息数
  getUnreadCount: async () => {
    const response = await apiClient.get('/rider/v1/messages/unread/count');
    return response.data;
  }
};
