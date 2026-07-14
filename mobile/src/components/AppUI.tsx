import { Pressable, StyleSheet, Text, View } from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import { colors, radius, shadow } from '../theme';

export function IconButton({
  name,
  onPress,
}: {
  name: keyof typeof Ionicons.glyphMap;
  onPress?: () => void;
}) {
  return (
    <Pressable
      accessibilityRole="button"
      onPress={onPress}
      style={({ pressed }) => [s.iconButton, pressed && s.pressed]}
    >
      <Ionicons name={name} size={21} color={colors.ink} />
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
        <Pressable onPress={onAction}>
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
        <Pressable onPress={onPress} style={s.stateButton}>
          <Text style={s.stateButtonText}>{action}</Text>
        </Pressable>
      ) : null}
    </View>
  );
}
export const page = StyleSheet.create({
  content: { paddingHorizontal: 20, paddingTop: 12, paddingBottom: 110 },
  title: { fontSize: 29, fontWeight: '900', letterSpacing: -1, color: colors.ink },
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
  pressed: { opacity: 0.72 },
  iconButton: {
    width: 42,
    height: 42,
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
});
