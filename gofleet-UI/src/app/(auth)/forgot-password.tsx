import React, { useState } from 'react';
import { StyleSheet, View } from 'react-native';

import Button from '@/design-system/components/Button';
import Input from '@/design-system/components/Input';
import { toast } from '@/design-system/components/Toast';
import { colors } from '@/design-system/colors';

export function ForgotPasswordScreen() {
  const [email, setEmail] = useState('');

  return (
    <View style={styles.container}>
      <Input label="Email" value={email} onChangeText={setEmail} autoCapitalize="none" keyboardType="email-address" />
      <Button label="Send reset link" size="full" onPress={() => toast.info('Forgot password flow pending backend route confirmation.')} />
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    gap: 16,
    padding: 24,
    justifyContent: 'center',
    backgroundColor: colors.background,
  },
});

export default ForgotPasswordScreen;
