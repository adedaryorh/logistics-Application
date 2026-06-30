export const ORDER_TYPES = ['ride', 'food', 'parcel'] as const;

export type OrderType = (typeof ORDER_TYPES)[number];

export const ORDER_TYPE_LABELS: Record<OrderType, string> = {
  ride: 'Ride',
  food: 'Food',
  parcel: 'Parcel',
};
