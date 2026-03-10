export const useKnowledgeBase = () => {
  const api = useApi()

  return {
    list: () => api.get<any>('/knowledge/documents'),
    upload: (file: File) => {
      const formData = new FormData()
      formData.append('file', file)
      return api.upload<any>('/knowledge/documents', formData)
    },
    findById: (id: string) => api.get<any>(`/knowledge/documents/${id}`),
    deleteDoc: (id: string) => api.del<any>(`/knowledge/documents/${id}`),
    reindex: (id: string) => api.post<any>(`/knowledge/documents/${id}/reindex`),
    search: (query: string, topK?: number) => api.post<any>('/knowledge/search', { query, topK }),
  }
}
