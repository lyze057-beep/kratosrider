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

export const incomeApi = {
  getIncomeList: async (riderId: number, limit: number) => {
    const response = await apiClient.get('/rider/v1/incomes', {
      params: { rider_id: riderId, limit }
    });
    return response.data;
  },

  getTotalIncome: async (riderId: number) => {
    const response = await apiClient.get('/rider/v1/income/total', {
      params: { rider_id: riderId }
    });
    return response.data;
  },

  applyWithdrawal: async (riderId: number, amount: number, platform: string, account: string) => {
    const response = await apiClient.post('/rider/v1/withdrawal/apply', {
      rider_id: riderId,
      amount,
      platform,
      account
    });
    return response.data;
  },

  getWithdrawalList: async (riderId: number, limit: number) => {
    const response = await apiClient.get('/rider/v1/withdrawals', {
      params: { rider_id: riderId, limit }
    });
    return response.data;
  },
};
