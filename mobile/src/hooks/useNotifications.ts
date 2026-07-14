import { useEffect } from 'react';
import { Platform } from 'react-native';
import * as Notifications from 'expo-notifications';

Notifications.setNotificationHandler({
  handleNotification: async () => ({
    shouldShowBanner: true,
    shouldShowList: true,
    shouldPlaySound: true,
    shouldSetBadge: false,
  }),
});
export function useNotifications(enabled: boolean) {
  useEffect(() => {
    if (!enabled) return;
    (async () => {
      if (Platform.OS === 'android')
        await Notifications.setNotificationChannelAsync('deliveries', {
          name: 'Delivery updates',
          importance: Notifications.AndroidImportance.HIGH,
        });
      const current = await Notifications.getPermissionsAsync();
      if (current.status !== 'granted') await Notifications.requestPermissionsAsync();
    })().catch(() => undefined);
  }, [enabled]);
}
