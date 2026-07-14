import { Pressable, ScrollView, StyleSheet, Text, View } from 'react-native';
import { LinearGradient } from 'expo-linear-gradient';
import { Ionicons } from '@expo/vector-icons';
import { IconButton, SectionTitle, page } from '../components/AppUI';
import { colors, radius, shadow } from '../theme';

export function WalletScreen() {
  return (
    <ScrollView contentContainerStyle={page.content}>
      <View style={page.header}>
        <Text style={page.title}>Wallet</Text>
        <IconButton name="help-circle-outline" />
      </View>
      <LinearGradient colors={['#183F46', '#0B7560']} style={s.card}>
        <Text style={s.label}>AVAILABLE BALANCE</Text>
        <Text style={s.balance}>
          ₦24,600<Text style={s.kobo}>.00</Text>
        </Text>
        <Text style={s.hint}>Logistics wallet</Text>
      </LinearGradient>
      <View style={s.actions}>
        {(
          [
            ['add-circle-outline', 'Add money'],
            ['send-outline', 'Send'],
            ['receipt-outline', 'History'],
          ] as const
        ).map(([icon, label]) => (
          <Pressable key={label} style={s.action}>
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
        <Text style={s.emptyTitle}>No wallet activity</Text>
        <Text style={s.emptyCopy}>Payments and refunds will appear here.</Text>
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
});
