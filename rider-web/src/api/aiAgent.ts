import apiClient from './client';

export interface AIAgentChatMessage {
  id: number;
  rider_id: number;
  content: string;
  message_type: number;
  content_type: number;
  created_at: string;
}

export const aiAgentApi = {
  sendMessage: async (riderId: number, content: string) => {
    const response = await apiClient.post('/rider/v1/ai-agent/send', {
      rider_id: riderId,
      content,
      message_type: 0
    });
    return response.data;
  },

  getChatHistory: async (riderId: number, limit: number) => {
    const response = await apiClient.get('/rider/v1/ai-agent/history', {
      params: { rider_id: riderId, limit }
    });
    return response.data;
  },

  getFAQList: async (category?: string, limit?: number) => {
    const response = await apiClient.get('/rider/v1/ai-agent/faq', {
      params: { category, limit }
    });
    return response.data;
  },
};

export const MESSAGE_TYPE = {
  USER: 0,
  AI: 1,
};
