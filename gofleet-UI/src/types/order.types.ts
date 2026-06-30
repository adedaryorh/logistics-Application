import type { OrderStatus } from '@/constants/orderStatuses';
import type { OrderType } from '@/constants/orderTypes';

export type Location = {
  lat: number;
  lng: number;
  address: string;
};

export type OrderItem = {
  id: string;
  name: string;
  quantity: number;
  unitPriceMinor: number;
};

export type PriceBreakdownItem = {
  label: string;
  amountMinor: number;
};

export type DriverSummary = {
  id: string;
  name: string;
  rating: number;
  vehicle: string;
};

export type Order = {
  id: string;
  type: OrderType;
  status: OrderStatus;
  pickup: Location;
  dropoff: Location;
  customerName?: string;
  merchantName?: string;
  driver?: DriverSummary;
  etaMinutes?: number;
  totalMinor: number;
  currency: string;
  createdAt: string;
  updatedAt: string;
  notes?: string;
  items?: OrderItem[];
  priceBreakdown?: PriceBreakdownItem[];
};
