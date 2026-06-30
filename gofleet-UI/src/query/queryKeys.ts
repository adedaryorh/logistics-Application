export const queryKeys = {
  auth: {
    me: ['auth', 'me'],
  },
  orders: {
    list: (filters?: { status?: string; type?: string }) => ['orders', 'list', filters],
    detail: (orderId: string) => ['orders', 'detail', orderId],
    create: ['orders', 'create'],
    track: (orderId: string) => ['orders', 'track', orderId],
  },
  driver: {
    status: ['driver', 'status'],
    earnings: ['driver', 'earnings'],
    history: (page?: number) => ['driver', 'history', page],
    offers: ['driver', 'offers'],
  },
  merchant: {
    dashboard: ['merchant', 'dashboard'],
    orders: (status?: string) => ['merchant', 'orders', status],
    menu: ['merchant', 'menu'],
    settings: ['merchant', 'settings'],
  },
  payments: {
    list: ['payments', 'list'],
    detail: (paymentId: string) => ['payments', 'detail', paymentId],
    initialize: ['payments', 'initialize'],
  },
};
