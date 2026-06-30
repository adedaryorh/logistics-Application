import React, { useEffect } from 'react';
import { View, Text } from 'react-native';
import Animated, {
  useSharedValue,
  useAnimatedStyle,
  withRepeat,
  withTiming,
  withSequence,
} from 'react-native-reanimated';
import { colors } from '../colors';

export type BadgeStatus =
  | 'online'
  | 'offline'
  | 'on_trip'
  | 'pending'
  | 'active'
  | 'completed';

export interface BadgeProps {
  status: BadgeStatus;
  label?: string;
  className?: string;
}

export const Badge: React.FC<BadgeProps> = ({ status, label, className = '' }) => {
  const pulseScale = useSharedValue(1);
  const pulseOpacity = useSharedValue(1);

  useEffect(() => {
    if (status === 'online') {
      pulseScale.value = withRepeat(
        withSequence(
          withTiming(1.3, { duration: 1000 }),
          withTiming(1.0, { duration: 1000 })
        ),
        -1, // infinite
        true
      );
      pulseOpacity.value = withRepeat(
        withSequence(
          withTiming(0.4, { duration: 1000 }),
          withTiming(1.0, { duration: 1000 })
        ),
        -1,
        true
      );
    } else {
      pulseScale.value = 1;
      pulseOpacity.value = 1;
    }
  }, [status]);

  const dotAnimatedStyle = useAnimatedStyle(() => {
    return {
      transform: [{ scale: pulseScale.value }],
      opacity: pulseOpacity.value,
    };
  });

  // Default display labels
  const displayLabel = label || status.replace('_', ' ').replace(/\b\w/g, (l) => l.toUpperCase());

  // Styling derivations
  let containerStyle = 'inline-flex flex-row items-center gap-1.5 rounded-full px-2.5 py-1 self-start';
  let dotStyle = 'w-1.5 h-1.5 rounded-full';
  let textStyle = 'text-xs font-semibold';

  switch (status) {
    case 'online':
      containerStyle += ' bg-success/10';
      dotStyle += ' bg-success';
      textStyle += ' text-success';
      break;
    case 'offline':
      containerStyle += ' bg-gray-200/50';
      dotStyle += ' bg-gray-400';
      textStyle += ' text-gray-500';
      break;
    case 'on_trip':
      containerStyle += ' bg-warning/10';
      dotStyle += ' bg-warning';
      textStyle += ' text-warning';
      break;
    case 'pending':
      containerStyle += ' bg-gray-200/50';
      dotStyle += ' bg-gray-400';
      textStyle += ' text-gray-600';
      break;
    case 'active':
      containerStyle += ' bg-accent/10';
      dotStyle += ' bg-accent';
      textStyle += ' text-accent';
      break;
    case 'completed':
      containerStyle += ' bg-success/10';
      dotStyle += ' bg-success';
      textStyle += ' text-success';
      break;
  }

  return (
    <View className={`${containerStyle} ${className}`}>
      <Animated.View
        className={dotStyle}
        style={status === 'online' ? dotAnimatedStyle : undefined}
      />
      <Text className={textStyle}>{displayLabel}</Text>
    </View>
  );
};

export default Badge;
