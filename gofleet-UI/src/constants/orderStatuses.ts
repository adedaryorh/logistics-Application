export const ORDER_STATUSES = [
  'pending',
  'accepted',
  'assigned',
  'arriving',
  'in_transit',
  'delivered',
  'completed',
  'cancelled',
] as const;

export type OrderStatus = (typeof ORDER_STATUSES)[number];

export const ORDER_STATUS_LABELS: Record<OrderStatus, string> = {
  pending: 'Pending',
  accepted: 'Accepted',
  assigned: 'Driver Assigned',
  arriving: 'Driver Arriving',
  in_transit: 'In Transit',
  delivered: 'Delivered',
  completed: 'Completed',
  cancelled: 'Cancelled',
};
