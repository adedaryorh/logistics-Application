import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import orderService from '@/services/orderService';
import { useAuthStore } from '@/features/auth/store/authStore';
import { toast } from '@/design-system/components/Toast';
import * as Haptics from '@/utils/haptics';
import { useCallback } from 'react';

// Query keys
export const ORDER_QUERY_KEYS = {
  ALL: ['orders'],
  LISTS: () => [...ORDER_QUERY_KEYS.ALL, 'list'],
  LIST: (filters: any) => [...ORDER_QUERY_KEYS.LISTS(), { filters }],
  DETAIL: () => [...ORDER_QUERY_KEYS.ALL, 'detail'],
  DETAIL_ID: (id: string) => [...ORDER_QUERY_KEYS.DETAIL(), id],
  TRACKING: (orderId: string) => ['order-tracking', orderId],
};

// Custom hook for getting orders
export const useOrders = (filters: any = {}) => {
  const { data, isLoading, error, refetch } = useQuery({
    queryKey: ORDER_QUERY_KEYS.LIST(filters),
    queryFn: () => orderService.getOrders(filters),
    keepPreviousData: true,
  });

  return {
    orders: data?.data || [],
    total: data?.total || 0,
    isLoading,
    error,
    refetch,
  };
};

// Custom hook for getting a single order
export const useOrder = (orderId: string) => {
  const { data, isLoading, error, refetch } = useQuery({
    queryKey: ORDER_QUERY_KEYS.DETAIL_ID(orderId),
    queryFn: () => orderService.getOrderById(orderId),
    enabled: !!orderId,
  });

  return {
    order: data,
    isLoading,
    error,
    refetch,
  };
};

// Custom hook for creating an order
export const useCreateOrder = () => {
  const queryClient = useQueryClient();
  const { user } = useAuthStore();
  
  return useMutation({
    mutationFn: (orderData: any) => orderService.createOrder(orderData),
    onSuccess: (data) => {
      // Invalidate and refetch queries
      queryClient.invalidateQueries({ queryKey: ORDER_QUERY_KEYS.ALL });
      
      // Show success message
      Haptics.haptics.success();
      toast.success('Order placed successfully!');
    },
    onError: (error: any) => {
      Haptics.haptics.error();
      toast.error(error.message || 'Failed to place order');
    },
  });
};

// Custom hook for canceling an order
export const useCancelOrder = () => {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: ({ orderId, reason }: { orderId: string; reason: string }) => 
      orderService.cancelOrder(orderId, reason),
    onSuccess: () => {
      // Invalidate and refetch queries
      queryClient.invalidateQueries({ queryKey: ORDER_QUERY_KEYS.ALL });
      
      // Show success message
      Haptics.haptics.success();
      toast.success('Order cancelled successfully!');
    },
    onError: (error: any) => {
      Haptics.haptics.error();
      toast.error(error.message || 'Failed to cancel order');
    },
  });
};

// Custom hook for reordering
export const useReorder = () => {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: (orderId: string) => orderService.reorder(orderId),
    onSuccess: (data) => {
      // Invalidate and refetch queries
      queryClient.invalidateQueries({ queryKey: ORDER_QUERY_KEYS.ALL });
      
      // Show success message
      Haptics.haptics.success();
      toast.success('Item added to cart!');
    },
    onError: (error: any) => {
      Haptics.haptics.error();
      toast.error(error.message || 'Failed to reorder');
    },
  });
};

// Custom hook for starting order tracking
export const useStartOrderTracking = () => {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: (orderId: string) => orderService.startTracking(orderId),
    onSuccess: () => {
      // Invalidate and refetch queries
      queryClient.invalidateQueries({ queryKey: ORDER_QUERY_KEYS.ALL });
      
      // Show success message
      Haptics.haptics.success();
      toast.success('Tracking started!');
    },
    onError: (error: any) => {
      Haptics.haptics.error();
      toast.error(error.message || 'Failed to start tracking');
    },
  });
};

// Custom hook for stopping order tracking
export const useStopOrderTracking = () => {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: (orderId: string) => orderService.stopTracking(orderId),
    onSuccess: () => {
      // Invalidate and refetch queries
      queryClient.invalidateQueries({ queryKey: ORDER_QUERY_KEYS.ALL });
      
      // Show success message
      Haptics.haptics.success();
      toast.success('Tracking stopped!');
    },
    onError: (error: any) => {
      Haptics.haptics.error();
      toast.error(error.message || 'Failed to stop tracking');
    },
  });
};

export {
  useOrders,
  useOrder,
  useCreateOrder,
  useCancelOrder,
  useReorder,
  useStartOrderTracking,
  useStopOrderTracking,
};
