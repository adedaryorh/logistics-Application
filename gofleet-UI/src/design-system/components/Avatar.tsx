import React from 'react';
import { View, Text, Image, ImageSourcePropType } from 'react-native';
import { colors } from '../colors';

export interface AvatarProps {
  source?: ImageSourcePropType | string;
  name?: string;
  size?: 'sm' | 'md' | 'lg' | 'xl';
  isOnline?: boolean;
  className?: string;
}

export const Avatar: React.FC<AvatarProps> = ({
  source,
  name = '',
  size = 'md',
  isOnline = false,
  className = '',
}) => {
  // Extract initials
  const getInitials = (fullName: string) => {
    const parts = fullName.trim().split(' ');
    if (parts.length === 0 || !parts[0]) return '?';
    if (parts.length === 1) return parts[0].substring(0, 2).toUpperCase();
    return (parts[0][0] + parts[1][0]).toUpperCase();
  };

  // Dimensions
  let containerSize = 'w-10 h-10';
  let textSize = 'text-sm';
  let dotSize = 'w-2.5 h-2.5';
  let ringSize = 'ring-2';

  switch (size) {
    case 'sm':
      containerSize = 'w-8 h-8';
      textSize = 'text-xs';
      dotSize = 'w-2 h-2';
      ringSize = 'ring-2';
      break;
    case 'md':
      containerSize = 'w-10 h-10';
      textSize = 'text-sm';
      dotSize = 'w-2.5 h-2.5';
      ringSize = 'ring-2';
      break;
    case 'lg':
      containerSize = 'w-14 h-14';
      textSize = 'text-base';
      dotSize = 'w-3.5 h-3.5';
      ringSize = 'ring-[3px]';
      break;
    case 'xl':
      containerSize = 'w-20 h-20';
      textSize = 'text-xl';
      dotSize = 'w-5 h-5';
      ringSize = 'ring-4';
      break;
  }

  // Resolve Image Source
  const imageSource = typeof source === 'string' ? { uri: source } : source;

  return (
    <View className={`relative inline-flex shrink-0 ${containerSize} ${className}`}>
      <View
        className={`w-full h-full rounded-full bg-accent items-center justify-center overflow-hidden`}
      >
        {imageSource ? (
          <Image
            source={imageSource}
            className="w-full h-full object-cover"
            accessibilityLabel={name}
          />
        ) : (
          <Text className={`font-semibold text-white ${textSize}`}>
            {getInitials(name)}
          </Text>
        )}
      </View>

      {isOnline && (
        <View
          className={`absolute right-0 bottom-0 rounded-full bg-success ring-white ${dotSize} ${ringSize}`}
          accessibilityLabel="Online status indicator"
        />
      )}
    </View>
  );
};

export default Avatar;
