import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import orderService from '@/services/orderService';
import { useAuthStore } from '@/features/auth/store/authStore';
import { toast } from '@/design-system/components/Toast';
import * as Haptics from '@/utils/haptics';

// Query keys
export const DRIVER_QUERY_KEYS = {
  STATUS: ['driver', 'status'],
  EARNINGS: ['driver', 'earnings'],
  HISTORY: () => [...DRIVER_QUERY_KEYS.ALL, 'history'],
  HISTORY_PAGE: (page: number) => [...DRIVER_QUERY_KEYS.HISTORY(), page],
  OFFERS: ['driver', 'offers'],
};

// Custom hook for getting driver status
export const useDriverStatus = () => {
  const { data, isLoading, error, refetch } = useQuery({
    queryKey: DRIVER_QUERY_KEYS.STATUS,
    queryFn: () => orderService.getDriverStatus(),
    refetchInterval: 30000, // Refetch every 30 seconds
  });

  return {
    status: data?.status,
    isOnline: data?.isOnline,
    isLoading,
    error,
    refetch,
  };
};

// Custom hook for going online/offline
export const useDriverToggleStatus = () => {
  const queryClient = useQueryClient();
  const { status } = useDriverStatus(); // Get current status
  
  const toggleStatus = useMutation({
    mutationFn: async (isOnline: boolean) => {
      if (isOnline) {
        await orderService.goOnline();
      } else {
        await orderService.goOffline();
      }
    },
    onSuccess: () => {
      // Refetch status
      queryClient.invalidateQueries({ queryKey: DRIVER_QUERY_KEYS.STATUS });
      
      // Show success message
      Haptics.haptics.success();
      toast.success(`You are now ${status === 'online' ? 'offline' : 'online'}`);
    },
    onError: (error: any) => {
      Haptics.haptics.error();
      toast.error(error.message || 'Failed to update status');
    },
  });

  return {
    toggleStatus: toggleStatus.mutate,
    isLoading: toggleStatus.isLoading,
  };
};

// Custom hook for getting driver earnings
export const useDriverEarnings = () => {
  const { data, isLoading, error, refetch } = useQuery({
    queryKey: DRIVER_QUERY_KEYS.EARNINGS,
    queryFn: () => orderService.getDriverEarnings(),
    refetchInterval: 300000, // Refetch every 5 minutes
  });

  return {
    earnings: data,
    isLoading,
    error,
    refetch,
  };
};

// Custom hook for getting driver history
export const useDriverHistory = (page: number = 1) => {
  const { data, isLoading, error, refetch } = useQuery({
    queryKey: DRIVER_QUERY_KEYS.HISTORY_PAGE(page),
    queryFn: () => orderService.getDriverHistory(page),
    keepPreviousData: true,
  });

  return {
    history: data?.data || [],
    total: data?.total || 0,
    isLoading,
    error,
    refetch,
  };
};

// Custom hook for getting driver offers
export const useDriverOffers = () => {
  const { data, isLoading, error, refetch } = useQuery({
    queryKey: DRIVER_QUERY_KEYS.OFFERS,
    queryFn: () => orderService.getDriverOffers(),
    refetchInterval: 10000, // Refetch every 10 seconds for new offers
  });

  return {
    offers: data || [],
    isLoading,
    error,
    refetch,
  };
};

// Custom hook for accepting an offer
export const useAcceptOffer = () => {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: (offerId: string) => orderService.acceptOffer(offerId),
    onSuccess: () => {
      // Invalidate and refetch queries
      queryClient.invalidateQueries({ queryKey: DRIVER_QUERY_KEYS.OFFERS });
      queryClient.invalidateQueries({ queryKey: DRIVER_QUERY_KEYS.STATUS });
      
      // Show success message
      Haptics.haptics.success();
      toast.success('Offer accepted!');
    },
    onError: (error: any) => {
      Haptics.haptics.error();
      toast.error(error.message || 'Failed to accept offer');
    },
  });
};

// Custom hook for rejecting an offer
export const useRejectOffer = () => {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: (offerId: string) => orderService.rejectOffer(offerId),
    onSuccess: () => {
      // Invalidate and refetch queries
      queryClient.invalidateQueries({ queryKey: DRIVER_QUERY_KEYS.OFFERS });
      
      // Show success message
      Haptics.haptics.success();
      toast.success('Offer rejected!');
    },
    onError: (error: any) => {
      Haptics.haptics.error();
      toast.error(error.message || 'Failed to reject offer');
    },
  });
};

// Custom hook for starting a trip
export const useStartTrip = () => {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: (tripId: string) => orderService.startTrip(tripId),
    onSuccess: () => {
      // Invalidate and refetch queries
      queryClient.invalidateQueries({ queryKey: DRIVER_QUERY_KEYS.STATUS });
      
      // Show success message
      Haptics.haptics.success();
      toast.success('Trip started!');
    },
    onError: (error: any) => {
      Haptics.haptics.error();
      toast.error(error.message || 'Failed to start trip');
    },
  });
};

// Custom hook for ending a trip
export const useEndTrip = () => {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: (tripId: string) => orderService.endTrip(tripId),
    onSuccess: () => {
      // Invalidate and refetch queries
      queryClient.invalidateQueries({ queryKey: DRIVER_QUERY_KEYS.STATUS });
      
      // Show success message
      Haptics.haptics.success();
      toast.success('Trip ended!');
    },
    onError: (error: any) => {
      Haptics.haptics.error();
      toast.error(error.message || 'Failed to end trip');
    },
  });
};

export {
  useDriverStatus,
  useDriverToggleStatus,
  useDriverEarnings,
  useDriverHistory,
  useDriverOffers,
  useAcceptOffer,
  useRejectOffer,
  useStartTrip,
  useEndTrip,
};
