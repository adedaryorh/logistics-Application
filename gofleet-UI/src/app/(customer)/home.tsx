import React from 'react';
import { ScrollView, StyleSheet, Text, View } from 'react-native';

import OrderCard from '@/design-system/components/OrderCard';
import { colors } from '@/design-system/colors';

const quickActions = [
  { title: 'Book ride', detail: 'Airport pickup in under 4 min', tone: colors.info },
  { title: 'Order food', detail: '22 restaurants with live prep times', tone: colors.warning },
  { title: 'Send parcel', detail: 'Same-city express and scheduled drops', tone: colors.success },
];

const highlights = [
  { value: '4 min', label: 'Nearest driver ETA' },
  { value: '11', label: 'Fast delivery merchants' },
  { value: '2', label: 'Active orders today' },
];

export function HomeScreen() {
  return (
    <ScrollView contentContainerStyle={styles.container}>
      <View style={styles.heroCard}>
        <Text style={styles.eyebrow}>Customer command</Text>
        <Text style={styles.title}>Move around the city, eat well, and keep every drop visible.</Text>
        <Text style={styles.subtitle}>
          Your home screen now surfaces the three core logistics modes instead of a generic placeholder shell.
        </Text>
        <View style={styles.metricRow}>
          {highlights.map((item) => (
            <View key={item.label} style={styles.metricCard}>
              <Text style={styles.metricValue}>{item.value}</Text>
              <Text style={styles.metricLabel}>{item.label}</Text>
            </View>
          ))}
        </View>
      </View>

      <View style={styles.section}>
        <Text style={styles.sectionTitle}>Quick actions</Text>
        {quickActions.map((item) => (
          <View key={item.title} style={styles.quickActionCard}>
            <View style={[styles.actionDot, { backgroundColor: item.tone }]} />
            <View style={styles.actionCopy}>
              <Text style={styles.actionTitle}>{item.title}</Text>
              <Text style={styles.actionDetail}>{item.detail}</Text>
            </View>
          </View>
        ))}
      </View>

      <View style={styles.section}>
        <Text style={styles.sectionTitle}>Active order</Text>
        <OrderCard
          type="food"
          status="active"
          pickupAddress="Kiln House Kitchen, Victoria Island"
          dropoffAddress="12 Admiralty Way, Lekki Phase 1"
          priceFormatted="NGN 8,400"
        />
      </View>
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: {
    flexGrow: 1,
    padding: 20,
    backgroundColor: colors.background,
    gap: 20,
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
    letterSpacing: 1.2,
    textTransform: 'uppercase',
  },
  title: {
    color: colors.white,
    fontSize: 30,
    lineHeight: 37,
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
  quickActionCard: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 14,
    backgroundColor: colors.card,
    borderRadius: 22,
    padding: 18,
    borderWidth: 1,
    borderColor: colors.gray200,
  },
  actionDot: {
    width: 14,
    height: 14,
    borderRadius: 999,
  },
  actionCopy: {
    flex: 1,
    gap: 4,
  },
  actionTitle: {
    color: colors.primary,
    fontSize: 16,
    fontWeight: '700',
    textTransform: 'capitalize',
  },
  actionDetail: {
    color: colors.gray600,
    fontSize: 14,
    lineHeight: 20,
  },
});

export default HomeScreen;
