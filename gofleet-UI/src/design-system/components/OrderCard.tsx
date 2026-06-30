import React from 'react';
import { View, Text } from 'react-native';
import { Car, ShoppingBag, Package, RotateCcw } from 'lucide-react-native';
import Badge, { BadgeStatus } from './Badge';
import Button from './Button';
import Card from './Card';
import { colors } from '../colors';

export type OrderCardType = 'ride' | 'food' | 'parcel';

export interface OrderCardProps {
  type: OrderCardType;
  status: BadgeStatus;
  pickupAddress: string;
  dropoffAddress: string;
  priceFormatted: string;
  onReorder?: () => void;
  onPress?: () => void;
}

export const OrderCard: React.FC<OrderCardProps> = ({
  type,
  status,
  pickupAddress,
  dropoffAddress,
  priceFormatted,
  onReorder,
  onPress,
}) => {
  // Select Type Icon
  let TypeIcon = Car;
  switch (type) {
    case 'ride':
      TypeIcon = Car;
      break;
    case 'food':
      TypeIcon = ShoppingBag;
      break;
    case 'parcel':
      TypeIcon = Package;
      break;
  }

  return (
    <Card variant="default" onPress={onPress} className="flex-col gap-3 mb-3">
      {/* Top Header */}
      <View className="flex-row items-center justify-between gap-2">
        <View className="w-10 h-10 items-center justify-center rounded-xl bg-accent/10">
          <TypeIcon size={20} color={colors.accent} />
        </View>
        <Badge status={status} />
      </View>

      {/* Address Timeline */}
      <View className="flex-row items-center gap-3 py-1">
        {/* Route visualization dots */}
        <View className="flex-col items-center justify-center w-4">
          <View className="w-2.5 h-2.5 rounded-full border-2 border-accent" />
          <View className="w-px h-6 bg-gray-200 my-1" />
          <View className="w-2.5 h-2.5 rounded-full bg-highlight" />
        </View>

        {/* Address texts */}
        <View className="flex-1 gap-3">
          <Text className="text-sm font-medium text-gray-900" numberOfLines={1}>
            {pickupAddress}
          </Text>
          <Text className="text-sm font-medium text-gray-900" numberOfLines={1}>
            {dropoffAddress}
          </Text>
        </View>
      </View>

      {/* Footer Price & Reorder Action */}
      <View className="flex-row items-center justify-between border-t border-gray-100 pt-3 mt-1">
        <Text className="text-lg font-bold text-gray-900">{priceFormatted}</Text>
        {onReorder && (
          <Button
            label="Reorder"
            onPress={onReorder}
            variant="outline"
            size="sm"
            leftIcon={<RotateCcw size={14} color={colors.accent} />}
          />
        )}
      </View>
    </Card>
  );
};

export default OrderCard;
