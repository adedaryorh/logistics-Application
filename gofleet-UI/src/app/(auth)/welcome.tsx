import React from 'react';
import { ScrollView, StyleSheet, Text, View } from 'react-native';

import Button from '@/design-system/components/Button';
import { colors } from '@/design-system/colors';

type AuthNavigation = {
  navigate: (screen: string) => void;
};

type WelcomeScreenProps = {
  navigation: AuthNavigation;
};

const roleHighlights = [
  {
    title: 'Customer mode',
    description: 'Book rides, order food, and track parcels from one calm command center.',
  },
  {
    title: 'Driver mode',
    description: 'Accept offers faster, stay online confidently, and keep earnings visible.',
  },
  {
    title: 'Merchant mode',
    description: 'Manage prep queues, menu health, and store revenue without switching apps.',
  },
];

const trustStats = [
  { value: '3', label: 'Role flows' },
  { value: '24/7', label: 'Dispatch ready' },
  { value: 'Live', label: 'Tracking events' },
];

export function WelcomeScreen({ navigation }: WelcomeScreenProps) {
  return (
    <ScrollView contentContainerStyle={styles.container}>
      <View style={styles.heroCard}>
        <Text style={styles.eyebrow}>GoFleet Platform</Text>
        <Text style={styles.title}>One polished mobile surface for rides, food, and parcel logistics.</Text>
        <Text style={styles.subtitle}>
          Built for customers, drivers, and merchants on a single codebase with backend-ready auth, payments, and live operations.
        </Text>

        <View style={styles.trustRow}>
          {trustStats.map((item) => (
            <View key={item.label} style={styles.trustPill}>
              <Text style={styles.trustValue}>{item.value}</Text>
              <Text style={styles.trustLabel}>{item.label}</Text>
            </View>
          ))}
        </View>
      </View>

      <View style={styles.roleGrid}>
        {roleHighlights.map((item) => (
          <View key={item.title} style={styles.roleCard}>
            <Text style={styles.roleTitle}>{item.title}</Text>
            <Text style={styles.roleDescription}>{item.description}</Text>
          </View>
        ))}
      </View>

      <View style={styles.actionPanel}>
        <Button label="Create account" size="full" onPress={() => navigation.navigate('Register')} />
        <Button
          label="I already have an account"
          variant="outline"
          size="full"
          onPress={() => navigation.navigate('Login')}
        />
        <Text style={styles.helperText}>
          Continue with your backend account and switch roles after authentication.
        </Text>
      </View>
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: {
    flexGrow: 1,
    padding: 24,
    backgroundColor: colors.backgroundDark,
    gap: 20,
  },
  heroCard: {
    backgroundColor: colors.secondary,
    borderRadius: 28,
    padding: 24,
    borderWidth: 1,
    borderColor: 'rgba(255,255,255,0.08)',
    gap: 16,
    marginTop: 16,
  },
  eyebrow: {
    color: colors.highlight,
    fontSize: 13,
    fontWeight: '800',
    letterSpacing: 1.4,
    textTransform: 'uppercase',
  },
  title: {
    fontSize: 34,
    lineHeight: 42,
    fontWeight: '800',
    color: colors.white,
  },
  subtitle: {
    fontSize: 16,
    lineHeight: 24,
    color: colors.gray300,
  },
  trustRow: {
    flexDirection: 'row',
    gap: 12,
    flexWrap: 'wrap',
  },
  trustPill: {
    flex: 1,
    minWidth: 88,
    backgroundColor: 'rgba(233,69,96,0.12)',
    borderRadius: 18,
    paddingVertical: 14,
    paddingHorizontal: 12,
  },
  trustValue: {
    color: colors.white,
    fontSize: 18,
    fontWeight: '800',
  },
  trustLabel: {
    color: colors.gray300,
    fontSize: 12,
    marginTop: 4,
  },
  roleGrid: {
    gap: 12,
  },
  roleCard: {
    backgroundColor: 'rgba(255,255,255,0.04)',
    borderRadius: 22,
    padding: 18,
    borderWidth: 1,
    borderColor: 'rgba(255,255,255,0.06)',
  },
  roleTitle: {
    color: colors.white,
    fontSize: 17,
    fontWeight: '700',
    marginBottom: 6,
  },
  roleDescription: {
    color: colors.gray300,
    fontSize: 14,
    lineHeight: 21,
  },
  actionPanel: {
    marginTop: 'auto',
    gap: 12,
    paddingBottom: 12,
  },
  helperText: {
    color: colors.gray400,
    fontSize: 13,
    lineHeight: 20,
    textAlign: 'center',
  },
});

export default WelcomeScreen;
