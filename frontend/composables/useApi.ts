interface ApiErrorBody {
  error: string
}

export function useApi() {
  const config = useRuntimeConfig()
  const baseURL = process.server ? config.apiBaseServer : config.public.apiBase

  async function request<T>(path: string, options: { method?: 'GET' | 'POST' | 'PATCH'; body?: unknown } = {}): Promise<T> {
    try {
      return await $fetch<T>(path, {
        baseURL,
        method: options.method ?? 'GET',
        body: options.body
      })
    } catch (err: any) {
      const message = err?.data?.error ?? err?.message ?? 'Request failed'
      throw new Error(message)
    }
  }

  return {
    get: <T>(path: string) => request<T>(path),
    post: <T>(path: string, body: unknown) => request<T>(path, { method: 'POST', body }),
    patch: <T>(path: string, body: unknown) => request<T>(path, { method: 'PATCH', body })
  }
}
