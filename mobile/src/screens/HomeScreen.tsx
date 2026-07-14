import { Pressable, ScrollView, StyleSheet, Text, View } from 'react-native';
import { LinearGradient } from 'expo-linear-gradient';
import { Ionicons } from '@expo/vector-icons';
import { DeliveryType, Order, User } from '../types';
import { colors, radius, shadow } from '../theme';
import { IconButton, Progress, SectionTitle, page } from '../components/AppUI';

export function HomeScreen({
  user,
  orders,
  onBook,
  onTrack,
  providerMode,
}: {
  user: User;
  orders: Order[];
  onBook: (t: DeliveryType) => void;
  onTrack: (o: Order) => void;
  providerMode: boolean;
}) {
  const active = orders.find((o) => !['delivered', 'cancelled', 'failed'].includes(o.status));
  const firstName = user.email.split('@')[0];
  return (
    <ScrollView showsVerticalScrollIndicator={false} contentContainerStyle={page.content}>
      <View style={s.top}>
        <View>
          <Text style={s.eyebrow}>{providerMode ? 'FIELD OPERATIONS' : 'GOOD MORNING'}</Text>
          <Text style={s.greeting}>Hello, {firstName}</Text>
        </View>
        <View style={s.actions}>
          <IconButton name="search-outline" label="Search jobs" />
          <IconButton name="notifications-outline" label="Notifications" />
        </View>
      </View>
      <LinearGradient colors={['#173D45', '#0F5C50']} style={s.hero}>
        <Text style={s.badge}>
          {providerMode ? 'DRIVER STATUS • READY' : 'DELIVERING ACROSS LAGOS'}
        </Text>
        <Text style={s.heroTitle}>
          {providerMode
            ? `Move every load safely.${`\n`}Proof at every step.`
            : `Anything, anywhere.${`\n`}We’ll get it there.`}
        </Text>
        <Text style={s.heroCopy}>
          {providerMode
            ? 'Review load handling, follow the route, and capture clear proof.'
            : 'Fast, trackable delivery with trusted riders.'}
        </Text>
        <Pressable
          accessibilityRole="button"
          accessibilityLabel={providerMode ? 'View delivery jobs' : 'Book a delivery'}
          onPress={() => (providerMode ? active && onTrack(active) : onBook('parcel'))}
          style={s.heroButton}
        >
          <Text style={s.heroButtonText}>
            {providerMode
              ? active
                ? 'Continue active job'
                : 'View delivery jobs'
              : 'Book a delivery'}
          </Text>
          <Ionicons name="arrow-forward" size={18} />
        </Pressable>
      </LinearGradient>
      {!providerMode ? (
        <>
          <SectionTitle title="What do you need?" />
          <View style={s.services}>
            {(
              [
                ['parcel', 'cube-outline', 'Send parcel', '#DDF5EA'],
                ['food', 'fast-food-outline', 'Get food', '#FFE3D2'],
                ['ride', 'car-sport-outline', 'Book ride', '#E4E6FF'],
              ] as const
            ).map(([type, icon, label, bg]) => (
              <Pressable
                accessibilityRole="button"
                accessibilityLabel={label}
                accessibilityHint="Starts a new request"
                key={type}
                onPress={() => onBook(type)}
                style={s.service}
              >
                <View style={[s.serviceIcon, { backgroundColor: bg }]}>
                  <Ionicons name={icon} size={23} color={colors.ink} />
                </View>
                <Text style={s.serviceTitle}>{label}</Text>
                <Text style={s.serviceCopy}>Fast & reliable</Text>
              </Pressable>
            ))}
          </View>
        </>
      ) : null}
      {active ? (
        <>
          <SectionTitle title="In progress" action="Track" onAction={() => onTrack(active)} />
          <Pressable
            accessibilityRole="button"
            accessibilityLabel={`Continue ${active.type} job`}
            accessibilityHint="Opens route, load details, and proof actions"
            onPress={() => onTrack(active)}
            style={s.active}
          >
            <View style={s.activeTop}>
              <View style={s.activeIcon}>
                <Ionicons name="bicycle" size={23} color={colors.greenDark} />
              </View>
              <View style={{ flex: 1 }}>
                <Text style={s.activeLabel}>
                  {active.type.toUpperCase()} • {active.id}
                </Text>
                <Text style={s.activeTitle}>{statusText(active.status)}</Text>
              </View>
              <Text style={s.eta}>12 min</Text>
            </View>
            <Progress />
            <Text numberOfLines={1} style={s.route}>
              {active.pickup.address} → {active.dropoff.address}
            </Text>
          </Pressable>
        </>
      ) : null}
    </ScrollView>
  );
}
const statusText = (v: string) => v.replaceAll('_', ' ').replace(/\b\w/g, (c) => c.toUpperCase());
const s = StyleSheet.create({
  top: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 22,
  },
  eyebrow: { fontSize: 11, fontWeight: '800', letterSpacing: 1.5, color: colors.muted },
  greeting: {
    fontSize: 27,
    fontWeight: '800',
    letterSpacing: -0.8,
    color: colors.ink,
    marginTop: 4,
  },
  actions: { flexDirection: 'row', gap: 9 },
  hero: { borderRadius: radius.lg, padding: 22, minHeight: 250, ...shadow },
  badge: { fontSize: 10, fontWeight: '800', letterSpacing: 1, color: '#D7F35D' },
  heroTitle: { fontSize: 29, lineHeight: 35, fontWeight: '800', color: 'white', marginTop: 25 },
  heroCopy: { fontSize: 14, color: '#C7DBD4', marginTop: 8 },
  heroButton: {
    marginTop: 24,
    alignSelf: 'flex-start',
    backgroundColor: colors.lime,
    borderRadius: 14,
    paddingHorizontal: 17,
    paddingVertical: 13,
    flexDirection: 'row',
    gap: 18,
  },
  heroButtonText: { fontSize: 14, fontWeight: '800', color: colors.ink },
  services: { flexDirection: 'row', gap: 10 },
  service: {
    flex: 1,
    backgroundColor: 'white',
    borderRadius: radius.md,
    padding: 12,
    borderWidth: 1,
    borderColor: colors.line,
  },
  serviceIcon: {
    width: 43,
    height: 43,
    borderRadius: 14,
    alignItems: 'center',
    justifyContent: 'center',
    marginBottom: 13,
  },
  serviceTitle: { fontSize: 13, fontWeight: '800', color: colors.ink },
  serviceCopy: { fontSize: 10, color: colors.muted, marginTop: 3 },
  active: {
    backgroundColor: 'white',
    borderRadius: radius.lg,
    padding: 17,
    borderWidth: 1,
    borderColor: colors.line,
    ...shadow,
  },
  activeTop: { flexDirection: 'row', alignItems: 'center', gap: 12 },
  activeIcon: {
    width: 45,
    height: 45,
    borderRadius: 15,
    backgroundColor: colors.greenSoft,
    alignItems: 'center',
    justifyContent: 'center',
  },
  activeLabel: { fontSize: 10, fontWeight: '800', letterSpacing: 1, color: colors.greenDark },
  activeTitle: { fontSize: 15, fontWeight: '800', color: colors.ink, marginTop: 4 },
  eta: { fontSize: 16, fontWeight: '900', color: colors.ink },
  route: { fontSize: 12, color: colors.muted, marginTop: 14 },
});
