import React from 'react';
import { createBottomTabNavigator } from '@react-navigation/bottom-tabs';
import MaterialCommunityIcons from 'react-native-vector-icons/MaterialCommunityIcons';

import { colors } from '@/design-system/colors';
import { DashboardScreen } from './dashboard';
import { OrdersScreen } from './orders';
import { MenuScreen } from './menu';
import { SettingsScreen } from './settings';

const Tab = createBottomTabNavigator();

type TabBarIconProps = {
  color: string;
  size: number;
};

const renderDashboardIcon = ({ color, size }: TabBarIconProps) => (
  <MaterialCommunityIcons name="view-dashboard-outline" color={color} size={size} />
);

const renderOrdersIcon = ({ color, size }: TabBarIconProps) => (
  <MaterialCommunityIcons name="clipboard-list-outline" color={color} size={size} />
);

const renderMenuIcon = ({ color, size }: TabBarIconProps) => (
  <MaterialCommunityIcons name="food-outline" color={color} size={size} />
);

const renderSettingsIcon = ({ color, size }: TabBarIconProps) => (
  <MaterialCommunityIcons name="cog-outline" color={color} size={size} />
);

export function MerchantTabNavigator() {
  return (
    <Tab.Navigator
      screenOptions={{
        headerShown: false,
        tabBarActiveTintColor: colors.highlight,
        tabBarInactiveTintColor: colors.gray500,
      }}
    >
      <Tab.Screen name="Dashboard" component={DashboardScreen} options={{ tabBarIcon: renderDashboardIcon }} />
      <Tab.Screen name="Orders" component={OrdersScreen} options={{ tabBarIcon: renderOrdersIcon }} />
      <Tab.Screen name="Menu" component={MenuScreen} options={{ tabBarIcon: renderMenuIcon }} />
      <Tab.Screen name="Settings" component={SettingsScreen} options={{ tabBarIcon: renderSettingsIcon }} />
    </Tab.Navigator>
  );
}

export default MerchantTabNavigator;
