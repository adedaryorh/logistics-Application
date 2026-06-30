import apiClient from '@/services/api/client';
import { API_ENDPOINTS } from '@/services/api/types';
import type { Location, Order, OrderItem } from '@/types/order.types';

export interface CreateOrderInput {
  type: 'ride' | 'food' | 'parcel';
  pickup: Location;
  dropoff: Location;
  items?: OrderItem[];
  paymentMethod: 'cash' | 'card' | 'wallet';
  notes?: string;
}

export interface OrderFilters {
  status?: string;
  type?: string;
  page?: number;
  limit?: number;
}

export const orderService = {
  createOrder: async (data: CreateOrderInput): Promise<Order> => {
    const response = await apiClient.post(API_ENDPOINTS.ORDERS.CREATE, data);
    return response.data;
  },
  getOrders: async (filters: OrderFilters = {}): Promise<{ data: Order[]; total: number }> => {
    const response = await apiClient.get(API_ENDPOINTS.ORDERS.LIST, { params: filters });
    return response.data;
  },
  getOrderById: async (orderId: string): Promise<Order> => {
    const response = await apiClient.get(API_ENDPOINTS.ORDERS.GET(orderId));
    return response.data;
  },
  cancelOrder: async (orderId: string, reason: string): Promise<void> => {
    await apiClient.post(API_ENDPOINTS.ORDERS.CANCEL(orderId), { reason });
  },
  reorder: async (orderId: string): Promise<Order> => {
    const response = await apiClient.post(API_ENDPOINTS.ORDERS.REORDER(orderId));
    return response.data;
  },
  startTracking: async (orderId: string): Promise<void> => {
    await apiClient.post(API_ENDPOINTS.ORDERS.TRACK(orderId), { action: 'start' });
  },
  stopTracking: async (orderId: string): Promise<void> => {
    await apiClient.post(API_ENDPOINTS.ORDERS.TRACK(orderId), { action: 'stop' });
  },
  getDriverStatus: async (): Promise<{ status: string; isOnline: boolean }> => {
    const response = await apiClient.get(API_ENDPOINTS.DRIVER.STATUS);
    return response.data;
  },
  getDriverOffers: async () => {
    const response = await apiClient.get(API_ENDPOINTS.DRIVER.OFFERS);
    return response.data;
  },
  getDriverEarnings: async () => {
    const response = await apiClient.get(API_ENDPOINTS.DRIVER.EARNINGS);
    return response.data;
  },
  getDriverHistory: async (page = 1) => {
    const response = await apiClient.get(API_ENDPOINTS.DRIVER.HISTORY, { params: { page } });
    return response.data;
  },
  getMerchantDashboard: async () => {
    const response = await apiClient.get(API_ENDPOINTS.MERCHANT.DASHBOARD);
    return response.data;
  },
  getMerchantOrders: async (status?: string) => {
    const response = await apiClient.get(API_ENDPOINTS.MERCHANT.ORDERS, { params: { status } });
    return response.data;
  },
  getMerchantMenu: async () => {
    const response = await apiClient.get(API_ENDPOINTS.MERCHANT.MENU);
    return response.data;
  },
  updateMerchantSettings: async (settings: Record<string, unknown>): Promise<void> => {
    await apiClient.patch(API_ENDPOINTS.MERCHANT.SETTINGS, settings);
  },
};

export default orderService;
