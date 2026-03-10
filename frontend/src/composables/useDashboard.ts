export const useDashboard = () => {
  const api = useApi()

  return {
    stats: () => api.get<any>('/dashboard/stats'),
    costAnalytics: (from: string, to: string) => api.get<any>('/dashboard/costs', { from, to }),
    callVolume: (from: string, to: string) => api.get<any>('/dashboard/volume', { from, to }),
  }
}
