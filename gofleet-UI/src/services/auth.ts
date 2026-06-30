type MockUser = {
  id: string;
  email: string;
  name: string;
  role: 'customer' | 'driver' | 'merchant';
};

export const authService = {
  login: async (credentials: { email: string }): Promise<MockUser> =>
    new Promise((resolve) => {
      setTimeout(() => {
        resolve({
          id: '1',
          email: credentials.email,
          name: 'John Doe',
          role: 'customer',
        });
      }, 100);
    }),

  logout: async (): Promise<void> =>
    new Promise((resolve) => {
      setTimeout(() => {
        resolve();
      }, 50);
    }),

  getCurrentUser: async (): Promise<MockUser | null> =>
    new Promise((resolve) => {
      setTimeout(() => {
        resolve(null);
      }, 50);
    }),
};

export default authService;
