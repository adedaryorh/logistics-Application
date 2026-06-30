import React from 'react';
import { View, Text, StyleSheet, Platform, Dimensions, TouchableOpacity } from 'react-native';
import { Navigation, MapPin, Truck, Compass, Plus, Minus } from 'lucide-react-native';
import { colors } from '../colors';

// Try to import react-native-maps, fallback to Mock Map on web or if it errors
let MapViewComponent: any = null;
let MarkerComponent: any = null;
let PolylineComponent: any = null;

try {
  if (Platform.OS !== 'web') {
    const RNMaps = require('react-native-maps');
    MapViewComponent = RNMaps.default;
    MarkerComponent = RNMaps.Marker;
    PolylineComponent = RNMaps.Polyline;
  }
} catch (e) {
  // react-native-maps not linked or errored
}

export interface MapViewProps {
  showRoute?: boolean;
  pickupCoords?: { latitude: number; longitude: number };
  dropoffCoords?: { latitude: number; longitude: number };
  driverCoords?: { latitude: number; longitude: number };
  driverName?: string;
  dark?: boolean;
  className?: string;
}

export const MapView: React.FC<MapViewProps> = ({
  showRoute = true,
  pickupCoords = { latitude: 6.4281, longitude: 3.4219 }, // VI coordinates
  dropoffCoords = { latitude: 6.4589, longitude: 3.5186 }, // Lekki coordinates
  driverCoords = { latitude: 6.4400, longitude: 3.4600 },
  driverName = 'Emeka O.',
  dark = true,
  className = '',
}) => {
  const isNativeMapAvailable = MapViewComponent !== null;

  if (isNativeMapAvailable) {
    // Native Maps implementation
    const customMapStyle = dark ? [
      {
        "elementType": "geometry",
        "stylers": [{ "color": "#1A1A2E" }]
      },
      {
        "elementType": "labels.text.fill",
        "stylers": [{ "color": "#746855" }]
      },
      {
        "elementType": "labels.text.stroke",
        "stylers": [{ "color": "#242f3e" }]
      },
      {
        "featureType": "administrative.locality",
        "elementType": "labels.text.fill",
        "stylers": [{ "color": "#d59563" }]
      },
      {
        "featureType": "poi",
        "elementType": "labels.text.fill",
        "stylers": [{ "color": "#d59563" }]
      },
      {
        "featureType": "road",
        "elementType": "geometry",
        "stylers": [{ "color": "#16213E" }]
      },
      {
        "featureType": "road",
        "elementType": "geometry.stroke",
        "stylers": [{ "color": "#212a37" }]
      },
      {
        "featureType": "road",
        "elementType": "labels.text.fill",
        "stylers": [{ "color": "#9ca3af" }]
      },
      {
        "featureType": "water",
        "elementType": "geometry",
        "stylers": [{ "color": "#0F3460" }]
      }
    ] : [];

    return (
      <View style={styles.container} className={className}>
        <MapViewComponent
          style={styles.map}
          customMapStyle={customMapStyle}
          initialRegion={{
            latitude: (pickupCoords.latitude + dropoffCoords.latitude) / 2,
            longitude: (pickupCoords.longitude + dropoffCoords.longitude) / 2,
            latitudeDelta: Math.abs(pickupCoords.latitude - dropoffCoords.latitude) * 1.5,
            longitudeDelta: Math.abs(pickupCoords.longitude - dropoffCoords.longitude) * 1.5,
          }}
        >
          {driverCoords && (
            <MarkerComponent coordinate={driverCoords} title={driverName} description="Driver Location">
              <View className="p-2 bg-success rounded-full border-2 border-white shadow-lg">
                <Truck size={16} color="white" />
              </View>
            </MarkerComponent>
          )}

          {pickupCoords && (
            <MarkerComponent coordinate={pickupCoords} title="Pickup Point">
              <View className="p-2 bg-info rounded-full border-2 border-white shadow-lg">
                <MapPin size={16} color="white" />
              </View>
            </MarkerComponent>
          )}

          {dropoffCoords && (
            <MarkerComponent coordinate={dropoffCoords} title="Dropoff Point">
              <View className="p-2 bg-highlight rounded-full border-2 border-white shadow-lg">
                <MapPin size={16} color="white" />
              </View>
            </MarkerComponent>
          )}

          {showRoute && pickupCoords && dropoffCoords && (
            <PolylineComponent
              coordinates={[pickupCoords, driverCoords, dropoffCoords].filter(Boolean)}
              strokeColor={colors.routeLine}
              strokeWidth={4}
              lineDashPattern={[5, 5]}
            />
          )}
        </MapViewComponent>
      </View>
    );
  }

  // Beautiful visual placeholder for web, mock showcases, or missing configuration
  const mapBg = dark ? 'bg-primary' : 'bg-gray-100';
  const gridLine = dark ? 'border-secondary/20' : 'border-gray-200';
  const textTitle = dark ? 'text-gray-400' : 'text-gray-600';

  return (
    <View className={`relative overflow-hidden w-full h-full ${mapBg} ${className}`} style={styles.container}>
      {/* Grid Overlay for Premium Technical Aesthetic */}
      <View className="absolute inset-0 flex-row flex-wrap opacity-40">
        {Array.from({ length: 48 }).map((_, i) => (
          <View key={i} className={`w-[16.66%] h-20 border-r border-b ${gridLine}`} />
        ))}
      </View>

      {/* Stylized Mock Map Roads */}
      <View className="absolute inset-0" pointerEvents="none">
        {/* Main Road 1 */}
        <View
          style={{
            position: 'absolute',
            top: '40%',
            left: '-10%',
            width: '120%',
            height: 24,
            transform: [{ rotate: '-12deg' }],
            backgroundColor: dark ? '#16213E' : '#E5E7EB',
            borderTopWidth: 1,
            borderBottomWidth: 1,
            borderColor: dark ? '#0F3460' : '#D1D5DB',
          }}
        />
        {/* Main Road 2 */}
        <View
          style={{
            position: 'absolute',
            top: '20%',
            left: '30%',
            width: '10%',
            height: '100%',
            transform: [{ rotate: '35deg' }],
            backgroundColor: dark ? '#16213E' : '#E5E7EB',
            borderLeftWidth: 1,
            borderRightWidth: 1,
            borderColor: dark ? '#0F3460' : '#D1D5DB',
          }}
        />
        {/* River overlay */}
        <View
          style={{
            position: 'absolute',
            bottom: '-10%',
            left: '-10%',
            width: '60%',
            height: '45%',
            borderRadius: 180,
            backgroundColor: dark ? 'rgba(15, 52, 96, 0.4)' : 'rgba(59, 130, 246, 0.15)',
          }}
        />
      </View>

      {/* Styled Route Line Path */}
      {showRoute && (
        <View className="absolute inset-0 items-center justify-center" pointerEvents="none">
          <svg
            style={{ width: '100%', height: '100%', position: 'absolute' }}
            viewBox="0 0 400 400"
          >
            {/* Draw curve connecting points */}
            <path
              d="M 120 180 Q 220 150 280 250"
              fill="none"
              stroke={colors.routeLine}
              strokeWidth="4"
              strokeDasharray="6,6"
            />
          </svg>
        </View>
      )}

      {/* Driver Marker */}
      <View
        className="absolute p-2 bg-success rounded-full border-2 border-white shadow-lg"
        style={{ top: '38%', left: '48%' }}
      >
        <Truck size={14} color="white" />
        <View className="absolute -top-6 -left-4 bg-primary px-2 py-0.5 rounded shadow border border-gray-800">
          <Text className="text-[9px] text-white font-bold">{driverName}</Text>
        </View>
      </View>

      {/* Pickup Marker */}
      <View
        className="absolute p-2 bg-info rounded-full border-2 border-white shadow-lg"
        style={{ top: '40%', left: '26%' }}
      >
        <MapPin size={14} color="white" />
        <View className="absolute -top-6 -left-3 bg-secondary px-2 py-0.5 rounded shadow border border-gray-800">
          <Text className="text-[9px] text-white font-bold">Pickup</Text>
        </View>
      </View>

      {/* Dropoff Marker */}
      <View
        className="absolute p-2 bg-highlight rounded-full border-2 border-white shadow-lg"
        style={{ top: '60%', left: '66%' }}
      >
        <MapPin size={14} color="white" />
        <View className="absolute -top-6 -left-3 bg-secondary px-2 py-0.5 rounded shadow border border-gray-800">
          <Text className="text-[9px] text-white font-bold">Dropoff</Text>
        </View>
      </View>

      {/* Map Control Utilities */}
      <View className="absolute bottom-4 right-4 flex-col gap-2">
        <TouchableOpacity className={`p-2.5 rounded-full ${dark ? 'bg-secondary' : 'bg-white'} shadow-md`}>
          <Compass size={18} color={colors.highlight} />
        </TouchableOpacity>
        <View className={`rounded-xl ${dark ? 'bg-secondary' : 'bg-white'} shadow-md overflow-hidden`}>
          <TouchableOpacity className="p-2.5 border-b border-gray-800/10">
            <Plus size={18} color={dark ? 'white' : 'black'} />
          </TouchableOpacity>
          <TouchableOpacity className="p-2.5">
            <Minus size={18} color={dark ? 'white' : 'black'} />
          </TouchableOpacity>
        </View>
      </View>

      {/* Bottom overlay warning badge */}
      <View className="absolute bottom-4 left-4 bg-primary/95 border border-secondary px-3 py-1.5 rounded-xl flex-row items-center gap-1.5 shadow-md">
        <Navigation size={12} color={colors.success} className="animate-pulse" />
        <Text className="text-[10px] text-gray-300 font-medium">GPS Live Simulation Active</Text>
      </View>
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    width: '100%',
    height: '100%',
  },
  map: {
    ...StyleSheet.absoluteFill,
  },
});

export default MapView;
