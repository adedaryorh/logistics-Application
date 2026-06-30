import { create } from 'zustand';

import { loginUser, logoutUser, refreshAccessToken, registerUser } from '@/services/api/authService';
import { tokenStorage } from '@/features/auth/utils/tokenStorage';

export type UserRole = 'customer' | 'driver' | 'merchant';

export type AuthUser = {
  id: string;
  email: string;
  name: string;
  role: UserRole;
  isVerified?: boolean;
};

type LoginCredentials = {
  email: string;
  password: string;
};

type RegisterPayload = {
  name: string;
  email: string;
  password: string;
  role: UserRole;
};

type AuthState = {
  user: AuthUser | null;
  accessToken: string | null;
  refreshToken: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  hydrate: () => Promise<void>;
  login: (credentials: LoginCredentials) => Promise<void>;
  register: (payload: RegisterPayload) => Promise<void>;
  logout: () => Promise<void>;
  refreshSession: () => Promise<boolean>;
  setUser: (user: AuthUser | null) => void;
};

const applySession = (user: AuthUser, accessToken: string, refreshToken: string | null) => {
  tokenStorage.setUser(JSON.stringify(user));
  tokenStorage.setAccessToken(accessToken);
  if (refreshToken) {
    tokenStorage.setRefreshToken(refreshToken);
  }
};

const clearSession = () => {
  tokenStorage.clearAll();
};

export const useAuthStore = create<AuthState>((set, get) => ({
  user: null,
  accessToken: null,
  refreshToken: null,
  isAuthenticated: false,
  isLoading: true,
  hydrate: async () => {
    try {
      const accessToken = tokenStorage.getAccessToken();
      const refreshToken = tokenStorage.getRefreshToken();
      const userJson = tokenStorage.getUser();
      const user = userJson ? (JSON.parse(userJson) as AuthUser) : null;

      set({
        user,
        accessToken,
        refreshToken,
        isAuthenticated: Boolean(accessToken && user),
        isLoading: false,
      });
    } catch {
      clearSession();
      set({
        user: null,
        accessToken: null,
        refreshToken: null,
        isAuthenticated: false,
        isLoading: false,
      });
    }
  },
  login: async (credentials) => {
    set({ isLoading: true });
    try {
      const response = await loginUser(credentials);
      applySession(response.user, response.access_token, response.refresh_token ?? null);
      set({
        user: response.user,
        accessToken: response.access_token,
        refreshToken: response.refresh_token ?? null,
        isAuthenticated: true,
        isLoading: false,
      });
    } catch (error) {
      set({ isLoading: false });
      throw error;
    }
  },
  register: async (payload) => {
    set({ isLoading: true });
    try {
      const response = await registerUser(payload);
      applySession(response.user, response.access_token, response.refresh_token ?? null);
      set({
        user: response.user,
        accessToken: response.access_token,
        refreshToken: response.refresh_token ?? null,
        isAuthenticated: true,
        isLoading: false,
      });
    } catch (error) {
      set({ isLoading: false });
      throw error;
    }
  },
  logout: async () => {
    try {
      await logoutUser();
    } catch {
      // local session should still be cleared
    }
    clearSession();
    set({
      user: null,
      accessToken: null,
      refreshToken: null,
      isAuthenticated: false,
      isLoading: false,
    });
  },
  refreshSession: async () => {
    const refreshToken = get().refreshToken ?? tokenStorage.getRefreshToken();
    if (!refreshToken) {
      return false;
    }
    try {
      const refreshed = await refreshAccessToken(refreshToken);
      tokenStorage.setAccessToken(refreshed.access_token);
      if (refreshed.refresh_token) {
        tokenStorage.setRefreshToken(refreshed.refresh_token);
      }
      set({
        accessToken: refreshed.access_token,
        refreshToken: refreshed.refresh_token ?? refreshToken,
        isAuthenticated: true,
      });
      return true;
    } catch {
      clearSession();
      set({
        user: null,
        accessToken: null,
        refreshToken: null,
        isAuthenticated: false,
      });
      return false;
    }
  },
  setUser: (user) => {
    if (user) {
      tokenStorage.setUser(JSON.stringify(user));
    } else {
      tokenStorage.removeUser();
    }
    set({ user, isAuthenticated: Boolean(user && get().accessToken) });
  },
}));

export default useAuthStore;
