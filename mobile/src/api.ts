import * as SecureStore from 'expo-secure-store';
import { CreateOrderInput, Order, Tracking, User } from './types';

const API_URL = process.env.EXPO_PUBLIC_API_URL ?? 'http://localhost:8080';
const TOKEN_KEY = 'logistics_access_token';

type Envelope<T> = {
  success: boolean;
  data: T;
  error?: { code: string; message: string };
};

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const token = await SecureStore.getItemAsync(TOKEN_KEY);
  const response = await fetch(`${API_URL}${path}`, {
    ...init,
    headers: {
      Accept: 'application/json',
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...init?.headers,
    },
  });
  const body = (await response.json()) as Envelope<T>;
  if (!response.ok || !body.success) {
    throw new Error(body.error?.message ?? 'Something went wrong. Please try again.');
  }
  return body.data;
}

export const api = {
  async login(email: string, password: string) {
    const data = await request<{
      user: User;
      tokens: { access_token: string; refresh_token: string };
    }>('/api/v1/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email, password, device_id: 'mobile' }),
    });
    await SecureStore.setItemAsync(TOKEN_KEY, data.tokens.access_token);
    return data;
  },
  register: (email: string, password: string) =>
    request('/api/v1/auth/register', { method: 'POST', body: JSON.stringify({ email, password }) }),
  me: async () => (await request<{ user: User }>('/api/v1/users/me')).user,
  orders: async () => (await request<{ orders: Order[]; next?: string }>('/api/v1/orders')).orders,
  order: async (id: string) => (await request<{ order: Order }>(`/api/v1/orders/${id}`)).order,
  // The backend tracking route is an SSE stream. Polling the order resource is
  // the reliable React Native fallback until a native EventSource is added.
  track: async (id: string): Promise<Tracking> => {
    const order = (await request<{ order: Order }>(`/api/v1/orders/${id}`)).order;
    return { order_id: order.id, status: order.status };
  },
  createOrder: async (payload: CreateOrderInput) =>
    (
      await request<{ order: Order }>('/api/v1/orders', {
        method: 'POST',
        body: JSON.stringify(payload),
      })
    ).order,
  async logout() {
    await SecureStore.deleteItemAsync(TOKEN_KEY);
  },
};
