/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./App.{js,jsx,ts,tsx}",
    "./src/**/*.{js,jsx,ts,tsx}"
  ],
  presets: [require("nativewind/preset")],
  theme: {
    extend: {
      colors: {
        primary: '#1A1A2E',
        secondary: '#16213E',
        accent: '#0F3460',
        highlight: '#E94560',
        success: '#10B981',
        warning: '#F59E0B',
        error: '#EF4444',
        info: '#3B82F6',
        cardDark: '#16213E',
        backgroundDark: '#1A1A2E',
      },
      fontFamily: {
        sans: ["System"],
      }
    },
  },
  plugins: [],
}
