import { useCallback, useEffect, useState } from 'react';
import { api } from '../api';
import { demoOrders } from '../data/demo';
import { Order } from '../types';

export function useOrders(demoMode: boolean) {
  const [orders, setOrders] = useState<Order[]>(demoMode ? demoOrders : []);
  const [loading, setLoading] = useState(!demoMode);
  const [error, setError] = useState<string>();
  const refresh = useCallback(async () => {
    if (demoMode) {
      setOrders(demoOrders);
      setLoading(false);
      return;
    }
    try {
      setLoading(true);
      setError(undefined);
      const data = await api.orders();
      setOrders(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Could not load orders.');
    } finally {
      setLoading(false);
    }
  }, [demoMode]);
  useEffect(() => {
    refresh();
  }, [refresh]);
  return { orders, loading, error, refresh, setOrders };
}
