import { MMKV } from 'react-native-mmkv';

const storage = new MMKV({
  id: 'auth-storage',
  encryptionKey: 'dev-key-32-chars-long!!',
});

export const tokenStorage = {
  setAccessToken: (token: string) => storage.set('access_token', token),
  getAccessToken: (): string | null => storage.getString('access_token') ?? null,
  removeAccessToken: () => storage.delete('access_token'),
  setRefreshToken: (token: string) => storage.set('refresh_token', token),
  getRefreshToken: (): string | null => storage.getString('refresh_token') ?? null,
  removeRefreshToken: () => storage.delete('refresh_token'),
  setUser: (user: string) => storage.set('user', user),
  getUser: (): string | null => storage.getString('user') ?? null,
  removeUser: () => storage.delete('user'),
  setRefreshTokenExpiresAt: (timestamp: number) => storage.set('refresh_token_expires_at', String(timestamp)),
  getRefreshTokenExpiresAt: (): number | null => {
    const value = storage.getString('refresh_token_expires_at');
    return value ? Number.parseInt(value, 10) : null;
  },
  clearAll: () => storage.clearAll(),
  isAuthenticated: (): boolean => Boolean(storage.getString('access_token')),
};

export default tokenStorage;
