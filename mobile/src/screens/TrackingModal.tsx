import { useEffect, useState } from 'react';
import {
  ActivityIndicator,
  Modal,
  Pressable,
  SafeAreaView,
  StyleSheet,
  Text,
  TextInput,
  View,
} from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import MapView, { Marker } from 'react-native-maps';
import { api } from '../api';
import { demoTracking } from '../data/demo';
import {
  IconButton,
  Notice,
  PrimaryButton,
  Progress,
  StateCard,
  StatusBadge,
} from '../components/AppUI';
import { Order, Tracking } from '../types';
import { colors, radius, shadow } from '../theme';

export function TrackingModal({
  order,
  demoMode,
  providerMode,
  onClose,
}: {
  order?: Order;
  demoMode: boolean;
  providerMode: boolean;
  onClose: () => void;
}) {
  const [data, setData] = useState<Tracking>();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string>();
  const [evidenceURL, setEvidenceURL] = useState('');
  const [recipientName, setRecipientName] = useState('');
  const [savingProof, setSavingProof] = useState(false);
  useEffect(() => {
    if (!order) return;
    if (demoMode) {
      setData({ ...demoTracking, order_id: order.id, status: order.status });
      return;
    }
    setLoading(true);
    let close: undefined | (() => void);
    api
      .track(order.id)
      .then(setData)
      .then(() =>
        api.subscribeToOrder(
          order.id,
          (update) => {
            setData((current) => ({ ...current, ...update }));
            setError(undefined);
          },
          setError,
        ),
      )
      .then((stop) => {
        close = stop;
      })
      .catch((e) => setError(e instanceof Error ? e.message : 'Tracking unavailable'))
      .finally(() => setLoading(false));
    return () => close?.();
  }, [order, demoMode]);
  const currentStatus = data?.status ?? order?.status;
  return (
    <Modal visible={!!order} animationType="slide" onRequestClose={onClose}>
      <SafeAreaView style={s.page}>
        <View style={s.header}>
          <IconButton name="arrow-back" label="Close job details" onPress={onClose} />
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
              <MapView
                style={StyleSheet.absoluteFill}
                initialRegion={{
                  latitude: order?.pickup.lat ?? 6.4474,
                  longitude: order?.pickup.lng ?? 3.4723,
                  latitudeDelta: 0.08,
                  longitudeDelta: 0.08,
                }}
              >
                {order ? (
                  <>
                    <Marker
                      coordinate={{ latitude: order.pickup.lat, longitude: order.pickup.lng }}
                      title="Pickup"
                    />
                    <Marker
                      coordinate={{ latitude: order.dropoff.lat, longitude: order.dropoff.lng }}
                      title="Dropoff"
                      pinColor={colors.greenDark}
                    />
                  </>
                ) : null}
              </MapView>
              <View style={s.eta}>
                <Text style={s.etaValue}>{data?.eta_minutes ?? 12} min</Text>
                <Text style={s.etaCopy}>3.2 km away</Text>
              </View>
            </View>
            <View style={s.sheet}>
              <View style={s.handle} />
              <Text style={s.eyebrow}>
                {providerMode ? 'ACTIVE DELIVERY JOB' : 'DELIVERY PROGRESS'}
              </Text>
              <Text style={s.title}>
                {(data?.status ?? order?.status ?? 'dispatching').replaceAll('_', ' ')}
              </Text>
              {currentStatus ? <StatusBadge status={currentStatus} /> : null}
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
              {order?.agricultural_shipment ? (
                <View style={s.jobCard}>
                  <Text style={s.jobTitle}>Agricultural load</Text>
                  <Text style={s.jobLine}>
                    {order.agricultural_shipment.produce_type} •{' '}
                    {order.agricultural_shipment.quantity}{' '}
                    {order.agricultural_shipment.quantity_unit} •{' '}
                    {order.agricultural_shipment.packaging}
                  </Text>
                  <Text style={s.jobLine}>
                    {order.agricultural_shipment.requires_refrigeration
                      ? `Cold chain ${order.agricultural_shipment.cold_chain_min_c}–${order.agricultural_shipment.cold_chain_max_c}°C`
                      : 'No refrigeration required'}
                  </Text>
                  <Text style={s.jobLine}>
                    Pickup{' '}
                    {new Date(
                      order.agricultural_shipment.pickup_window.start_at,
                    ).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                    –
                    {new Date(order.agricultural_shipment.pickup_window.end_at).toLocaleTimeString(
                      [],
                      { hour: '2-digit', minute: '2-digit' },
                    )}
                  </Text>
                  {order.agricultural_shipment.handling_notes ? (
                    <Text style={s.jobNote}>
                      Handle: {order.agricultural_shipment.handling_notes}
                    </Text>
                  ) : null}
                  {order.agricultural_shipment.loading_notes ? (
                    <Text style={s.jobNote}>
                      Loading: {order.agricultural_shipment.loading_notes}
                    </Text>
                  ) : null}
                </View>
              ) : null}
              {providerMode &&
              order?.agricultural_shipment &&
              (currentStatus === 'assigned' || currentStatus === 'picked_up') ? (
                <View style={s.proofCard}>
                  <Text style={s.jobTitle}>
                    {currentStatus === 'assigned' ? 'Record pickup proof' : 'Record delivery proof'}
                  </Text>
                  <View style={s.checklist}>
                    <Check text="Load matches job quantity and packaging" />
                    <Check text="Photo clearly shows load and location" />
                    <Check text="Handling and cold-chain requirements checked" />
                  </View>
                  {error ? <Notice tone="danger" title="Proof not saved" copy={error} /> : null}
                  <TextInput
                    accessibilityLabel="Evidence upload URL"
                    accessibilityHint="Paste the durable URL returned after uploading the proof photo"
                    value={evidenceURL}
                    onChangeText={setEvidenceURL}
                    placeholder="Evidence upload URL"
                    autoCapitalize="none"
                    style={s.input}
                  />
                  {currentStatus === 'picked_up' ? (
                    <TextInput
                      accessibilityLabel="Recipient name"
                      value={recipientName}
                      onChangeText={setRecipientName}
                      placeholder="Recipient name"
                      style={s.input}
                    />
                  ) : null}
                  <PrimaryButton
                    disabled={
                      savingProof ||
                      !evidenceURL ||
                      (currentStatus === 'picked_up' && !recipientName)
                    }
                    label={
                      savingProof
                        ? 'Saving proof…'
                        : currentStatus === 'assigned'
                          ? 'Confirm pickup'
                          : 'Confirm delivery'
                    }
                    icon={
                      currentStatus === 'assigned' ? 'camera-outline' : 'checkmark-circle-outline'
                    }
                    onPress={async () => {
                      if (!order) return;
                      setSavingProof(true);
                      setError(undefined);
                      try {
                        const updated = await api.recordProof(
                          order.id,
                          currentStatus === 'assigned' ? 'pickup' : 'delivery',
                          {
                            evidence_url: evidenceURL,
                            recipient_name: recipientName,
                            coordinate: currentStatus === 'assigned' ? order.pickup : order.dropoff,
                          },
                        );
                        setData((current) => ({ ...current, status: updated.status }));
                        setEvidenceURL('');
                        setRecipientName('');
                      } catch (e) {
                        setError(e instanceof Error ? e.message : 'Could not save proof');
                      } finally {
                        setSavingProof(false);
                      }
                    }}
                  />
                </View>
              ) : null}
            </View>
          </>
        )}
      </SafeAreaView>
    </Modal>
  );
}
function Check({ text }: { text: string }) {
  return (
    <View style={s.check}>
      <Ionicons name="checkmark-circle" size={20} color={colors.greenDark} />
      <Text style={s.checkText}>{text}</Text>
    </View>
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
  jobCard: { backgroundColor: '#EEF6EC', borderRadius: radius.md, padding: 14, gap: 5 },
  proofCard: {
    backgroundColor: 'white',
    borderRadius: radius.md,
    padding: 14,
    marginTop: 10,
    gap: 9,
  },
  jobTitle: { fontSize: 13, fontWeight: '900', color: colors.ink },
  jobLine: { fontSize: 12, color: colors.ink },
  jobNote: { fontSize: 11, color: colors.muted },
  input: {
    borderWidth: 1,
    borderColor: colors.line,
    borderRadius: 12,
    paddingHorizontal: 12,
    height: 44,
    color: colors.ink,
  },
  checklist: { gap: 7, paddingVertical: 4 },
  check: { flexDirection: 'row', alignItems: 'center', gap: 8 },
  checkText: { flex: 1, fontSize: 13, lineHeight: 18, color: colors.ink },
});
