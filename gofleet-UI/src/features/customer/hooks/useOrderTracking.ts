import { useState, useEffect, useCallback } from 'react';
import { TrackingSocket } from '@/services/websocket/trackingSocket';
import { useAuth } from '@/hooks/useAuth';
import { formatDistance } from '@/utils/formatDistance';
import { formatDuration } from '@/utils/formatDuration';

interface UseOrderTrackingReturn {
  driverLocation: { latitude: number; longitude: number } | null;
  driverName: string;
  driverRating: number;
  vehicleInfo: string;
  etaToPickup: string;
  etaToDestination: string;
  status: string;
  isTracking: boolean;
  error: string | null;
  refresh: () => void;
}

export const useOrderTracking = (orderId: string): UseOrderTrackingReturn => {
  const [driverLocation, setDriverLocation] = useState<{ latitude: number; longitude: number } | null>(null);
  const [driverName, setDriverName] = useState('');
  const [driverRating, setDriverRating] = useState(0);
  const [vehicleInfo, setVehicleInfo] = useState('');
  const [etaToPickup, setEtaToPickup] = useState('0 min');
  [etaToDestination, setEtaToDestination] = useState('0 min');
  const [status, setStatus] = useState('');
  const [isTracking, setIsTracking] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const { user } = useAuth();

  // Tracking socket instance
  const trackingSocketRef = React.useRef<TrackingSocket | null>(null);

  // Initialize tracking when orderId changes
  useEffect(() => {
    if (!orderId) return;

    // Create new tracking socket
    trackingSocketRef.current = new TrackingSocket(orderId, {
      onLocationUpdate: (data) => {
        setDriverLocation({
          latitude: data.latitude,
          longitude: data.longitude,
        });
      },
      onStatusChange: (newStatus) => {
        setStatus(newStatus);
      },
      onError: (err) => {
        console.error('Tracking error:', err);
        setError(err.message || 'Tracking error');
      },
      onConnectionStatus: (connected) => {
        setIsTracking(connected);
        if (!connected) {
          setError('Connection lost. Reconnecting...');
        }
      },
    });

    // Connect to WebSocket
    try {
      trackingSocketRef.current?.connect();
    } catch (err) {
      setError('Failed to connect to tracking service');
    }

    // Cleanup on unmount or orderId change
    return () => {
      trackingSocketRef.current?.disconnect();
      trackingSocketRef.current = null;
    };
  }, [orderId]);

  // Mock data for demo - in real app, this would come from API/WebSocket
  useEffect(() => {
    if (!driverLocation) {
      // Set some mock data for demonstration
      setDriverName('Emeka O.');
      setDriverRating(4.8);
      setVehicleInfo('Toyota Corolla • ABC-123XY');
      setEtaToPickup('4 min');
      setEtaToDestination('18 min');
      setStatus('assigned');
    }
  }, [driverLocation]);

  // Format ETAs for display
  const formattedEtas = {
    etaToPickup: etaToPickup,
    etaToDestination: etaToDestination,
  };

  // Manual refresh function
  const refresh = useCallback(() => {
    // In a real implementation, this would trigger a refresh of the connection
    // or fetch latest data from API
    if (trackingSocketRef.current) {
      // Reconnect
      trackingSocketRef.current.disconnect();
      setTimeout(() => {
        try {
          trackingSocketRef.current?.connect();
        } catch (err) {
          setError('Failed to refresh connection');
        }
      }, 1000);
    }
  }, []);

  return {
    driverLocation,
    driverName,
    driverRating: driverRating || 0,
    vehicleInfo,
    etaToPickup: formatDuration(parseFloat(etaToPickup.replace(' min', '')) || 0),
    etaToDestination: formatDuration(parseFloat(etaToDestination.replace(' min', '')) || 0),
    status,
    isTracking,
    error,
    refresh,
  };
};

export default useOrderTracking;
