import React, { useState, useEffect, useRef } from 'react';
import { View, Text, TouchableOpacity, StyleSheet, Dimensions, Platform } from 'react-native';
import Animated, {
  useSharedValue,
  useAnimatedStyle,
  withSpring,
  withTiming,
  runOnJS,
} from 'react-native-reanimated';
import { Gesture, GestureDetector } from 'react-native-gesture-handler';
import { CheckCircle2, XCircle, AlertTriangle, Info, X } from 'lucide-react-native';
import { colors } from '../colors';

export type ToastType = 'success' | 'error' | 'warning' | 'info';

export interface ToastData {
  id: string;
  title: string;
  message?: string;
  type: ToastType;
  duration?: number;
}

type ToastListener = (toasts: ToastData[]) => void;
const listeners = new Set<ToastListener>();
let activeToasts: ToastData[] = [];

const notifyListeners = () => {
  listeners.forEach((listener) => listener([...activeToasts]));
};

// Global Toast Controller
export const toast = {
  show: (title: string, message?: string, type: ToastType = 'info', duration = 3000) => {
    const id = Math.random().toString(36).substring(7);
    const newToast: ToastData = { id, title, message, type, duration };
    activeToasts = [...activeToasts, newToast];
    notifyListeners();
    return id;
  },
  success: (title: string, message?: string, duration?: number) =>
    toast.show(title, message, 'success', duration),
  error: (title: string, message?: string, duration?: number) =>
    toast.show(title, message, 'error', duration),
  warning: (title: string, message?: string, duration?: number) =>
    toast.show(title, message, 'warning', duration),
  info: (title: string, message?: string, duration?: number) =>
    toast.show(title, message, 'info', duration),
  dismiss: (id: string) => {
    activeToasts = activeToasts.filter((t) => t.id !== id);
    notifyListeners();
  },
};

// Toast Container component to put at the top of App hierarchy
export const ToastContainer: React.FC = () => {
  const [toasts, setToasts] = useState<ToastData[]>([]);

  useEffect(() => {
    const handleUpdate = (updatedToasts: ToastData[]) => {
      setToasts(updatedToasts);
    };
    listeners.add(handleUpdate);
    // Initial sync
    setToasts(activeToasts);
    return () => {
      listeners.delete(handleUpdate);
    };
  }, []);

  return (
    <View style={styles.container} pointerEvents="box-none">
      {toasts.map((item, index) => (
        <ToastItem key={item.id} data={item} index={index} />
      ))}
    </View>
  );
};

// Individual Toast Item component with Reanimated + Gesture Handler
const ToastItem: React.FC<{ data: ToastData; index: number }> = ({ data, index }) => {
  const translateY = useSharedValue(-100);
  const opacity = useSharedValue(0);
  const translateX = useSharedValue(0);

  const isDismissed = useRef(false);

  const dismiss = () => {
    if (isDismissed.current) return;
    isDismissed.current = true;
    opacity.value = withTiming(0, { duration: 200 });
    translateY.value = withTiming(-100, { duration: 250 }, () => {
      runOnJS(toast.dismiss)(data.id);
    });
  };

  useEffect(() => {
    // Entrance Animation
    translateY.value = withSpring(0, { damping: 12 });
    opacity.value = withTiming(1, { duration: 200 });

    // Auto dismiss
    const timer = setTimeout(() => {
      dismiss();
    }, data.duration || 3000);

    return () => clearTimeout(timer);
  }, []);

  // Gesture setup for swiping left/right to dismiss
  const panGesture = Gesture.Pan()
    .onChange((event) => {
      translateX.value = event.translationX;
    })
    .onEnd((event) => {
      if (Math.abs(event.velocityX) > 500 || Math.abs(translateX.value) > 150) {
        // Swipe to dismiss
        translateX.value = withTiming(
          translateX.value > 0 ? Dimensions.get('window').width : -Dimensions.get('window').width,
          { duration: 200 },
          () => {
            runOnJS(dismiss)();
          }
        );
      } else {
        // Snap back
        translateX.value = withSpring(0);
      }
    });

  const animatedStyle = useAnimatedStyle(() => {
    return {
      transform: [
        { translateY: translateY.value },
        { translateX: translateX.value },
      ],
      opacity: opacity.value,
    };
  });

  // Derived icon and styles
  let typeColor: string = colors.info;
  let IconComponent = Info;

  switch (data.type) {
    case 'success':
      typeColor = colors.success;
      IconComponent = CheckCircle2;
      break;
    case 'error':
      typeColor = colors.error;
      IconComponent = XCircle;
      break;
    case 'warning':
      typeColor = colors.warning;
      IconComponent = AlertTriangle;
      break;
    case 'info':
      typeColor = colors.info;
      IconComponent = Info;
      break;
  }

  return (
    <GestureDetector gesture={panGesture}>
      <Animated.View
        style={[styles.toastItem, animatedStyle, { borderLeftColor: typeColor }]}
        className="flex-row items-start bg-white border border-gray-200 p-3.5 shadow-lg rounded-xl mb-3"
      >
        <View className="mr-3 mt-0.5">
          <IconComponent size={20} color={typeColor} />
        </View>

        <View className="flex-1 min-w-0">
          <Text className="text-sm font-semibold text-gray-900 leading-tight">
            {data.title}
          </Text>
          {data.message && (
            <Text className="mt-0.5 text-xs text-gray-500 leading-normal">
              {data.message}
            </Text>
          )}
        </View>

        <TouchableOpacity onPress={dismiss} className="ml-3 self-center p-1">
          <X size={16} color={colors.gray400} />
        </TouchableOpacity>
      </Animated.View>
    </GestureDetector>
  );
};

const styles = StyleSheet.create({
  container: {
    position: 'absolute',
    top: Platform.OS === 'ios' ? 50 : 20,
    left: 16,
    right: 16,
    zIndex: 9999,
  },
  toastItem: {
    borderLeftWidth: 4,
    elevation: 5,
  },
});

export default toast;
