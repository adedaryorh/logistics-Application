import React from 'react';
import { createBottomTabNavigator } from '@react-navigation/bottom-tabs';
import MaterialCommunityIcons from 'react-native-vector-icons/MaterialCommunityIcons';

import { colors } from '@/design-system/colors';
import { HomeScreen } from './home';
import { BookRideScreen } from './book-ride';
import { OrderFoodScreen } from './order-food';
import { TrackingScreen } from './tracking';
import { OrdersScreen } from './orders';

const Tab = createBottomTabNavigator();

type TabBarIconProps = {
  color: string;
  size: number;
};

const renderHomeIcon = ({ color, size }: TabBarIconProps) => (
  <MaterialCommunityIcons name="home-variant-outline" color={color} size={size} />
);

const renderRideIcon = ({ color, size }: TabBarIconProps) => (
  <MaterialCommunityIcons name="car-outline" color={color} size={size} />
);

const renderFoodIcon = ({ color, size }: TabBarIconProps) => (
  <MaterialCommunityIcons name="silverware-fork-knife" color={color} size={size} />
);

const renderTrackingIcon = ({ color, size }: TabBarIconProps) => (
  <MaterialCommunityIcons name="crosshairs-gps" color={color} size={size} />
);

const renderOrdersIcon = ({ color, size }: TabBarIconProps) => (
  <MaterialCommunityIcons name="clipboard-text-outline" color={color} size={size} />
);

export function CustomerTabNavigator() {
  return (
    <Tab.Navigator
      screenOptions={{
        headerShown: false,
        tabBarActiveTintColor: colors.highlight,
        tabBarInactiveTintColor: colors.gray500,
      }}
    >
      <Tab.Screen name="Home" component={HomeScreen} options={{ tabBarIcon: renderHomeIcon }} />
      <Tab.Screen name="Ride" component={BookRideScreen} options={{ tabBarIcon: renderRideIcon }} />
      <Tab.Screen name="Food" component={OrderFoodScreen} options={{ tabBarIcon: renderFoodIcon }} />
      <Tab.Screen name="Tracking" component={TrackingScreen} options={{ tabBarIcon: renderTrackingIcon }} />
      <Tab.Screen name="Orders" component={OrdersScreen} options={{ tabBarIcon: renderOrdersIcon }} />
    </Tab.Navigator>
  );
}

export default CustomerTabNavigator;
