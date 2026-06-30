import React, { useState, useEffect } from 'react';
import { View, Text, TouchableOpacity, StyleSheet } from 'react-native';
import Animated, {
  useSharedValue,
  useAnimatedStyle,
  withTiming,
  interpolate,
} from 'react-native-reanimated';
import { ChevronDown } from 'lucide-react-native';
import Card from './Card';
import { colors } from '../colors';

export interface PriceBreakdownProps {
  baseFareFormatted: string;
  surgeMultiplier: number;
  surgeAmountFormatted?: string;
  totalFormatted: string;
  className?: string;
}

export const PriceBreakdown: React.FC<PriceBreakdownProps> = ({
  baseFareFormatted,
  surgeMultiplier,
  surgeAmountFormatted,
  totalFormatted,
  className = '',
}) => {
  const [isExpanded, setIsExpanded] = useState(true);
  const expansionProgress = useSharedValue(1);

  useEffect(() => {
    expansionProgress.value = withTiming(isExpanded ? 1 : 0, { duration: 250 });
  }, [isExpanded]);

  // Animate Chevron Rotation
  const chevronStyle = useAnimatedStyle(() => {
    const rotate = interpolate(expansionProgress.value, [0, 1], [0, 180]);
    return {
      transform: [{ rotate: `${rotate}deg` }],
    };
  });

  // Animate Content Height/Scale
  const contentStyle = useAnimatedStyle(() => {
    return {
      opacity: expansionProgress.value,
      height: interpolate(expansionProgress.value, [0, 1], [0, 105]), // estimated height of inner details
      overflow: 'hidden',
    };
  });

  const showSurge = surgeMultiplier > 1;

  return (
    <Card variant="default" className={`p-0 overflow-hidden ${className}`}>
      {/* Header Button */}
      <TouchableOpacity
        onPress={() => setIsExpanded(!isExpanded)}
        activeOpacity={0.9}
        className="flex-row w-full items-center justify-between px-4 py-3.5"
      >
        <Text className="font-semibold text-gray-900 text-sm">Price breakdown</Text>
        <View className="flex-row items-center gap-2">
          <Text className="font-bold text-gray-900 text-base">{totalFormatted}</Text>
          <Animated.View style={chevronStyle}>
            <ChevronDown size={16} color={colors.gray500} />
          </Animated.View>
        </View>
      </TouchableOpacity>

      {/* Expandable Details Container */}
      <Animated.View style={contentStyle}>
        <View className="flex-col gap-2.5 border-t border-gray-100 px-4 py-3.5">
          {/* Base Fare Row */}
          <View className="flex-row items-center justify-between text-sm">
            <Text className="text-gray-600">Base fare</Text>
            <Text className="font-medium text-gray-900">{baseFareFormatted}</Text>
          </View>

          {/* Surge Pricing Row */}
          {showSurge && (
            <View className="flex-row items-center justify-between text-sm">
              <View className="flex-row items-center gap-2">
                <Text className="text-gray-600">Surge</Text>
                <View className="bg-warning/15 px-1.5 py-0.5 rounded">
                  <Text className="text-[10px] font-bold text-warning">
                    {surgeMultiplier}x
                  </Text>
                </View>
              </View>
              <Text className="font-medium text-warning">
                {surgeAmountFormatted || '₦0'}
              </Text>
            </View>
          )}

          {/* Divider & Total */}
          <View className="mt-1 flex-row items-center justify-between border-t border-gray-100 pt-3">
            <Text className="font-semibold text-gray-900 text-sm">Total</Text>
            <Text className="font-bold text-gray-900 text-base">{totalFormatted}</Text>
          </View>
        </View>
      </Animated.View>
    </Card>
  );
};

export default PriceBreakdown;
