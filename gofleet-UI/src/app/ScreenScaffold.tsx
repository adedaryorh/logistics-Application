import React from 'react';
import { ScrollView, StyleSheet, Text, View } from 'react-native';

import { colors } from '@/design-system/colors';

export type ScreenScaffoldProps = {
  title: string;
  subtitle: string;
  children?: React.ReactNode;
};

export function ScreenScaffold({ title, subtitle, children }: ScreenScaffoldProps) {
  return (
    <ScrollView contentContainerStyle={styles.container}>
      <Text style={styles.title}>{title}</Text>
      <Text style={styles.subtitle}>{subtitle}</Text>
      <View style={styles.content}>{children}</View>
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: {
    flexGrow: 1,
    padding: 20,
    backgroundColor: colors.background,
  },
  title: {
    fontSize: 28,
    fontWeight: '700',
    color: colors.primary,
    marginBottom: 8,
  },
  subtitle: {
    fontSize: 15,
    lineHeight: 22,
    color: colors.gray600,
    marginBottom: 20,
  },
  content: {
    gap: 12,
  },
});
