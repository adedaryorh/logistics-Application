import React from 'react';
import { ScrollView, StyleSheet, Text, View } from 'react-native';

import Badge from '@/design-system/components/Badge';
import { colors } from '@/design-system/colors';

const revenueCards = [
  { value: 'NGN 248k', label: 'Today sales' },
  { value: '31', label: 'Orders accepted' },
  { value: '14 min', label: 'Avg prep time' },
];

const queueStates = [
  { title: 'New orders', count: '5', status: 'pending' as const },
  { title: 'In kitchen', count: '8', status: 'active' as const },
  { title: 'Ready for pickup', count: '3', status: 'completed' as const },
];

export function DashboardScreen() {
  return (
    <ScrollView contentContainerStyle={styles.container}>
      <View style={styles.heroCard}>
        <Text style={styles.eyebrow}>Merchant control</Text>
        <Text style={styles.title}>Watch your prep line, order queue, and revenue from one store cockpit.</Text>
        <Text style={styles.subtitle}>
          The merchant app now feels like an operations dashboard with clearer service levels and store health signals.
        </Text>
        <View style={styles.metricRow}>
          {revenueCards.map((item) => (
            <View key={item.label} style={styles.metricCard}>
              <Text style={styles.metricValue}>{item.value}</Text>
              <Text style={styles.metricLabel}>{item.label}</Text>
            </View>
          ))}
        </View>
      </View>

      <View style={styles.section}>
        <Text style={styles.sectionTitle}>Kitchen queue</Text>
        {queueStates.map((item) => (
          <View key={item.title} style={styles.queueCard}>
            <View style={styles.queueCopy}>
              <Text style={styles.queueTitle}>{item.title}</Text>
              <Text style={styles.queueCount}>{item.count} orders</Text>
            </View>
            <Badge status={item.status} label={item.title} />
          </View>
        ))}
      </View>
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: {
    flexGrow: 1,
    padding: 20,
    backgroundColor: colors.background,
    gap: 18,
  },
  heroCard: {
    backgroundColor: colors.primary,
    borderRadius: 28,
    padding: 22,
    gap: 12,
  },
  eyebrow: {
    color: colors.highlight,
    fontSize: 13,
    fontWeight: '800',
    textTransform: 'uppercase',
    letterSpacing: 1.2,
  },
  title: {
    color: colors.white,
    fontSize: 29,
    lineHeight: 36,
    fontWeight: '800',
  },
  subtitle: {
    color: colors.gray300,
    fontSize: 15,
    lineHeight: 22,
  },
  metricRow: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: 12,
  },
  metricCard: {
    flex: 1,
    minWidth: 96,
    backgroundColor: 'rgba(255,255,255,0.08)',
    borderRadius: 18,
    padding: 14,
  },
  metricValue: {
    color: colors.white,
    fontSize: 18,
    fontWeight: '800',
  },
  metricLabel: {
    color: colors.gray300,
    fontSize: 12,
    marginTop: 4,
  },
  section: {
    gap: 12,
  },
  sectionTitle: {
    color: colors.primary,
    fontSize: 18,
    fontWeight: '800',
  },
  queueCard: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    gap: 12,
    backgroundColor: colors.card,
    borderRadius: 24,
    padding: 18,
    borderWidth: 1,
    borderColor: colors.gray200,
  },
  queueCopy: {
    flex: 1,
    gap: 4,
  },
  queueTitle: {
    color: colors.primary,
    fontSize: 16,
    fontWeight: '800',
    textTransform: 'capitalize',
  },
  queueCount: {
    color: colors.gray600,
    fontSize: 14,
  },
});

export default DashboardScreen;
