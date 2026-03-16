import apiClient from './client';

export interface RatingInfo {
  id: number;
  rider_id: number;
  source_type: number;
  source_id: number;
  score: number;
  content: string;
  tags: string[];
  is_anonymous: boolean;
  reply_content: string;
  reply_time: string;
  created_at: string;
}

export interface RatingSummary {
  average_score: number;
  total_count: number;
  five_star_count: number;
  four_star_count: number;
  three_star_count: number;
  two_star_count: number;
  one_star_count: number;
}

export const ratingApi = {
  // 获取骑手评分
  getRiderRating: async (riderId: number) => {
    const response = await apiClient.get(`/rider/v1/rating/${riderId}`);
    return response.data;
  },

  // 获取评分记录列表
  getRatingRecords: async (params: {
    rider_id?: number;
    source_type?: number;
    page?: number;
    page_size?: number;
  }) => {
    const response = await apiClient.get('/rider/v1/rating/records', { params });
    return response.data;
  },

  // 获取评分详情
  getRatingDetail: async (recordId: number) => {
    const response = await apiClient.get(`/rider/v1/rating/detail/${recordId}`);
    return response.data;
  },

  // 回复评价
  replyToRating: async (recordId: number, content: string) => {
    const response = await apiClient.post('/rider/v1/rating/reply', {
      record_id: recordId,
      content
    });
    return response.data;
  },

  // 获取评分汇总
  getRatingSummary: async (riderId: number) => {
    const response = await apiClient.get(`/rider/v1/rating/summary/${riderId}`);
    return response.data;
  }
};

export const RATING_SOURCE = {
  USER: 1,
  MERCHANT: 2,
  SYSTEM: 3,
  PLATFORM: 4
};

export const RATING_SOURCE_TEXT: Record<number, string> = {
  [RATING_SOURCE.USER]: '用户',
  [RATING_SOURCE.MERCHANT]: '商家',
  [RATING_SOURCE.SYSTEM]: '系统',
  [RATING_SOURCE.PLATFORM]: '平台'
};
