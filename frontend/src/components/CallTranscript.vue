<script setup lang="ts">
defineProps<{
  transcript: Array<{ speaker: string; text: string; at: string }>
}>()
</script>

<template>
  <div class="space-y-3">
    <div v-for="(entry, i) in transcript" :key="i"
      :class="['p-3 rounded-lg max-w-[80%]', entry.speaker === 'assistant' ? 'bg-primary-50 ml-auto' : 'bg-gray-100']">
      <div class="flex justify-between items-center mb-1">
        <span class="text-xs font-semibold" :class="entry.speaker === 'assistant' ? 'text-primary-600' : 'text-gray-600'">
          {{ entry.speaker === 'assistant' ? 'Alex (AI)' : 'Caller' }}
        </span>
        <span class="text-xs text-gray-400">{{ new Date(entry.at).toLocaleTimeString() }}</span>
      </div>
      <p class="text-sm text-gray-800">{{ entry.text }}</p>
    </div>
  </div>
</template>
