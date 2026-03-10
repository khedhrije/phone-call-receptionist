<script setup lang="ts">
const kb = useKnowledgeBase()
const docs = ref<any[]>([])
const searchQuery = ref('')
const searchResults = ref<any>(null)
const uploading = ref(false)

const columns = [
  { key: 'filename', label: 'Filename' },
  { key: 'mimeType', label: 'Type' },
  { key: 'chunkCount', label: 'Chunks' },
  { key: 'status', label: 'Status' },
  { key: 'createdAt', label: 'Uploaded' },
]

const load = async () => {
  try {
    docs.value = await kb.list()
  } catch {}
}

const onUpload = async (file: File) => {
  uploading.value = true
  try {
    await kb.upload(file)
    await load()
  } catch {} finally {
    uploading.value = false
  }
}

const onSearch = async () => {
  if (!searchQuery.value) return
  try {
    searchResults.value = await kb.search(searchQuery.value, 5)
  } catch {}
}

const onDelete = async (id: string) => {
  if (confirm('Delete this document?')) {
    await kb.deleteDoc(id)
    await load()
  }
}

onMounted(load)
</script>

<template>
  <div>
    <h1 class="text-2xl font-bold text-gray-900 mb-6">Knowledge Base</h1>

    <FileUpload @upload="onUpload" class="mb-6" />
    <p v-if="uploading" class="text-sm text-primary-600 mb-4">Uploading and indexing...</p>

    <div class="mb-6">
      <div class="flex gap-2">
        <input v-model="searchQuery" @keyup.enter="onSearch" placeholder="Search knowledge base..."
          class="flex-1 px-4 py-2 border rounded-lg focus:ring-2 focus:ring-primary-500" />
        <button @click="onSearch" class="px-4 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700">Search</button>
      </div>
      <div v-if="searchResults" class="mt-4 bg-white rounded-xl shadow-sm border p-4">
        <p class="font-semibold mb-2">Answer ({{ searchResults.provider }}):</p>
        <p class="text-gray-700">{{ searchResults.answer }}</p>
      </div>
    </div>

    <DataTable :columns="columns" :rows="docs">
      <template #status="{ value }"><StatusBadge :status="value" /></template>
      <template #createdAt="{ value }">{{ new Date(value).toLocaleDateString() }}</template>
    </DataTable>
  </div>
</template>
