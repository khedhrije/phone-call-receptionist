<script setup lang="ts">
defineProps<{
  data: Array<{ date: string; twilioCost: number; llmCost: number }>
}>()
</script>

<template>
  <div class="bg-white rounded-xl shadow-sm border p-6">
    <h3 class="text-lg font-semibold mb-4">Cost Breakdown</h3>
    <div v-if="data?.length" class="space-y-2">
      <div v-for="item in data" :key="item.date" class="flex items-center gap-4">
        <span class="text-sm text-gray-600 w-24">{{ item.date }}</span>
        <div class="flex-1 flex gap-1 h-6">
          <div class="bg-blue-400 rounded" :style="{ width: `${(item.twilioCost / 1) * 100}%` }" :title="`Twilio: $${item.twilioCost.toFixed(4)}`"></div>
          <div class="bg-purple-400 rounded" :style="{ width: `${(item.llmCost / 1) * 100}%` }" :title="`LLM: $${item.llmCost.toFixed(4)}`"></div>
        </div>
        <span class="text-sm font-medium w-20 text-right">${{ (item.twilioCost + item.llmCost).toFixed(4) }}</span>
      </div>
    </div>
    <p v-else class="text-sm text-gray-400">No cost data available</p>
    <div class="flex gap-4 mt-4 text-xs text-gray-500">
      <span class="flex items-center gap-1"><span class="w-3 h-3 bg-blue-400 rounded"></span> Twilio</span>
      <span class="flex items-center gap-1"><span class="w-3 h-3 bg-purple-400 rounded"></span> LLM</span>
    </div>
  </div>
</template>
