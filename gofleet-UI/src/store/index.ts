import { create } from 'zustand';

import { useAuthStore } from '@/features/auth/store/authStore';

type AppStore = {
  hydratedAt: number | null;
  setHydratedAt: (timestamp: number) => void;
};

export const useStore = create<AppStore>((set) => ({
  hydratedAt: null,
  setHydratedAt: (timestamp) => set({ hydratedAt: timestamp }),
}));

export { useAuthStore };
export default useStore;
