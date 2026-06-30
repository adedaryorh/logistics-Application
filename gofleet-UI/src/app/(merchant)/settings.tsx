import React from 'react';
import { ScrollView, StyleSheet, Text, View } from 'react-native';

import Badge from '@/design-system/components/Badge';
import { colors } from '@/design-system/colors';

const settingsRows = [
  { title: 'Store visibility', value: 'Visible to customers', status: 'active' as const },
  { title: 'Auto-accept small orders', value: 'Enabled for lunch rush', status: 'completed' as const },
  { title: 'Payout account review', value: 'Needs finance confirmation', status: 'pending' as const },
];

export function SettingsScreen() {
  return (
    <ScrollView contentContainerStyle={styles.container}>
      <View style={styles.heroCard}>
        <Text style={styles.eyebrow}>Merchant settings</Text>
        <Text style={styles.title}>Control fulfillment defaults, store visibility, and payout readiness with fewer taps.</Text>
      </View>

      {settingsRows.map((item) => (
        <View key={item.title} style={styles.settingsCard}>
          <View style={styles.settingsCopy}>
            <Text style={styles.settingsTitle}>{item.title}</Text>
            <Text style={styles.settingsValue}>{item.value}</Text>
          </View>
          <Badge status={item.status} />
        </View>
      ))}
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: { flexGrow: 1, padding: 20, backgroundColor: colors.background, gap: 18 },
  heroCard: { backgroundColor: colors.secondary, borderRadius: 28, padding: 22, gap: 10 },
  eyebrow: { color: colors.highlight, fontSize: 13, fontWeight: '800', textTransform: 'uppercase', letterSpacing: 1.2 },
  title: { color: colors.white, fontSize: 28, lineHeight: 35, fontWeight: '800' },
  settingsCard: { flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between', gap: 12, backgroundColor: colors.card, borderRadius: 24, padding: 18, borderWidth: 1, borderColor: colors.gray200 },
  settingsCopy: { flex: 1, gap: 4 },
  settingsTitle: { color: colors.primary, fontSize: 16, fontWeight: '800' },
  settingsValue: { color: colors.gray600, fontSize: 14, lineHeight: 20 },
});

export default SettingsScreen;
