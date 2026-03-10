<script setup lang="ts">
const api = useApi()
const settings = ref<any>({})
const saving = ref(false)
const saved = ref(false)

onMounted(async () => {
  try {
    settings.value = await api.get('/settings')
  } catch {}
})

const save = async () => {
  saving.value = true
  saved.value = false
  try {
    settings.value = await api.put('/settings', settings.value)
    saved.value = true
    setTimeout(() => saved.value = false, 3000)
  } catch {} finally {
    saving.value = false
  }
}
</script>

<template>
  <div>
    <h1 class="text-2xl font-bold text-gray-900 mb-6">Settings</h1>

    <div class="bg-white rounded-xl shadow-sm border p-6 max-w-2xl">
      <div class="space-y-4">
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Default LLM Provider</label>
          <select v-model="settings.defaultLlmProvider" class="w-full border rounded-lg px-3 py-2">
            <option value="gemini">Gemini</option>
            <option value="claude">Claude</option>
            <option value="openai">OpenAI</option>
            <option value="mistral">Mistral</option>
            <option value="deepseek">DeepSeek</option>
            <option value="glm">GLM</option>
          </select>
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Default Voice ID</label>
          <input v-model="settings.defaultVoiceId" class="w-full border rounded-lg px-3 py-2" />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Top K (Vector Search Results)</label>
          <input v-model.number="settings.topK" type="number" min="1" max="20" class="w-full border rounded-lg px-3 py-2" />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Max Call Duration (seconds)</label>
          <input v-model.number="settings.maxCallDurationSecs" type="number" min="60" max="1800" class="w-full border rounded-lg px-3 py-2" />
        </div>

        <div class="flex items-center gap-4 pt-4">
          <button @click="save" :disabled="saving"
            class="px-4 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700 disabled:opacity-50">
            {{ saving ? 'Saving...' : 'Save Settings' }}
          </button>
          <span v-if="saved" class="text-sm text-green-600">Settings saved!</span>
        </div>
      </div>
    </div>
  </div>
</template>
