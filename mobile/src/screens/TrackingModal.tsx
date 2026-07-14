import { useEffect, useState } from 'react';
import { ActivityIndicator, Modal, SafeAreaView, StyleSheet, Text, View } from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import { api } from '../api';
import { demoTracking } from '../data/demo';
import { IconButton, Progress, StateCard } from '../components/AppUI';
import { Order, Tracking } from '../types';
import { colors, radius, shadow } from '../theme';

export function TrackingModal({
  order,
  demoMode,
  onClose,
}: {
  order?: Order;
  demoMode: boolean;
  onClose: () => void;
}) {
  const [data, setData] = useState<Tracking>();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string>();
  useEffect(() => {
    if (!order) return;
    if (demoMode) {
      setData({ ...demoTracking, order_id: order.id, status: order.status });
      return;
    }
    setLoading(true);
    api
      .track(order.id)
      .then(setData)
      .catch((e) => setError(e instanceof Error ? e.message : 'Tracking unavailable'))
      .finally(() => setLoading(false));
  }, [order, demoMode]);
  return (
    <Modal visible={!!order} animationType="slide" onRequestClose={onClose}>
      <SafeAreaView style={s.page}>
        <View style={s.header}>
          <IconButton name="arrow-back" onPress={onClose} />
          <Text style={s.headerTitle}>{order?.id}</Text>
          <IconButton name="ellipsis-horizontal" />
        </View>
        {loading ? (
          <ActivityIndicator style={{ flex: 1 }} color={colors.green} />
        ) : error ? (
          <View style={s.error}>
            <StateCard icon="navigate-outline" title="Tracking unavailable" copy={error} />
          </View>
        ) : (
          <>
            <View style={s.map}>
              <View style={s.road} />
              <View style={s.roadTwo} />
              <View style={s.pin}>
                <Ionicons name="bicycle" size={23} color="white" />
              </View>
              <View style={s.eta}>
                <Text style={s.etaValue}>{data?.eta_minutes ?? 12} min</Text>
                <Text style={s.etaCopy}>3.2 km away</Text>
              </View>
            </View>
            <View style={s.sheet}>
              <View style={s.handle} />
              <Text style={s.eyebrow}>ARRIVING SOON</Text>
              <Text style={s.title}>
                {(data?.status ?? order?.status ?? 'dispatching').replaceAll('_', ' ')}
              </Text>
              <Progress />
              <View style={s.rider}>
                <View style={s.avatar}>
                  <Text style={s.initial}>DA</Text>
                </View>
                <View style={{ flex: 1 }}>
                  <Text style={s.riderName}>{data?.driver?.full_name ?? 'Finding your rider'}</Text>
                  <Text style={s.riderMeta}>
                    Verified rider {data?.driver?.rating ? `• ★ ${data.driver.rating}` : ''}
                  </Text>
                </View>
                <IconButton name="call-outline" />
              </View>
              <Text style={s.dropoff}>Dropoff • {order?.dropoff.address}</Text>
            </View>
          </>
        )}
      </SafeAreaView>
    </Modal>
  );
}
const s = StyleSheet.create({
  page: { flex: 1, backgroundColor: colors.canvas },
  header: {
    height: 66,
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    paddingHorizontal: 18,
  },
  headerTitle: { fontSize: 14, fontWeight: '800', color: colors.ink },
  error: { flex: 1, justifyContent: 'center', padding: 20 },
  map: { flex: 1, backgroundColor: '#DCE8DF', overflow: 'hidden' },
  road: {
    position: 'absolute',
    width: '140%',
    height: 26,
    backgroundColor: '#F8F7F1',
    top: '38%',
    left: '-20%',
    transform: [{ rotate: '-12deg' }],
  },
  roadTwo: {
    position: 'absolute',
    width: 24,
    height: '130%',
    backgroundColor: '#F8F7F1',
    top: '-15%',
    left: '33%',
    transform: [{ rotate: '16deg' }],
  },
  pin: {
    position: 'absolute',
    left: '43%',
    top: '42%',
    width: 52,
    height: 52,
    borderRadius: 26,
    backgroundColor: colors.green,
    alignItems: 'center',
    justifyContent: 'center',
    borderWidth: 4,
    borderColor: 'white',
    ...shadow,
  },
  eta: {
    position: 'absolute',
    left: 20,
    top: 22,
    backgroundColor: 'white',
    borderRadius: 15,
    padding: 12,
    ...shadow,
  },
  etaValue: { fontSize: 18, fontWeight: '900', color: colors.ink },
  etaCopy: { fontSize: 10, color: colors.muted },
  sheet: {
    backgroundColor: colors.canvas,
    borderTopLeftRadius: 30,
    borderTopRightRadius: 30,
    padding: 20,
    marginTop: -28,
  },
  handle: {
    width: 40,
    height: 5,
    borderRadius: 3,
    backgroundColor: '#C9CECA',
    alignSelf: 'center',
    marginBottom: 18,
  },
  eyebrow: { fontSize: 10, fontWeight: '800', letterSpacing: 1.4, color: colors.greenDark },
  title: {
    fontSize: 22,
    fontWeight: '900',
    textTransform: 'capitalize',
    color: colors.ink,
    marginTop: 5,
  },
  rider: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 10,
    backgroundColor: 'white',
    borderRadius: radius.md,
    padding: 13,
    marginTop: 18,
  },
  avatar: {
    width: 46,
    height: 46,
    borderRadius: 16,
    backgroundColor: '#FFE3D2',
    alignItems: 'center',
    justifyContent: 'center',
  },
  initial: { fontWeight: '900', color: colors.ink },
  riderName: { fontSize: 14, fontWeight: '900', color: colors.ink },
  riderMeta: { fontSize: 10, color: colors.muted, marginTop: 4 },
  dropoff: { fontSize: 12, color: colors.muted, padding: 15 },
});
