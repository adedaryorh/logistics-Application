import React from 'react';
import { createBottomTabNavigator } from '@react-navigation/bottom-tabs';
import MaterialCommunityIcons from 'react-native-vector-icons/MaterialCommunityIcons';

import { colors } from '@/design-system/colors';
import { HomeScreen } from './home';
import { EarningsScreen } from './earnings';
import { HistoryScreen } from './history';
import { ProfileScreen } from './profile';

const Tab = createBottomTabNavigator();

type TabBarIconProps = {
  color: string;
  size: number;
};

const renderDashboardIcon = ({ color, size }: TabBarIconProps) => (
  <MaterialCommunityIcons name="steering" color={color} size={size} />
);

const renderEarningsIcon = ({ color, size }: TabBarIconProps) => (
  <MaterialCommunityIcons name="cash-multiple" color={color} size={size} />
);

const renderHistoryIcon = ({ color, size }: TabBarIconProps) => (
  <MaterialCommunityIcons name="history" color={color} size={size} />
);

const renderProfileIcon = ({ color, size }: TabBarIconProps) => (
  <MaterialCommunityIcons name="account-circle-outline" color={color} size={size} />
);

export function DriverTabNavigator() {
  return (
    <Tab.Navigator
      screenOptions={{
        headerShown: false,
        tabBarActiveTintColor: colors.highlight,
        tabBarInactiveTintColor: colors.gray500,
      }}
    >
      <Tab.Screen name="Dashboard" component={HomeScreen} options={{ tabBarIcon: renderDashboardIcon }} />
      <Tab.Screen name="Earnings" component={EarningsScreen} options={{ tabBarIcon: renderEarningsIcon }} />
      <Tab.Screen name="History" component={HistoryScreen} options={{ tabBarIcon: renderHistoryIcon }} />
      <Tab.Screen name="Profile" component={ProfileScreen} options={{ tabBarIcon: renderProfileIcon }} />
    </Tab.Navigator>
  );
}

export default DriverTabNavigator;
