import { Pressable, ScrollView, StyleSheet, Text, View } from 'react-native';
import { LinearGradient } from 'expo-linear-gradient';
import { Ionicons } from '@expo/vector-icons';
import { IconButton, SectionTitle, page } from '../components/AppUI';
import { colors, radius, shadow } from '../theme';
import { Wallet, WalletEntry } from '../types';

export function WalletScreen({
  wallet,
  ledger,
  loading,
  error,
  onRefresh,
}: {
  wallet: Wallet;
  ledger: WalletEntry[];
  loading: boolean;
  error?: string;
  onRefresh: () => void;
}) {
  return (
    <ScrollView contentContainerStyle={page.content}>
      <View style={page.header}>
        <Text accessibilityRole="header" style={page.title}>
          Earnings
        </Text>
        <IconButton name="help-circle-outline" label="Earnings help" />
      </View>
      <LinearGradient colors={['#183F46', '#0B7560']} style={s.card}>
        <Text style={s.label}>AVAILABLE TO WITHDRAW</Text>
        <Text style={s.balance}>
          {new Intl.NumberFormat('en-NG', {
            style: 'currency',
            currency: wallet.currency,
            maximumFractionDigits: 0,
          }).format(wallet.balance_minor / 100)}
        </Text>
        <Text style={s.hint}>FarmSense provider earnings</Text>
      </LinearGradient>
      <View style={s.actions}>
        {(
          [
            ['arrow-down-circle-outline', 'Withdraw'],
            ['calendar-outline', 'Statements'],
            ['receipt-outline', 'Job history'],
          ] as const
        ).map(([icon, label]) => (
          <Pressable
            accessibilityRole="button"
            accessibilityLabel={label}
            key={label}
            style={({ pressed }) => [s.action, pressed && { opacity: 0.72 }]}
          >
            <View style={s.actionIcon}>
              <Ionicons name={icon} size={22} color={colors.greenDark} />
            </View>
            <Text style={s.actionText}>{label}</Text>
          </Pressable>
        ))}
      </View>
      <SectionTitle title="Recent activity" />
      <View style={s.empty}>
        <Ionicons name="receipt-outline" size={28} color={colors.muted} />
        <Text style={s.emptyTitle}>
          {loading
            ? 'Loading wallet…'
            : error
              ? 'Wallet unavailable'
              : ledger.length
                ? `${ledger.length} recent transaction${ledger.length === 1 ? '' : 's'}`
                : 'No wallet activity'}
        </Text>
        <Text style={s.emptyCopy}>
          {error ?? 'Completed-job earnings and withdrawals will appear here.'}
        </Text>
        {error ? (
          <Pressable
            accessibilityRole="button"
            accessibilityLabel="Retry earnings"
            hitSlop={12}
            onPress={onRefresh}
          >
            <Text style={s.retry}>Try again</Text>
          </Pressable>
        ) : null}
      </View>
    </ScrollView>
  );
}
const s = StyleSheet.create({
  card: { borderRadius: radius.lg, padding: 22, minHeight: 190, ...shadow },
  label: { fontSize: 10, fontWeight: '800', letterSpacing: 1.5, color: '#BFD7D0' },
  balance: { fontSize: 38, fontWeight: '900', letterSpacing: -1.4, color: 'white', marginTop: 15 },
  kobo: { fontSize: 21, color: '#C4D9D3' },
  hint: { fontSize: 12, color: '#C4D9D3', marginTop: 'auto' },
  actions: { flexDirection: 'row', justifyContent: 'space-around', marginTop: 20 },
  action: { alignItems: 'center', gap: 8 },
  actionIcon: {
    width: 50,
    height: 50,
    borderRadius: 17,
    backgroundColor: colors.greenSoft,
    alignItems: 'center',
    justifyContent: 'center',
  },
  actionText: { fontSize: 12, fontWeight: '700', color: colors.ink },
  empty: {
    alignItems: 'center',
    backgroundColor: 'white',
    borderRadius: radius.lg,
    padding: 25,
    borderWidth: 1,
    borderColor: colors.line,
  },
  emptyTitle: { fontSize: 15, fontWeight: '800', color: colors.ink, marginTop: 9 },
  emptyCopy: { fontSize: 11, color: colors.muted, marginTop: 4 },
  retry: { fontSize: 12, fontWeight: '800', color: colors.greenDark, marginTop: 10 },
});
