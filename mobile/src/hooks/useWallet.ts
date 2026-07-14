import { useCallback, useEffect, useState } from 'react';
import { api } from '../api';
import { Wallet, WalletEntry } from '../types';

const empty: Wallet = { user_id: 'demo', balance_minor: 2460000, currency: 'NGN' };
export function useWallet(enabled: boolean, demoMode: boolean) {
  const [wallet, setWallet] = useState<Wallet>(empty);
  const [ledger, setLedger] = useState<WalletEntry[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string>();
  const refresh = useCallback(async () => {
    if (!enabled || demoMode) {
      setWallet(empty);
      return;
    }
    try {
      setLoading(true);
      setError(undefined);
      const data = await api.wallet();
      setWallet(data.wallet);
      setLedger(data.ledger);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Could not load wallet.');
    } finally {
      setLoading(false);
    }
  }, [enabled, demoMode]);
  useEffect(() => {
    refresh();
  }, [refresh]);
  return { wallet, ledger, loading, error, refresh };
}
