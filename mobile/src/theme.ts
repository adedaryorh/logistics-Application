// FarmSense Field Logistics design tokens. Components consume semantic names;
// raw values stay here so contrast and density can be changed centrally.
export const colors = {
  ink: '#10231B',
  muted: '#52625B',
  canvas: '#F4F6F1',
  surface: '#FFFFFF',
  line: '#CBD5CE',
  green: '#087A54',
  greenDark: '#075F43',
  greenSoft: '#DDF3E8',
  lime: '#D7F35D',
  navy: '#123C43',
  peach: '#FFE3D2',
  amber: '#B86B00',
  amberSoft: '#FFF1D6',
  danger: '#B42318',
  dangerSoft: '#FEE4E2',
  info: '#175CD3',
  infoSoft: '#EAF2FF',
  disabled: '#89958F',
  scrim: 'rgba(8,22,16,.56)',
} as const;
export const spacing = { xs: 4, sm: 8, md: 12, lg: 16, xl: 20, xxl: 28, xxxl: 36 } as const;
export const type = { caption: 12, body: 16, label: 15, title: 24, display: 34 } as const;
export const radius = { sm: 10, md: 14, lg: 20, xl: 28 } as const;
export const touch = { minimum: 48, primary: 56 } as const;
export const motion = { quick: 150, standard: 220, deliberate: 300 } as const;
export const layout = { maxContent: 760, tabletBreakpoint: 720 } as const;
export const shadow = {
  shadowColor: '#10231B',
  shadowOffset: { width: 0, height: 5 },
  shadowOpacity: 0.1,
  shadowRadius: 12,
  elevation: 3,
};
export const statusTone = {
  pending: ['Requested', colors.info, colors.infoSoft],
  awaiting_payment: ['Quote ready', colors.amber, colors.amberSoft],
  paid: ['Confirmed', colors.greenDark, colors.greenSoft],
  dispatching: ['Booked', colors.info, colors.infoSoft],
  assigned: ['Provider assigned', colors.greenDark, colors.greenSoft],
  picked_up: ['Picked up', colors.amber, colors.amberSoft],
  delivered: ['Delivered', colors.greenDark, colors.greenSoft],
  cancelled: ['Cancelled', colors.danger, colors.dangerSoft],
  failed: ['Failed', colors.danger, colors.dangerSoft],
} as const;
