import { Order, Tracking, User } from '../types';

export const demoUser: User = {
  id: 'demo-user',
  email: 'tobi@example.com',
  phone: '+234 801 234 5678',
};
export const demoOrders: Order[] = [
  {
    id: 'LG-2048',
    type: 'parcel',
    status: 'picked_up',
    pickup: { lat: 6.4474, lng: 3.4723, address: '12 Admiralty Way, Lekki' },
    dropoff: { lat: 6.4281, lng: 3.4219, address: '14A Adeola Odeku, Victoria Island' },
    price_minor: 385000,
    currency: 'NGN',
    created_at: new Date().toISOString(),
    driver_id: 'demo-rider',
  },
  {
    id: 'LG-1921',
    type: 'food',
    status: 'delivered',
    pickup: { lat: 6.45, lng: 3.47, address: 'Lekki Phase 1' },
    dropoff: { lat: 6.44, lng: 3.43, address: 'Ikoyi' },
    price_minor: 240000,
    currency: 'NGN',
    created_at: '2026-07-12T18:14:00Z',
  },
  {
    id: 'LG-1877',
    type: 'ride',
    status: 'delivered',
    pickup: { lat: 6.43, lng: 3.42, address: 'Victoria Island' },
    dropoff: { lat: 6.52, lng: 3.37, address: 'Surulere' },
    price_minor: 610000,
    currency: 'NGN',
    created_at: '2026-07-08T15:20:00Z',
  },
];
export const demoTracking: Tracking = {
  order_id: 'LG-2048',
  status: 'picked_up',
  eta_minutes: 12,
  driver: { full_name: 'Damilola A.', phone: '+234 800 000 0000', rating: 4.9 },
};
