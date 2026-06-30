import React, { forwardRef, useCallback } from 'react';
import { View, StyleSheet } from 'react-native';
import GorhomBottomSheet, {
  BottomSheetView,
  BottomSheetBackdrop,
  BottomSheetBackdropProps,
} from '@gorhom/bottom-sheet';

export interface BottomSheetProps {
  children: React.ReactNode;
  snapPoints?: (string | number)[];
  index?: number;
  onChange?: (index: number) => void;
  enablePanDownToClose?: boolean;
}

export const BottomSheet = forwardRef<GorhomBottomSheet, BottomSheetProps>(
  (
    {
      children,
      snapPoints = ['25%', '50%', '90%'],
      index = 0,
      onChange,
      enablePanDownToClose = true,
    },
    ref
  ) => {
    // Render Backdrop backdrop dismiss functionality
    const renderBackdrop = useCallback(
      (props: BottomSheetBackdropProps) => (
        <BottomSheetBackdrop
          {...props}
          disappearsOnIndex={-1}
          appearsOnIndex={0}
          opacity={0.5}
          pressBehavior="close"
        />
      ),
      []
    );

    return (
      <GorhomBottomSheet
        ref={ref}
        index={index}
        snapPoints={snapPoints}
        onChange={onChange}
        enablePanDownToClose={enablePanDownToClose}
        backdropComponent={renderBackdrop}
        handleIndicatorStyle={styles.handleIndicator}
        backgroundStyle={styles.background}
      >
        <BottomSheetView style={styles.contentContainer}>
          {children}
        </BottomSheetView>
      </GorhomBottomSheet>
    );
  }
);

BottomSheet.displayName = 'BottomSheet';

const styles = StyleSheet.create({
  handleIndicator: {
    backgroundColor: '#D1D5DB', // gray300
    width: 40,
    height: 4,
  },
  background: {
    backgroundColor: '#FFFFFF',
    borderTopLeftRadius: 24,
    borderTopRightRadius: 24,
    shadowColor: '#000000',
    shadowOffset: { width: 0, height: -4 },
    shadowOpacity: 0.1,
    shadowRadius: 10,
    elevation: 10,
  },
  contentContainer: {
    flex: 1,
    padding: 16,
  },
});

export default BottomSheet;
