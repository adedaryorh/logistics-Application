import React from 'react';
import { ActivityIndicator, StyleSheet, Text, View } from 'react-native';
import { NavigationContainer } from '@react-navigation/native';
import { createNativeStackNavigator } from '@react-navigation/native-stack';

import { AuthStack } from './(auth)/_layout';
import { CustomerTabNavigator } from './(customer)/_layout';
import { DriverTabNavigator } from './(driver)/_layout';
import { MerchantTabNavigator } from './(merchant)/_layout';
import { AuthProvider } from '../context/auth-context';
import { useAuthContext } from '../hooks/useAuthContext';
import { colors } from '../design-system/colors';

const Stack = createNativeStackNavigator();

function RootNavigation() {
  const { isAuthenticated, isLoading, role } = useAuthContext();

  if (isLoading) {
    return (
      <View style={styles.loadingContainer}>
        <ActivityIndicator size="large" color={colors.highlight} />
        <Text style={styles.loadingText}>Preparing GoFleet...</Text>
      </View>
    );
  }

  return (
    <NavigationContainer>
      <Stack.Navigator screenOptions={{ headerShown: false }}>
        {!isAuthenticated ? (
          <Stack.Screen name="Auth" component={AuthStack} />
        ) : role === 'driver' ? (
          <Stack.Screen name="Driver" component={DriverTabNavigator} />
        ) : role === 'merchant' ? (
          <Stack.Screen name="Merchant" component={MerchantTabNavigator} />
        ) : (
          <Stack.Screen name="Customer" component={CustomerTabNavigator} />
        )}
      </Stack.Navigator>
    </NavigationContainer>
  );
}

export function RootNavigator() {
  return (
    <AuthProvider>
      <RootNavigation />
    </AuthProvider>
  );
}

const styles = StyleSheet.create({
  loadingContainer: {
    flex: 1,
    alignItems: 'center',
    justifyContent: 'center',
    backgroundColor: colors.background,
  },
  loadingText: {
    marginTop: 12,
    color: colors.gray600,
    fontSize: 16,
  },
});

export default RootNavigator;
