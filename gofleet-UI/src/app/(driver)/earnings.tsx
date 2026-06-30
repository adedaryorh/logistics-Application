import React from 'react';
import { ScrollView, StyleSheet, Text, View } from 'react-native';

import { colors } from '@/design-system/colors';

const summaries = [
  { value: 'NGN 86,400', label: 'Today' },
  { value: 'NGN 412,000', label: 'This week' },
  { value: 'NGN 1.28m', label: 'This month' },
];

const payouts = [
  { title: 'Wallet balance', value: 'NGN 118,200', tone: colors.success },
  { title: 'Pending settlement', value: 'NGN 26,500', tone: colors.warning },
];

export function EarningsScreen() {
  return (
    <ScrollView contentContainerStyle={styles.container}>
      <View style={styles.heroCard}>
        <Text style={styles.eyebrow}>Driver earnings</Text>
        <Text style={styles.title}>Watch revenue, wallet balance, and settlement timing without leaving the trip flow.</Text>
        <View style={styles.metricRow}>
          {summaries.map((item) => (
            <View key={item.label} style={styles.metricCard}>
              <Text style={styles.metricValue}>{item.value}</Text>
              <Text style={styles.metricLabel}>{item.label}</Text>
            </View>
          ))}
        </View>
      </View>

      {payouts.map((item) => (
        <View key={item.title} style={styles.summaryCard}>
          <View style={[styles.summaryDot, { backgroundColor: item.tone }]} />
          <View style={styles.summaryCopy}>
            <Text style={styles.summaryTitle}>{item.title}</Text>
            <Text style={styles.summaryValue}>{item.value}</Text>
          </View>
        </View>
      ))}
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: { flexGrow: 1, padding: 20, backgroundColor: colors.background, gap: 18 },
  heroCard: { backgroundColor: colors.primary, borderRadius: 28, padding: 22, gap: 16 },
  eyebrow: { color: colors.highlight, fontSize: 13, fontWeight: '800', textTransform: 'uppercase', letterSpacing: 1.2 },
  title: { color: colors.white, fontSize: 28, lineHeight: 35, fontWeight: '800' },
  metricRow: { flexDirection: 'row', flexWrap: 'wrap', gap: 12 },
  metricCard: { flex: 1, minWidth: 96, backgroundColor: 'rgba(255,255,255,0.08)', borderRadius: 18, padding: 14 },
  metricValue: { color: colors.white, fontSize: 18, fontWeight: '800' },
  metricLabel: { color: colors.gray300, fontSize: 12, marginTop: 4 },
  summaryCard: { flexDirection: 'row', alignItems: 'center', gap: 14, backgroundColor: colors.card, borderRadius: 24, padding: 18, borderWidth: 1, borderColor: colors.gray200 },
  summaryDot: { width: 14, height: 14, borderRadius: 999 },
  summaryCopy: { gap: 4 },
  summaryTitle: { color: colors.gray600, fontSize: 13, textTransform: 'uppercase', letterSpacing: 0.8 },
  summaryValue: { color: colors.primary, fontSize: 20, fontWeight: '800' },
});

export default EarningsScreen;
