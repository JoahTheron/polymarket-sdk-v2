import { defineConfig } from 'vitepress'

export default defineConfig({
  lang: 'en-US',
  title: 'Polymarket Go Client',
  description: 'Go SDK for Polymarket — CLOB v2, Relayer, Data, Gamma, and Bridge APIs',
  base: '/polymarket-client/',
  themeConfig: {
    logo: '/logo.svg',
    nav: [
      { text: 'Guide', link: '/guide/what-is' },
      { text: 'API Reference', link: '/api/' },
      { text: 'GitHub', link: 'https://github.com/bububa/polymarket-client' },
    ],
    sidebar: [
      {
        text: 'Getting Started',
        items: [
          { text: 'What is Polymarket Client', link: '/guide/what-is' },
          { text: 'Installation', link: '/guide/installation' },
          { text: 'Quick Start', link: '/guide/quick-start' },
        ],
      },
      {
        text: 'Authentication',
        items: [
          { text: 'Auth Levels Overview', link: '/guide/auth-levels' },
          { text: 'L1 — Wallet Signatures', link: '/guide/auth-l1' },
          { text: 'L2 — Full Trading', link: '/guide/auth-l2' },
          { text: 'API Key Lifecycle', link: '/guide/api-key-lifecycle' },
        ],
      },
      {
        text: 'Guides',
        items: [
          { text: 'Trading Orders', link: '/guide/trading' },
          { text: 'WebSocket Streams', link: '/guide/websocket' },
          { text: 'Working with Types', link: '/guide/types' },
          { text: 'RFQ (Request for Quote)', link: '/guide/rfq' },
          { text: 'Relayer Integration', link: '/guide/relayer' },
        ],
      },
      {
        text: 'API Reference',
        items: [
          { text: 'clob', link: '/api/clob' },
          { text: 'clob/ws', link: '/api/ws' },
          { text: 'relayer', link: '/api/relayer' },
          { text: 'data', link: '/api/data' },
          { text: 'gamma', link: '/api/gamma' },
          { text: 'bridge', link: '/api/bridge' },
          { text: 'types', link: '/api/types' },
        ],
      },
      {
        text: 'Resources',
        items: [
          { text: 'Architecture', link: '/architecture' },
          { text: 'Troubleshooting', link: '/troubleshooting' },
          { text: 'Changelog', link: '/changelog' },
        ],
      },
    ],
    socialLinks: [
      { icon: 'github', link: 'https://github.com/bububa/polymarket-client' },
    ],
    search: {
      provider: 'local',
    },
  },
})
