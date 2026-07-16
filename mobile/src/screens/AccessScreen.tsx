import { useState } from 'react';
import {
  ActivityIndicator,
  Alert,
  KeyboardAvoidingView,
  Platform,
  Pressable,
  SafeAreaView,
  ScrollView,
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
  const [mode, setMode] = useState<'login' | 'register' | 'reset'>('login');
  const [fullName, setFullName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [loading, setLoading] = useState(false);
  const submit = async () => {
    if (
      !email.trim() ||
      (mode === 'register' && fullName.trim().length < 2) ||
      (mode !== 'reset' && password.length < 8)
    ) {
      Alert.alert(
        'Check your details',
        'Enter your full name, a valid email, and a password of at least 8 characters.',
      );
      return;
    }
    try {
      setLoading(true);
      if (mode === 'reset') {
        await api.requestPasswordReset(email.trim());
        Alert.alert(
          'Check your email',
          'If the account exists, password reset instructions are on the way.',
        );
        setMode('login');
        return;
      }
      if (mode === 'register') await api.register(fullName.trim(), email.trim(), password);
      await api.login(email.trim(), password);
      await Haptics.notificationAsync(Haptics.NotificationFeedbackType.Success);
      onAuthenticated();
    } catch (e) {
      Alert.alert('Could not continue', e instanceof Error ? e.message : 'Please try again.');
    } finally {
      setLoading(false);
    }
  };
  return (
    <SafeAreaView style={s.page}>
      <KeyboardAvoidingView
        style={s.keyboardView}
        behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
      >
        <ScrollView
          contentContainerStyle={s.scrollContent}
          keyboardShouldPersistTaps="handled"
          keyboardDismissMode={Platform.OS === 'ios' ? 'interactive' : 'on-drag'}
          showsVerticalScrollIndicator={false}
        >
          <LinearGradient colors={['#123C43', '#087A61']} style={s.hero}>
        <View style={s.orb} />
        <View style={s.mark}>
          <Ionicons name="navigate" size={25} color={colors.ink} />
        </View>
        <View style={s.heroBottom}>
          <Text style={s.kicker}>FARMSENSE FIELD LOGISTICS</Text>
          <Text accessibilityRole="header" style={s.title}>
            Every load,{`\n`}handled right.
          </Text>
          <Text style={s.copy}>
            Your driver workspace for assigned jobs, safe handling, route progress, and delivery
            proof.
          </Text>
        </View>
          </LinearGradient>
          <View style={s.panel}>
        <Text style={s.panelTitle}>
          {mode === 'login'
            ? 'Welcome back'
            : mode === 'register'
              ? 'Create your account'
              : 'Reset password'}
        </Text>
        <Text style={s.panelCopy}>
          {mode === 'reset'
            ? 'We’ll email you a secure reset link.'
            : mode === 'register'
              ? 'Create your provider account to receive assigned work.'
              : 'Sign in to review and complete your delivery jobs.'}
        </Text>
        {mode === 'register' ? (
          <Field
            icon="person-outline"
            value={fullName}
            onChangeText={setFullName}
            placeholder="Full name"
          />
        ) : null}
        <Field
          icon="mail-outline"
          value={email}
          onChangeText={setEmail}
          placeholder="Email address"
        />
        {mode !== 'reset' ? (
          <Field
            icon="lock-closed-outline"
            value={password}
            onChangeText={setPassword}
            placeholder="Password"
            secure
          />
        ) : null}
        {mode === 'login' ? (
          <Pressable
            accessibilityRole="button"
            accessibilityLabel="Forgot password"
            hitSlop={10}
            onPress={() => setMode('reset')}
          >
            <Text style={s.forgot}>Forgot password?</Text>
          </Pressable>
        ) : null}
        <Pressable
          accessibilityRole="button"
          accessibilityLabel={
            mode === 'login'
              ? 'Sign in'
              : mode === 'register'
                ? 'Create account'
                : 'Send reset link'
          }
          accessibilityState={{ disabled: loading }}
          onPress={submit}
          disabled={loading}
          style={({ pressed }) => [s.button, pressed && { opacity: 0.78 }]}
        >
          {loading ? (
            <ActivityIndicator color="white" />
          ) : (
            <>
              <Text style={s.buttonText}>
                {mode === 'login'
                  ? 'Sign in'
                  : mode === 'register'
                    ? 'Create account'
                    : 'Send reset link'}
              </Text>
              <Ionicons name="arrow-forward" size={20} color="white" />
            </>
          )}
        </Pressable>
        <Pressable
          accessibilityRole="button"
          accessibilityLabel={
            mode === 'register'
              ? 'Sign in instead'
              : mode === 'login'
                ? 'Create an account'
                : 'Back to sign in'
          }
          onPress={() =>
            setMode(mode === 'register' ? 'login' : mode === 'login' ? 'register' : 'login')
          }
          style={s.switchMode}
        >
          <Text style={s.switchText}>
            {mode === 'register'
              ? 'Already have an account? Sign in'
              : mode === 'login'
                ? 'New here? Create an account'
                : 'Back to sign in'}
          </Text>
        </Pressable>
        <Pressable
          accessibilityRole="button"
          accessibilityLabel="Explore provider demo"
          onPress={onDemo}
          style={s.demo}
        >
          <Text style={s.demoText}>Explore the demo</Text>
        </Pressable>
        <Text style={s.legal}>By continuing, you agree to our Terms and Privacy Policy.</Text>
          </View>
        </ScrollView>
      </KeyboardAvoidingView>
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
  const isEmail = icon === 'mail-outline';
  const [passwordVisible, setPasswordVisible] = useState(false);
  return (
    <View style={s.field}>
      <Ionicons name={icon} size={19} color={colors.muted} />
      <TextInput
        {...props}
        secureTextEntry={secure && !passwordVisible}
        autoCapitalize={isEmail || secure ? 'none' : 'words'}
        style={s.input}
        placeholderTextColor="#9AA39F"
        accessibilityLabel={props.placeholder}
        accessibilityHint={
          secure ? 'Enter your secure password' : isEmail ? 'Enter your account email' : 'Enter your full name'
        }
        autoComplete={secure ? 'current-password' : isEmail ? 'email' : 'name'}
        keyboardType={isEmail ? 'email-address' : 'default'}
      />
      {secure ? (
        <Pressable
          accessibilityRole="button"
          accessibilityLabel={passwordVisible ? 'Hide password' : 'Show password'}
          accessibilityHint={
            passwordVisible ? 'Hides the password characters' : 'Shows the password characters'
          }
          accessibilityState={{ selected: passwordVisible }}
          hitSlop={8}
          onPress={() => setPasswordVisible((visible) => !visible)}
          style={({ pressed }) => [s.passwordToggle, pressed && s.passwordTogglePressed]}
        >
          <Ionicons
            name={passwordVisible ? 'eye-off-outline' : 'eye-outline'}
            size={21}
            color={colors.muted}
          />
        </Pressable>
      ) : null}
    </View>
  );
}
const s = StyleSheet.create({
  page: { flex: 1, backgroundColor: colors.canvas },
  keyboardView: { flex: 1 },
  scrollContent: { flexGrow: 1 },
  hero: { minHeight: 390, flexGrow: 1, padding: 24, overflow: 'hidden' },
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
  passwordToggle: {
    width: 44,
    height: 44,
    marginRight: -10,
    alignItems: 'center',
    justifyContent: 'center',
  },
  passwordTogglePressed: { opacity: 0.55 },
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
  forgot: {
    alignSelf: 'flex-end',
    fontSize: 11,
    fontWeight: '700',
    color: colors.greenDark,
    marginBottom: 9,
  },
  switchMode: { alignItems: 'center', paddingTop: 13 },
  switchText: { fontSize: 12, fontWeight: '700', color: colors.ink },
  demo: { alignItems: 'center', paddingVertical: 14 },
  demoText: { fontSize: 13, fontWeight: '800', color: colors.greenDark },
  legal: { fontSize: 10, color: '#909A95', textAlign: 'center' },
});
