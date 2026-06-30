import { tokenStorage } from '@/features/auth/utils/tokenStorage';
import { API_ENDPOINTS } from '@/services/api/types';

export interface LocationUpdate {
  latitude: number;
  longitude: number;
  accuracy?: number;
  heading?: number;
  speed?: number;
  timestamp: number;
}

export interface StatusUpdate {
  status: string;
  timestamp: number;
  data?: Record<string, any>;
}

export interface TrackingCallbacks {
  onLocationUpdate: (data: LocationUpdate) => void;
  onStatusChange: (status: string) => void;
  onError?: (error: any) => void;
  onConnectionStatus?: (connected: boolean) => void;
}

export class TrackingSocket {
  private ws: WebSocket | null = null;
  private orderId: string;
  private callbacks: TrackingCallbacks;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectTimeout: NodeJS.Timeout | null = null;
  private heartbeatInterval: NodeJS.Timeout | null = null;
  private isConnected = false;
  private readonly HEARTBEAT_INTERVAL = 30000; // 30 seconds

  constructor(orderId: string, callbacks: TrackingCallbacks) {
    this.orderId = orderId;
    this.callbacks = callbacks;
  }

  /**
   * Connect to the WebSocket server
   */
  connect(): void {
    // Disconnect any existing connection
    this.disconnect();
    
    const token = tokenStorage.getAccessToken();
    if (!token) {
      throw new Error('No authentication token available');
    }
    
    // Construct WebSocket URL
    const wsUrl = `${process.env.EXPO_PUBLIC_WS_URL?.replace(/^http/, 'ws') || 'ws://localhost:8080'}/api/v1/orders/${this.orderId}/track/ws`;
    
    try {
      this.ws = new WebSocket(wsUrl, [
        `Authorization: Bearer ${token}`
      ]);
      
      this.setupEventListeners();
    } catch (error) {
      console.error('Failed to create WebSocket connection:', error);
      this.handleError(error);
    }
  }

  /**
   * Disconnect from WebSocket server
   */
  disconnect(): void {
    if (this.heartbeatInterval) {
      clearInterval(this.heartbeatInterval);
      this.heartbeatInterval = null;
    }
    
    if (this.reconnectTimeout) {
      clearTimeout(this.reconnectTimeout);
      this.reconnectTimeout = null;
    }
    
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
    
    this.isConnected = false;
    this.callbacks.onConnectionStatus?.(false);
  }

  /**
   * Set up WebSocket event listeners
   */
  private setupEventListeners(): void {
    if (!this.ws) return;
    
    this.ws.onopen = () => {
      console.log('WebSocket connected for order tracking:', this.orderId);
      this.isConnected = true;
      this.reconnectAttempts = 0; // Reset retry count on successful connection
      this.callbacks.onConnectionStatus?.(true);
      
      // Start heartbeat to keep connection alive
      this.startHeartbeat();
    };
    
    this.ws.onmessage = (event: MessageEvent) => {
      try {
        const data = JSON.parse(event.data);
        
        if (data.type === 'location') {
          this.callbacks.onLocationUpdate({
            latitude: data.payload.latitude,
            longitude: data.payload.longitude,
            accuracy: data.payload.accuracy,
            heading: data.payload.heading,
            speed: data.payload.speed,
            timestamp: data.payload.timestamp || Date.now(),
          });
        } else if (data.type === 'status') {
          this.callbacks.onStatusChange(data.payload.status);
        } else if (data.type === 'heartbeat_ack') {
          // Heartbeat acknowledged
        }
      } catch (error) {
        console.error('Error parsing WebSocket message:', error);
        this.callbacks.onError?.(error);
      }
    };
    
    this.ws.onclose = (event) => {
      console.log('WebSocket disconnected:', event.code, event.reason);
      this.isConnected = false;
      this.callbacks.onConnectionStatus?.(false);
      
      // Stop heartbeat
      if (this.heartbeatInterval) {
        clearInterval(this.heartbeatInterval);
        this.heartbeatInterval = null;
      }
      
      // Attempt reconnection with exponential backoff
      this.scheduleReconnect();
    };
    
    this.ws.onerror = (error) => {
      console.error('WebSocket error:', error);
      this.handleError(error);
    };
  }

  /**
   * Start heartbeat to keep connection alive
   */
  private startHeartbeat(): void {
    if (this.heartbeatInterval) return;
    
    this.heartbeatInterval = setInterval(() => {
      if (this.ws && this.ws.readyState === WebSocket.OPEN) {
        this.ws.send(JSON.stringify({ type: 'heartbeat' }));
      }
    }, this.HEARTBEAT_INTERVAL);
  }

  /**
   * Schedule reconnection attempt with exponential backoff
   */
  private scheduleReconnect(): void {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.error('Max reconnection attempts reached');
      this.callbacks.onError?.(new Error('Failed to reconnect after multiple attempts'));
      return;
    }
    
    // Exponential backoff: 1s, 2s, 4s, 8s, 16s
    const delay = Math.min(1000 * Math.pow(2, this.reconnectAttempts), 30000);
    
    this.reconnectTimeout = setTimeout(() => {
      this.reconnectAttempts++;
      console.log(`Reconnecting attempt ${this.reconnectAttempts}/${this.maxReconnectAttempts}`);
      this.connect();
    }, delay);
  }

  /**
   * Handle WebSocket errors
   */
  private handleError(error: any): void {
    console.error('WebSocket error:', error);
    this.callbacks.onError?.(error);
    
    // Try to reconnect on error
    this.scheduleReconnect();
  }

  /**
   * Send a message through the WebSocket
   */
  sendMessage(data: any): void {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(data));
    } else {
      console.warn('WebSocket not connected, cannot send message');
    }
  }

  /**
   * Get connection status
   */
  isConnected(): boolean {
    return this.isConnected;
  }
}

/**
 * Helper function to create a tracking socket instance
 */
export const createTrackingSocket = (orderId: string, callbacks: TrackingCallbacks) => {
  return new TrackingSocket(orderId, callbacks);
};

export default TrackingSocket;
