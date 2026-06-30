export type DriverStatus = 'online' | 'offline' | 'on_trip' | 'suspended';

export interface Driver {
  id: string;
  userId: string;
  status: DriverStatus;
  ratingAvg: number;
  totalTrips: number;
}

export interface Vehicle {
  id: string;
  type: 'car' | 'bike' | 'van' | 'truck';
  model: string;
  licensePlate: string;
}

export interface DispatchOffer {
  assignmentId: string;
  orderId: string;
  orderType: 'ride' | 'food' | 'parcel';
  pickupAddress: string;
  pickupEtaMinutes: number;
  dropoffAddress: string;
  tripEtaMinutes: number;
  fareMinor: number;
  currency: string;
}

export interface Location {
  lat: number;
  lng: number;
  address?: string;
}

export interface Trip {
  id: string;
  orderId: string;
  status: 'accepted' | 'en_route_pickup' | 'at_pickup' | 'en_route_dropoff' | 'completed';
  pickupLocation: Location;
  dropoffLocation: Location;
}

export interface EarningsSummary {
  today: number;
  week: number;
  month: number;
  total: number;
}
