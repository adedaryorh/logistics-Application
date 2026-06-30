import apiClient from './client';
import { tokenStorage } from '@/features/auth/utils/tokenStorage';
import { API_ENDPOINTS } from './types';
import type { AuthResponse, User } from '@/types/auth.types';

export type LoginResponse = AuthResponse;
export type RegisterResponse = AuthResponse;

export const loginUser = async (credentials: { email: string; password: string }): Promise<LoginResponse> => {
  const response = await apiClient.post(API_ENDPOINTS.AUTH.LOGIN, credentials);
  const data = response.data as LoginResponse;
  tokenStorage.setAccessToken(data.access_token);
  tokenStorage.setRefreshToken(data.refresh_token);
  tokenStorage.setUser(JSON.stringify(data.user));
  return data;
};

export const registerUser = async (userData: {
  name: string;
  email: string;
  password: string;
  role: 'customer' | 'driver' | 'merchant';
}): Promise<RegisterResponse> => {
  const response = await apiClient.post(API_ENDPOINTS.AUTH.REGISTER, userData);
  const data = response.data as RegisterResponse;
  tokenStorage.setAccessToken(data.access_token);
  tokenStorage.setRefreshToken(data.refresh_token);
  tokenStorage.setUser(JSON.stringify(data.user));
  return data;
};

export const logoutUser = async (): Promise<void> => {
  try {
    await apiClient.post(API_ENDPOINTS.AUTH.LOGOUT);
  } finally {
    tokenStorage.clearAll();
  }
};

export const refreshAccessToken = async (refreshToken: string): Promise<{ access_token: string; refresh_token?: string }> => {
  const response = await apiClient.post(API_ENDPOINTS.AUTH.REFRESH, { refresh_token: refreshToken });
  const data = response.data as { access_token: string; refresh_token?: string };
  tokenStorage.setAccessToken(data.access_token);
  if (data.refresh_token) {
    tokenStorage.setRefreshToken(data.refresh_token);
  }
  return data;
};

export const forgotPassword = async (email: string) => apiClient.post(API_ENDPOINTS.AUTH.FORGOT_PASSWORD, { email });
export const resetPassword = async (token: string, password: string) => apiClient.post(API_ENDPOINTS.AUTH.RESET_PASSWORD, { token, password });
export const verifyEmail = async (token: string) => apiClient.post(API_ENDPOINTS.AUTH.VERIFY_EMAIL, { token });
export const resendVerification = async (email: string) => apiClient.post(API_ENDPOINTS.AUTH.RESEND_VERIFICATION, { email });
export const getCurrentUser = async (): Promise<User> => {
  const response = await apiClient.get(API_ENDPOINTS.AUTH.ME);
  return response.data as User;
};
export const updateUserProfile = async (data: Partial<{ name: string; phone: string; avatar: string }>) => apiClient.patch(API_ENDPOINTS.AUTH.ME, data);
export const changePassword = async (currentPassword: string, newPassword: string) => apiClient.post(API_ENDPOINTS.AUTH.RESET_PASSWORD, { current_password: currentPassword, new_password: newPassword });
export const getUserPreferences = async () => apiClient.get(API_ENDPOINTS.AUTH.ME);
export const updateUserPreferences = async (data: Record<string, unknown>) => apiClient.patch(API_ENDPOINTS.AUTH.ME, data);
