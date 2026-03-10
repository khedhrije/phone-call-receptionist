export const useCalls = () => {
  const api = useApi()

  return {
    list: (params?: Record<string, any>) => api.get<any>('/calls', params),
    detail: (id: string) => api.get<any>(`/calls/${id}`),
    ragQueries: (id: string) => api.get<any>(`/calls/${id}/rag-queries`),
    stats: () => api.get<any>('/calls/stats'),
  }
}
