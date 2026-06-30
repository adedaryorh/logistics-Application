import React, { useEffect, useState, useRef } from 'react';
import { View, Text, ActivityIndicator, StyleSheet, BackHandler, Platform } from 'react-native';
import { WebView } from 'react-native-webview';
import { useNavigation, useRoute } from '@react-navigation/native';
import { useAuthStore } from '@/features/auth/store/authStore';
import paymentService from '@/services/payment/paymentService';
import { Button } from '@/design-system/components/Button';
import { toast } from '@/design-system/components/Toast';
import * as Haptics from '@/utils/haptics';
import { formatMoney } from '@/utils/formatMoney';

interface PaymentWebViewProps {
  orderId: string;
  amount: number; // Amount in kobo/cents
  currency?: string;
}

interface RouteParams {
  paymentId?: string;
  amount?: number;
}

const PaymentWebView: React.FC<PaymentWebViewProps> = ({ orderId, amount, currency = 'NGN' }) => {
  const navigation = useNavigation();
  const route = useRoute<RouteParams>();
  const { user } = useAuthStore();
  const [paymentId, setPaymentId] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  [error, setError] = useState<string | null>(null);
  [status, setStatus] = useState<'pending' | 'completed' | 'failed' | null>(null);
  [webViewRef, setWebViewRef] = useState<WebView | null>(null);
  
  const backHandler = BackHandler.addEventListener('hardwareBackPress', () => {
    if (webViewRef && webViewRef.canGoBack) {
      webViewRef.goBack();
      return true;
    }
    return false;
  });

  useEffect(() => {
    return () => backHandler.remove();
  }, []);

  // Initialize payment when component mounts
  useEffect(() => {
    const initPayment = async () => {
      try {
        setLoading(true);
        const paymentData = await paymentService.initializePayment({
          orderId,
          amountMinor: amount,
          currency,
          method: 'card', // Default to card, could be made configurable
          metadata: { 
            userId: user?.id,
            timestamp: new Date().toISOString()
          }
        });
        
        setPaymentId(paymentData.paymentId);
        setLoading(false);
      } catch (err: any) {
        setError(err.message || 'Failed to initialize payment');
        setLoading(false);
        Haptics.haptics.error();
        toast.error('Payment initialization failed');
        
        // Auto-go back after error
        setTimeout(() => {
          navigation.goBack();
        }, 2000);
      }
    };

    initPayment();
  }, [orderId, amount, currency, user?.id, navigation]);

  // Handle WebView message events (for payment callbacks)
  const handleMessage = (event: any) => {
    try {
      const data = JSON.parse(event.nativeEvent.data);
      
      if (data.type === 'payment_success') {
        handlePaymentSuccess(data.paymentId);
      } else if (data.type === 'payment_failed') {
        handlePaymentFailure(data.paymentId, data.error);
      } else if (data.type === 'redirect_to_app') {
        // Handle deep link back to app
        handlePaymentSuccess(data.paymentId);
      }
    } catch (e) {
      // Ignore parsing errors - not all messages are JSON
    }
  };

  const handlePaymentSuccess = async (paymentId: string) => {
    try {
      setStatus('completed');
      const payment = await paymentService.verifyPayment(paymentId);
      
      Haptics.haptics.success();
      toast.success('Payment successful!');
      
      // Navigate back to order screen or tracking
      setTimeout(() => {
        navigation.navigate('OrderTracking', { orderId });
      }, 1500);
    } catch (err: any) {
      setStatus('failed');
      setError('Payment verification failed');
      Haptics.haptics.error();
      toast.error('Payment verification failed');
    }
  };

  const handlePaymentFailure = (paymentId: string, errorMsg: string) => {
    setStatus('failed');
    setError(errorMsg || 'Payment failed');
    Haptics.haptics.error();
    toast.error('Payment failed');
    
    // Auto-go back after failure
    setTimeout(() => {
      navigation.goBack();
    }, 2000);
  };

  // Handle loading states
  if (loading && !paymentId) {
    return (
      <View style={styles.container}>
        <ActivityIndicator size="large" color={Colors.accent} />
        <Text style={styles.loadingText}>Initializing payment...</Text>
      </View>
    );
  }

  if (error && !paymentId) {
    return (
      <View style={styles.container}>
        <Text style={styles.errorText}>Error: {error}</Text>
        <Button 
          title="Go Back" 
          onPress={() => navigation.goBack()}
          variant="outline"
        />
      </View>
    );
  }

  return (
    <View style={styles.container}>
      <WebView
        ref={setWebViewRef}
        source={{
          uri: `${process.env.EXPO_PUBLIC_API_URL}/api/v1/payments/${paymentId}/webview`,
        }}
        javaScriptEnabled={true}
        domStorageEnabled={true}
        allowFileAccess={true}
        mixedContentMode="compatibility"
        onMessage={handleMessage}
        loadingStart={() => setStatus('pending')}
        loadingFinish={() => {
          // Optionally check URL for success/failure patterns
        }}
        error={(e) => {
          console.error('WebView error:', e);
          setError('Failed to load payment page');
        }}
        style={styles.webView}
      />
      
      {/* Fallback buttons in case WebView fails */}
      {status === 'pending' && (
        <View style={styles.buttonContainer}>
          <Button 
            title="Check Payment Status" 
            onPress={() => {
              if (paymentId) {
                paymentService.verifyPayment(paymentId)
                  .then(() => {
                    handlePaymentSuccess(paymentId);
                  })
                  .catch(() => {
                    handlePaymentFailure(paymentId, 'Payment check failed');
                  });
              }
            }}
          />
          <Button 
            title="Cancel" 
            onPress={() => {
              if (paymentId) {
                // In a real app, you might cancel the payment
                navigation.goBack();
              }
            }}
            variant="outline"
          />
        </View>
      )}
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: Colors.background,
  },
  webView: {
    flex: 1,
  },
  buttonContainer: {
    position: 'absolute',
    bottom: 20,
    left: 20,
    right: 20,
    flexDirection: 'row',
    justifyContent: 'space-between',
  },
  loadingText: {
    marginTop: 20,
    textAlign: 'center',
    color: Colors.gray600,
  },
  errorText: {
    marginTop: 20,
    textAlign: 'center',
    color: Colors.error,
  },
});

export default PaymentWebView;
