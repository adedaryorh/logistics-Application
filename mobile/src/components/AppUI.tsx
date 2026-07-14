import { Pressable, StyleSheet, Text, View } from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import { colors, layout, radius, shadow, statusTone, touch } from '../theme';

export function IconButton({
  name,
  onPress,
  label,
}: {
  name: keyof typeof Ionicons.glyphMap;
  onPress?: () => void;
  label?: string;
}) {
  return (
    <Pressable
      accessibilityRole="button"
      accessibilityLabel={label}
      accessibilityHint={onPress ? 'Activates this control' : undefined}
      hitSlop={8}
      onPress={onPress}
      style={({ pressed }) => [s.iconButton, pressed && s.pressed]}
    >
      <Ionicons name={name} size={21} color={colors.ink} />
    </Pressable>
  );
}
export function StatusBadge({ status }: { status: keyof typeof statusTone }) {
  const [label, foreground, background] = statusTone[status];
  return (
    <View
      accessible
      accessibilityLabel={`Status: ${label}`}
      style={[s.badge, { backgroundColor: background }]}
    >
      <View style={[s.dot, { backgroundColor: foreground }]} />
      <Text style={[s.badgeText, { color: foreground }]}>{label}</Text>
    </View>
  );
}
export function Notice({
  tone = 'info',
  title,
  copy,
}: {
  tone?: 'info' | 'warning' | 'danger';
  title: string;
  copy: string;
}) {
  const icon =
    tone === 'danger' ? 'alert-circle' : tone === 'warning' ? 'warning' : 'information-circle';
  const fg = tone === 'danger' ? colors.danger : tone === 'warning' ? colors.amber : colors.info;
  const bg =
    tone === 'danger' ? colors.dangerSoft : tone === 'warning' ? colors.amberSoft : colors.infoSoft;
  return (
    <View accessibilityRole="alert" style={[s.notice, { backgroundColor: bg }]}>
      <Ionicons name={icon} size={22} color={fg} />
      <View style={{ flex: 1 }}>
        <Text style={[s.noticeTitle, { color: fg }]}>{title}</Text>
        <Text style={s.noticeCopy}>{copy}</Text>
      </View>
    </View>
  );
}
export function PrimaryButton({
  label,
  icon = 'arrow-forward',
  onPress,
  disabled = false,
  danger = false,
}: {
  label: string;
  icon?: keyof typeof Ionicons.glyphMap;
  onPress: () => void;
  disabled?: boolean;
  danger?: boolean;
}) {
  return (
    <Pressable
      accessibilityRole="button"
      accessibilityLabel={label}
      accessibilityState={{ disabled }}
      disabled={disabled}
      onPress={onPress}
      style={({ pressed }) => [
        s.primary,
        danger && s.primaryDanger,
        disabled && s.disabled,
        pressed && s.pressed,
      ]}
    >
      <Text style={s.primaryText}>{label}</Text>
      <Ionicons name={icon} size={21} color="white" />
    </Pressable>
  );
}
export function SectionTitle({
  title,
  action,
  onAction,
}: {
  title: string;
  action?: string;
  onAction?: () => void;
}) {
  return (
    <View style={s.section}>
      <Text style={s.heading}>{title}</Text>
      {action ? (
        <Pressable
          accessibilityRole="button"
          accessibilityLabel={action}
          hitSlop={10}
          onPress={onAction}
        >
          <Text style={s.action}>{action}</Text>
        </Pressable>
      ) : null}
    </View>
  );
}
export function Progress({ value = 0.68 }: { value?: number }) {
  return (
    <View style={s.track}>
      <View style={[s.fill, { width: `${Math.max(0, Math.min(1, value)) * 100}%` }]} />
    </View>
  );
}
export function StateCard({
  icon,
  title,
  copy,
  action,
  onPress,
}: {
  icon: keyof typeof Ionicons.glyphMap;
  title: string;
  copy: string;
  action?: string;
  onPress?: () => void;
}) {
  return (
    <View style={s.state}>
      <View style={s.stateIcon}>
        <Ionicons name={icon} size={24} color={colors.greenDark} />
      </View>
      <Text style={s.stateTitle}>{title}</Text>
      <Text style={s.stateCopy}>{copy}</Text>
      {action ? (
        <Pressable
          accessibilityRole="button"
          accessibilityLabel={action}
          onPress={onPress}
          style={s.stateButton}
        >
          <Text style={s.stateButtonText}>{action}</Text>
        </Pressable>
      ) : null}
    </View>
  );
}
export const page = StyleSheet.create({
  content: {
    width: '100%',
    maxWidth: layout.maxContent,
    alignSelf: 'center',
    paddingHorizontal: 20,
    paddingTop: 12,
    paddingBottom: 118,
  },
  title: {
    fontSize: 29,
    lineHeight: 36,
    fontWeight: '900',
    letterSpacing: -0.6,
    color: colors.ink,
  },
  header: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    marginBottom: 22,
  },
  card: {
    backgroundColor: colors.surface,
    borderRadius: radius.lg,
    borderWidth: 1,
    borderColor: colors.line,
    ...shadow,
  },
});
const s = StyleSheet.create({
  pressed: { opacity: 0.76, transform: [{ scale: 0.985 }] },
  iconButton: {
    width: touch.minimum,
    height: touch.minimum,
    borderRadius: 21,
    backgroundColor: colors.surface,
    alignItems: 'center',
    justifyContent: 'center',
    borderWidth: 1,
    borderColor: colors.line,
  },
  section: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginTop: 28,
    marginBottom: 13,
  },
  heading: { fontSize: 19, fontWeight: '800', letterSpacing: -0.4, color: colors.ink },
  action: { fontSize: 13, fontWeight: '700', color: colors.greenDark },
  track: {
    height: 5,
    borderRadius: 4,
    backgroundColor: '#E7ECE8',
    marginTop: 16,
    overflow: 'hidden',
  },
  fill: { height: '100%', borderRadius: 4, backgroundColor: colors.green },
  state: {
    alignItems: 'center',
    padding: 28,
    backgroundColor: colors.surface,
    borderRadius: radius.lg,
    borderWidth: 1,
    borderColor: colors.line,
  },
  stateIcon: {
    width: 52,
    height: 52,
    borderRadius: 18,
    backgroundColor: colors.greenSoft,
    alignItems: 'center',
    justifyContent: 'center',
  },
  stateTitle: { fontSize: 17, fontWeight: '900', color: colors.ink, marginTop: 14 },
  stateCopy: {
    fontSize: 12,
    lineHeight: 18,
    color: colors.muted,
    textAlign: 'center',
    marginTop: 5,
  },
  stateButton: {
    backgroundColor: colors.ink,
    borderRadius: 12,
    paddingHorizontal: 16,
    paddingVertical: 10,
    marginTop: 15,
  },
  stateButtonText: { fontSize: 12, fontWeight: '800', color: colors.surface },
  badge: {
    minHeight: 30,
    alignSelf: 'flex-start',
    borderRadius: 999,
    paddingHorizontal: 10,
    flexDirection: 'row',
    alignItems: 'center',
    gap: 6,
  },
  dot: { width: 8, height: 8, borderRadius: 4 },
  badgeText: { fontSize: 12, fontWeight: '800' },
  notice: {
    flexDirection: 'row',
    gap: 10,
    padding: 14,
    borderRadius: radius.md,
    alignItems: 'flex-start',
  },
  noticeTitle: { fontSize: 14, fontWeight: '900' },
  noticeCopy: { fontSize: 13, lineHeight: 19, color: colors.ink, marginTop: 2 },
  primary: {
    minHeight: touch.primary,
    borderRadius: radius.md,
    paddingHorizontal: 18,
    backgroundColor: colors.green,
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
  },
  primaryDanger: { backgroundColor: colors.danger },
  primaryText: { fontSize: 16, fontWeight: '900', color: 'white' },
  disabled: { opacity: 0.48 },
});
