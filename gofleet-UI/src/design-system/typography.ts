// src/design-system/typography.ts
export const typography = {
  // Font families
  fontFamily: {
    regular:    'System', // Fallback to System to ensure it works out of the box in React Native
    medium:     'System',
    semibold:   'System',
    bold:       'System',
    mono:       'System',
  },

  // Size scale
  fontSize: {
    xs:   10,
    sm:   12,
    base: 14,
    md:   16,
    lg:   18,
    xl:   20,
    '2xl': 24,
    '3xl': 30,
    '4xl': 36,
  },

  // Line heights
  lineHeight: {
    tight:   1.2,
    normal:  1.5,
    relaxed: 1.75,
  },
} as const;
