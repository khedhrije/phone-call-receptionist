<script setup lang="ts">
defineProps<{
  columns: Array<{ key: string; label: string }>
  rows: any[]
  total?: number
  page?: number
  pageSize?: number
}>()

const emit = defineEmits<{
  (e: 'page-change', page: number): void
  (e: 'row-click', row: any): void
}>()
</script>

<template>
  <div class="bg-white rounded-xl shadow-sm border overflow-hidden">
    <div class="overflow-x-auto">
      <table class="w-full">
        <thead class="bg-gray-50 border-b">
          <tr>
            <th v-for="col in columns" :key="col.key" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
              {{ col.label }}
            </th>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-200">
          <tr v-for="(row, i) in rows" :key="i" @click="emit('row-click', row)" class="hover:bg-gray-50 cursor-pointer">
            <td v-for="col in columns" :key="col.key" class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
              <slot :name="col.key" :row="row" :value="row[col.key]">
                {{ row[col.key] }}
              </slot>
            </td>
          </tr>
          <tr v-if="!rows?.length">
            <td :colspan="columns.length" class="px-6 py-8 text-center text-sm text-gray-500">No data available</td>
          </tr>
        </tbody>
      </table>
    </div>
    <div v-if="total && pageSize" class="px-6 py-3 border-t flex justify-between items-center">
      <span class="text-sm text-gray-500">{{ total }} total</span>
      <div class="flex gap-2">
        <button @click="emit('page-change', (page || 1) - 1)" :disabled="(page || 1) <= 1"
          class="px-3 py-1 text-sm border rounded disabled:opacity-50">Prev</button>
        <button @click="emit('page-change', (page || 1) + 1)" :disabled="(page || 1) * (pageSize || 20) >= (total || 0)"
          class="px-3 py-1 text-sm border rounded disabled:opacity-50">Next</button>
      </div>
    </div>
  </div>
</template>
