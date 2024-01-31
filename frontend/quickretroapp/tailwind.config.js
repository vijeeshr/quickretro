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
  plugins: [],
  safelist: [
    {
      pattern: /bg-(red|green|yellow)-(100|400)/,
      variants: ['hover'],
    },
    {
      pattern: /border-(red|green|yellow)-(300)/
    },
    {
      pattern: /text-(red|green|yellow)-(600)/
    },
    {
      pattern: /w-(6|8)/,
    },
    {
      pattern: /h-(6|8)/,
    }
  ],  
}