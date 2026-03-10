<script setup lang="ts">
const route = useRoute()
const calls = useCalls()
const call = ref<any>(null)

onMounted(async () => {
  try {
    call.value = await calls.detail(route.params.id as string)
  } catch {}
})
</script>

<template>
  <div v-if="call">
    <div class="flex items-center gap-4 mb-6">
      <button @click="navigateTo('/calls')" class="text-primary-600 hover:text-primary-800">&larr; Back</button>
      <h1 class="text-2xl font-bold text-gray-900">Call Detail</h1>
      <StatusBadge :status="call.status" />
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6 mb-6">
      <div class="bg-white rounded-xl shadow-sm border p-4">
        <p class="text-sm text-gray-500">Caller</p>
        <p class="text-lg font-semibold">{{ call.callerPhone }}</p>
      </div>
      <div class="bg-white rounded-xl shadow-sm border p-4">
        <p class="text-sm text-gray-500">Duration</p>
        <p class="text-lg font-semibold">{{ Math.floor(call.durationSeconds / 60) }}m {{ call.durationSeconds % 60 }}s</p>
      </div>
      <div class="bg-white rounded-xl shadow-sm border p-4">
        <p class="text-sm text-gray-500">Total Cost</p>
        <p class="text-lg font-semibold">${{ call.totalCostUsd?.toFixed(4) }}</p>
      </div>
    </div>

    <div class="bg-white rounded-xl shadow-sm border p-6">
      <h2 class="text-lg font-semibold mb-4">Transcript</h2>
      <CallTranscript :transcript="call.transcript || []" />
    </div>
  </div>
  <div v-else class="text-center text-gray-500 py-12">Loading...</div>
</template>
