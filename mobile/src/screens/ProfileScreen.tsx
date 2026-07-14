import { Alert, Pressable, ScrollView, StyleSheet, Text, View } from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import { IconButton, page } from '../components/AppUI';
import { User } from '../types';
import { colors, radius } from '../theme';

export function ProfileScreen({
  user,
  demoMode,
  onLogout,
}: {
  user: User;
  demoMode: boolean;
  onLogout: () => void;
}) {
  const menu = [
    ['person-outline', 'Personal details'],
    ['document-text-outline', 'Driver documents'],
    ['car-outline', 'Vehicle and capacity'],
    ['notifications-outline', 'Job notifications'],
    ['shield-checkmark-outline', 'Safety & privacy'],
    ['headset-outline', 'Help centre'],
  ] as const;
  return (
    <ScrollView contentContainerStyle={page.content}>
      <View style={page.header}>
        <Text style={page.title}>Profile</Text>
        <IconButton name="settings-outline" />
      </View>
      <View style={s.profile}>
        <View style={s.avatar}>
          <Text style={s.initial}>{user.email.slice(0, 2).toUpperCase()}</Text>
        </View>
        <View style={{ flex: 1 }}>
          <Text style={s.name}>{user.email.split('@')[0]}</Text>
          <Text style={s.email}>{user.email}</Text>
          <Text style={s.verified}>
            {demoMode ? 'Demo provider' : `Verified ${user.role ?? 'provider'}`}
          </Text>
        </View>
      </View>
      <View style={s.list}>
        {menu.map(([icon, label]) => (
          <Pressable
            accessibilityRole="button"
            accessibilityLabel={label}
            accessibilityHint="Opens account settings"
            key={label}
            style={({ pressed }) => [s.row, pressed && { opacity: 0.72 }]}
          >
            <View style={s.icon}>
              <Ionicons name={icon} size={20} color={colors.ink} />
            </View>
            <Text style={s.menu}>{label}</Text>
            <Ionicons name="chevron-forward" size={19} color="#A1AAA6" />
          </Pressable>
        ))}
      </View>
      <Pressable
        accessibilityRole="button"
        accessibilityLabel="Sign out"
        onPress={() =>
          Alert.alert(
            'Sign out?',
            'Saved job data will remain on this device, but live updates will stop.',
            [
              { text: 'Stay signed in', style: 'cancel' },
              { text: 'Sign out', style: 'destructive', onPress: onLogout },
            ],
          )
        }
        style={s.logout}
      >
        <Ionicons name="log-out-outline" size={20} color={colors.danger} />
        <Text style={s.logoutText}>Sign out</Text>
      </Pressable>
    </ScrollView>
  );
}
const s = StyleSheet.create({
  profile: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 16,
    backgroundColor: 'white',
    borderRadius: radius.lg,
    padding: 18,
    borderWidth: 1,
    borderColor: colors.line,
    marginBottom: 18,
  },
  avatar: {
    width: 64,
    height: 64,
    borderRadius: 22,
    backgroundColor: colors.lime,
    alignItems: 'center',
    justifyContent: 'center',
  },
  initial: { fontSize: 21, fontWeight: '900', color: colors.ink },
  name: { fontSize: 19, fontWeight: '900', textTransform: 'capitalize', color: colors.ink },
  email: { fontSize: 12, color: colors.muted, marginTop: 3 },
  verified: { fontSize: 11, fontWeight: '700', color: colors.greenDark, marginTop: 8 },
  list: {
    backgroundColor: 'white',
    borderRadius: radius.lg,
    paddingHorizontal: 16,
    borderWidth: 1,
    borderColor: colors.line,
  },
  row: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 13,
    paddingVertical: 15,
    borderBottomWidth: 1,
    borderBottomColor: colors.line,
  },
  icon: {
    width: 38,
    height: 38,
    borderRadius: 12,
    backgroundColor: '#F0F1EC',
    alignItems: 'center',
    justifyContent: 'center',
  },
  menu: { flex: 1, fontSize: 14, fontWeight: '700', color: colors.ink },
  logout: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    gap: 8,
    padding: 16,
    marginTop: 18,
  },
  logoutText: { fontSize: 13, fontWeight: '800', color: colors.danger },
});
