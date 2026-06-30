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

type RegisterScreenProps = {
  navigation: AuthNavigation;
};

const defaultRoles = ['Customer account', 'Driver onboarding', 'Merchant onboarding'];

export function RegisterScreen({ navigation }: RegisterScreenProps) {
  const { register, isLoading } = useAuth();
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');

  const handleSubmit = async () => {
    try {
      await register({ name, email, password, role: 'customer' });
      toast.success('Account created');
    } catch {
      toast.error('Registration failed', 'Please verify the backend auth routes are available.');
    }
  };

  return (
    <ScrollView contentContainerStyle={styles.container} keyboardShouldPersistTaps="handled">
      <View style={styles.hero}>
        <Text style={styles.eyebrow}>Create account</Text>
        <Text style={styles.title}>Start with a secure identity and grow into every logistics flow.</Text>
        <Text style={styles.subtitle}>
          Registration starts as a customer profile, then your backend roles and permissions can expand after approval.
        </Text>
        <View style={styles.roleList}>
          {defaultRoles.map((item) => (
            <Text key={item} style={styles.roleItem}>{'• '}{item}</Text>
          ))}
        </View>
      </View>

      <View style={styles.formCard}>
        <Input label="Full name" value={name} onChangeText={setName} helperText="Shown on driver, merchant, or customer profiles." />
        <Input label="Email" value={email} onChangeText={setEmail} autoCapitalize="none" keyboardType="email-address" />
        <Input label="Password" value={password} onChangeText={setPassword} secureTextEntry helperText="Use a strong password before biometric sign-in is added." />
        <Button label="Create account" size="full" loading={isLoading} onPress={handleSubmit} />
        <Button label="Already have an account" variant="ghost" size="full" onPress={() => navigation.navigate('Login')} />
      </View>
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: {
    flexGrow: 1,
    padding: 24,
    backgroundColor: colors.background,
    gap: 18,
    justifyContent: 'center',
  },
  hero: {
    backgroundColor: colors.secondary,
    borderRadius: 28,
    padding: 24,
    gap: 10,
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
    fontSize: 28,
    lineHeight: 35,
    fontWeight: '800',
  },
  subtitle: {
    color: colors.gray300,
    fontSize: 15,
    lineHeight: 22,
  },
  roleList: {
    marginTop: 6,
    gap: 6,
  },
  roleItem: {
    color: colors.white,
    fontSize: 14,
    lineHeight: 20,
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

export default RegisterScreen;
