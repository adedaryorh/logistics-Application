import React from 'react';
import { ScrollView, StyleSheet, Text, View } from 'react-native';

import DriverCard from '@/design-system/components/DriverCard';
import ETABadge from '@/design-system/components/ETABadge';
import { colors } from '@/design-system/colors';

const progressSteps = [
  { label: 'Order confirmed', status: 'done' },
  { label: 'Driver heading to pickup', status: 'active' },
  { label: 'Package in transit', status: 'pending' },
  { label: 'Delivered', status: 'pending' },
] as const;

export function TrackingScreen() {
  return (
    <ScrollView contentContainerStyle={styles.container}>
      <View style={styles.mapCard}>
        <Text style={styles.mapLabel}>Live route</Text>
        <Text style={styles.mapHeadline}>Driver assigned and moving toward your pickup point.</Text>
        <ETABadge minutes={6} type="parcel" />
        <View style={styles.routePreview}>
          <View style={styles.routeLine} />
          <View style={[styles.marker, styles.pickupMarker]} />
          <View style={[styles.marker, styles.driverMarker]} />
          <View style={[styles.marker, styles.dropoffMarker]} />
        </View>
      </View>

      <DriverCard
        name="Amina Yusuf"
        rating={4.9}
        vehicleModel="Toyota Hiace"
        licensePlate="LSD 248 KD"
        etaMinutes={6}
        showActions
      />

      <View style={styles.progressCard}>
        <Text style={styles.sectionTitle}>Delivery progress</Text>
        {progressSteps.map((step, index) => (
          <View key={step.label} style={styles.progressRow}>
            <View style={styles.progressRail}>
              <View
                style={[
                  styles.progressDot,
                  step.status === 'done' ? styles.progressDotDone : undefined,
                  step.status === 'active' ? styles.progressDotActive : undefined,
                ]}
              />
              {index < progressSteps.length - 1 ? <View style={styles.progressLine} /> : null}
            </View>
            <Text
              style={[
                styles.progressText,
                step.status === 'done' ? styles.progressTextDone : undefined,
                step.status === 'active' ? styles.progressTextActive : undefined,
              ]}
            >
              {step.label}
            </Text>
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
  mapCard: {
    backgroundColor: colors.primary,
    borderRadius: 28,
    padding: 22,
    gap: 12,
  },
  mapLabel: {
    color: colors.highlight,
    fontSize: 13,
    fontWeight: '800',
    textTransform: 'uppercase',
    letterSpacing: 1.2,
  },
  mapHeadline: {
    color: colors.white,
    fontSize: 26,
    lineHeight: 32,
    fontWeight: '800',
  },
  routePreview: {
    height: 180,
    borderRadius: 22,
    backgroundColor: colors.mapOverlay,
    marginTop: 8,
    overflow: 'hidden',
    justifyContent: 'center',
  },
  routeLine: {
    position: 'absolute',
    left: 42,
    right: 42,
    top: 88,
    height: 4,
    borderRadius: 999,
    backgroundColor: colors.routeLine,
  },
  marker: {
    position: 'absolute',
    width: 18,
    height: 18,
    borderRadius: 999,
    borderWidth: 3,
    borderColor: colors.white,
  },
  pickupMarker: {
    left: 32,
    top: 79,
    backgroundColor: colors.pickupMarker,
  },
  driverMarker: {
    left: '46%',
    top: 70,
    backgroundColor: colors.driverMarker,
  },
  dropoffMarker: {
    right: 32,
    top: 79,
    backgroundColor: colors.dropoffMarker,
  },
  progressCard: {
    backgroundColor: colors.card,
    borderRadius: 24,
    padding: 20,
    borderWidth: 1,
    borderColor: colors.gray200,
    gap: 12,
  },
  sectionTitle: {
    color: colors.primary,
    fontSize: 18,
    fontWeight: '800',
  },
  progressRow: {
    flexDirection: 'row',
    gap: 12,
  },
  progressRail: {
    alignItems: 'center',
  },
  progressDot: {
    width: 14,
    height: 14,
    borderRadius: 999,
    backgroundColor: colors.gray300,
  },
  progressDotDone: {
    backgroundColor: colors.success,
  },
  progressDotActive: {
    backgroundColor: colors.highlight,
  },
  progressLine: {
    width: 2,
    flex: 1,
    minHeight: 24,
    backgroundColor: colors.gray200,
    marginVertical: 4,
  },
  progressText: {
    flex: 1,
    color: colors.gray500,
    fontSize: 15,
    lineHeight: 22,
    paddingBottom: 14,
  },
  progressTextDone: {
    color: colors.primary,
    fontWeight: '700',
  },
  progressTextActive: {
    color: colors.highlight,
    fontWeight: '800',
  },
});

export default TrackingScreen;
