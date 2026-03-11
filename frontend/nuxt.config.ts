export default defineNuxtConfig({
  srcDir: 'src/',
  devtools: { enabled: true },
  devServer: { port: 3000 },
  modules: [
    '@nuxtjs/tailwindcss',
  ],
  runtimeConfig: {
    public: {
      apiBase: process.env.API_BASE_URL || 'http://localhost:8082/api',
      wsBase: process.env.WS_BASE_URL || 'ws://localhost:8082/ws',
    },
  },
  tailwindcss: {
    configPath: 'tailwind.config.ts',
  },
})
