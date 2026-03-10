export default defineNuxtConfig({
  srcDir: 'src/',
  devtools: { enabled: true },
  modules: [
    '@nuxtjs/tailwindcss',
  ],
  runtimeConfig: {
    public: {
      apiBase: process.env.API_BASE_URL || 'http://localhost:8080/api',
      wsBase: process.env.WS_BASE_URL || 'ws://localhost:8080/ws',
    },
  },
  tailwindcss: {
    configPath: 'tailwind.config.ts',
  },
})
