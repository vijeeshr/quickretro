/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{vue,js,ts,jsx,tsx}'],
  theme: {
    extend: {
      minHeight: (theme) => ({
        ...theme('spacing'),
      }),
    }
  },
  darkMode: 'class',
  plugins: [],
  safelist: [
    // Backgrounds
    {
      pattern: /^(bg|hover:bg|dark:bg|dark:hover:bg)-(red|green|yellow|fuchsia|orange)-(100|400|600|800)/,
      variants: ['hover', 'dark', 'dark:hover'],
    },
    // Borders
    {
      pattern: /^(border|dark:border)-(red|green|yellow|fuchsia|orange)-(300|700)/,
      variants: ['dark'],
    },
    // Text
    {
      pattern: /^(text|dark:text)-(red|green|yellow|fuchsia|orange)-(100|600)/,
      variants: ['dark'],
    }
    // Sizes
    // {
    //   pattern: /w-(6|8)/,
    // },
    // {
    //   pattern: /h-(6|8)/,
    // }
  ],
}