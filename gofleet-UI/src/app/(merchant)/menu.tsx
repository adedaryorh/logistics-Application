import React from 'react';
import { ScrollView, StyleSheet, Text, View } from 'react-native';

import { colors } from '@/design-system/colors';

const menuItems = [
  { name: 'Signature Jollof Bowl', status: 'Best seller', price: 'NGN 5,500' },
  { name: 'Chicken Suya Wrap', status: 'Low stock alert', price: 'NGN 4,100' },
  { name: 'Plantain Dessert Box', status: 'High margin', price: 'NGN 3,200' },
];

export function MenuScreen() {
  return (
    <ScrollView contentContainerStyle={styles.container}>
      <View style={styles.heroCard}>
        <Text style={styles.eyebrow}>Menu manager</Text>
        <Text style={styles.title}>Keep availability, pricing, and featured dishes visible from your store app.</Text>
      </View>

      {menuItems.map((item) => (
        <View key={item.name} style={styles.menuCard}>
          <View style={styles.menuCopy}>
            <Text style={styles.menuName}>{item.name}</Text>
            <Text style={styles.menuStatus}>{item.status}</Text>
          </View>
          <Text style={styles.menuPrice}>{item.price}</Text>
        </View>
      ))}
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: { flexGrow: 1, padding: 20, backgroundColor: colors.background, gap: 18 },
  heroCard: { backgroundColor: colors.primary, borderRadius: 28, padding: 22, gap: 10 },
  eyebrow: { color: colors.highlight, fontSize: 13, fontWeight: '800', textTransform: 'uppercase', letterSpacing: 1.2 },
  title: { color: colors.white, fontSize: 28, lineHeight: 35, fontWeight: '800' },
  menuCard: { flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between', gap: 12, backgroundColor: colors.card, borderRadius: 24, padding: 18, borderWidth: 1, borderColor: colors.gray200 },
  menuCopy: { flex: 1, gap: 4 },
  menuName: { color: colors.primary, fontSize: 16, fontWeight: '800' },
  menuStatus: { color: colors.gray600, fontSize: 14 },
  menuPrice: { color: colors.highlight, fontSize: 16, fontWeight: '800' },
});

export default MenuScreen;
