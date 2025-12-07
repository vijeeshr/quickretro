import { defineConfig } from 'vitepress'

// https://vitepress.dev/reference/site-config
export default defineConfig({
  title: "QuickRetro",
  description: "Self-hosted, Free & Open-source Sprint retro app",

  transformHead({ pageData }) {
    const relativePath = pageData.relativePath.replace(/\.md$/, '');
    const relativePathForCanonicalUrl = relativePath === 'index' ? '' : `/${relativePath}`;
    const canonicalUrl = `https://quickretro.app${relativePathForCanonicalUrl}`;
    return [
      ['link', { rel: 'canonical', href: canonicalUrl }]
    ];
  },

  head: [
    // Favicon
    // ['link', { rel: 'icon', type: 'image/svg+xml', href: "data:image/svg+xml,<svg xmlns=%22http://www.w3.org/2000/svg%22 width=%22256%22 height=%22256%22 viewBox=%220 0 100 100%22><rect width=%22100%22 height=%22100%22 rx=%2220%22 fill=%22%230a0a0a%22></rect><path fill=%22%23fff%22 d=%22M42 36.60L42 81.80Q41.60 81.96 40.76 82.16Q39.92 82.36 39.04 82.36L39.04 82.36Q35.60 82.36 35.60 79.40L35.60 79.40L35.60 65.88Q34 66.84 31.72 67.52Q29.44 68.20 26.32 68.20L26.32 68.20Q22.40 68.20 19.08 67.08Q15.76 65.96 13.36 63.64Q10.96 61.32 9.60 57.64Q8.24 53.96 8.24 48.92L8.24 48.92Q8.24 43.88 9.68 40.16Q11.12 36.44 13.68 34.00Q16.24 31.56 19.72 30.36Q23.20 29.16 27.20 29.16L27.20 29.16Q31.04 29.16 34.16 30.12Q37.28 31.08 39.36 32.36L39.36 32.36Q40.56 33.08 41.28 34.12Q42 35.16 42 36.60L42 36.60ZM27.28 62.92L27.28 62.92Q29.84 62.92 32 62.16Q34.16 61.40 35.60 60.20L35.60 60.20L35.60 36.68Q34.24 35.80 32.20 35.12Q30.16 34.44 27.36 34.44L27.36 34.44Q24.80 34.44 22.52 35.24Q20.24 36.04 18.52 37.80Q16.80 39.56 15.80 42.28Q14.80 45 14.80 48.92L14.80 48.92Q14.80 56.28 18.20 59.60Q21.60 62.92 27.28 62.92ZM61.68 48.76L61.68 67Q61.20 67.16 60.40 67.36Q59.60 67.56 58.64 67.56L58.64 67.56Q55.12 67.56 55.12 64.60L55.12 64.60L55.12 22.04Q55.12 20.76 55.76 20.08Q56.40 19.40 57.76 19.00L57.76 19.00Q60.08 18.28 63.48 17.96Q66.88 17.64 70.08 17.64L70.08 17.64Q80.16 17.64 85.08 21.68Q90 25.72 90 33.16L90 33.16Q90 38.68 87.16 42.44Q84.32 46.20 78.56 47.80L78.56 47.80Q80.56 50.36 82.56 52.92Q84.56 55.48 86.36 57.76Q88.16 60.04 89.56 61.88Q90.96 63.72 91.76 64.76L91.76 64.76Q91.44 66.12 90.36 66.92Q89.28 67.72 88.16 67.72L88.16 67.72Q86.80 67.72 85.92 67.08Q85.04 66.44 84.08 65.08L84.08 65.08L72.40 48.76L61.68 48.76ZM61.60 43.32L70.80 43.32Q76.72 43.32 80.08 40.76Q83.44 38.20 83.44 33.16L83.44 33.16Q83.44 28.12 79.92 25.60Q76.40 23.08 69.84 23.08L69.84 23.08Q67.68 23.08 65.48 23.28Q63.28 23.48 61.60 23.80L61.60 23.80L61.60 43.32Z%22></path></svg>" }],
    ['link', { rel: 'icon', href: '/favicon.ico', type: 'image/x-icon' }],
    ['link', { rel: 'apple-touch-icon', href: '/apple-touch-icon.png' }],
    // Canonical Link
    // ['link', { rel: 'canonical', href: 'https://quickretro.app' }],
    // Meta Tags
    ['meta', { name: 'title', content: 'QuickRetro - Free and Open-Source Sprint Retrospective Meeting App' }],
    ['meta', { name: 'robots', content: 'index, follow' }],
    ['meta', { name: 'language', content: 'English' }],
    ['meta', { name: 'keywords', content: 'sprint, agile, restrospective, meeting, websocket, opensource, free' }],
    // OpenGraph tags
    ['meta', { property: 'og:title', content: 'QuickRetro - Free and Open-Source Sprint Retrospective Meeting App' }],
    ['meta', { property: 'og:url', content: 'https://quickretro.app' }],    
    ['meta', { property: 'og:site_name', content: 'QuickRetro' }],
    ['meta', { property: 'og:description', content: 'Easily conduct a Sprint retrospective online with this free and open-source app' }],
    ['meta', { property: 'og:image', content: 'https://quickretro.app/logo.png' }],
    ['meta', { property: 'og:type', content: 'website' }],
    // Twitter tags
    ['meta', { name: 'twitter:card', content: 'summary_large_image' }],
    ['meta', { name: 'twitter:title', content: 'QuickRetro - Free and Open-Source Sprint Retrospective Meeting App' }],
    ['meta', { name: 'twitter:description', content: 'Easily conduct a Sprint retrospective online with this free and open-source app' }],
    ['meta', { name: 'twitter:image', content: 'https://quickretro.app/logo.png' }],
    ['meta', { property: 'twitter:domain', content: 'quickretro.app' }],
    ['meta', { property: 'twitter:url', content: 'https://quickretro.app/' }],
    // JSON-LD Structured Data for SEO
    ['script', { type: 'application/ld+json' }, `
      {
        "@context": "https://schema.org",
        "@type": "WebSite",
        "name": "QuickRetro - Free and Open-Source Sprint Retrospective Meeting App",
        "url": "https://quickretro.app",
        "description": "Easily conduct a Sprint retrospective online with this free and open-source app",
        "logo": "https://quickretro.app/logo.png"
      }
    `]
  ],

  lastUpdated: true,
  cleanUrls: true,
  ignoreDeadLinks: [
    // ignore all localhost links
    /^https?:\/\/localhost/
  ],
  sitemap: {
    hostname: 'https://quickretro.app',
    transformItems: (items) => {
      // modify/filter existing items
      items.forEach(item => {
        // Default priority
        item.priority = 0.5
        // Dashboard
        if (item.url === 'guide/dashboard') {
          item.priority = 1
          item.img = {
            url: 'https://quickretro.app/dashboard_owner.png',
            caption: 'Dashboard features',
            title: 'Dashboard features'
          }
        }
        // Home page
        if (item.url === '') {
          item.priority = 1
          item.img = {
            url: 'https://quickretro.app/logo.png',
            caption: 'Free and Open-Source Sprint Retrospective App',
            title: 'QuickRetro | Free and Open-Source Sprint Retrospective Meeting App'
          }
        }
        // Create-Board
        if (item.url === 'guide/create-board') {
          item.priority = 0.9
          item.img = {
            url: 'https://quickretro.app/createboard.png',
            caption: 'Create board',
            title: 'Create board'
          }
        }
        // Development
        if (item.url === 'guide/development' || item.url === 'guide/getting-started') {
          item.priority = 0.9
        }
      })

      // add new items
      // Demo link
      items.push({
        url: 'https://demo.quickretro.app',
        priority: 1
      })

      return items
    }
  },
  
  themeConfig: {
    // https://vitepress.dev/reference/default-theme-config
    logo : { light: "logo_large_light.png", dark: "logo_large_dark.png", width: 24, height: 24 },

    nav: [
      { text: 'Home', link: '/' },
      { text: 'Features', link: '/guide/dashboard' },
      { text: 'Development', link: '/guide/development' }
    ],

    sidebar: [
      {
        text: 'Guide',
        items: [
          { text: 'Getting Started', link: '/guide/getting-started' },
          { text: 'Create Board', link: '/guide/create-board' },
          { text: 'Dashboard', link: '/guide/dashboard' }
        ]
      },
      {
        // text: 'Development',
        items: [
          { text: 'Development', link: '/guide/development' },
          { text: 'Configurations', link: '/guide/configurations' }
        ]
      },
      {
        // text: 'Self-hosting',
        items: [
          { text: 'Self-Hosting', link: '/guide/self-hosting' }
        ]
      }             
    ],

    socialLinks: [
      { icon: 'github', link: 'https://github.com/vijeeshr/quickretro' }
    ],

    footer: {
      message: 'Released under the AGPL-3.0 license',
      copyright: 'Copyright © 2024-present QuickRetro™'
    }

  }
})
