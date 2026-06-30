import React from 'react';
import { ScrollView, StyleSheet, Text, View } from 'react-native';

import Badge from '@/design-system/components/Badge';
import ETABadge from '@/design-system/components/ETABadge';
import { colors } from '@/design-system/colors';

const dispatchOffers = [
  { route: 'Lekki Phase 1 to Victoria Island', payout: 'NGN 6,400', eta: 4 },
  { route: 'Ikeja GRA to Maryland Mall', payout: 'NGN 4,100', eta: 6 },
];

const driverStats = [
  { value: 'NGN 86k', label: 'Today earnings' },
  { value: '4.94', label: 'Rating' },
  { value: '92%', label: 'Acceptance' },
];

export function HomeScreen() {
  return (
    <ScrollView contentContainerStyle={styles.container}>
      <View style={styles.heroCard}>
        <View style={styles.heroTopRow}>
          <View style={styles.heroCopy}>
            <Text style={styles.eyebrow}>Driver dashboard</Text>
            <Text style={styles.title}>Stay online, win better trips, and keep your numbers visible.</Text>
          </View>
          <Badge status="online" />
        </View>
        <View style={styles.metricRow}>
          {driverStats.map((item) => (
            <View key={item.label} style={styles.metricCard}>
              <Text style={styles.metricValue}>{item.value}</Text>
              <Text style={styles.metricLabel}>{item.label}</Text>
            </View>
          ))}
        </View>
      </View>

      <View style={styles.section}>
        <Text style={styles.sectionTitle}>Nearby dispatch offers</Text>
        {dispatchOffers.map((offer, index) => (
          <View key={offer.route} style={[styles.offerCard, index === 0 ? styles.featuredOfferCard : undefined]}>
            <View style={styles.offerTopRow}>
              <Text style={styles.offerRoute}>{offer.route}</Text>
              <ETABadge minutes={offer.eta} type="ride" pulse={index === 0} />
            </View>
            <View style={styles.offerFooter}>
              <Text style={styles.offerPayout}>{offer.payout}</Text>
              <Text style={styles.offerMeta}>Pickup bonus included</Text>
            </View>
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
    backgroundColor: colors.secondary,
    borderRadius: 28,
    padding: 22,
    gap: 16,
  },
  heroTopRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    gap: 16,
  },
  heroCopy: {
    flex: 1,
    gap: 10,
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
  offerCard: {
    backgroundColor: colors.card,
    borderRadius: 24,
    padding: 18,
    gap: 14,
    borderWidth: 1,
    borderColor: colors.gray200,
  },
  featuredOfferCard: {
    borderColor: colors.highlight,
  },
  offerTopRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    gap: 12,
  },
  offerRoute: {
    flex: 1,
    color: colors.primary,
    fontSize: 16,
    lineHeight: 22,
    fontWeight: '800',
  },
  offerFooter: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    gap: 10,
  },
  offerPayout: {
    color: colors.primary,
    fontSize: 19,
    fontWeight: '800',
  },
  offerMeta: {
    color: colors.gray600,
    fontSize: 12,
    fontWeight: '700',
  },
});

export default HomeScreen;
