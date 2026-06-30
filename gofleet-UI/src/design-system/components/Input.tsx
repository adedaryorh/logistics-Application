import React, { useState, forwardRef } from 'react';
import {
  View,
  Text,
  TextInput,
  TouchableOpacity,
  TextInputProps,
  Platform,
} from 'react-native';
import { Eye, EyeOff } from 'lucide-react-native';
import { colors } from '../colors';

export interface InputProps extends Omit<TextInputProps, 'secureTextEntry'> {
  label?: string;
  helperText?: string;
  error?: string;
  variant?: 'default' | 'filled' | 'outlined';
  leftIcon?: React.ReactNode;
  rightIcon?: React.ReactNode;
  secureTextEntry?: boolean;
  className?: string;
  containerClassName?: string;
}

export const Input = forwardRef<TextInput, InputProps>(
  (
    {
      label,
      helperText,
      error,
      variant = 'default',
      leftIcon,
      rightIcon,
      secureTextEntry,
      className = '',
      containerClassName = '',
      ...props
    },
    ref
  ) => {
    const [isFocused, setIsFocused] = useState(false);
    const [isPasswordVisible, setIsPasswordVisible] = useState(false);

    const isPassword = !!secureTextEntry;

    // Outer container styling
    let inputContainerStyle =
      'flex-row h-12 items-center rounded-xl px-3.5 border transition-all duration-150';

    // Apply variants
    switch (variant) {
      case 'default':
        inputContainerStyle += error
          ? ' border-error bg-white'
          : isFocused
          ? ' border-accent bg-white shadow-sm'
          : ' border-gray-200 bg-white';
        break;
      case 'filled':
        inputContainerStyle += error
          ? ' border-error bg-error/5'
          : isFocused
          ? ' border-accent bg-white shadow-sm'
          : ' border-transparent bg-gray-100';
        break;
      case 'outlined':
        inputContainerStyle += error
          ? ' border-error border-2'
          : isFocused
          ? ' border-accent border-2'
          : ' border-gray-300 border-2 bg-transparent';
        break;
    }

    return (
      <View className={`w-full ${containerClassName}`}>
        {label && (
          <Text className="mb-1.5 text-sm font-medium text-gray-700">
            {label}
          </Text>
        )}

        <View className={inputContainerStyle}>
          {leftIcon && (
            <View className="mr-2 text-gray-400 items-center justify-center">
              {leftIcon}
            </View>
          )}

          <TextInput
            ref={ref}
            className={`flex-1 h-full text-sm text-gray-900 outline-none p-0 ${
              Platform.OS === 'web' ? 'outline-none' : ''
            } ${className}`}
            placeholderTextColor={colors.gray400}
            secureTextEntry={isPassword && !isPasswordVisible}
            onFocus={(e) => {
              setIsFocused(true);
              props.onFocus?.(e);
            }}
            onBlur={(e) => {
              setIsFocused(false);
              props.onBlur?.(e);
            }}
            {...props}
          />

          {isPassword ? (
            <TouchableOpacity
              onPress={() => setIsPasswordVisible(!isPasswordVisible)}
              className="ml-2 text-gray-400"
              activeOpacity={0.7}
              accessibilityLabel={isPasswordVisible ? 'Hide password' : 'Show password'}
            >
              {isPasswordVisible ? (
                <EyeOff size={20} color={colors.gray400} />
              ) : (
                <Eye size={20} color={colors.gray400} />
              )}
            </TouchableOpacity>
          ) : (
            rightIcon && (
              <View className="ml-2 text-gray-400 items-center justify-center">
                {rightIcon}
              </View>
            )
          )}
        </View>

        {error ? (
          <Text className="mt-1.5 text-xs text-error font-medium">{error}</Text>
        ) : helperText ? (
          <Text className="mt-1.5 text-xs text-gray-500">{helperText}</Text>
        ) : null}
      </View>
    );
  }
);

Input.displayName = 'Input';

export default Input;
