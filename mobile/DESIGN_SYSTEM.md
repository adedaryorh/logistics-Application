# Field Logistics design system

Direction: a sunlight-readable, trustworthy execution tool for drivers and service providers moving agricultural loads. The interface uses white field cards, deep forest text, high-contrast green actions, restrained status colors, plain language, and Ionicons only.

- Minimum target: 48×48 dp; primary action: 56 dp.
- Text supports Dynamic Type and avoids fixed-height text containers.
- Status is always color plus a written label and icon.
- Screens center at 760 dp on tablets; lists virtualize; bottom actions remain one-handed.
- Motion lasts 150–300 ms and is functional. Native reduced-motion settings take precedence.
- Forms retain local state when errors occur or modals remain open.
- Offline/error states explain whether data is cached, retryable, or requires support.
- Destructive actions require a native confirmation dialog.
- Semantic tokens live in `src/theme.ts`; reusable interaction/state primitives live in `src/components/AppUI.tsx`.
