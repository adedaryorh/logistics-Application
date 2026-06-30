import ReactNativeHapticFeedback from 'react-native-haptic-feedback';

const options = {
  enableVibrateFallback: true,
  ignoreAndroidSystemSettings: false,
};

export const haptics = {
  light: () => ReactNativeHapticFeedback.trigger('impactLight', options),
  medium: () => ReactNativeHapticFeedback.trigger('impactMedium', options),
  heavy: () => ReactNativeHapticFeedback.trigger('impactHeavy', options),
  success: () => ReactNativeHapticFeedback.trigger('notificationSuccess', options),
  error: () => ReactNativeHapticFeedback.trigger('notificationError', options),
  warning: () => ReactNativeHapticFeedback.trigger('notificationWarning', options),
  selection: () => ReactNativeHapticFeedback.trigger('selection', options),
  impact: () => ReactNativeHapticFeedback.trigger('impactLight', options),
  selectionSoft: () => ReactNativeHapticFeedback.trigger('selection', options),
};

export const hapticPatterns = {
  buttonPress: () => haptics.impact(),
  error: () => haptics.error(),
  success: () => haptics.success(),
  warning: () => haptics.warning(),
  selection: () => haptics.selection(),
  pullRefresh: () => haptics.light(),
  longPress: () => haptics.medium(),
};

export default haptics;
