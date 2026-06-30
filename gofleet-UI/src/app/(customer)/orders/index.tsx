import React from 'react';
import { ScrollView, StyleSheet, Text, View } from 'react-native';

import OrderCard from '@/design-system/components/OrderCard';
import { colors } from '@/design-system/colors';

export function OrdersScreen() {
  return (
    <ScrollView contentContainerStyle={styles.container}>
      <View style={styles.heroCard}>
        <Text style={styles.eyebrow}>Order history</Text>
        <Text style={styles.title}>Keep rides, meals, and parcels in one clean timeline instead of splitting history by service.</Text>
      </View>

      <OrderCard
        type="ride"
        status="completed"
        pickupAddress="The Palms, Lekki"
        dropoffAddress="Eko Hotel, Victoria Island"
        priceFormatted="NGN 5,100"
      />
      <OrderCard
        type="food"
        status="active"
        pickupAddress="Spice Route Kitchen"
        dropoffAddress="12 Fola Osibo, Lekki"
        priceFormatted="NGN 9,800"
      />
      <OrderCard
        type="parcel"
        status="pending"
        pickupAddress="Sabo Yaba"
        dropoffAddress="Chevron Drive"
        priceFormatted="NGN 7,400"
      />
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: { flexGrow: 1, padding: 20, backgroundColor: colors.background, gap: 18 },
  heroCard: { backgroundColor: colors.primary, borderRadius: 28, padding: 22, gap: 10 },
  eyebrow: { color: colors.highlight, fontSize: 13, fontWeight: '800', textTransform: 'uppercase', letterSpacing: 1.2 },
  title: { color: colors.white, fontSize: 28, lineHeight: 35, fontWeight: '800' },
});

export default OrdersScreen;
