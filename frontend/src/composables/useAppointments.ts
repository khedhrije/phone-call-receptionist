export const useAppointments = () => {
  const api = useApi()

  return {
    list: (params?: Record<string, any>) => api.get<any>('/appointments', params),
    create: (data: any) => api.post<any>('/appointments', data),
    reschedule: (id: string, data: any) => api.put<any>(`/appointments/${id}`, data),
    cancel: (id: string) => api.del<any>(`/appointments/${id}`),
    availability: (from: string, to: string) => api.get<any>('/appointments/availability', { from, to }),
  }
}
