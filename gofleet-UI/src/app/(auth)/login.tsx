import React, { useState } from 'react';
import { ScrollView, StyleSheet, Text, View } from 'react-native';

import Button from '@/design-system/components/Button';
import Input from '@/design-system/components/Input';
import { toast } from '@/design-system/components/Toast';
import { colors } from '@/design-system/colors';
import { useAuth } from '@/hooks/useAuth';

type AuthNavigation = {
  navigate: (screen: string) => void;
};

type LoginScreenProps = {
  navigation: AuthNavigation;
};

const quickRoles = ['Customer', 'Driver', 'Merchant'];

export function LoginScreen({ navigation }: LoginScreenProps) {
  const { login, isLoading } = useAuth();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');

  const handleSubmit = async () => {
    try {
      await login({ email, password });
      toast.success('Welcome back');
    } catch {
      toast.error('Login failed', 'Check your backend connection and credentials.');
    }
  };

  return (
    <ScrollView contentContainerStyle={styles.container} keyboardShouldPersistTaps="handled">
      <View style={styles.hero}>
        <Text style={styles.eyebrow}>Sign in</Text>
        <Text style={styles.title}>Continue into the operations network.</Text>
        <Text style={styles.subtitle}>
          Use the same backend credentials for customer, driver, or merchant access and let role-based navigation take over.
        </Text>
        <View style={styles.roleRow}>
          {quickRoles.map((role) => (
            <View key={role} style={styles.roleChip}>
              <Text style={styles.roleChipText}>{role}</Text>
            </View>
          ))}
        </View>
      </View>

      <View style={styles.formCard}>
        <Input
          label="Email"
          value={email}
          onChangeText={setEmail}
          autoCapitalize="none"
          keyboardType="email-address"
          helperText="Use the account linked to your logistics role."
        />
        <Input
          label="Password"
          value={password}
          onChangeText={setPassword}
          secureTextEntry
          helperText="JWT tokens are stored securely in MMKV after sign-in."
        />
        <Button label="Sign in" size="full" loading={isLoading} onPress={handleSubmit} />
        <Button label="Create a new account" variant="ghost" size="full" onPress={() => navigation.navigate('Register')} />
      </View>
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: {
    flexGrow: 1,
    padding: 24,
    justifyContent: 'center',
    backgroundColor: colors.background,
    gap: 18,
  },
  hero: {
    backgroundColor: colors.primary,
    borderRadius: 28,
    padding: 24,
    gap: 12,
  },
  eyebrow: {
    color: colors.highlight,
    fontSize: 13,
    fontWeight: '800',
    textTransform: 'uppercase',
    letterSpacing: 1.3,
  },
  title: {
    color: colors.white,
    fontSize: 30,
    lineHeight: 36,
    fontWeight: '800',
  },
  subtitle: {
    color: colors.gray300,
    fontSize: 15,
    lineHeight: 22,
  },
  roleRow: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: 10,
    marginTop: 4,
  },
  roleChip: {
    backgroundColor: 'rgba(255,255,255,0.08)',
    borderRadius: 999,
    paddingHorizontal: 12,
    paddingVertical: 8,
  },
  roleChipText: {
    color: colors.white,
    fontSize: 12,
    fontWeight: '700',
  },
  formCard: {
    backgroundColor: colors.card,
    borderRadius: 24,
    padding: 20,
    gap: 16,
    borderWidth: 1,
    borderColor: colors.gray200,
  },
});

export default LoginScreen;
