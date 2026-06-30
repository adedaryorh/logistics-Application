import axios, { AxiosError, AxiosHeaders, InternalAxiosRequestConfig } from 'axios';

import { refreshAccessToken } from '@/services/api/authService';
import { tokenStorage } from '@/features/auth/utils/tokenStorage';

const apiClient = axios.create({
  baseURL: process.env.REACT_NATIVE_API_URL || 'http://localhost:8080',
  timeout: 15000,
  headers: {
    Accept: 'application/json',
    'Content-Type': 'application/json',
  },
});

apiClient.interceptors.request.use((config: InternalAxiosRequestConfig) => {
  const token = tokenStorage.getAccessToken();
  if (token) {
    config.headers = config.headers ?? new AxiosHeaders();
    config.headers.set('Authorization', `Bearer ${token}`);
  }
  config.headers = config.headers ?? new AxiosHeaders();
  config.headers.set('X-Request-ID', Math.random().toString(36).slice(2, 12));
  return config;
});

apiClient.interceptors.response.use(
  (response) => response,
  async (error: AxiosError) => {
    const originalRequest = error.config as (InternalAxiosRequestConfig & { _retry?: boolean }) | undefined;

    if (error.response?.status === 401 && originalRequest && !originalRequest._retry) {
      originalRequest._retry = true;
      const refreshToken = tokenStorage.getRefreshToken();
      if (!refreshToken) {
        tokenStorage.clearAll();
        return Promise.reject(error);
      }

      try {
        const refreshed = await refreshAccessToken(refreshToken);
        tokenStorage.setAccessToken(refreshed.access_token);
        if (refreshed.refresh_token) {
          tokenStorage.setRefreshToken(refreshed.refresh_token);
        }
        originalRequest.headers = originalRequest.headers ?? new AxiosHeaders();
        originalRequest.headers.set('Authorization', `Bearer ${refreshed.access_token}`);
        return apiClient(originalRequest);
      } catch (refreshError) {
        tokenStorage.clearAll();
        return Promise.reject(refreshError);
      }
    }

    return Promise.reject(error);
  },
);

export const setBaseURL = (url: string) => {
  apiClient.defaults.baseURL = url;
};

export default apiClient;
