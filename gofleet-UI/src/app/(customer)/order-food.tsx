import React from 'react';
import { ScrollView, StyleSheet, Text, View } from 'react-native';

import Button from '@/design-system/components/Button';
import ETABadge from '@/design-system/components/ETABadge';
import { colors } from '@/design-system/colors';

const cuisines = ['African', 'Burgers', 'Healthy', 'Desserts'];
const featuredMerchants = [
  { name: 'Citrus Grill', eta: 18, tag: 'Top rated', highlight: 'Buy 1 get 1 wraps' },
  { name: 'Mama B Kitchen', eta: 24, tag: 'Fast prep', highlight: 'Hot jollof bowls all afternoon' },
  { name: 'Bento Express', eta: 31, tag: 'New', highlight: 'Asian fusion lunch bundles' },
];

export function OrderFoodScreen() {
  return (
    <ScrollView contentContainerStyle={styles.container}>
      <View style={styles.heroCard}>
        <Text style={styles.eyebrow}>Food delivery</Text>
        <Text style={styles.title}>Discover merchants with live prep speed and checkout-ready menus.</Text>
        <Text style={styles.subtitle}>
          This screen now feels like a real marketplace instead of a scaffold, with cuisine discovery and merchant cards.
        </Text>
      </View>

      <View style={styles.cuisineRow}>
        {cuisines.map((item) => (
          <View key={item} style={styles.cuisineChip}>
            <Text style={styles.cuisineChipText}>{item}</Text>
          </View>
        ))}
      </View>

      <View style={styles.section}>
        <Text style={styles.sectionTitle}>Featured merchants</Text>
        {featuredMerchants.map((merchant, index) => (
          <View key={merchant.name} style={[styles.merchantCard, index === 0 ? styles.featuredMerchantCard : undefined]}>
            <View style={styles.merchantTopRow}>
              <View style={styles.merchantCopy}>
                <Text style={styles.merchantName}>{merchant.name}</Text>
                <Text style={styles.merchantTag}>{merchant.tag}</Text>
              </View>
              <ETABadge minutes={merchant.eta} type="food" pulse={index === 0} />
            </View>
            <Text style={styles.merchantHighlight}>{merchant.highlight}</Text>
          </View>
        ))}
      </View>

      <View style={styles.cartCard}>
        <Text style={styles.cartTitle}>Cart snapshot</Text>
        <Text style={styles.cartBody}>2 meals, 1 drink, delivery fee included. Total estimated checkout: NGN 14,200.</Text>
        <Button label="Open checkout" size="full" />
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
  cuisineRow: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: 10,
  },
  cuisineChip: {
    backgroundColor: colors.white,
    borderRadius: 999,
    paddingHorizontal: 14,
    paddingVertical: 10,
    borderWidth: 1,
    borderColor: colors.gray200,
  },
  cuisineChipText: {
    color: colors.primary,
    fontSize: 13,
    fontWeight: '700',
  },
  section: {
    gap: 12,
  },
  sectionTitle: {
    color: colors.primary,
    fontSize: 18,
    fontWeight: '800',
  },
  merchantCard: {
    backgroundColor: colors.card,
    borderRadius: 24,
    padding: 18,
    gap: 12,
    borderWidth: 1,
    borderColor: colors.gray200,
  },
  featuredMerchantCard: {
    borderColor: colors.highlight,
  },
  merchantTopRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    gap: 12,
  },
  merchantCopy: {
    flex: 1,
    gap: 4,
  },
  merchantName: {
    color: colors.primary,
    fontSize: 17,
    fontWeight: '800',
  },
  merchantTag: {
    color: colors.highlight,
    fontSize: 12,
    fontWeight: '700',
    textTransform: 'uppercase',
    letterSpacing: 0.8,
  },
  merchantHighlight: {
    color: colors.gray600,
    fontSize: 14,
    lineHeight: 21,
  },
  cartCard: {
    backgroundColor: colors.secondary,
    borderRadius: 24,
    padding: 20,
    gap: 12,
  },
  cartTitle: {
    color: colors.white,
    fontSize: 18,
    fontWeight: '800',
  },
  cartBody: {
    color: colors.gray300,
    fontSize: 14,
    lineHeight: 21,
  },
});

export default OrderFoodScreen;
