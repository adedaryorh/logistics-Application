import { useEffect, useMemo, useState } from 'react';
import { ActivityIndicator, Pressable, SafeAreaView, StyleSheet, Text, View } from 'react-native';
import { StatusBar } from 'expo-status-bar';
import { Ionicons } from '@expo/vector-icons';
import * as Haptics from 'expo-haptics';
import { SafeAreaProvider } from 'react-native-safe-area-context';
import { api } from './src/api';
import { demoUser } from './src/data/demo';
import { useOrders } from './src/hooks/useOrders';
import { useNotifications } from './src/hooks/useNotifications';
import { useWallet } from './src/hooks/useWallet';
import { AccessScreen } from './src/screens/AccessScreen';
import { BookingModal } from './src/screens/BookingModal';
import { HomeScreen } from './src/screens/HomeScreen';
import { OrdersScreen } from './src/screens/OrdersScreen';
import { ProfileScreen } from './src/screens/ProfileScreen';
import { TrackingModal } from './src/screens/TrackingModal';
import { WalletScreen } from './src/screens/WalletScreen';
import { colors, shadow } from './src/theme';
import { DeliveryType, Order, Tab, User } from './src/types';

const tabs: {
  key: Tab;
  icon: keyof typeof Ionicons.glyphMap;
  activeIcon: keyof typeof Ionicons.glyphMap;
  label: string;
}[] = [
  { key: 'home', icon: 'home-outline', activeIcon: 'home', label: 'Home' },
  { key: 'orders', icon: 'receipt-outline', activeIcon: 'receipt', label: 'Orders' },
  { key: 'wallet', icon: 'wallet-outline', activeIcon: 'wallet', label: 'Wallet' },
  { key: 'profile', icon: 'person-outline', activeIcon: 'person', label: 'Profile' },
];

export default function App() {
  const [entered, setEntered] = useState(false);
  const [restoring, setRestoring] = useState(true);
  const [demoMode, setDemoMode] = useState(false);
  const [user, setUser] = useState<User>(demoUser);
  const [tab, setTab] = useState<Tab>('home');
  const [bookingType, setBookingType] = useState<DeliveryType>();
  const [trackingOrder, setTrackingOrder] = useState<Order>();
  const { orders, loading, error, refresh, setOrders } = useOrders(demoMode || !entered);
  const walletState = useWallet(entered, demoMode);
  useNotifications(entered && !demoMode);

  useEffect(() => {
    api
      .hasSession()
      .then((hasSession) => setEntered(hasSession))
      .finally(() => setRestoring(false));
  }, []);

  useEffect(() => {
    if (!entered || demoMode) return;
    api
      .me()
      .then(setUser)
      .catch(() => undefined);
  }, [entered, demoMode]);

  const currentScreen = useMemo(() => {
    if (tab === 'home')
      return (
        <HomeScreen
          user={user}
          orders={orders}
          onBook={setBookingType}
          onTrack={setTrackingOrder}
        />
      );
    if (tab === 'orders')
      return (
        <OrdersScreen
          orders={orders}
          loading={loading}
          error={error}
          onRefresh={refresh}
          onTrack={setTrackingOrder}
          onBook={() => setBookingType('parcel')}
        />
      );
    if (tab === 'wallet')
      return (
        <WalletScreen
          wallet={walletState.wallet}
          ledger={walletState.ledger}
          loading={walletState.loading}
          error={walletState.error}
          onRefresh={walletState.refresh}
        />
      );
    return <ProfileScreen user={user} demoMode={demoMode} onLogout={logout} />;
  }, [tab, user, orders, loading, error, refresh, demoMode, walletState]);

  function enterDemo() {
    setDemoMode(true);
    setUser(demoUser);
    setEntered(true);
  }
  function enterAuthenticated() {
    setDemoMode(false);
    setEntered(true);
  }
  async function logout() {
    if (!demoMode) await api.logout();
    setEntered(false);
    setDemoMode(false);
    setTab('home');
  }
  function orderCreated(order: Order) {
    setOrders((current) => [order, ...current]);
    setBookingType(undefined);
    setTrackingOrder(order);
  }

  if (restoring)
    return (
      <SafeAreaProvider>
        <View style={s.restoring}>
          <ActivityIndicator size="large" color={colors.green} />
        </View>
      </SafeAreaProvider>
    );
  return (
    <SafeAreaProvider>
      {!entered ? (
        <AccessScreen onAuthenticated={enterAuthenticated} onDemo={enterDemo} />
      ) : (
        <SafeAreaView style={s.app}>
          <StatusBar style="dark" />
          {currentScreen}
          <View style={s.tabs}>
            {tabs.map((item) => {
              const active = tab === item.key;
              return (
                <Pressable
                  key={item.key}
                  onPress={() => {
                    setTab(item.key);
                    Haptics.selectionAsync();
                  }}
                  style={s.tab}
                >
                  <View style={active && s.activeIcon}>
                    <Ionicons
                      name={active ? item.activeIcon : item.icon}
                      size={22}
                      color={active ? colors.greenDark : '#89938E'}
                    />
                  </View>
                  <Text style={[s.label, active && s.activeLabel]}>{item.label}</Text>
                </Pressable>
              );
            })}
          </View>
          <BookingModal
            visible={!!bookingType}
            initialType={bookingType ?? 'parcel'}
            demoMode={demoMode}
            onClose={() => setBookingType(undefined)}
            onCreated={orderCreated}
          />
          <TrackingModal
            order={trackingOrder}
            demoMode={demoMode}
            onClose={() => setTrackingOrder(undefined)}
          />
        </SafeAreaView>
      )}
    </SafeAreaProvider>
  );
}

const s = StyleSheet.create({
  restoring: {
    flex: 1,
    alignItems: 'center',
    justifyContent: 'center',
    backgroundColor: colors.canvas,
  },
  app: { flex: 1, backgroundColor: colors.canvas },
  tabs: {
    position: 'absolute',
    left: 14,
    right: 14,
    bottom: 10,
    height: 76,
    backgroundColor: 'white',
    borderRadius: 24,
    flexDirection: 'row',
    alignItems: 'center',
    paddingHorizontal: 8,
    borderWidth: 1,
    borderColor: colors.line,
    ...shadow,
  },
  tab: { flex: 1, alignItems: 'center', gap: 4 },
  activeIcon: {
    width: 38,
    height: 32,
    borderRadius: 12,
    backgroundColor: colors.greenSoft,
    alignItems: 'center',
    justifyContent: 'center',
  },
  label: { fontSize: 10, fontWeight: '700', color: '#89938E' },
  activeLabel: { color: colors.greenDark },
});
