export type PaymentStatus = 'pending' | 'processing' | 'completed' | 'failed' | 'refunded';

export type PaymentMethod = 'card' | 'bank_transfer' | 'ussd' | 'wallet' | 'cash';

export interface Payment {
  id: string;
  orderId: string;
  amountMinor: number;
  currency: string;
  status: PaymentStatus;
  method: PaymentMethod;
  provider: string;
  providerPaymentId?: string;
  feesMinor?: number;
  netAmountMinor?: number;
  initiatedAt: string;
  processedAt?: string;
  completedAt?: string;
  failedAt?: string;
  failureReason?: string;
  metadata?: Record<string, unknown>;
}

export interface PaymentInitRequest {
  orderId: string;
  amountMinor: number;
  currency: string;
  method: PaymentMethod;
  metadata?: Record<string, unknown>;
}

export interface PaymentInitResponse {
  paymentId: string;
  status: PaymentStatus;
  redirectUrl?: string;
  clientSecret?: string;
  requiresAction: boolean;
}

export interface PaymentVerificationRequest {
  paymentId: string;
}

export interface PaymentVerificationResponse {
  status: PaymentStatus;
  amountMinor: number;
  currency: string;
  method: PaymentMethod;
  feesMinor?: number;
  netAmountMinor?: number;
  processedAt: string;
}

export interface Wallet {
  id: string;
  userId: string;
  balanceMinor: number;
  currency: string;
  transactions: WalletTransaction[];
}

export interface WalletTransaction {
  id: string;
  type: 'credit' | 'debit';
  amountMinor: number;
  currency: string;
  description: string;
  relatedEntityId?: string;
  createdAt: string;
}

export interface PaymentMethodInfo {
  id: PaymentMethod;
  name: string;
  icon: string;
  isAvailable: boolean;
  fees: {
    fixedMinor: number;
    percentage: number;
  };
  limits: {
    min: number;
    max: number;
  };
}
