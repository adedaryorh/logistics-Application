import { MMKV } from 'react-native-mmkv';

const mmkv = new MMKV({
  id: 'zustand-storage',
  encryptionKey: 'dev-key-32-chars-long!!',
});

export const MMKVStorage = {
  setItem: (key: string, value: string) => {
    mmkv.set(key, value);
  },
  getItem: (key: string): string | null => {
    return mmkv.getString(key) ?? null;
  },
  removeItem: (key: string) => {
    mmkv.delete(key);
  },
};
