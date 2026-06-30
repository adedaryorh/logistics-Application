import React from 'react';
import { View, Text } from 'react-native';
import { Box } from 'lucide-react-native';
import Button from './Button';
import { colors } from '../colors';

export interface EmptyStateProps {
  title: string;
  description: string;
  icon?: React.ReactNode;
  actionLabel?: string;
  onAction?: () => void;
  className?: string;
}

export const EmptyState: React.FC<EmptyStateProps> = ({
  title,
  description,
  icon,
  actionLabel,
  onAction,
  className = '',
}) => {
  return (
    <View className={`flex-1 items-center justify-center p-6 bg-white rounded-2xl border border-gray-200 shadow-sm ${className}`}>
      <View className="items-center text-center max-w-sm gap-4">
        {/* Styled Circle with Vector Icon */}
        <View className="size-16 items-center justify-center rounded-full bg-accent/5 mb-2">
          {icon || <Box size={32} color={colors.accent} />}
        </View>

        <Text className="text-lg font-bold text-gray-900 text-center">
          {title}
        </Text>

        <Text className="text-sm text-gray-500 text-center leading-normal">
          {description}
        </Text>

        {actionLabel && onAction && (
          <Button
            label={actionLabel}
            onPress={onAction}
            variant="outline"
            size="sm"
            className="mt-2"
          />
        )}
      </View>
    </View>
  );
};

export default EmptyState;
