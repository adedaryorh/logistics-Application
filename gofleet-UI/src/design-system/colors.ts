// src/design-system/colors.ts
export const colors = {
  // Primary brand
  primary:    '#1A1A2E',   // deep navy
  secondary:  '#16213E',   // darker navy
  accent:     '#0F3460',   // royal blue
  highlight:  '#E94560',   // electric red (CTAs, active states)

  // Semantic
  success:    '#10B981',
  warning:    '#F59E0B',
  error:      '#EF4444',
  info:       '#3B82F6',

  // Neutrals
  white:      '#FFFFFF',
  black:      '#000000',
  gray50:     '#F9FAFB',
  gray100:    '#F3F4F6',
  gray200:    '#E5E7EB',
  gray300:    '#D1D5DB',
  gray400:    '#9CA3AF',
  gray500:    '#6B7280',
  gray600:    '#4B5563',
  gray700:    '#374151',
  gray800:    '#1F2937',
  gray900:    '#111827',

  // Map overlays
  mapOverlay:       'rgba(26,26,46,0.85)',
  driverMarker:     '#10B981',
  pickupMarker:     '#3B82F6',
  dropoffMarker:    '#E94560',
  routeLine:        '#0F3460',

  // Driver app
  onlineBadge:  '#10B981',
  offlineBadge: '#9CA3AF',
  onTripBadge:  '#F59E0B',

  // Backgrounds
  background:       '#F9FAFB',
  backgroundDark:   '#1A1A2E',
  card:             '#FFFFFF',
  cardDark:         '#16213E',

  // Transparent
  overlay: 'rgba(0,0,0,0.5)',
} as const;
