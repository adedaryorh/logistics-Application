import * as SecureStore from 'expo-secure-store';
import EventSource from 'react-native-sse';
import { CreateOrderInput, Order, Tracking, User, Wallet, WalletEntry } from './types';

const API_URL = process.env.EXPO_PUBLIC_API_URL ?? 'http://localhost:8080';
const TOKEN_KEY = 'logistics_access_token';
const REFRESH_KEY = 'logistics_refresh_token';

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
    await SecureStore.setItemAsync(REFRESH_KEY, data.tokens.refresh_token);
    return data;
  },
  register: async (email: string, password: string) =>
    (
      await request<{ user: User }>('/api/v1/auth/register', {
        method: 'POST',
        body: JSON.stringify({ email, password }),
      })
    ).user,
  requestPasswordReset: (email: string) =>
    request('/api/v1/auth/password/reset-request', {
      method: 'POST',
      body: JSON.stringify({ email }),
    }),
  hasSession: async () => Boolean(await SecureStore.getItemAsync(TOKEN_KEY)),
  me: async () => (await request<{ user: User }>('/api/v1/users/me')).user,
  orders: async () => (await request<{ orders: Order[]; next?: string }>('/api/v1/orders')).orders,
  order: async (id: string) => (await request<{ order: Order }>(`/api/v1/orders/${id}`)).order,
  // The backend tracking route is an SSE stream. Polling the order resource is
  // the reliable React Native fallback until a native EventSource is added.
  track: async (id: string): Promise<Tracking> => {
    const order = (await request<{ order: Order }>(`/api/v1/orders/${id}`)).order;
    return { order_id: order.id, status: order.status };
  },
  async subscribeToOrder(
    id: string,
    onUpdate: (tracking: Tracking) => void,
    onError: (message: string) => void,
  ) {
    const token = await SecureStore.getItemAsync(TOKEN_KEY);
    const source = new EventSource<'order.status'>(`${API_URL}/api/v1/orders/${id}/track`, {
      headers: token ? { Authorization: `Bearer ${token}` } : {},
    });
    source.addEventListener('order.status', (event) => {
      try {
        onUpdate(JSON.parse(event.data ?? '{}') as Tracking);
      } catch {
        onError('Received an invalid tracking update.');
      }
    });
    source.addEventListener('error', () => onError('Live updates disconnected. Retrying…'));
    return () => source.close();
  },
  wallet: () => request<{ wallet: Wallet; ledger: WalletEntry[] }>('/api/v1/payments/wallet/me'),
  createOrder: async (payload: CreateOrderInput) =>
    (
      await request<{ order: Order }>('/api/v1/orders', {
        method: 'POST',
        body: JSON.stringify(payload),
      })
    ).order,
  recordProof: async (id: string, kind: 'pickup' | 'delivery', payload: { evidence_url: string; notes?: string; recipient_name?: string; coordinate: { lat: number; lng: number } }) =>
    (await request<{ order: Order }>(`/api/v1/orders/${id}/proofs/${kind}`, { method: 'POST', body: JSON.stringify({ ...payload, captured_at: new Date().toISOString() }) })).order,
  async logout() {
    await SecureStore.deleteItemAsync(TOKEN_KEY);
    await SecureStore.deleteItemAsync(REFRESH_KEY);
  },
};
