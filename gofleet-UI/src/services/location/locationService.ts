import { tokenStorage } from '@/features/auth/utils/tokenStorage';
import { API_ENDPOINTS } from '@/services/api/types';
import { Platform, AppState } from 'react-native';
import Geolocation from 'react-native-geolocation-service';
// For background tracking, you would typically use:
// import BackgroundGeolocation from 'react-native-background-geolocation';
// But for simplicity, we'll use react-native-geolocation-service with app state monitoring

export interface LocationData {
  latitude: number;
  longitude: number;
  accuracy: number;
  heading?: number;
  speed?: number;
  timestamp: number;
}

export class LocationService {
  private watchId: number | null = null;
  private isTracking = false;
  private readonly MIN_DISTANCE_METERS = 10; // Minimum distance to update
  private lastLocation: LocationData | null = null;
  private readonly UPDATE_INTERVAL_MS = 3000; // 3 seconds as specified
  private AppState: any;

  constructor() {
    // For background tracking, you would initialize BackgroundGeolocation here
    // This is a simplified version using standard geolocation
  }

  /**
   * Start tracking location
   */
  startTracking = async (): Promise<void> => {
    if (this.isTracking) return;
    
    try {
      // Request location permission
      const granted = await this.requestPermission();
      if (!granted) {
        throw new Error('Location permission not granted');
      }
      
      // Set up location watch
      this.watchId = Geolocation.watchPosition(
        (position) => this.handleLocationUpdate(position),
        (error) => this.handleLocationError(error),
        {
          enableHighAccuracy: true,
          distanceFilter: this.MIN_DISTANCE_METERS,
          interval: this.UPDATE_INTERVAL_MS,
          fastestInterval: this.UPDATE_INTERVAL_MS,
        }
      );
      
      this.isTracking = true;
      console.log('Location tracking started');
      
      // Monitor app state to handle background/foreground transitions
      this.setupAppStateListener();
    } catch (error) {
      console.error('Failed to start location tracking:', error);
      throw error;
    }
  };

  /**
   * Stop tracking location
   */
  stopTracking = (): void => {
    if (this.watchId !== null) {
      Geolocation.clearWatch(this.watchId);
      this.watchId = null;
    }
    
    this.isTracking = false;
    this.lastLocation = null;
    
    // Remove app state listener
    if (this.AppState) {
      this.AppState.removeEventListener('change', this.handleAppStateChange);
    }
    
    console.log('Location tracking stopped');
  };

  /**
   * Request location permission
   */
  private requestPermission = async (): Promise<boolean> => {
    try {
      const granted = await Geolocation.requestAuthorization('whenInUse');
      if (granted === 'granted' || granted === true) {
        return true;
      }
      
      // If not granted, try to get permission permanently
      const status = await Geolocation.requestPermission();
      return status === 'granted';
    } catch (error) {
      console.error('Error requesting location permission:', error);
      return false;
    }
  };

  /**
   * Handle location updates
   */
  private handleLocationUpdate = (position: any) => {
    const newLocation: LocationData = {
      latitude: position.coords.latitude,
      longitude: position.coords.longitude,
      accuracy: position.coords.accuracy,
      heading: position.coords.heading ?? null,
      speed: position.coords.speed ?? null,
      timestamp: position.timestamp,
    };
    
    // Only send update if moved significantly or first update
    if (
      !this.lastLocation ||
      this.isSignificantMovement(this.lastLocation, newLocation)
    ) {
      this.lastLocation = newLocation;
      this.sendLocationUpdate(newLocation);
    }
  };

  /**
   * Check if location has moved significantly
   */
  private isSignificantMovement = (oldLoc: LocationData, newLoc: LocationData): boolean => {
    // Haversine formula to calculate distance between two points
    const R = 6371e3; // Earth's radius in meters
    const φ1 = (oldLoc.latitude * Math.PI) / 180;
    const φ2 = (newLoc.latitude * Math.PI) / 180;
    const Δφ = ((newLoc.latitude - oldLoc.latitude) * Math.PI) / 180;
    const Δλ = ((newLoc.longitude - oldLoc.longitude) * Math.PI) / 180;
    
    const a =
      Math.sin(Δφ / 2) * Math.sin(Δφ / 2) +
      Math.cos(φ1) * Math.cos(φ2) *
      Math.sin(Δλ / 2) * Math.sin(Δλ / 2);
    const c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a));
    
    const distance = R * c; // Distance in meters
    return distance >= this.MIN_DISTANCE_METERS;
  };

  /**
   * Send location update to server
   */
  private sendLocationUpdate = async (location: LocationData) => {
    try {
      const token = tokenStorage.getAccessToken();
      if (!token) {
        console.warn('No auth token available for location update');
        return;
      }
      
      await fetch(`${process.env.EXPO_PUBLIC_API_URL || 'http://localhost:8080'}${API_ENDPOINTS.DRIVER.LOCATION}`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({
          latitude: location.latitude,
          longitude: location.longitude,
          accuracy: location.accuracy,
          heading: location.heading,
          speed: location.speed,
          timestamp: Date.now(),
        }),
      });
    } catch (error) {
      console.error('Failed to send location update:', error);
      // Don't throw - location updates should not break the app
    }
  };

  /**
   * Handle location errors
   */
  private handleLocationError = (error: any) => {
    console.error('Location error:', error);
    
    // Handle specific error cases
    if (error.code === 1) { // Permission denied
      this.stopTracking();
    } else if (error.code === 2) { // Position unavailable
      // Try again after a short delay
      setTimeout(() => {
        if (this.isTracking) {
          // Restart watch
          if (this.watchId !== null) {
            Geolocation.clearWatch(this.watchId);
          }
          this.watchId = Geolocation.watchPosition(
            this.handleLocationUpdate,
            this.handleLocationError,
            {
              enableHighAccuracy: true,
              distanceFilter: this.MIN_DISTANCE_METERS,
              interval: this.UPDATE_INTERVAL_MS,
            }
          );
        }
      }, 5000);
    }
  };

  /**
   * Setup app state listener to handle background/foreground transitions
   */
  private setupAppStateListener = () => {
    if (typeof AppState !== 'undefined') {
      this.AppState = require('react-native').AppState;
      this.AppState.addEventListener('change', this.handleAppStateChange);
    }
  };

  /**
   * Handle app state changes
   */
  private handleAppStateChange = (nextState: string) => {
    if (this.isTracking) {
      if (nextState === 'background') {
        // App went to background - reduce accuracy to save battery
        console.log('App entered background mode');
        // In a real implementation, you would switch to background-specific settings
      } else if (nextState === 'active') {
        // App came to foreground - restore full accuracy
        console.log('App entered foreground mode');
      }
    }
  };

  /**
   * Get last known location
   */
  getLastLocation = (): LocationData | null => {
    return this.lastLocation;
  };

  /**
   * Check if currently tracking
   */
  isTrackingActive = (): boolean => {
    return this.isTracking;
  };
}

// Singleton instance
export const locationService = new LocationService();

export default locationService;
