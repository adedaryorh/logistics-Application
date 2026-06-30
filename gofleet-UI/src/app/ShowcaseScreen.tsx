import React, { useRef, useState } from 'react';
import {
  ScrollView,
  View,
  Text,
  SafeAreaView,
  TouchableOpacity,
  StyleSheet,
  Dimensions,
  Platform,
} from 'react-native';
import GorhomBottomSheet from '@gorhom/bottom-sheet';
import {
  Sparkles,
  Smartphone,
  Layers,
  MapPin,
  Clock,
  Terminal,
  Activity,
} from 'lucide-react-native';

// Import design system tokens & components
import { colors } from '../design-system/colors';
import Button from '../design-system/components/Button';
import Input from '../design-system/components/Input';
import Card from '../design-system/components/Card';
import Badge from '../design-system/components/Badge';
import Avatar from '../design-system/components/Avatar';
import BottomSheet from '../design-system/components/BottomSheet';
import MapView from '../design-system/components/MapView';
import StatusBar from '../design-system/components/StatusBar';
import { toast } from '../design-system/components/Toast';
import { LoadingScreen, SkeletonCard, SkeletonList } from '../design-system/components/LoadingScreen';
import EmptyState from '../design-system/components/EmptyState';
import DriverCard from '../design-system/components/DriverCard';
import OrderCard from '../design-system/components/OrderCard';
import PriceBreakdown from '../design-system/components/PriceBreakdown';
import ETABadge from '../design-system/components/ETABadge';

export const ShowcaseScreen: React.FC = () => {
  const bottomSheetRef = useRef<GorhomBottomSheet>(null);
  
  // States for interactive components
  const [inputText, setInputText] = useState('');
  const [passwordText, setPasswordText] = useState('');
  const [showLoading, setShowLoading] = useState(false);
  const [showSkeletons, setShowSkeletons] = useState(false);

  // Trigger loading state for 2 seconds
  const triggerLoading = () => {
    setShowLoading(true);
    setTimeout(() => {
      setShowLoading(false);
    }, 2000);
  };

  const handleOpenBottomSheet = () => {
    bottomSheetRef.current?.expand();
  };

  if (showLoading) {
    return <LoadingScreen message="Simulating dispatch flow..." dark={true} />;
  }

  return (
    <SafeAreaView style={styles.safeArea} className="flex-1 bg-gray-100">
      <StatusBar barStyle="light-content" backgroundColor={colors.primary} />
      
      {/* Premium Design System Header */}
      <View className="bg-primary px-6 py-6 border-b border-secondary">
        <View className="flex-row items-center gap-3">
          <View className="w-10 h-10 items-center justify-center rounded-xl bg-highlight">
            <Sparkles size={20} color="white" />
          </View>
          <View>
            <Text className="text-xl font-bold text-white">Logiflow</Text>
            <Text className="text-xs font-semibold text-gray-400">Design System Library</Text>
          </View>
        </View>
        <Text className="mt-4 text-xs font-medium text-gray-300 leading-relaxed">
          Production-ready components for ride-hailing, food delivery, and merchant services.
        </Text>
      </View>

      <ScrollView className="flex-1 px-4 py-4" contentContainerStyle={styles.scrollContent}>
        
        {/* BUTTONS SECTION */}
        <View className="mb-8">
          <View className="mb-3">
            <Text className="text-lg font-bold text-gray-900">Button</Text>
            <Text className="text-xs text-gray-500">5 variants × 4 sizes, loading and disabled states.</Text>
          </View>
          <Card variant="default" className="gap-5">
            <View>
              <Text className="mb-2 text-xs font-bold uppercase tracking-wider text-gray-400">Variants</Text>
              <View className="flex-row flex-wrap gap-2">
                <Button label="Primary" variant="primary" size="sm" />
                <Button label="Secondary" variant="secondary" size="sm" />
                <Button label="Outline" variant="outline" size="sm" />
                <Button label="Ghost" variant="ghost" size="sm" />
                <Button label="Danger" variant="danger" size="sm" />
              </View>
            </View>

            <View>
              <Text className="mb-2 text-xs font-bold uppercase tracking-wider text-gray-400">Sizes</Text>
              <View className="flex-row flex-wrap items-center gap-2">
                <Button label="Small" variant="primary" size="sm" />
                <Button label="Medium" variant="primary" size="md" />
                <Button label="Large" variant="primary" size="lg" />
              </View>
              <View className="mt-2">
                <Button label="Full Width Button" variant="primary" size="full" />
              </View>
            </View>

            <View>
              <Text className="mb-2 text-xs font-bold uppercase tracking-wider text-gray-400">States</Text>
              <View className="flex-row flex-wrap gap-2">
                <Button label="Loading" variant="primary" loading={true} size="md" />
                <Button label="Disabled" variant="primary" disabled={true} size="md" />
              </View>
            </View>
          </Card>
        </View>

        {/* INPUTS SECTION */}
        <View className="mb-8">
          <View className="mb-3">
            <Text className="text-lg font-bold text-gray-900">Input</Text>
            <Text className="text-xs text-gray-500">Standard configurations with validations, toggles and helper text.</Text>
          </View>
          <Card variant="default" className="gap-4">
            <Input
              label="Email Address"
              placeholder="you@example.com"
              value={inputText}
              onChangeText={setInputText}
              helperText="We'll send receipts and updates here."
              variant="default"
            />
            <Input
              label="Search Destination"
              placeholder="Where to?"
              variant="filled"
              leftIcon={<MapPin size={18} color={colors.gray400} />}
            />
            <Input
              label="Secure Password"
              placeholder="Enter password"
              value={passwordText}
              onChangeText={setPasswordText}
              secureTextEntry={true}
              variant="outlined"
            />
            <Input
              label="Phone Number"
              placeholder="0801 234 5678"
              value="0801"
              error="Enter a valid 11-digit phone number."
            />
          </Card>
        </View>

        {/* CARDS SECTION */}
        <View className="mb-8">
          <View className="mb-3">
            <Text className="text-lg font-bold text-gray-900">Card</Text>
            <Text className="text-xs text-gray-500">Structured content containers with distinct elevations.</Text>
          </View>
          <View className="gap-3">
            <Card variant="default">
              <Text className="font-bold text-gray-900">Default Card</Text>
              <Text className="text-xs text-gray-500 mt-1">Light grey border with dynamic hover feedback support.</Text>
            </Card>
            <Card variant="elevated">
              <Text className="font-bold text-gray-900">Elevated Card</Text>
              <Text className="text-xs text-gray-500 mt-1">Soft dark shadow rendering for top-level modules.</Text>
            </Card>
            <Card variant="outlined">
              <Text className="font-bold text-gray-900">Outlined Card</Text>
              <Text className="text-xs text-gray-500 mt-1">2px distinct borders for visual separations.</Text>
            </Card>
            <Card variant="flat">
              <Text className="font-bold text-gray-900">Flat Card</Text>
              <Text className="text-xs text-gray-500 mt-1">Simple neutral grey filling backdrop.</Text>
            </Card>
          </View>
        </View>

        {/* BADGES & AVATARS */}
        <View className="mb-8">
          <View className="mb-3">
            <Text className="text-lg font-bold text-gray-900">Badges & Avatars</Text>
            <Text className="text-xs text-gray-500">Live pulsing statuses and image/initial avatars.</Text>
          </View>
          <Card variant="default" className="gap-5">
            <View>
              <Text className="mb-2 text-xs font-bold uppercase tracking-wider text-gray-400">Status Badges</Text>
              <View className="flex-row flex-wrap gap-2">
                <Badge status="online" />
                <Badge status="offline" />
                <Badge status="on_trip" />
                <Badge status="pending" />
                <Badge status="active" />
                <Badge status="completed" />
              </View>
            </View>

            <View>
              <Text className="mb-2 text-xs font-bold uppercase tracking-wider text-gray-400">Avatars with Image</Text>
              <View className="flex-row items-end gap-3">
                <Avatar source="https://randomuser.me/api/portraits/men/32.jpg" name="Emeka O." size="sm" isOnline={true} />
                <Avatar source="https://randomuser.me/api/portraits/men/32.jpg" name="Emeka O." size="md" isOnline={true} />
                <Avatar source="https://randomuser.me/api/portraits/men/32.jpg" name="Emeka O." size="lg" isOnline={true} />
                <Avatar source="https://randomuser.me/api/portraits/men/32.jpg" name="Emeka O." size="xl" isOnline={true} />
              </View>
            </View>

            <View>
              <Text className="mb-2 text-xs font-bold uppercase tracking-wider text-gray-400">Initials Fallback</Text>
              <View className="flex-row items-end gap-3">
                <Avatar name="Ada Nnamdi" size="sm" />
                <Avatar name="Ada Nnamdi" size="md" />
                <Avatar name="Ada Nnamdi" size="lg" isOnline={false} />
                <Avatar name="Ada Nnamdi" size="xl" isOnline={true} />
              </View>
            </View>
          </Card>
        </View>

        {/* INTERACTIVE UTILITIES */}
        <View className="mb-8">
          <View className="mb-3">
            <Text className="text-lg font-bold text-gray-900">Global Toasts & Sheets</Text>
            <Text className="text-xs text-gray-500">Interactive triggers to test global notification actions.</Text>
          </View>
          <Card variant="default" className="gap-3">
            <View className="flex-row flex-wrap gap-2">
              <Button
                label="Trigger Success Toast"
                variant="primary"
                size="sm"
                onPress={() => toast.success('Driver Arrived', 'Emeka is waiting at the pickup zone.')}
              />
              <Button
                label="Trigger Error Toast"
                variant="danger"
                size="sm"
                onPress={() => toast.error('Payment Failed', 'Insufficient funds on selected card.')}
              />
              <Button
                label="Trigger Warning Toast"
                variant="secondary"
                size="sm"
                onPress={() => toast.warning('Surge Pricing Active', 'Fares are slightly higher due to demand.')}
              />
            </View>
            <Button
              label="Trigger App Loading Screen"
              variant="outline"
              size="full"
              onPress={triggerLoading}
            />
            <Button
              label="Open Bottom Sheet overlay"
              variant="secondary"
              size="full"
              onPress={handleOpenBottomSheet}
            />
          </Card>
        </View>

        {/* SHIMMER SKELETONS */}
        <View className="mb-8">
          <View className="mb-3 flex-row items-center justify-between">
            <View>
              <Text className="text-lg font-bold text-gray-900">Shimmer Skeletons</Text>
              <Text className="text-xs text-gray-500">Pulsing content skeleton blocks during data fetching.</Text>
            </View>
            <TouchableOpacity
              onPress={() => setShowSkeletons(!showSkeletons)}
              className="px-3 py-1 bg-accent/15 rounded-lg"
            >
              <Text className="text-xs font-bold text-accent">
                {showSkeletons ? 'Hide' : 'Show list'}
              </Text>
            </TouchableOpacity>
          </View>
          {showSkeletons ? (
            <SkeletonList count={2} />
          ) : (
            <SkeletonCard />
          )}
        </View>

        {/* MAPVIEW RENDERING */}
        <View className="mb-8">
          <View className="mb-3">
            <Text className="text-lg font-bold text-gray-900">MapView</Text>
            <Text className="text-xs text-gray-500">Visual mapping panel drawing driver location and routing routes.</Text>
          </View>
          <Card variant="default" className="p-0 overflow-hidden h-72">
            <MapView showRoute={true} dark={true} />
          </Card>
        </View>

        {/* COMPLEX LOGISTICS CARDS */}
        <View className="mb-8">
          <View className="mb-3">
            <Text className="text-lg font-bold text-gray-900">Logistics Components</Text>
            <Text className="text-xs text-gray-500">Domain-specific modules for active statuses and orders.</Text>
          </View>
          <View className="gap-4">
            {/* Driver Card */}
            <Text className="text-xs font-bold uppercase tracking-wider text-gray-400 mb-1">Driver Card</Text>
            <DriverCard
              avatarSource="https://randomuser.me/api/portraits/men/32.jpg"
              name="Emeka O."
              rating={4.8}
              vehicleModel="Toyota Corolla"
              licensePlate="ABC-123XY"
              etaMinutes={4}
              showActions={true}
              onCall={() => toast.info('Simulating Call', 'Dialing driver Emeka...')}
              onMessage={() => toast.info('Simulating Chat', 'Opening thread with driver...')}
            />

            {/* Order Card */}
            <Text className="text-xs font-bold uppercase tracking-wider text-gray-400 mb-1">Order Card</Text>
            <OrderCard
              type="ride"
              status="completed"
              pickupAddress="Victoria Island"
              dropoffAddress="Lekki"
              priceFormatted="₦1,200"
              onReorder={() => toast.success('Ride Reordered', 'Searching for a driver again...')}
            />

            {/* ETA Badges */}
            <Text className="text-xs font-bold uppercase tracking-wider text-gray-400 mb-1">ETA Badges</Text>
            <View className="flex-row gap-2 flex-wrap">
              <ETABadge minutes={4} type="ride" />
              <ETABadge minutes={12} type="food" />
              <ETABadge minutes="Arriving" type="parcel" />
            </View>

            {/* Price Breakdown */}
            <Text className="text-xs font-bold uppercase tracking-wider text-gray-400 mb-1">Price Breakdown (Expandable)</Text>
            <PriceBreakdown
              baseFareFormatted="₦600"
              surgeMultiplier={1.2}
              surgeAmountFormatted="₦120"
              totalFormatted="₦720"
            />
          </View>
        </View>

        {/* EMPTY STATES */}
        <View className="mb-12">
          <View className="mb-3">
            <Text className="text-lg font-bold text-gray-900">Empty State</Text>
            <Text className="text-xs text-gray-500">Standard fallback when data lists are empty.</Text>
          </View>
          <EmptyState
            title="No Active Orders"
            description="You don't have any bookings running. Start by searching for a ride or restaurant."
            actionLabel="Book a Ride"
            onAction={() => toast.info('Booking Simulation', 'Redirecting to booking screen...')}
          />
        </View>

      </ScrollView>

      {/* BOTTOM SHEET TRIGGER OVERLAY */}
      <BottomSheet ref={bottomSheetRef} index={-1} snapPoints={['25%', '60%']}>
        <View className="flex-col gap-4">
          <View className="flex-row items-center gap-2.5 pb-3 border-b border-gray-100">
            <Layers size={20} color={colors.highlight} />
            <Text className="text-lg font-bold text-gray-900">Logiflow Bottom Sheet</Text>
          </View>
          <Text className="text-sm text-gray-500 leading-normal">
            This sheet is running on react-native-reanimated and uses native snap points. Swipe down to dismiss or drag it to expand.
          </Text>
          <Button
            label="Dismiss Sheet"
            variant="primary"
            size="full"
            onPress={() => bottomSheetRef.current?.close()}
          />
        </View>
      </BottomSheet>
    </SafeAreaView>
  );
};

const styles = StyleSheet.create({
  safeArea: {
    paddingTop: Platform.OS === 'android' ? 24 : 0,
  },
  scrollContent: {
    paddingBottom: 40,
  },
});

export default ShowcaseScreen;
