export const useLeads = () => {
  const api = useApi()

  return {
    list: (params?: Record<string, any>) => api.get<any>('/leads', params),
    findById: (id: string) => api.get<any>(`/leads/${id}`),
    update: (id: string, data: any) => api.put<any>(`/leads/${id}`, data),
  }
}
