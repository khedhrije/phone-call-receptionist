<script setup lang="ts">
const leads = useLeads()
const data = ref<any>({ items: [], total: 0 })
const page = ref(1)

const columns = [
  { key: 'phone', label: 'Phone' },
  { key: 'name', label: 'Name' },
  { key: 'email', label: 'Email' },
  { key: 'status', label: 'Status' },
  { key: 'createdAt', label: 'Created' },
]

const load = async () => {
  try {
    data.value = await leads.list({ page: page.value, pageSize: 20 })
  } catch {}
}

onMounted(load)
</script>

<template>
  <div>
    <h1 class="text-2xl font-bold text-gray-900 mb-6">Leads</h1>

    <DataTable :columns="columns" :rows="data.items || []" :total="data.total" :page="page" :page-size="20"
      @page-change="page = $event">
      <template #status="{ value }"><StatusBadge :status="value" /></template>
      <template #createdAt="{ value }">{{ new Date(value).toLocaleDateString() }}</template>
    </DataTable>
  </div>
</template>
