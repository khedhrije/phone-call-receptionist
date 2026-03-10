<script setup lang="ts">
const dashboard = useDashboard()
const stats = ref<any>(null)
const costData = ref<any[]>([])

onMounted(async () => {
  try {
    stats.value = await dashboard.stats()
    const now = new Date()
    const from = new Date(now.getTime() - 30 * 24 * 60 * 60 * 1000).toISOString()
    const to = now.toISOString()
    const costs = await dashboard.costAnalytics(from, to)
    costData.value = costs.days || []
  } catch {}
})
</script>

<template>
  <div>
    <h1 class="text-2xl font-bold text-gray-900 mb-6">Dashboard</h1>

    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
      <StatsCard title="Total Calls" :value="stats?.totalCalls || 0" />
      <StatsCard title="Appointments" :value="stats?.totalAppointments || 0" />
      <StatsCard title="Leads" :value="stats?.totalLeads || 0" />
      <StatsCard title="Avg Cost/Call" :value="`$${(stats?.avgCostPerCall || 0).toFixed(4)}`" />
    </div>

    <CostChart :data="costData" />
  </div>
</template>
