import React from 'react';
import {
  TouchableOpacity,
  Text,
  ActivityIndicator,
  View,
  GestureResponderEvent,
} from 'react-native';
import { colors } from '../colors';
import { haptics } from '../../utils/haptics';

export interface ButtonProps {
  label: string;
  onPress?: (event: GestureResponderEvent) => void;
  variant?: 'primary' | 'secondary' | 'outline' | 'ghost' | 'danger';
  size?: 'sm' | 'md' | 'lg' | 'full';
  loading?: boolean;
  disabled?: boolean;
  leftIcon?: React.ReactNode;
  rightIcon?: React.ReactNode;
  className?: string;
  accessibilityLabel?: string;
}

export const Button: React.FC<ButtonProps> = ({
  label,
  onPress,
  variant = 'primary',
  size = 'md',
  loading = false,
  disabled = false,
  leftIcon,
  rightIcon,
  className = '',
  accessibilityLabel,
}) => {
  const handlePress = (event: GestureResponderEvent) => {
    if (disabled || loading) {
      return;
    }
    haptics.impact();
    onPress?.(event);
  };

  let baseStyle = 'flex-row items-center justify-center rounded-xl font-semibold';
  let variantStyle = '';
  let textStyle = '';
  let loaderColor: string = colors.white;

  switch (variant) {
    case 'primary':
      variantStyle = 'bg-highlight';
      textStyle = 'text-white';
      break;
    case 'secondary':
      variantStyle = 'bg-secondary';
      textStyle = 'text-white';
      break;
    case 'outline':
      variantStyle = 'border border-accent/40 bg-transparent';
      textStyle = 'text-accent';
      loaderColor = colors.accent;
      break;
    case 'ghost':
      variantStyle = 'bg-transparent';
      textStyle = 'text-accent';
      loaderColor = colors.accent;
      break;
    case 'danger':
      variantStyle = 'bg-error';
      textStyle = 'text-white';
      break;
  }

  let sizeStyle = '';
  let textSizeStyle = '';
  switch (size) {
    case 'sm':
      sizeStyle = 'h-9 px-3.5';
      textSizeStyle = 'text-xs';
      break;
    case 'md':
      sizeStyle = 'h-11 px-5';
      textSizeStyle = 'text-sm';
      break;
    case 'lg':
      sizeStyle = 'h-13 px-6';
      textSizeStyle = 'text-base';
      break;
    case 'full':
      sizeStyle = 'h-12 w-full px-6';
      textSizeStyle = 'text-sm';
      break;
  }

  if (disabled || loading) {
    variantStyle += ' opacity-50';
  }

  return (
    <TouchableOpacity
      onPress={handlePress}
      activeOpacity={0.8}
      disabled={disabled || loading}
      accessibilityRole="button"
      accessibilityLabel={accessibilityLabel || label}
      accessibilityState={{ disabled: disabled || loading, busy: loading }}
      className={`${baseStyle} ${variantStyle} ${sizeStyle} ${className}`}
    >
      {loading ? (
        <View className="flex-row items-center justify-center gap-2">
          <ActivityIndicator size="small" color={loaderColor} />
          <Text className={`${textStyle} ${textSizeStyle}`}>{label}</Text>
        </View>
      ) : (
        <View className="flex-row items-center justify-center gap-2">
          {leftIcon ? <View className="mr-1">{leftIcon}</View> : null}
          <Text className={`${textStyle} ${textSizeStyle} font-semibold`}>{label}</Text>
          {rightIcon ? <View className="ml-1">{rightIcon}</View> : null}
        </View>
      )}
    </TouchableOpacity>
  );
};

export default Button;
