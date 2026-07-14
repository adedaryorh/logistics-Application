import { useEffect, useState } from 'react';
import {
  ActivityIndicator,
  Alert,
  Modal,
  Pressable,
  SafeAreaView,
  StyleSheet,
  Text,
  TextInput,
  View,
} from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import * as Haptics from 'expo-haptics';
import { api } from '../api';
import { IconButton } from '../components/AppUI';
import { CreateOrderInput, DeliveryType, Order } from '../types';
import { colors, radius } from '../theme';

export function BookingModal({
  visible,
  initialType,
  demoMode,
  onClose,
  onCreated,
}: {
  visible: boolean;
  initialType: DeliveryType;
  demoMode: boolean;
  onClose: () => void;
  onCreated: (o: Order) => void;
}) {
  const [type, setType] = useState(initialType);
  const [pickup, setPickup] = useState('Current location');
  const [dropoff, setDropoff] = useState('');
  const [loading, setLoading] = useState(false);
  useEffect(() => setType(initialType), [initialType]);
  const submit = async () => {
    const input: CreateOrderInput = {
      type,
      pickup: { lat: 6.4474, lng: 3.4723, address: pickup },
      dropoff: { lat: 6.4281, lng: 3.4219, address: dropoff },
      items: [],
      idempotency_key: `mobile-${Date.now()}`,
    };
    try {
      setLoading(true);
      const order = demoMode
        ? {
            ...input,
            id: `LG-${Date.now().toString().slice(-4)}`,
            status: 'pending' as const,
            price_minor: 350000,
            currency: 'NGN',
            created_at: new Date().toISOString(),
          }
        : await api.createOrder(input);
      await Haptics.notificationAsync(Haptics.NotificationFeedbackType.Success);
      setDropoff('');
      onCreated(order);
    } catch (e) {
      Alert.alert(
        'Could not create delivery',
        e instanceof Error ? e.message : 'Please try again.',
      );
    } finally {
      setLoading(false);
    }
  };
  return (
    <Modal visible={visible} animationType="slide" transparent onRequestClose={onClose}>
      <View style={s.backdrop}>
        <SafeAreaView style={s.sheet}>
          <View style={s.handle} />
          <View style={s.header}>
            <View>
              <Text style={s.eyebrow}>NEW REQUEST</Text>
              <Text style={s.title}>Where are we going?</Text>
            </View>
            <IconButton name="close" onPress={onClose} />
          </View>
          <View style={s.types}>
            {(['parcel', 'food', 'ride'] as DeliveryType[]).map((item) => (
              <Pressable
                key={item}
                onPress={() => setType(item)}
                style={[s.type, type === item && s.typeActive]}
              >
                <Text style={[s.typeText, type === item && s.typeTextActive]}>{item}</Text>
              </Pressable>
            ))}
          </View>
          <View style={s.address}>
            <Text style={s.label}>PICKUP</Text>
            <TextInput value={pickup} onChangeText={setPickup} style={s.input} />
            <View style={s.divider} />
            <Text style={s.label}>DROPOFF</Text>
            <TextInput
              value={dropoff}
              onChangeText={setDropoff}
              style={s.input}
              placeholder="Where to?"
              placeholderTextColor="#9AA39F"
            />
          </View>
          <Pressable
            disabled={!dropoff.trim() || loading}
            onPress={submit}
            style={[s.button, (!dropoff.trim() || loading) && s.disabled]}
          >
            {loading ? (
              <ActivityIndicator color="white" />
            ) : (
              <>
                <Text style={s.buttonText}>Create delivery</Text>
                <Ionicons name="arrow-forward" size={20} color="white" />
              </>
            )}
          </Pressable>
        </SafeAreaView>
      </View>
    </Modal>
  );
}
const s = StyleSheet.create({
  backdrop: { flex: 1, backgroundColor: 'rgba(8,22,16,.35)', justifyContent: 'flex-end' },
  sheet: {
    backgroundColor: colors.canvas,
    borderTopLeftRadius: 32,
    borderTopRightRadius: 32,
    padding: 20,
  },
  handle: {
    width: 40,
    height: 5,
    borderRadius: 3,
    backgroundColor: '#C9CECA',
    alignSelf: 'center',
    marginBottom: 18,
  },
  header: { flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center' },
  eyebrow: { fontSize: 10, fontWeight: '800', letterSpacing: 1.4, color: colors.greenDark },
  title: { fontSize: 24, fontWeight: '900', color: colors.ink, marginTop: 4 },
  types: { flexDirection: 'row', gap: 9, marginTop: 20 },
  type: {
    flex: 1,
    alignItems: 'center',
    padding: 12,
    borderRadius: 14,
    backgroundColor: 'white',
    borderWidth: 1,
    borderColor: colors.line,
  },
  typeActive: { backgroundColor: colors.ink },
  typeText: { fontSize: 12, fontWeight: '800', textTransform: 'capitalize', color: colors.ink },
  typeTextActive: { color: 'white' },
  address: {
    backgroundColor: 'white',
    borderRadius: radius.lg,
    padding: 17,
    marginTop: 16,
    borderWidth: 1,
    borderColor: colors.line,
  },
  label: { fontSize: 9, fontWeight: '800', letterSpacing: 1.2, color: colors.muted },
  input: { fontSize: 15, fontWeight: '700', color: colors.ink, paddingVertical: 9 },
  divider: { height: 1, backgroundColor: colors.line, marginVertical: 6 },
  button: {
    height: 55,
    backgroundColor: colors.green,
    borderRadius: 16,
    paddingHorizontal: 18,
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    marginTop: 22,
  },
  buttonText: { fontSize: 15, fontWeight: '900', color: 'white' },
  disabled: { opacity: 0.4 },
});
