import apiClient from '@/services/api/client';
import { API_ENDPOINTS } from '@/services/api/types';
import type { Payment, PaymentInitRequest, PaymentInitResponse, PaymentVerificationResponse } from '@/types/payment.types';

export const paymentService = {
  initializePayment: async (payload: PaymentInitRequest): Promise<PaymentInitResponse> => {
    const response = await apiClient.post(API_ENDPOINTS.PAYMENTS.CREATE, payload);
    return response.data;
  },
  confirmPayment: async (paymentId: string): Promise<Payment> => {
    const response = await apiClient.post(API_ENDPOINTS.PAYMENTS.CONFIRM(paymentId));
    return response.data;
  },
  getPayment: async (paymentId: string): Promise<PaymentVerificationResponse> => {
    const response = await apiClient.get(API_ENDPOINTS.PAYMENTS.GET(paymentId));
    return response.data;
  },
};

export default paymentService;
