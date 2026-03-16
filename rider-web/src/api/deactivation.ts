import apiClient from './client';

export interface DeactivationInfo {
  id: number;
  rider_id: number;
  reason_type: number;
  reason_detail: string;
  status: number;
  applied_at: string;
  reviewed_at: string;
  review_remark: string;
}

export const deactivationApi = {
  // 申请注销
  applyDeactivation: async (data: {
    reason_type: number;
    reason_detail?: string;
  }) => {
    const response = await apiClient.post('/rider/v1/deactivation/apply', data);
    return response.data;
  },

  // 获取注销状态
  getDeactivationStatus: async () => {
    const response = await apiClient.get('/rider/v1/deactivation/status');
    return response.data;
  },

  // 取消注销申请
  cancelDeactivation: async () => {
    const response = await apiClient.post('/rider/v1/deactivation/cancel');
    return response.data;
  },

  // 获取注销原因列表
  getDeactivationReasons: async () => {
    const response = await apiClient.get('/rider/v1/deactivation/reasons');
    return response.data;
  }
};

export const DEACTIVATION_REASON = {
  PERSONAL: 1,
  HEALTH: 2,
  OTHER_JOB: 3,
  INCOME: 4,
  TIME: 5,
  OTHER: 6
};

export const DEACTIVATION_REASON_TEXT: Record<number, string> = {
  [DEACTIVATION_REASON.PERSONAL]: '个人原因',
  [DEACTIVATION_REASON.HEALTH]: '健康原因',
  [DEACTIVATION_REASON.OTHER_JOB]: '找到其他工作',
  [DEACTIVATION_REASON.INCOME]: '收入不满意',
  [DEACTIVATION_REASON.TIME]: '时间不合适',
  [DEACTIVATION_REASON.OTHER]: '其他原因'
};

export const DEACTIVATION_STATUS = {
  PENDING: 1,
  APPROVED: 2,
  REJECTED: 3,
  COMPLETED: 4
};

export const DEACTIVATION_STATUS_TEXT: Record<number, string> = {
  [DEACTIVATION_STATUS.PENDING]: '待审核',
  [DEACTIVATION_STATUS.APPROVED]: '已批准',
  [DEACTIVATION_STATUS.REJECTED]: '已拒绝',
  [DEACTIVATION_STATUS.COMPLETED]: '已完成'
};
