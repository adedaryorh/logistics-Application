import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import orderService from '@/services/orderService';
import { useAuthStore } from '@/features/auth/store/authStore';
import { toast } from '@/design-system/components/Toast';
import * as Haptics from '@/utils/haptics';

// Query keys
export const MERCHANT_QUERY_KEYS = {
  DASHBOARD: ['merchant', 'dashboard'],
  ORDERS: () => [...MERCHANT_QUERY_KEYS.ALL, 'orders'],
  ORDERS_STATUS: (status: string) => [...MERCHANT_QUERY_KEYS.ORDERS(), status],
  MENU: ['merchant', 'menu'],
  SETTINGS: ['merchant', 'settings'],
};

// Custom hook for getting merchant dashboard
export const useMerchantDashboard = () => {
  const { data, isLoading, error, refetch } = useQuery({
    queryKey: MERCHANT_QUERY_KEYS.DASHBOARD,
    queryFn: () => orderService.getMerchantDashboard(),
    refetchInterval: 60000, // Refetch every minute
  });

  return {
    dashboard: data,
    isLoading,
    error,
    refetch,
  };
};

// Custom hook for getting merchant orders
export const useMerchantOrders = (status?: string) => {
  const { data, isLoading, error, refetch } = useQuery({
    queryKey: MERCHANT_QUERY_KEYS.ORDERS_STATUS(status || 'all'),
    queryFn: () => orderService.getMerchantOrders(status),
    keepPreviousData: true,
  });

  return {
    orders: data || [],
    isLoading,
    error,
    refetch,
  };
};

// Custom hook for updating order status
export const useUpdateOrderStatus = () => {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: ({ orderId, status }: { orderId: string; status: string }) => 
      orderService.updateOrderStatus(orderId, status),
    onSuccess: (_, variables) => {
      // Invalidate and refetch queries
      queryClient.invalidateQueries({ queryKey: MERCHANT_QUERY_KEYS.ORDERS() });
      
      // Show success message
      Haptics.haptics.success();
      toast.success('Order status updated!');
    },
    onError: (error: any) => {
      Haptics.haptics.error();
      toast.error(error.message || 'Failed to update order status');
    },
  });
};

// Custom hook for getting merchant menu
export const useMerchantMenu = () => {
  const { data, isLoading, error, refetch } = useQuery({
    queryKey: MERCHANT_QUERY_KEYS.MENU,
    queryFn: () => orderService.getMerchantMenu(),
  });

  return {
    menu: data || [],
    isLoading,
    error,
    refetch,
  };
};

// Custom hook for updating merchant settings
export const useUpdateMerchantSettings = () => {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: (settings: any) => orderService.updateMerchantSettings(settings),
    onSuccess: () => {
      // Invalidate and refetch queries
      queryClient.invalidateQueries({ queryKey: MERCHANT_QUERY_KEYS.SETTINGS });
      
      // Show success message
      Haptics.haptics.success();
      toast.success('Settings updated!');
    },
    onError: (error: any) => {
      Haptics.haptics.error();
      toast.error(error.message || 'Failed to update settings');
    },
  });
};

export {
  useMerchantDashboard,
  useMerchantOrders,
  useUpdateOrderStatus,
  useMerchantMenu,
  useUpdateMerchantSettings,
};
