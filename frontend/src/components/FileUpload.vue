<script setup lang="ts">
const emit = defineEmits<{
  (e: 'upload', file: File): void
}>()

const dragging = ref(false)

const onDrop = (e: DragEvent) => {
  dragging.value = false
  const file = e.dataTransfer?.files[0]
  if (file) emit('upload', file)
}

const onFileSelect = (e: Event) => {
  const file = (e.target as HTMLInputElement).files?.[0]
  if (file) emit('upload', file)
}
</script>

<template>
  <div
    @dragover.prevent="dragging = true"
    @dragleave="dragging = false"
    @drop.prevent="onDrop"
    :class="['border-2 border-dashed rounded-xl p-8 text-center transition-colors cursor-pointer',
      dragging ? 'border-primary-500 bg-primary-50' : 'border-gray-300 hover:border-gray-400']"
    @click="($refs.fileInput as HTMLInputElement)?.click()"
  >
    <input ref="fileInput" type="file" class="hidden" @change="onFileSelect" />
    <p class="text-gray-500">Drop a file here or click to upload</p>
    <p class="text-xs text-gray-400 mt-2">Supports PDF, TXT, DOCX</p>
  </div>
</template>
