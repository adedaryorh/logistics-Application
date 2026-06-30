import React from 'react';
import { StatusBar as RNStatusBar, Platform } from 'react-native';
import { colors } from '../colors';

export interface StatusBarProps {
  barStyle?: 'light-content' | 'dark-content' | 'default';
  backgroundColor?: string;
  translucent?: boolean;
}

export const StatusBar: React.FC<StatusBarProps> = ({
  barStyle = 'light-content',
  backgroundColor = colors.primary,
  translucent = false,
}) => {
  return (
    <RNStatusBar
      barStyle={barStyle}
      backgroundColor={Platform.OS === 'android' ? backgroundColor : undefined}
      translucent={translucent}
    />
  );
};

export default StatusBar;
