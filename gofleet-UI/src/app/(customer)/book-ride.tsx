import React from 'react';
import { ScrollView, StyleSheet, Text, View } from 'react-native';

import Button from '@/design-system/components/Button';
import ETABadge from '@/design-system/components/ETABadge';
import { colors } from '@/design-system/colors';

const rideOptions = [
  { name: 'GoMini', eta: '3 min', price: 'NGN 4,800', note: 'Affordable solo rides' },
  { name: 'GoComfort', eta: '5 min', price: 'NGN 7,200', note: 'Extra space and air-conditioned comfort' },
  { name: 'GoXL', eta: '7 min', price: 'NGN 9,600', note: 'Best for groups, luggage, and airport runs' },
];

export function BookRideScreen() {
  return (
    <ScrollView contentContainerStyle={styles.container}>
      <View style={styles.heroCard}>
        <Text style={styles.eyebrow}>Ride booking</Text>
        <Text style={styles.title}>Choose a car class, confirm your route, and dispatch instantly.</Text>
        <Text style={styles.subtitle}>
          This view now mirrors a real ride-booking flow with pickup context, fare comparison, and checkout readiness.
        </Text>
      </View>

      <View style={styles.routeCard}>
        <Text style={styles.sectionTitle}>Route</Text>
        <View style={styles.routeRow}>
          <View style={[styles.routeDot, styles.pickupDot]} />
          <View style={styles.routeCopy}>
            <Text style={styles.routeLabel}>Pickup</Text>
            <Text style={styles.routeValue}>Civic Towers, Ozumba Mbadiwe</Text>
          </View>
        </View>
        <View style={styles.routeDivider} />
        <View style={styles.routeRow}>
          <View style={[styles.routeDot, styles.dropoffDot]} />
          <View style={styles.routeCopy}>
            <Text style={styles.routeLabel}>Dropoff</Text>
            <Text style={styles.routeValue}>MM2 Domestic Terminal, Ikeja</Text>
          </View>
        </View>
      </View>

      <View style={styles.sectionBlock}>
        <Text style={styles.sectionTitle}>Available rides</Text>
        {rideOptions.map((option, index) => (
          <View key={option.name} style={[styles.optionCard, index === 0 ? styles.optionCardFeatured : undefined]}>
            <View style={styles.optionTopRow}>
              <View>
                <Text style={styles.optionName}>{option.name}</Text>
                <Text style={styles.optionNote}>{option.note}</Text>
              </View>
              <ETABadge minutes={option.eta} pulse={index === 0} />
            </View>
            <View style={styles.optionFooter}>
              <Text style={styles.optionPrice}>{option.price}</Text>
              {index === 0 ? <Text style={styles.featuredLabel}>Best pickup speed</Text> : null}
            </View>
          </View>
        ))}
      </View>

      <View style={styles.checkoutCard}>
        <Text style={styles.checkoutTitle}>Payment</Text>
        <Text style={styles.checkoutBody}>Wallet balance NGN 12,300. Card fallback is ready for provider checkout when needed.</Text>
        <Button label="Confirm GoMini ride" size="full" />
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
    fontSize: 28,
    lineHeight: 35,
    fontWeight: '800',
  },
  subtitle: {
    color: colors.gray300,
    fontSize: 15,
    lineHeight: 22,
  },
  routeCard: {
    backgroundColor: colors.card,
    borderRadius: 24,
    padding: 20,
    borderWidth: 1,
    borderColor: colors.gray200,
    gap: 14,
  },
  sectionBlock: {
    gap: 12,
  },
  sectionTitle: {
    color: colors.primary,
    fontSize: 18,
    fontWeight: '800',
  },
  routeRow: {
    flexDirection: 'row',
    gap: 12,
    alignItems: 'center',
  },
  routeDot: {
    width: 12,
    height: 12,
    borderRadius: 999,
  },
  pickupDot: {
    backgroundColor: colors.pickupMarker,
  },
  dropoffDot: {
    backgroundColor: colors.dropoffMarker,
  },
  routeCopy: {
    flex: 1,
    gap: 2,
  },
  routeLabel: {
    color: colors.gray500,
    fontSize: 12,
    textTransform: 'uppercase',
    letterSpacing: 0.8,
  },
  routeValue: {
    color: colors.primary,
    fontSize: 15,
    fontWeight: '700',
  },
  routeDivider: {
    height: 1,
    backgroundColor: colors.gray200,
    marginLeft: 6,
  },
  optionCard: {
    backgroundColor: colors.card,
    borderRadius: 24,
    padding: 18,
    borderWidth: 1,
    borderColor: colors.gray200,
    gap: 14,
  },
  optionCardFeatured: {
    borderColor: colors.highlight,
    shadowColor: colors.highlight,
    shadowOpacity: 0.08,
    shadowRadius: 16,
    shadowOffset: { width: 0, height: 8 },
  },
  optionTopRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    gap: 12,
  },
  optionName: {
    color: colors.primary,
    fontSize: 17,
    fontWeight: '800',
  },
  optionNote: {
    marginTop: 4,
    color: colors.gray600,
    fontSize: 13,
    lineHeight: 19,
    maxWidth: 220,
  },
  optionFooter: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    gap: 12,
  },
  optionPrice: {
    color: colors.primary,
    fontSize: 19,
    fontWeight: '800',
  },
  featuredLabel: {
    color: colors.highlight,
    fontSize: 12,
    fontWeight: '700',
  },
  checkoutCard: {
    backgroundColor: colors.primary,
    borderRadius: 24,
    padding: 20,
    gap: 12,
    marginBottom: 8,
  },
  checkoutTitle: {
    color: colors.white,
    fontSize: 18,
    fontWeight: '800',
  },
  checkoutBody: {
    color: colors.gray300,
    fontSize: 14,
    lineHeight: 21,
  },
});

export default BookRideScreen;
