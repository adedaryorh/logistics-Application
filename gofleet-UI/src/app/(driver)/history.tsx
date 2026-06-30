import React from 'react';
import { ScrollView, StyleSheet, Text, View } from 'react-native';

import OrderCard from '@/design-system/components/OrderCard';
import { colors } from '@/design-system/colors';

export function HistoryScreen() {
  return (
    <ScrollView contentContainerStyle={styles.container}>
      <View style={styles.heroCard}>
        <Text style={styles.eyebrow}>Trip history</Text>
        <Text style={styles.title}>Look back at completed rides and delivery work with cleaner payout context.</Text>
      </View>

      <OrderCard
        type="ride"
        status="completed"
        pickupAddress="Ikoyi Club Road"
        dropoffAddress="Admiralty Way, Lekki"
        priceFormatted="NGN 5,900"
      />
      <OrderCard
        type="parcel"
        status="completed"
        pickupAddress="Yaba Tech Hub"
        dropoffAddress="Chevron Drive Estate"
        priceFormatted="NGN 8,300"
      />
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: { flexGrow: 1, padding: 20, backgroundColor: colors.background, gap: 18 },
  heroCard: { backgroundColor: colors.secondary, borderRadius: 28, padding: 22, gap: 10 },
  eyebrow: { color: colors.highlight, fontSize: 13, fontWeight: '800', textTransform: 'uppercase', letterSpacing: 1.2 },
  title: { color: colors.white, fontSize: 28, lineHeight: 35, fontWeight: '800' },
});

export default HistoryScreen;
