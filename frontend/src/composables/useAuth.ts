interface User {
  id: string
  email: string
  role: string
  isBlocked: boolean
}

const token = ref<string>('')
const user = ref<User | null>(null)

export const useAuth = () => {
  const config = useRuntimeConfig()

  const isAuthenticated = computed(() => !!token.value)

  const init = () => {
    if (import.meta.client) {
      token.value = localStorage.getItem('token') || ''
      const stored = localStorage.getItem('user')
      if (stored) user.value = JSON.parse(stored)
    }
  }

  const signIn = async (email: string, password: string) => {
    const data = await $fetch<{ token: string; user: User }>(`${config.public.apiBase}/auth/signin`, {
      method: 'POST',
      body: { email, password },
    })
    token.value = data.token
    user.value = data.user
    if (import.meta.client) {
      localStorage.setItem('token', data.token)
      localStorage.setItem('user', JSON.stringify(data.user))
    }
  }

  const signUp = async (email: string, password: string) => {
    const data = await $fetch<{ token: string; user: User }>(`${config.public.apiBase}/auth/signup`, {
      method: 'POST',
      body: { email, password },
    })
    token.value = data.token
    user.value = data.user
    if (import.meta.client) {
      localStorage.setItem('token', data.token)
      localStorage.setItem('user', JSON.stringify(data.user))
    }
  }

  const logout = () => {
    token.value = ''
    user.value = null
    if (import.meta.client) {
      localStorage.removeItem('token')
      localStorage.removeItem('user')
    }
    navigateTo('/')
  }

  init()

  return { token, user, isAuthenticated, signIn, signUp, logout }
}
