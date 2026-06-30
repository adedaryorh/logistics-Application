import React from 'react';
import { View, TouchableOpacity, ViewStyle } from 'react-native';

export interface CardProps {
  children: React.ReactNode;
  variant?: 'default' | 'elevated' | 'outlined' | 'flat';
  onPress?: () => void;
  className?: string;
  style?: ViewStyle;
}

export const Card: React.FC<CardProps> = ({
  children,
  variant = 'default',
  onPress,
  className = '',
  style,
}) => {
  let cardStyle = 'rounded-2xl p-4 text-gray-900 overflow-hidden';

  switch (variant) {
    case 'default':
      cardStyle += ' bg-white border border-gray-200 shadow-sm';
      break;
    case 'elevated':
      cardStyle += ' bg-white shadow-lg shadow-gray-900/10';
      break;
    case 'outlined':
      cardStyle += ' bg-white border-2 border-gray-300';
      break;
    case 'flat':
      cardStyle += ' bg-gray-100';
      break;
  }

  if (onPress) {
    return (
      <TouchableOpacity
        onPress={onPress}
        activeOpacity={0.9}
        className={`${cardStyle} ${className}`}
        style={style}
      >
        {children}
      </TouchableOpacity>
    );
  }

  return (
    <View className={`${cardStyle} ${className}`} style={style}>
      {children}
    </View>
  );
};

export default Card;
