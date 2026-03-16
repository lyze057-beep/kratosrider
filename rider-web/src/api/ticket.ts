import apiClient from './client';

export interface TicketInfo {
  id: number;
  user_id: number;
  ticket_type: string;
  title: string;
  description: string;
  status: string;
  order_id: number;
  attachments: string[];
  reply_count: number;
  last_reply_at: string;
  created_at: string;
  updated_at: string;
}

export interface TicketReplyInfo {
  id: number;
  ticket_id: number;
  user_id: number;
  content: string;
  is_staff: boolean;
  attachments: string[];
  created_at: string;
}

export const ticketApi = {
  // 创建工单
  createTicket: async (data: {
    ticket_type: string;
    title: string;
    description: string;
    order_id?: number;
    attachments?: string[];
  }) => {
    const response = await apiClient.post('/rider/v1/ticket/create', data);
    return response.data;
  },

  // 获取工单列表
  getTicketList: async (params: {
    status?: string;
    page?: number;
    page_size?: number;
  }) => {
    const response = await apiClient.get('/rider/v1/tickets', { params });
    return response.data;
  },

  // 获取工单详情
  getTicketDetail: async (ticketId: number) => {
    const response = await apiClient.get(`/rider/v1/ticket/${ticketId}`);
    return response.data;
  },

  // 添加工单回复
  addTicketReply: async (ticketId: number, content: string, attachments?: string[]) => {
    const response = await apiClient.post('/rider/v1/ticket/reply', {
      ticket_id: ticketId,
      content,
      attachments
    });
    return response.data;
  },

  // 更新工单状态
  updateTicketStatus: async (ticketId: number, status: string) => {
    const response = await apiClient.post('/rider/v1/ticket/status', {
      ticket_id: ticketId,
      status
    });
    return response.data;
  },

  // 获取工单统计
  getTicketStatistics: async () => {
    const response = await apiClient.get('/rider/v1/ticket/statistics');
    return response.data;
  }
};

export const TICKET_STATUS = {
  PENDING: 'pending',
  PROCESSING: 'processing',
  RESOLVED: 'resolved',
  CLOSED: 'closed'
};

export const TICKET_STATUS_TEXT: Record<string, string> = {
  [TICKET_STATUS.PENDING]: '待处理',
  [TICKET_STATUS.PROCESSING]: '处理中',
  [TICKET_STATUS.RESOLVED]: '已解决',
  [TICKET_STATUS.CLOSED]: '已关闭'
};

export const TICKET_TYPE_TEXT: Record<string, string> = {
  'order_issue': '订单问题',
  'income_issue': '收入问题',
  'account_issue': '账号问题',
  'appeal': '申诉',
  'suggestion': '建议反馈',
  'other': '其他'
};
