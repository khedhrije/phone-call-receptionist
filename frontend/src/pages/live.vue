<script setup lang="ts">
definePageMeta({ layout: 'default' })

const { connect, disconnect, events, connected } = useWebSocket()

interface TranscriptEntry {
  speaker: string
  text: string
  at: string
}

interface ActiveCall {
  callId: string
  callerPhone: string
  startedAt: string
  transcript: TranscriptEntry[]
}

const activeCall = ref<ActiveCall | null>(null)
const ringing = ref(false)
const callHistory = ref<Array<{ callId: string; callerPhone: string; duration: number; cost: number; endedAt: string }>>([])

const transcriptContainer = ref<HTMLElement | null>(null)

const scrollToBottom = () => {
  nextTick(() => {
    if (transcriptContainer.value) {
      transcriptContainer.value.scrollTop = transcriptContainer.value.scrollHeight
    }
  })
}

watch(events, (evts) => {
  if (!evts.length) return
  const event = evts[0]

  if (event.type === 'call_started') {
    ringing.value = true
    activeCall.value = {
      callId: event.callId,
      callerPhone: event.callerPhone,
      startedAt: event.at,
      transcript: []
    }
    // Stop ringing animation after 3s
    setTimeout(() => { ringing.value = false }, 3000)

    // Vibrate if supported
    if (navigator.vibrate) {
      navigator.vibrate([200, 100, 200, 100, 200])
    }
  }

  if (event.type === 'call_speech' && activeCall.value?.callId === event.callId) {
    activeCall.value.transcript.push(
      { speaker: 'caller', text: event.transcript, at: event.at },
      { speaker: 'assistant', text: event.response, at: event.at }
    )
    scrollToBottom()
  }

  if (event.type === 'call_ended') {
    if (activeCall.value?.callId === event.callId) {
      callHistory.value.unshift({
        callId: event.callId,
        callerPhone: activeCall.value.callerPhone,
        duration: event.duration,
        cost: event.cost,
        endedAt: event.at
      })
      activeCall.value = null
    }
    ringing.value = false
  }
}, { deep: true })

const callDuration = ref(0)
let durationInterval: ReturnType<typeof setInterval> | null = null

watch(activeCall, (call) => {
  if (call) {
    callDuration.value = 0
    durationInterval = setInterval(() => {
      callDuration.value = Math.floor((Date.now() - new Date(call.startedAt).getTime()) / 1000)
    }, 1000)
  } else {
    if (durationInterval) clearInterval(durationInterval)
    callDuration.value = 0
  }
})

const formatDuration = (seconds: number) => {
  const m = Math.floor(seconds / 60)
  const s = seconds % 60
  return `${m}:${s.toString().padStart(2, '0')}`
}

onMounted(() => { connect() })
onUnmounted(() => { disconnect() })
</script>

<template>
  <div class="p-6 max-w-4xl mx-auto">
    <div class="flex items-center justify-between mb-8">
      <h1 class="text-2xl font-bold text-gray-900">Live Monitor</h1>
      <div class="flex items-center gap-2">
        <span class="w-2.5 h-2.5 rounded-full" :class="connected ? 'bg-green-500' : 'bg-red-500'"></span>
        <span class="text-sm text-gray-500">{{ connected ? 'Connected' : 'Disconnected' }}</span>
      </div>
    </div>

    <!-- Phone visual -->
    <div class="flex flex-col items-center mb-10">
      <div
        class="relative w-48 h-80 bg-gray-900 rounded-[2.5rem] border-4 border-gray-700 shadow-2xl flex flex-col items-center justify-center transition-all"
        :class="{ 'animate-vibrate': ringing }"
      >
        <!-- Notch -->
        <div class="absolute top-3 w-20 h-5 bg-gray-800 rounded-full"></div>

        <!-- Screen -->
        <div class="w-40 h-64 bg-gray-800 rounded-2xl mt-4 flex flex-col items-center justify-center overflow-hidden">
          <template v-if="activeCall">
            <!-- Active call screen -->
            <div class="flex flex-col items-center gap-3 text-center px-3">
              <div class="w-14 h-14 rounded-full bg-green-500/20 flex items-center justify-center">
                <svg class="w-7 h-7 text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                    d="M3 5a2 2 0 012-2h3.28a1 1 0 01.948.684l1.498 4.493a1 1 0 01-.502 1.21l-2.257 1.13a11.042 11.042 0 005.516 5.516l1.13-2.257a1 1 0 011.21-.502l4.493 1.498a1 1 0 01.684.949V19a2 2 0 01-2 2h-1C9.716 21 3 14.284 3 6V5z" />
                </svg>
              </div>
              <p class="text-green-400 text-xs font-medium">In Call</p>
              <p class="text-white text-sm font-mono">{{ activeCall.callerPhone }}</p>
              <p class="text-gray-400 text-lg font-mono">{{ formatDuration(callDuration) }}</p>
              <div class="flex gap-1 mt-1">
                <span class="w-1.5 h-1.5 rounded-full bg-green-400 animate-pulse"></span>
                <span class="w-1.5 h-1.5 rounded-full bg-green-400 animate-pulse" style="animation-delay: 0.2s"></span>
                <span class="w-1.5 h-1.5 rounded-full bg-green-400 animate-pulse" style="animation-delay: 0.4s"></span>
              </div>
            </div>
          </template>

          <template v-else-if="ringing">
            <!-- Ringing screen -->
            <div class="flex flex-col items-center gap-3">
              <div class="w-14 h-14 rounded-full bg-green-500/30 flex items-center justify-center animate-ping-slow">
                <svg class="w-7 h-7 text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                    d="M3 5a2 2 0 012-2h3.28a1 1 0 01.948.684l1.498 4.493a1 1 0 01-.502 1.21l-2.257 1.13a11.042 11.042 0 005.516 5.516l1.13-2.257a1 1 0 011.21-.502l4.493 1.498a1 1 0 01.684.949V19a2 2 0 01-2 2h-1C9.716 21 3 14.284 3 6V5z" />
                </svg>
              </div>
              <p class="text-green-400 text-sm font-medium animate-pulse">Incoming Call...</p>
            </div>
          </template>

          <template v-else>
            <!-- Idle screen -->
            <div class="flex flex-col items-center gap-3">
              <div class="w-14 h-14 rounded-full bg-gray-700 flex items-center justify-center">
                <svg class="w-7 h-7 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                    d="M3 5a2 2 0 012-2h3.28a1 1 0 01.948.684l1.498 4.493a1 1 0 01-.502 1.21l-2.257 1.13a11.042 11.042 0 005.516 5.516l1.13-2.257a1 1 0 011.21-.502l4.493 1.498a1 1 0 01.684.949V19a2 2 0 01-2 2h-1C9.716 21 3 14.284 3 6V5z" />
                </svg>
              </div>
              <p class="text-gray-500 text-xs">Waiting for calls...</p>
            </div>
          </template>
        </div>

        <!-- Home button -->
        <div class="absolute bottom-2 w-10 h-1 bg-gray-600 rounded-full"></div>
      </div>
    </div>

    <!-- Live transcript -->
    <div v-if="activeCall" class="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
      <div class="px-6 py-4 border-b border-gray-100 flex items-center justify-between">
        <div class="flex items-center gap-3">
          <span class="w-2.5 h-2.5 rounded-full bg-red-500 animate-pulse"></span>
          <h2 class="font-semibold text-gray-900">Live Transcript</h2>
        </div>
        <span class="text-sm text-gray-500 font-mono">{{ activeCall.callerPhone }}</span>
      </div>
      <div ref="transcriptContainer" class="p-6 max-h-96 overflow-y-auto space-y-3">
        <div
          v-for="(entry, i) in activeCall.transcript"
          :key="i"
          :class="['p-3 rounded-lg max-w-[80%] transition-all', entry.speaker === 'assistant' ? 'bg-primary-50 ml-auto' : 'bg-gray-100']"
        >
          <div class="flex justify-between items-center mb-1">
            <span class="text-xs font-semibold" :class="entry.speaker === 'assistant' ? 'text-primary-600' : 'text-gray-600'">
              {{ entry.speaker === 'assistant' ? 'Alex (AI)' : 'Caller' }}
            </span>
            <span class="text-xs text-gray-400">{{ new Date(entry.at).toLocaleTimeString() }}</span>
          </div>
          <p class="text-sm text-gray-800">{{ entry.text }}</p>
        </div>
        <div v-if="!activeCall.transcript.length" class="text-center text-gray-400 text-sm py-8">
          Waiting for conversation to begin...
        </div>
      </div>
    </div>

    <!-- No active call -->
    <div v-else class="bg-white rounded-xl shadow-sm border border-gray-200 p-12 text-center">
      <p class="text-gray-400">No active call. The transcript will appear here when a call comes in.</p>
    </div>

    <!-- Recent calls -->
    <div v-if="callHistory.length" class="mt-6">
      <h3 class="text-sm font-semibold text-gray-500 uppercase tracking-wide mb-3">Recent Calls (this session)</h3>
      <div class="space-y-2">
        <div v-for="call in callHistory" :key="call.callId"
          class="bg-white rounded-lg border border-gray-200 px-4 py-3 flex items-center justify-between">
          <div class="flex items-center gap-3">
            <span class="w-8 h-8 bg-gray-100 rounded-full flex items-center justify-center">
              <svg class="w-4 h-4 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                  d="M3 5a2 2 0 012-2h3.28a1 1 0 01.948.684l1.498 4.493a1 1 0 01-.502 1.21l-2.257 1.13a11.042 11.042 0 005.516 5.516l1.13-2.257a1 1 0 011.21-.502l4.493 1.498a1 1 0 01.684.949V19a2 2 0 01-2 2h-1C9.716 21 3 14.284 3 6V5z" />
              </svg>
            </span>
            <div>
              <p class="text-sm font-medium text-gray-900">{{ call.callerPhone }}</p>
              <p class="text-xs text-gray-400">{{ new Date(call.endedAt).toLocaleTimeString() }}</p>
            </div>
          </div>
          <div class="flex items-center gap-4 text-sm text-gray-500">
            <span>{{ formatDuration(call.duration) }}</span>
            <span>${{ call.cost.toFixed(4) }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
@keyframes vibrate {
  0%, 100% { transform: translate(0, 0) rotate(0deg); }
  10% { transform: translate(-2px, -1px) rotate(-1deg); }
  20% { transform: translate(2px, 1px) rotate(1deg); }
  30% { transform: translate(-2px, 1px) rotate(-1deg); }
  40% { transform: translate(2px, -1px) rotate(1deg); }
  50% { transform: translate(-1px, 2px) rotate(0deg); }
  60% { transform: translate(1px, -2px) rotate(1deg); }
  70% { transform: translate(-1px, -1px) rotate(-1deg); }
  80% { transform: translate(1px, 1px) rotate(1deg); }
  90% { transform: translate(-1px, 0) rotate(-1deg); }
}

.animate-vibrate {
  animation: vibrate 0.3s linear infinite;
}

@keyframes ping-slow {
  0% { transform: scale(1); opacity: 1; }
  75% { transform: scale(1.3); opacity: 0.5; }
  100% { transform: scale(1); opacity: 1; }
}

.animate-ping-slow {
  animation: ping-slow 1.5s ease-in-out infinite;
}
</style>
