export const useWebSocket = () => {
  const config = useRuntimeConfig()
  let socket: WebSocket | null = null
  const events = ref<any[]>([])
  const connected = ref(false)

  const connect = () => {
    if (import.meta.server) return

    socket = new WebSocket(config.public.wsBase)

    socket.onopen = () => {
      connected.value = true
    }

    socket.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data)
        events.value.unshift(data)
        if (events.value.length > 100) events.value.pop()
      } catch {}
    }

    socket.onclose = () => {
      connected.value = false
      setTimeout(connect, 3000)
    }

    socket.onerror = () => {
      socket?.close()
    }
  }

  const disconnect = () => {
    socket?.close()
    socket = null
  }

  return { connect, disconnect, events, connected }
}
