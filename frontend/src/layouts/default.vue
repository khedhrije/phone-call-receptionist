<script setup lang="ts">
const { isAuthenticated } = useAuth()

onMounted(() => {
  if (!isAuthenticated.value) navigateTo('/')
})
</script>

<template>
  <div class="flex h-screen bg-gray-100">
    <Sidebar />
    <div class="flex-1 flex flex-col overflow-hidden">
      <header class="bg-white shadow-sm border-b px-6 py-4 flex justify-between items-center">
        <h1 class="text-xl font-semibold text-gray-800">Phone Call Receptionist</h1>
        <div class="flex items-center gap-4">
          <span class="text-sm text-gray-600">{{ useAuth().user.value?.email }}</span>
          <button @click="useAuth().logout()" class="text-sm text-red-600 hover:text-red-800">Logout</button>
        </div>
      </header>
      <main class="flex-1 overflow-y-auto p-6">
        <slot />
      </main>
    </div>
  </div>
</template>
