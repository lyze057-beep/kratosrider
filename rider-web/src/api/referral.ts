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

export interface InviteRecord {
  id: number;
  invited_user_id: number;
  invited_user_phone: string;
  invited_user_name: string;
  status: number;
  reward_amount: number;
  created_at: string;
}

export interface TaskInfo {
  id: number;
  task_name: string;
  task_description: string;
  task_type: string;
  target_count: number;
  reward_amount: number;
  status: number;
}

export interface TaskProgress {
  task_id: number;
  current_count: number;
  target_count: number;
  is_completed: boolean;
  is_claimed: boolean;
}

export interface RewardRecord {
  id: number;
  task_id: number;
  reward_type: string;
  reward_amount: number;
  status: number;
  created_at: string;
}

export const referralApi = {
  // 生成邀请码
  generateInviteCode: async () => {
    const response = await apiClient.post('/rider/v1/referral/invite-code/generate');
    return response.data;
  },

  // 获取我的邀请信息
  getMyReferralInfo: async () => {
    const response = await apiClient.get('/rider/v1/referral/my-info');
    return response.data;
  },

  // 获取邀请记录列表
  getInviteRecordList: async (page: number = 1, pageSize: number = 20) => {
    const response = await apiClient.get('/rider/v1/referral/invite-records', {
      params: { page, page_size: pageSize }
    });
    return response.data;
  },

  // 获取任务列表
  getTaskList: async () => {
    const response = await apiClient.get('/rider/v1/referral/tasks');
    return response.data;
  },

  // 获取任务进度
  getTaskProgress: async (taskId: number) => {
    const response = await apiClient.get(`/rider/v1/referral/task-progress/${taskId}`);
    return response.data;
  },

  // 领取任务奖励
  claimTaskReward: async (taskId: number) => {
    const response = await apiClient.post(`/rider/v1/referral/tasks/${taskId}/claim`);
    return response.data;
  },

  // 获取奖励记录
  getRewardRecordList: async (page: number = 1, pageSize: number = 20) => {
    const response = await apiClient.get('/rider/v1/referral/reward-records', {
      params: { page, page_size: pageSize }
    });
    return response.data;
  },

  // 获取拉新统计数据
  getReferralStatistics: async () => {
    const response = await apiClient.get('/rider/v1/referral/statistics');
    return response.data;
  }
};
