<script setup lang="ts">
const route = useRoute()

const links = [
  { path: '/dashboard', label: 'Dashboard', icon: 'D' },
  { path: '/live', label: 'Live Monitor', icon: 'P' },
  { path: '/calls', label: 'Calls', icon: 'C' },
  { path: '/appointments', label: 'Appointments', icon: 'A' },
  { path: '/leads', label: 'Leads', icon: 'L' },
  { path: '/knowledge-base', label: 'Knowledge Base', icon: 'K' },
  { path: '/settings', label: 'Settings', icon: 'S' },
]
</script>

<template>
  <aside class="w-64 bg-gray-900 text-white flex flex-col">
    <div class="p-6 border-b border-gray-700">
      <h2 class="text-lg font-bold">AI Receptionist</h2>
      <p class="text-xs text-gray-400 mt-1">IT Services</p>
    </div>
    <nav class="flex-1 p-4 space-y-1">
      <NuxtLink
        v-for="link in links"
        :key="link.path"
        :to="link.path"
        :class="[
          'flex items-center gap-3 px-4 py-3 rounded-lg text-sm transition-colors',
          route.path.startsWith(link.path)
            ? 'bg-primary-600 text-white'
            : 'text-gray-300 hover:bg-gray-800 hover:text-white'
        ]"
      >
        <span class="w-6 h-6 bg-gray-700 rounded flex items-center justify-center text-xs font-bold">{{ link.icon }}</span>
        {{ link.label }}
      </NuxtLink>
    </nav>
  </aside>
</template>
