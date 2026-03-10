import apiClient from './client';

export interface InviteCodeInfo {
  invite_code: string;
  invite_link: string;
  created_at: string;
  expire_at: string;
  total_invited: number;
  valid_invited: number;
  total_rewards: number;
}

export interface InvitedRiderInfo {
  rider_id: number;
  phone: string;
  nickname: string;
  status: number;
  register_at: string;
  first_order_at: string;
  task_completed_at: string;
  reward_amount: number;
}

export interface TaskInfo {
  task_id: number;
  task_name: string;
  task_desc: string;
  task_type: number;
  target_value: number;
  reward_amount: number;
  time_limit_days: number;
  start_time: string;
  end_time: string;
}

export interface TaskProgressInfo {
  task_id: number;
  task_name: string;
  current_value: number;
  target_value: number;
  progress_percent: number;
  status: number;
  deadline: string;
  is_claimed: boolean;
}

export interface RewardRecord {
  record_id: number;
  task_id: number;
  task_name: string;
  reward_amount: number;
  reward_type: number;
  status: number;
  created_at: string;
  issued_at: string;
}

export const referralApi = {
  generateInviteCode: async (riderId: number): Promise<{ invite_code_info: InviteCodeInfo }> => {
    const response = await apiClient.post('/rider/v1/referral/invite-code/generate', {
      rider_id: riderId,
    });
    return response.data;
  },

  getMyReferralInfo: async (riderId: number) => {
    const response = await apiClient.get('/rider/v1/referral/my-info', {
      params: { rider_id: riderId },
    });
    return response.data;
  },

  getInviteRecordList: async (riderId: number, page: number, pageSize: number, status?: number) => {
    const response = await apiClient.get('/rider/v1/referral/invite-records', {
      params: { rider_id: riderId, page, page_size: pageSize, status },
    });
    return response.data;
  },

  validateInviteCode: async (inviteCode: string) => {
    const response = await apiClient.post('/rider/v1/referral/invite-code/validate', {
      invite_code: inviteCode,
    });
    return response.data;
  },

  bindReferralRelation: async (newRiderId: number, inviteCode: string, phone: string) => {
    const response = await apiClient.post('/rider/v1/referral/bind', {
      new_rider_id: newRiderId,
      invite_code: inviteCode,
      phone,
    });
    return response.data;
  },

  getTaskList: async (riderId: number): Promise<{ tasks: TaskInfo[] }> => {
    const response = await apiClient.get('/rider/v1/referral/tasks', {
      params: { rider_id: riderId },
    });
    return response.data;
  },

  getTaskProgress: async (riderId: number, taskId: number): Promise<{ progress: TaskProgressInfo }> => {
    const response = await apiClient.get(`/rider/v1/referral/task-progress/${taskId}`, {
      params: { rider_id: riderId },
    });
    return response.data;
  },

  claimTaskReward: async (riderId: number, taskId: number) => {
    const response = await apiClient.post(`/rider/v1/referral/tasks/${taskId}/claim`, {
      rider_id: riderId,
    });
    return response.data;
  },

  getRewardRecordList: async (riderId: number, page: number, pageSize: number, status?: number) => {
    const response = await apiClient.get('/rider/v1/referral/reward-records', {
      params: { rider_id: riderId, page, page_size: pageSize, status },
    });
    return response.data;
  },
};

export const TASK_TYPE = {
  FIRST_ORDER: 1,
  ONLINE_TIME: 2,
  ORDER_COUNT: 3,
};

export const INVITED_RIDER_STATUS = {
  REGISTERED: 1,
  FIRST_ORDER: 2,
  TASK_COMPLETED: 3,
  REWARDED: 4,
};
