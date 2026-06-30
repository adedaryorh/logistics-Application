import React from 'react';
import { ScrollView, StyleSheet, Text, View } from 'react-native';

import Avatar from '@/design-system/components/Avatar';
import Badge from '@/design-system/components/Badge';
import { colors } from '@/design-system/colors';

const profileItems = [
  { label: 'Vehicle', value: 'Toyota Corolla 2020' },
  { label: 'Plate number', value: 'APP 482 KR' },
  { label: 'Preferred zone', value: 'Victoria Island and Lekki' },
];

export function ProfileScreen() {
  return (
    <ScrollView contentContainerStyle={styles.container}>
      <View style={styles.heroCard}>
        <Avatar name="Amina Yusuf" size="xl" isOnline />
        <View style={styles.heroCopy}>
          <Text style={styles.name}>Amina Yusuf</Text>
          <Text style={styles.subtitle}>Top-performing driver with backend-linked identity and trip stats.</Text>
        </View>
        <Badge status="online" />
      </View>

      {profileItems.map((item) => (
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
  heroCard: { backgroundColor: colors.primary, borderRadius: 28, padding: 22, gap: 16, alignItems: 'flex-start' },
  heroCopy: { gap: 6 },
  name: { color: colors.white, fontSize: 28, fontWeight: '800' },
  subtitle: { color: colors.gray300, fontSize: 15, lineHeight: 22 },
  detailCard: { backgroundColor: colors.card, borderRadius: 24, padding: 18, borderWidth: 1, borderColor: colors.gray200, gap: 6 },
  detailLabel: { color: colors.gray500, fontSize: 12, textTransform: 'uppercase', letterSpacing: 0.8 },
  detailValue: { color: colors.primary, fontSize: 16, fontWeight: '800' },
});

export default ProfileScreen;
