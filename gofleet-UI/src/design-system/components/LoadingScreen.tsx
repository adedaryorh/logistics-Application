import React, { useEffect } from 'react';
import { View, Text, ActivityIndicator, StyleSheet } from 'react-native';
import Animated, {
  useSharedValue,
  useAnimatedStyle,
  withRepeat,
  withTiming,
  withSequence,
} from 'react-native-reanimated';
import { Truck } from 'lucide-react-native';
import { colors } from '../colors';

export interface LoadingScreenProps {
  message?: string;
  dark?: boolean;
}

export const LoadingScreen: React.FC<LoadingScreenProps> = ({
  message = 'Loading...',
  dark = true,
}) => {
  return (
    <View
      className={`flex-1 items-center justify-center ${
        dark ? 'bg-primary' : 'bg-background'
      }`}
    >
      <View className="items-center gap-4">
        {/* Animated Brand Icon */}
        <View className="size-16 items-center justify-center rounded-2xl bg-highlight shadow-lg">
          <Truck size={32} color="white" />
        </View>

        <View className="mt-4 flex-row items-center gap-3">
          <ActivityIndicator size="large" color={colors.highlight} />
          <Text className={`text-base font-medium ${dark ? 'text-white' : 'text-gray-900'}`}>
            {message}
          </Text>
        </View>
      </View>
    </View>
  );
};

// Reusable Skeleton block with animated opacity shimmer
export interface SkeletonProps {
  width?: number | string;
  height?: number | string;
  borderRadius?: number;
  className?: string;
}

export const Skeleton: React.FC<SkeletonProps> = ({
  width = '100%',
  height = 20,
  borderRadius = 8,
  className = '',
}) => {
  const opacity = useSharedValue(0.3);

  useEffect(() => {
    opacity.value = withRepeat(
      withSequence(
        withTiming(0.7, { duration: 800 }),
        withTiming(0.3, { duration: 800 })
      ),
      -1,
      true
    );
  }, []);

  const animatedStyle = useAnimatedStyle(() => {
    return {
      opacity: opacity.value,
    };
  });

  return (
    <Animated.View
      className={`bg-gray-300 ${className}`}
      style={[
        styles.skeleton,
        animatedStyle,
        {
          width: width as any,
          height: height as any,
          borderRadius,
        },
      ]}
    />
  );
};

// Skeleton Loader Card (matches a typical UI card layout)
export const SkeletonCard: React.FC = () => {
  return (
    <View className="p-4 bg-white border border-gray-200 rounded-2xl mb-3 flex-col gap-3 shadow-sm">
      <View className="flex-row items-center gap-3">
        <Skeleton width={48} height={48} borderRadius={999} />
        <View className="flex-1 gap-2">
          <Skeleton width="40%" height={16} />
          <Skeleton width="60%" height={12} />
        </View>
      </View>
      <Skeleton width="100%" height={14} />
      <View className="flex-row justify-between pt-2 border-t border-gray-100">
        <Skeleton width="25%" height={24} />
        <Skeleton width="30%" height={32} borderRadius={12} />
      </View>
    </View>
  );
};

// Skeleton list view
export const SkeletonList: React.FC<{ count?: number }> = ({ count = 3 }) => {
  return (
    <View className="flex-1 p-4 gap-3 bg-gray-50">
      {Array.from({ length: count }).map((_, i) => (
        <SkeletonCard key={i} />
      ))}
    </View>
  );
};

const styles = StyleSheet.create({
  skeleton: {
    backgroundColor: '#E5E7EB', // default fallback gray200
  },
});

export default LoadingScreen;
