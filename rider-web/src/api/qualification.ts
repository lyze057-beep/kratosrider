import apiClient from './client';

export interface QualificationInfo {
  id: number;
  rider_id: number;
  real_name: string;
  id_card_number: string;
  id_card_front: string;
  id_card_back: string;
  id_card_status: number;
  id_card_reject_reason: string;
  driver_license_number: string;
  driver_license_image: string;
  driver_license_status: number;
  driver_license_reject_reason: string;
  health_certificate_number: string;
  health_certificate_image: string;
  health_certificate_status: number;
  health_certificate_reject_reason: string;
  verification_status: number;
  submitted_at: string;
  verified_at: string;
}

export const qualificationApi = {
  // 获取资质认证状态
  getVerificationStatus: async (riderId: number) => {
    const response = await apiClient.get('/rider/v1/qualification/status', {
      params: { rider_id: riderId }
    });
    return response.data;
  },

  // 获取资质认证详情
  getVerificationDetail: async (riderId: number) => {
    const response = await apiClient.get('/rider/v1/qualification/detail', {
      params: { rider_id: riderId }
    });
    return response.data;
  },

  // 提交身份证认证
  submitIDCard: async (data: {
    rider_id: number;
    real_name: string;
    id_card_number: string;
    id_card_front: string;
    id_card_back: string;
  }) => {
    const response = await apiClient.post('/rider/v1/qualification/idcard', data);
    return response.data;
  },

  // 提交驾驶证认证
  submitDriverLicense: async (data: {
    rider_id: number;
    driver_license_number: string;
    driver_license_image: string;
  }) => {
    const response = await apiClient.post('/rider/v1/qualification/driver-license', data);
    return response.data;
  },

  // 提交健康证认证
  submitHealthCertificate: async (data: {
    rider_id: number;
    health_certificate_number: string;
    health_certificate_image: string;
  }) => {
    const response = await apiClient.post('/rider/v1/qualification/health', data);
    return response.data;
  },

  // 重新提交认证
  resubmitVerification: async (data: {
    rider_id: number;
    qualification_type: string;
  }) => {
    const response = await apiClient.post('/rider/v1/qualification/resubmit', data);
    return response.data;
  }
};
