// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: '2025-07-15',
  devtools: { enabled: true },
  typescript: { strict: true },
  runtimeConfig: {
    // Server-side rendering runs inside this process, where "localhost" means
    // this container, not the backend one — so it needs its own base URL
    // when the two run as separate Docker Compose services.
    apiBaseServer: process.env.NUXT_API_BASE_SERVER || process.env.NUXT_PUBLIC_API_BASE || 'http://localhost:8081',
    public: {
      apiBase: process.env.NUXT_PUBLIC_API_BASE || 'http://localhost:8081'
    }
  }
})
