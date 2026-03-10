<script setup lang="ts">
const appointments = useAppointments()
const data = ref<any>({ items: [], total: 0 })
const page = ref(1)

const columns = [
  { key: 'callerName', label: 'Name' },
  { key: 'callerPhone', label: 'Phone' },
  { key: 'serviceType', label: 'Service' },
  { key: 'scheduledAt', label: 'Scheduled' },
  { key: 'status', label: 'Status' },
]

const load = async () => {
  try {
    data.value = await appointments.list({ page: page.value, pageSize: 20 })
  } catch {}
}

const cancelAppt = async (id: string) => {
  if (confirm('Cancel this appointment?')) {
    await appointments.cancel(id)
    await load()
  }
}

onMounted(load)
</script>

<template>
  <div>
    <h1 class="text-2xl font-bold text-gray-900 mb-6">Appointments</h1>

    <DataTable :columns="columns" :rows="data.items || []" :total="data.total" :page="page" :page-size="20"
      @page-change="page = $event">
      <template #scheduledAt="{ value }">{{ new Date(value).toLocaleString() }}</template>
      <template #status="{ value }"><StatusBadge :status="value" /></template>
    </DataTable>
  </div>
</template>
