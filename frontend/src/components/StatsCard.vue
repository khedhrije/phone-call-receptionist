<script setup lang="ts">
defineProps<{
  title: string
  value: string | number
  subtitle?: string
}>()
</script>

<template>
  <div class="bg-white rounded-xl shadow-sm p-6 border">
    <p class="text-sm text-gray-500 font-medium">{{ title }}</p>
    <p class="text-3xl font-bold text-gray-900 mt-2">{{ value }}</p>
    <p v-if="subtitle" class="text-sm text-gray-400 mt-1">{{ subtitle }}</p>
  </div>
</template>
