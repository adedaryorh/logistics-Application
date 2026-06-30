import React, { useState } from 'react';
import { StyleSheet, View } from 'react-native';

import Button from '@/design-system/components/Button';
import Input from '@/design-system/components/Input';
import { toast } from '@/design-system/components/Toast';
import { colors } from '@/design-system/colors';

export function MagicLinkScreen() {
  const [email, setEmail] = useState('');

  return (
    <View style={styles.container}>
      <Input label="Email" value={email} onChangeText={setEmail} autoCapitalize="none" keyboardType="email-address" />
      <Button label="Send magic link" size="full" onPress={() => toast.info('Magic link flow pending backend endpoint.')} />
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

export default MagicLinkScreen;
