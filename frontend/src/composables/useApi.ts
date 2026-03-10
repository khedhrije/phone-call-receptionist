export const useApi = () => {
  const config = useRuntimeConfig()
  const { token, logout } = useAuth()

  const headers = computed(() => {
    const h: Record<string, string> = { 'Content-Type': 'application/json' }
    if (token.value) h['Authorization'] = `Bearer ${token.value}`
    return h
  })

  const request = async <T>(path: string, options: any = {}): Promise<T> => {
    try {
      return await $fetch<T>(`${config.public.apiBase}${path}`, {
        ...options,
        headers: { ...headers.value, ...options.headers },
      })
    } catch (err: any) {
      if (err?.response?.status === 401) logout()
      throw err
    }
  }

  return {
    get: <T>(path: string, params?: Record<string, any>) =>
      request<T>(path, { method: 'GET', params }),
    post: <T>(path: string, body?: any) =>
      request<T>(path, { method: 'POST', body }),
    put: <T>(path: string, body?: any) =>
      request<T>(path, { method: 'PUT', body }),
    del: <T>(path: string) =>
      request<T>(path, { method: 'DELETE' }),
    upload: <T>(path: string, formData: FormData) =>
      request<T>(path, { method: 'POST', body: formData, headers: { Authorization: `Bearer ${token.value}` } }),
  }
}
