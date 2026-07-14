import {
  ActivityIndicator,
  Pressable,
  RefreshControl,
  FlatList,
  StyleSheet,
  Text,
  View,
} from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import { IconButton, StateCard, StatusBadge, Notice, page } from '../components/AppUI';
import { Order } from '../types';
import { colors, radius } from '../theme';

export function OrdersScreen({
  orders,
  loading,
  error,
  onRefresh,
  onTrack,
  onBook,
}: {
  orders: Order[];
  loading: boolean;
  error?: string;
  onRefresh: () => void;
  onTrack: (o: Order) => void;
  onBook: () => void;
}) {
  return (
    <FlatList
      data={orders}
      keyExtractor={(item) => item.id}
      refreshControl={<RefreshControl refreshing={loading} onRefresh={onRefresh} />}
      contentContainerStyle={page.content}
      ListHeaderComponent={
        <>
          <View style={page.header}>
            <View>
              <Text accessibilityRole="header" style={page.title}>
                Delivery jobs
              </Text>
              <Text style={s.subtitle}>Agricultural loads and assigned work</Text>
            </View>
            <IconButton name="options-outline" label="Filter jobs" />
          </View>
          {error && orders.length ? (
            <Notice
              tone="warning"
              title="Showing saved jobs"
              copy="Live updates are unavailable. Pull down to retry."
            />
          ) : null}
        </>
      }
      ListEmptyComponent={
        loading ? (
          <ActivityIndicator size="large" color={colors.green} />
        ) : error && !orders.length ? (
          <StateCard
            icon="cloud-offline-outline"
            title="Couldn’t load orders"
            copy={error}
            action="Try again"
            onPress={onRefresh}
          />
        ) : !orders.length ? (
          <StateCard
            icon="cube-outline"
            title="No jobs assigned"
            copy="New opportunities and accepted agricultural jobs will appear here."
            action="Refresh jobs"
            onPress={onBook}
          />
        ) : null
      }
      renderItem={({ item: order }) => (
        <Pressable
          accessibilityRole="button"
          accessibilityLabel={`${order.agricultural_shipment?.produce_type ?? order.type} job, ${order.status.replaceAll('_', ' ')}`}
          accessibilityHint="Opens job route and details"
          onPress={() => onTrack(order)}
          style={({ pressed }) => [s.row, pressed && s.pressed]}
        >
          <View style={s.icon}>
            <Ionicons
              name={
                order.type === 'agricultural'
                  ? 'leaf-outline'
                  : order.type === 'parcel'
                    ? 'cube-outline'
                    : order.type === 'food'
                      ? 'fast-food-outline'
                      : 'car-outline'
              }
              size={21}
              color={colors.ink}
            />
          </View>
          <View style={{ flex: 1 }}>
            <Text style={s.name}>
              {order.agricultural_shipment?.produce_type ??
                `${order.type[0].toUpperCase() + order.type.slice(1)} delivery`}
            </Text>
            <Text style={s.meta}>
              {order.agricultural_shipment
                ? `${order.agricultural_shipment.quantity} ${order.agricultural_shipment.quantity_unit} • ${order.pickup.address ?? 'Pickup ready'}`
                : `${order.id} • ${new Date(order.created_at).toLocaleDateString()}`}
            </Text>
          </View>
          <View style={s.right}>
            <Text style={s.amount}>{money(order.price_minor, order.currency)}</Text>
            <StatusBadge status={order.status} />
          </View>
        </Pressable>
      )}
      ItemSeparatorComponent={() => <View style={{ height: 12 }} />}
    />
  );
}
const money = (minor: number, currency: string) =>
  new Intl.NumberFormat('en-NG', {
    style: 'currency',
    currency: currency || 'NGN',
    maximumFractionDigits: 0,
  }).format(minor / 100);
const s = StyleSheet.create({
  subtitle: { fontSize: 13, color: colors.muted, marginTop: 3 },
  pressed: { opacity: 0.78, transform: [{ scale: 0.99 }] },
  row: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 12,
    paddingVertical: 16,
    backgroundColor: 'white',
    borderWidth: 1,
    borderColor: colors.line,
    borderRadius: radius.lg,
    padding: 16,
  },
  icon: {
    width: 42,
    height: 42,
    borderRadius: 14,
    backgroundColor: '#F0F1EC',
    alignItems: 'center',
    justifyContent: 'center',
  },
  name: { fontSize: 14, fontWeight: '800', color: colors.ink },
  meta: { fontSize: 11, color: colors.muted, marginTop: 4 },
  right: { alignItems: 'flex-end' },
  amount: { fontSize: 13, fontWeight: '800', color: colors.ink },
});
