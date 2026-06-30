export const API_ENDPOINTS = {
  AUTH: {
    LOGIN: '/api/v1/auth/login',
    REGISTER: '/api/v1/auth/register',
    ME: '/api/v1/auth/me',
    LOGOUT: '/api/v1/auth/logout',
    REFRESH: '/api/v1/auth/refresh',
    FORGOT_PASSWORD: '/api/v1/auth/forgot-password',
    VERIFY_RESET_CODE: '/api/v1/auth/verify-reset-code',
    RESET_PASSWORD: '/api/v1/auth/reset-password',
    VERIFY_EMAIL: '/api/v1/auth/verify-email',
    RESEND_VERIFICATION: '/api/v1/auth/resend-verification',
    GOOGLE_AUTH: '/api/v1/auth/google',
    APPLE_AUTH: '/api/v1/auth/apple',
  },
  ORDERS: {
    CREATE: '/api/v1/orders',
    LIST: '/api/v1/orders',
    GET: (id: string) => `/api/v1/orders/${id}`,
    CANCEL: (id: string) => `/api/v1/orders/${id}/cancel`,
    TRACK: (id: string) => `/api/v1/orders/${id}/track`,
    REORDER: (id: string) => `/api/v1/orders/${id}/reorder`,
  },
  DRIVER: {
    STATUS: '/api/v1/driver/status',
    LOCATION: '/api/v1/driver/location',
    OFFERS: '/api/v1/driver/offers',
    ACCEPT_OFFER: (id: string) => `/api/v1/driver/offers/${id}/accept`,
    REJECT_OFFER: (id: string) => `/api/v1/driver/offers/${id}/reject`,
    TRIP_START: (tripId: string) => `/api/v1/driver/trips/${tripId}/start`,
    TRIP_END: (tripId: string) => `/api/v1/driver/trips/${tripId}/end`,
    EARNINGS: '/api/v1/driver/earnings',
    HISTORY: '/api/v1/driver/history',
  },
  MERCHANT: {
    DASHBOARD: '/api/v1/merchant/dashboard',
    ORDERS: '/api/v1/merchant/orders',
    ORDER_UPDATE: (id: string) => `/api/v1/merchant/orders/${id}/status`,
    MENU: '/api/v1/merchant/menu',
    MENU_ITEM: (id: string) => `/api/v1/merchant/menu/${id}`,
    SETTINGS: '/api/v1/merchant/settings',
  },
  PAYMENTS: {
    CREATE: '/api/v1/payments',
    CONFIRM: (id: string) => `/api/v1/payments/${id}/confirm`,
    GET: (id: string) => `/api/v1/payments/${id}`,
  },
  LOCATION: {
    REVERSE_GEOCODE: '/api/v1/location/reverse-geocode',
    AUTOCOMPLETE: '/api/v1/location/autocomplete',
  },
} as const;

export type ApiError = {
  code: string;
  message: string;
  statusCode: number;
};
