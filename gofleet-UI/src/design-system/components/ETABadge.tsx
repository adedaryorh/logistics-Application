import React, { useEffect } from 'react';
import { View, Text } from 'react-native';
import Animated, {
  useSharedValue,
  useAnimatedStyle,
  withRepeat,
  withSequence,
  withTiming,
} from 'react-native-reanimated';
import { Car, ShoppingBag, Package } from 'lucide-react-native';
import { colors } from '../colors';

export type ETABadgeType = 'ride' | 'food' | 'parcel';

export interface ETABadgeProps {
  minutes: number | string;
  type?: ETABadgeType;
  pulse?: boolean;
  className?: string;
}

export const ETABadge: React.FC<ETABadgeProps> = ({
  minutes,
  type = 'ride',
  pulse = true,
  className = '',
}) => {
  const scale = useSharedValue(1);

  useEffect(() => {
    if (pulse) {
      scale.value = withRepeat(
        withSequence(
          withTiming(1.05, { duration: 1200 }),
          withTiming(1.0, { duration: 1200 })
        ),
        -1, // infinite
        true
      );
    } else {
      scale.value = 1;
    }
  }, [pulse]);

  const animatedStyle = useAnimatedStyle(() => {
    return {
      transform: [{ scale: scale.value }],
    };
  });

  // Select Type Icon
  let IconComponent = Car;
  switch (type) {
    case 'ride':
      IconComponent = Car;
      break;
    case 'food':
      IconComponent = ShoppingBag;
      break;
    case 'parcel':
      IconComponent = Package;
      break;
  }

  const label = typeof minutes === 'number' ? `${minutes} min` : minutes;

  return (
    <Animated.View
      style={pulse ? animatedStyle : undefined}
      className={`inline-flex flex-row items-center gap-1.5 rounded-full bg-accent/10 px-2.5 py-1 self-start ${className}`}
    >
      <IconComponent size={14} color={colors.accent} />
      <Text className="text-xs font-semibold text-accent">{label}</Text>
    </Animated.View>
  );
};

export default ETABadge;
