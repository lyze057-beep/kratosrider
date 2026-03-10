import apiClient from './client';

export interface IncomeInfo {
  id: number;
  rider_id: number;
  order_id: number;
  amount: number;
  income_type: number;
  income_status: number;
  created_at: string;
}

export interface WithdrawalInfo {
  id: number;
  rider_id: number;
  amount: number;
  platform: string;
  account: string;
  status: number;
  transaction_id: string;
  created_at: string;
  updated_at: string;
}

export interface CreateIncomeRequest {
  order_id: number;
  income_type?: number;
  income_status?: number;
}

export interface CreateIncomeReply {
  success: boolean;
  message: string;
  income_id: number;
  amount: number;
  distance: number;
}

export interface GetIncomeListRequest {
  rider_id: number;
  limit: number;
}

export interface GetIncomeListReply {
  incomes: IncomeInfo[];
  total: number;
}

export interface ApplyWithdrawalRequest {
  rider_id: number;
  amount: number;
  platform: string;
  account: string;
}

export interface ApplyWithdrawalReply {
  success: boolean;
  message: string;
  withdrawal_id: number;
}

export interface GetWithdrawalListRequest {
  rider_id: number;
  limit: number;
}

export interface GetWithdrawalListReply {
  withdrawals: WithdrawalInfo[];
  total: number;
}

export const incomeApi = {
  createIncome: async (data: CreateIncomeRequest): Promise<CreateIncomeReply> => {
    const response = await apiClient.post('/rider/v1/income/create', data);
    return response.data;
  },

  getIncomeList: async (params: GetIncomeListRequest): Promise<GetIncomeListReply> => {
    const response = await apiClient.get('/rider/v1/incomes', { params });
    return response.data;
  },

  getTotalIncome: async (riderId: number): Promise<{ total: number }> => {
    const response = await apiClient.get('/rider/v1/income/total', {
      params: { rider_id: riderId },
    });
    return response.data;
  },

  applyWithdrawal: async (data: ApplyWithdrawalRequest): Promise<ApplyWithdrawalReply> => {
    const response = await apiClient.post('/rider/v1/withdrawal/apply', data);
    return response.data;
  },

  getWithdrawalList: async (params: GetWithdrawalListRequest): Promise<GetWithdrawalListReply> => {
    const response = await apiClient.get('/rider/v1/withdrawals', { params });
    return response.data;
  },
};

export const INCOME_TYPE = {
  ORDER: 0,
  OTHER: 1,
};

export const INCOME_STATUS = {
  PENDING: 0,
  SETTLED: 1,
  WITHDRAWN: 2,
};

export const WITHDRAWAL_STATUS = {
  PENDING: 0,
  PROCESSING: 1,
  SUCCESS: 2,
  FAILED: 3,
};

export const WITHDRAWAL_STATUS_TEXT: Record<number, string> = {
  [WITHDRAWAL_STATUS.PENDING]: '待处理',
  [WITHDRAWAL_STATUS.PROCESSING]: '处理中',
  [WITHDRAWAL_STATUS.SUCCESS]: '成功',
  [WITHDRAWAL_STATUS.FAILED]: '失败',
};
