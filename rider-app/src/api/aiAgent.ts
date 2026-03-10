import apiClient from './client';

export interface AIResponse {
  content: string;
  response_type: number;
  actions: AIAction[];
  created_at: string;
}

export interface AIAction {
  action_type: string;
  action_name: string;
  action_params: string;
}

export interface AIAgentChatMessage {
  id: number;
  rider_id: number;
  content: string;
  message_type: number;
  content_type: number;
  created_at: string;
}

export interface AIAgentSendMessageRequest {
  rider_id: number;
  content: string;
  message_type: number;
}

export interface AIAgentSendMessageReply {
  success: boolean;
  message: string;
  ai_response: AIResponse;
}

export interface AIAgentGetChatHistoryRequest {
  rider_id: number;
  last_message_id?: number;
  limit: number;
}

export interface AIAgentGetChatHistoryReply {
  messages: AIAgentChatMessage[];
  has_more: boolean;
}

export interface AIAgentFAQItem {
  id: number;
  question: string;
  answer: string;
  category: string;
  view_count: number;
}

export interface AIAgentGetSuggestedQuestionsRequest {
  rider_id: number;
  context?: string;
}

export const aiAgentApi = {
  sendMessage: async (data: AIAgentSendMessageRequest): Promise<AIAgentSendMessageReply> => {
    const response = await apiClient.post('/rider/v1/ai-agent/send', data);
    return response.data;
  },

  getChatHistory: async (params: AIAgentGetChatHistoryRequest): Promise<AIAgentGetChatHistoryReply> => {
    const response = await apiClient.get('/rider/v1/ai-agent/history', { params });
    return response.data;
  },

  getFAQList: async (category?: string, limit?: number): Promise<{ faqs: AIAgentFAQItem[] }> => {
    const response = await apiClient.get('/rider/v1/ai-agent/faq', {
      params: { category, limit },
    });
    return response.data;
  },

  rateService: async (riderId: number, sessionId: number, rating: number, feedback?: string): Promise<{ success: boolean }> => {
    const response = await apiClient.post('/rider/v1/ai-agent/rate', {
      rider_id: riderId,
      session_id: sessionId,
      rating,
      feedback,
    });
    return response.data;
  },

  getSuggestedQuestions: async (params: AIAgentGetSuggestedQuestionsRequest): Promise<{ questions: string[] }> => {
    const response = await apiClient.get('/rider/v1/ai-agent/suggestions', { params });
    return response.data;
  },
};

export const MESSAGE_TYPE = {
  USER: 0,
  AI: 1,
};

export const CONTENT_TYPE = {
  TEXT: 0,
  IMAGE: 1,
  VOICE: 2,
};

export const RESPONSE_TYPE = {
  TEXT: 0,
  ORDER_QUERY: 1,
  INCOME_QUERY: 2,
  GUIDE: 3,
  TRANSFER_HUMAN: 4,
};
