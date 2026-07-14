import { useState } from 'react';
import {
  ActivityIndicator,
  Alert,
  Pressable,
  SafeAreaView,
  StyleSheet,
  Text,
  TextInput,
  View,
} from 'react-native';
import { LinearGradient } from 'expo-linear-gradient';
import { Ionicons } from '@expo/vector-icons';
import * as Haptics from 'expo-haptics';
import { api } from '../api';
import { colors } from '../theme';

export function AccessScreen({
  onAuthenticated,
  onDemo,
}: {
  onAuthenticated: () => void;
  onDemo: () => void;
}) {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [loading, setLoading] = useState(false);
  const signIn = async () => {
    if (!email.trim() || password.length < 8) {
      Alert.alert(
        'Check your details',
        'Enter a valid email and a password of at least 8 characters.',
      );
      return;
    }
    try {
      setLoading(true);
      await api.login(email.trim(), password);
      await Haptics.notificationAsync(Haptics.NotificationFeedbackType.Success);
      onAuthenticated();
    } catch (e) {
      Alert.alert('Could not sign in', e instanceof Error ? e.message : 'Please try again.');
    } finally {
      setLoading(false);
    }
  };
  return (
    <SafeAreaView style={s.page}>
      <LinearGradient colors={['#123C43', '#087A61']} style={s.hero}>
        <View style={s.orb} />
        <View style={s.mark}>
          <Ionicons name="navigate" size={25} color={colors.ink} />
        </View>
        <View style={s.heroBottom}>
          <Text style={s.kicker}>MOVE WITH CONFIDENCE</Text>
          <Text style={s.title}>Your city,{`\n`}within reach.</Text>
          <Text style={s.copy}>
            Send parcels, order favourites, and track every trip from one beautiful place.
          </Text>
        </View>
      </LinearGradient>
      <View style={s.panel}>
        <Text style={s.panelTitle}>Welcome back</Text>
        <Text style={s.panelCopy}>Sign in to keep things moving.</Text>
        <Field
          icon="mail-outline"
          value={email}
          onChangeText={setEmail}
          placeholder="Email address"
        />
        <Field
          icon="lock-closed-outline"
          value={password}
          onChangeText={setPassword}
          placeholder="Password"
          secure
        />
        <Pressable onPress={signIn} disabled={loading} style={s.button}>
          {loading ? (
            <ActivityIndicator color="white" />
          ) : (
            <>
              <Text style={s.buttonText}>Sign in</Text>
              <Ionicons name="arrow-forward" size={20} color="white" />
            </>
          )}
        </Pressable>
        <Pressable onPress={onDemo} style={s.demo}>
          <Text style={s.demoText}>Explore the demo</Text>
        </Pressable>
        <Text style={s.legal}>By continuing, you agree to our Terms and Privacy Policy.</Text>
      </View>
    </SafeAreaView>
  );
}
function Field({
  icon,
  secure,
  ...props
}: {
  icon: keyof typeof Ionicons.glyphMap;
  secure?: boolean;
  value: string;
  onChangeText: (v: string) => void;
  placeholder: string;
}) {
  return (
    <View style={s.field}>
      <Ionicons name={icon} size={19} color={colors.muted} />
      <TextInput
        {...props}
        secureTextEntry={secure}
        autoCapitalize="none"
        style={s.input}
        placeholderTextColor="#9AA39F"
      />
    </View>
  );
}
const s = StyleSheet.create({
  page: { flex: 1, backgroundColor: colors.canvas },
  hero: { flex: 1, padding: 24, overflow: 'hidden' },
  orb: {
    position: 'absolute',
    width: 320,
    height: 320,
    borderRadius: 160,
    backgroundColor: 'rgba(215,243,93,.13)',
    right: -120,
    top: -70,
  },
  mark: {
    width: 52,
    height: 52,
    borderRadius: 17,
    backgroundColor: colors.lime,
    alignItems: 'center',
    justifyContent: 'center',
  },
  heroBottom: { marginTop: 'auto', paddingBottom: 46 },
  kicker: { fontSize: 10, fontWeight: '900', letterSpacing: 1.7, color: colors.lime },
  title: {
    fontSize: 43,
    lineHeight: 47,
    fontWeight: '900',
    letterSpacing: -1.8,
    color: 'white',
    marginTop: 12,
  },
  copy: { fontSize: 14, lineHeight: 21, color: '#C6DDD5', marginTop: 14, maxWidth: 320 },
  panel: {
    backgroundColor: colors.canvas,
    borderTopLeftRadius: 32,
    borderTopRightRadius: 32,
    paddingHorizontal: 22,
    paddingTop: 25,
    paddingBottom: 18,
    marginTop: -28,
  },
  panelTitle: { fontSize: 24, fontWeight: '900', color: colors.ink },
  panelCopy: { fontSize: 13, color: colors.muted, marginTop: 4, marginBottom: 17 },
  field: {
    height: 54,
    flexDirection: 'row',
    alignItems: 'center',
    gap: 11,
    backgroundColor: 'white',
    borderRadius: 15,
    borderWidth: 1,
    borderColor: colors.line,
    paddingHorizontal: 15,
    marginBottom: 10,
  },
  input: { flex: 1, fontSize: 14, fontWeight: '600', color: colors.ink },
  button: {
    height: 55,
    backgroundColor: colors.green,
    borderRadius: 16,
    paddingHorizontal: 18,
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    marginTop: 4,
  },
  buttonText: { fontSize: 15, fontWeight: '900', color: 'white' },
  demo: { alignItems: 'center', paddingVertical: 14 },
  demoText: { fontSize: 13, fontWeight: '800', color: colors.greenDark },
  legal: { fontSize: 10, color: '#909A95', textAlign: 'center' },
});
