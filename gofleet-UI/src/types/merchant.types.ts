export interface Merchant {
  id: string;
  userId: string;
  name: string;
  address: string;
  phone: string;
  email: string;
  isOpen: boolean;
}

export interface MenuCategory {
  id: string;
  merchantId: string;
  name: string;
}

export interface MenuItem {
  id: string;
  categoryId: string;
  name: string;
  priceMinor: number;
  currency: string;
  isAvailable: boolean;
}

export interface OrderItem {
  menuItemId: string;
  quantity: number;
  priceAtTimeOfPurchaseMinor: number;
}

export interface Order {
  id: string;
  customerId: string;
  items: OrderItem[];
  totalAmountMinor: number;
  currency: string;
  status: 'pending' | 'accepted' | 'preparing' | 'ready' | 'completed' | 'cancelled';
}

export interface DashboardStats {
  todayOrders: number;
  todayRevenue: number;
  pendingOrders: number;
}

export interface Settings {
  orderNotifications: boolean;
  autoAcceptOrders: boolean;
  currency: string;
}
