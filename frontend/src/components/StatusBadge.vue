<script setup lang="ts">
const props = defineProps<{ status: string }>()

const colorMap: Record<string, string> = {
  completed: 'bg-green-100 text-green-800',
  in_progress: 'bg-blue-100 text-blue-800',
  ringing: 'bg-yellow-100 text-yellow-800',
  failed: 'bg-red-100 text-red-800',
  pending: 'bg-yellow-100 text-yellow-800',
  confirmed: 'bg-green-100 text-green-800',
  cancelled: 'bg-red-100 text-red-800',
  rescheduled: 'bg-orange-100 text-orange-800',
  new: 'bg-blue-100 text-blue-800',
  contacted: 'bg-purple-100 text-purple-800',
  qualified: 'bg-indigo-100 text-indigo-800',
  converted: 'bg-green-100 text-green-800',
  lost: 'bg-gray-100 text-gray-800',
  indexed: 'bg-green-100 text-green-800',
  indexing: 'bg-blue-100 text-blue-800',
}

const classes = computed(() => colorMap[props.status] || 'bg-gray-100 text-gray-800')
</script>

<template>
  <span :class="['inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium', classes]">
    {{ status.replace('_', ' ') }}
  </span>
</template>
