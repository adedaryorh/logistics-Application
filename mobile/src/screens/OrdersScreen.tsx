import {
  ActivityIndicator,
  Pressable,
  RefreshControl,
  ScrollView,
  StyleSheet,
  Text,
  View,
} from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import { IconButton, StateCard, page } from '../components/AppUI';
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
    <ScrollView
      refreshControl={<RefreshControl refreshing={loading} onRefresh={onRefresh} />}
      contentContainerStyle={page.content}
    >
      <View style={page.header}>
        <Text style={page.title}>Your orders</Text>
        <IconButton name="options-outline" />
      </View>
      {loading && !orders.length ? (
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
          title="No deliveries yet"
          copy="Your active and completed deliveries will appear here."
          action="Book delivery"
          onPress={onBook}
        />
      ) : (
        <View style={s.list}>
          {orders.map((order) => (
            <Pressable key={order.id} onPress={() => onTrack(order)} style={s.row}>
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
                  {order.type[0].toUpperCase() + order.type.slice(1)} delivery
                </Text>
                <Text style={s.meta}>
                  {order.id} • {new Date(order.created_at).toLocaleDateString()}
                </Text>
              </View>
              <View style={s.right}>
                <Text style={s.amount}>{money(order.price_minor, order.currency)}</Text>
                <Text
                  style={[
                    s.status,
                    { color: order.status === 'delivered' ? colors.greenDark : colors.muted },
                  ]}
                >
                  {order.status.replaceAll('_', ' ')}
                </Text>
              </View>
            </Pressable>
          ))}
        </View>
      )}
    </ScrollView>
  );
}
const money = (minor: number, currency: string) =>
  new Intl.NumberFormat('en-NG', {
    style: 'currency',
    currency: currency || 'NGN',
    maximumFractionDigits: 0,
  }).format(minor / 100);
const s = StyleSheet.create({
  list: {
    backgroundColor: 'white',
    borderRadius: radius.lg,
    paddingHorizontal: 16,
    borderWidth: 1,
    borderColor: colors.line,
  },
  row: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 12,
    paddingVertical: 16,
    borderBottomWidth: 1,
    borderBottomColor: colors.line,
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
  status: { fontSize: 10, fontWeight: '700', marginTop: 4, textTransform: 'capitalize' },
});
