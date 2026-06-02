/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      fontFamily: {
        sans: ['Nunito', 'system-ui', 'sans-serif'],
      },
      colors: {
        brand: {
          50: '#fffbeb',
          100: '#fef3c7',
          200: '#fde68a',
          500: '#f59e0b',
          600: '#d97706',
          700: '#b45309',
        },
        page: {
          warm: '#FFFDF7',
          cool: '#F9FAFB',
        },
      },
      borderRadius: {
        container: '1rem',
        card: '0.75rem',
        control: '0.5rem',
      },
    },
  },
  plugins: [],
}
