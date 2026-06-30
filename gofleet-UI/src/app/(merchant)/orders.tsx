import React from 'react';
import { ScrollView, StyleSheet, Text, View } from 'react-native';

import Badge from '@/design-system/components/Badge';
import { colors } from '@/design-system/colors';

const orders = [
  { customer: 'Tobi A.', items: '2 shawarmas, 1 drink', status: 'pending' as const },
  { customer: 'Chioma E.', items: 'Family rice pack', status: 'active' as const },
  { customer: 'David O.', items: 'Burger combo x2', status: 'completed' as const },
];

export function OrdersScreen() {
  return (
    <ScrollView contentContainerStyle={styles.container}>
      <View style={styles.heroCard}>
        <Text style={styles.eyebrow}>Merchant orders</Text>
        <Text style={styles.title}>See incoming food orders by state so kitchen and pickup teams stay aligned.</Text>
      </View>

      {orders.map((order) => (
        <View key={order.customer + order.items} style={styles.orderCard}>
          <View style={styles.orderCopy}>
            <Text style={styles.orderCustomer}>{order.customer}</Text>
            <Text style={styles.orderItems}>{order.items}</Text>
          </View>
          <Badge status={order.status} />
        </View>
      ))}
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: { flexGrow: 1, padding: 20, backgroundColor: colors.background, gap: 18 },
  heroCard: { backgroundColor: colors.secondary, borderRadius: 28, padding: 22, gap: 10 },
  eyebrow: { color: colors.highlight, fontSize: 13, fontWeight: '800', textTransform: 'uppercase', letterSpacing: 1.2 },
  title: { color: colors.white, fontSize: 28, lineHeight: 35, fontWeight: '800' },
  orderCard: { flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between', gap: 12, backgroundColor: colors.card, borderRadius: 24, padding: 18, borderWidth: 1, borderColor: colors.gray200 },
  orderCopy: { flex: 1, gap: 4 },
  orderCustomer: { color: colors.primary, fontSize: 16, fontWeight: '800' },
  orderItems: { color: colors.gray600, fontSize: 14, lineHeight: 20 },
});

export default OrdersScreen;
