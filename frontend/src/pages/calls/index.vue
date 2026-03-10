<script setup lang="ts">
const calls = useCalls()
const data = ref<any>({ items: [], total: 0 })
const page = ref(1)
const statusFilter = ref('')

const columns = [
  { key: 'callerPhone', label: 'Caller' },
  { key: 'status', label: 'Status' },
  { key: 'durationSeconds', label: 'Duration' },
  { key: 'totalCostUsd', label: 'Cost' },
  { key: 'createdAt', label: 'Date' },
]

const load = async () => {
  try {
    data.value = await calls.list({ page: page.value, pageSize: 20, status: statusFilter.value || undefined })
  } catch {}
}

onMounted(load)
watch([page, statusFilter], load)
</script>

<template>
  <div>
    <div class="flex justify-between items-center mb-6">
      <h1 class="text-2xl font-bold text-gray-900">Call History</h1>
      <select v-model="statusFilter" class="border rounded-lg px-3 py-2 text-sm">
        <option value="">All Status</option>
        <option value="completed">Completed</option>
        <option value="in_progress">In Progress</option>
        <option value="failed">Failed</option>
      </select>
    </div>

    <DataTable :columns="columns" :rows="data.items || []" :total="data.total" :page="page" :page-size="20"
      @page-change="page = $event" @row-click="navigateTo(`/calls/${$event.id}`)">
      <template #status="{ value }"><StatusBadge :status="value" /></template>
      <template #durationSeconds="{ value }">{{ Math.floor(value / 60) }}m {{ value % 60 }}s</template>
      <template #totalCostUsd="{ value }">${{ value?.toFixed(4) }}</template>
      <template #createdAt="{ value }">{{ new Date(value).toLocaleDateString() }}</template>
    </DataTable>
  </div>
</template>
