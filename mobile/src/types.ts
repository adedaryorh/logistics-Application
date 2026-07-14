export type Tab = 'home' | 'orders' | 'wallet' | 'profile';
export type DeliveryType = 'parcel' | 'food' | 'ride' | 'agricultural';
export type OrderStatus =
  | 'pending'
  | 'awaiting_payment'
  | 'paid'
  | 'dispatching'
  | 'assigned'
  | 'picked_up'
  | 'delivered'
  | 'cancelled'
  | 'failed';

export type Coordinate = { lat: number; lng: number; address?: string };
export type Order = {
  id: string;
  type: DeliveryType;
  status: OrderStatus;
  pickup: Coordinate;
  dropoff: Coordinate;
  price_minor: number;
  currency: string;
  created_at: string;
  driver_id?: string;
  platform_user_id?: string;
  marketplace_request_id?: string;
  agricultural_shipment?: AgriculturalShipment;
  proofs?: DeliveryProof[];
};

export type AgriculturalShipment = {
  produce_type: string;
  quantity: number;
  quantity_unit: string;
  weight_kg?: number;
  packaging: string;
  requires_refrigeration: boolean;
  cold_chain_min_c?: number;
  cold_chain_max_c?: number;
  pickup_window: { start_at: string; end_at: string };
  delivery_window: { start_at: string; end_at: string };
  handling_notes?: string;
  loading_notes?: string;
};
export type DeliveryProof = {
  id: string;
  type: 'pickup' | 'delivery';
  evidence_url: string;
  recipient_name?: string;
  captured_at: string;
};

export type User = { id: string; email: string; phone?: string; role?: string };
export type Tracking = {
  order_id?: string;
  status?: OrderStatus;
  eta_minutes?: number;
  driver?: { full_name?: string; phone?: string; rating?: number };
};
export type Wallet = { user_id: string; balance_minor: number; currency: string };
export type WalletEntry = {
  id: string;
  type: 'credit' | 'debit';
  amount_minor: number;
  description?: string;
  created_at: string;
};
export type CreateOrderInput = {
  type: DeliveryType;
  pickup: Coordinate;
  dropoff: Coordinate;
  items: unknown[];
  idempotency_key: string;
};
