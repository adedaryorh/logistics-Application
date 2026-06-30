import React, { useState } from 'react';
import { StyleSheet, View } from 'react-native';

import Button from '@/design-system/components/Button';
import Input from '@/design-system/components/Input';
import { toast } from '@/design-system/components/Toast';
import { colors } from '@/design-system/colors';

export function VerifyEmailScreen() {
  const [code, setCode] = useState('');

  return (
    <View style={styles.container}>
      <Input label="Verification code" value={code} onChangeText={setCode} keyboardType="number-pad" />
      <Button label="Verify email" size="full" onPress={() => toast.success('Verification flow connected once backend OTP endpoint is wired.')} />
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

export default VerifyEmailScreen;
