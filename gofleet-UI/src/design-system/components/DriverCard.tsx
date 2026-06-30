import React from 'react';
import { View, Text, TouchableOpacity, ImageSourcePropType } from 'react-native';
import { Star, Car, Phone, MessageSquare } from 'lucide-react-native';
import Avatar from './Avatar';
import Card from './Card';
import { colors } from '../colors';

export interface DriverCardProps {
  avatarSource?: ImageSourcePropType | string;
  name: string;
  rating: number;
  vehicleModel: string;
  licensePlate: string;
  etaMinutes?: number;
  onPress?: () => void;
  onCall?: () => void;
  onMessage?: () => void;
  showActions?: boolean;
}

export const DriverCard: React.FC<DriverCardProps> = ({
  avatarSource,
  name,
  rating,
  vehicleModel,
  licensePlate,
  etaMinutes,
  onPress,
  onCall,
  onMessage,
  showActions = false,
}) => {
  return (
    <Card variant="elevated" onPress={onPress} className="flex-col gap-3">
      <View className="flex-row items-center gap-3">
        {/* Avatar with Online Ring */}
        <Avatar source={avatarSource} name={name} size="lg" isOnline={true} />

        {/* Details Column */}
        <View className="flex-1 min-w-0">
          <View className="flex-row items-center justify-between gap-2">
            <Text className="truncate font-semibold text-gray-900 text-base">
              {name}
            </Text>

            {etaMinutes !== undefined && (
              <View className="flex-row items-center gap-1 bg-accent/10 px-2 py-0.5 rounded-full">
                <Car size={14} color={colors.accent} />
                <Text className="text-xs font-semibold text-accent">
                  {etaMinutes} min
                </Text>
              </View>
            )}
          </View>

          {/* Rating Row */}
          <View className="mt-0.5 flex-row items-center gap-1">
            <Star size={14} color={colors.warning} fill={colors.warning} />
            <Text className="font-medium text-gray-700 text-sm">{rating.toFixed(1)}</Text>
          </View>

          {/* Vehicle Info */}
          <Text className="mt-1 truncate text-sm text-gray-500">
            {vehicleModel} <Text className="text-gray-300">•</Text> {licensePlate}
          </Text>
        </View>
      </View>

      {/* Optional Quick Actions Row (Call / Message) */}
      {showActions && (onCall || onMessage) && (
        <View className="flex-row gap-2 mt-2 pt-3 border-t border-gray-100">
          {onCall && (
            <TouchableOpacity
              onPress={onCall}
              activeOpacity={0.7}
              className="flex-1 flex-row items-center justify-center gap-1.5 h-10 border border-gray-200 rounded-xl bg-gray-50 active:bg-gray-100"
            >
              <Phone size={14} color={colors.gray700} />
              <Text className="text-xs font-semibold text-gray-700">Call</Text>
            </TouchableOpacity>
          )}
          {onMessage && (
            <TouchableOpacity
              onPress={onMessage}
              activeOpacity={0.7}
              className="flex-1 flex-row items-center justify-center gap-1.5 h-10 border border-gray-200 rounded-xl bg-gray-50 active:bg-gray-100"
            >
              <MessageSquare size={14} color={colors.gray700} />
              <Text className="text-xs font-semibold text-gray-700">Message</Text>
            </TouchableOpacity>
          )}
        </View>
      )}
    </Card>
  );
};

export default DriverCard;
