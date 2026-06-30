import React from 'react';
import { ScrollView, StyleSheet, Text, View } from 'react-native';

import ETABadge from '@/design-system/components/ETABadge';
import { colors } from '@/design-system/colors';

const breakdown = [
  { label: 'Service type', value: 'Parcel delivery' },
  { label: 'Pickup', value: 'Sabo Yaba, Lagos' },
  { label: 'Dropoff', value: 'Chevron Drive, Lekki' },
  { label: 'Amount', value: 'NGN 7,400' },
];

export function OrderDetailsScreen() {
  return (
    <ScrollView contentContainerStyle={styles.container}>
      <View style={styles.heroCard}>
        <Text style={styles.eyebrow}>Order details</Text>
        <Text style={styles.title}>Inspect one order deeply with fulfillment, route, and pricing context.</Text>
        <ETABadge minutes={12} type="parcel" />
      </View>

      {breakdown.map((item) => (
        <View key={item.label} style={styles.detailCard}>
          <Text style={styles.detailLabel}>{item.label}</Text>
          <Text style={styles.detailValue}>{item.value}</Text>
        </View>
      ))}
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: { flexGrow: 1, padding: 20, backgroundColor: colors.background, gap: 18 },
  heroCard: { backgroundColor: colors.secondary, borderRadius: 28, padding: 22, gap: 12 },
  eyebrow: { color: colors.highlight, fontSize: 13, fontWeight: '800', textTransform: 'uppercase', letterSpacing: 1.2 },
  title: { color: colors.white, fontSize: 28, lineHeight: 35, fontWeight: '800' },
  detailCard: { backgroundColor: colors.card, borderRadius: 24, padding: 18, borderWidth: 1, borderColor: colors.gray200, gap: 6 },
  detailLabel: { color: colors.gray500, fontSize: 12, textTransform: 'uppercase', letterSpacing: 0.8 },
  detailValue: { color: colors.primary, fontSize: 16, lineHeight: 22, fontWeight: '800' },
});

export default OrderDetailsScreen;
