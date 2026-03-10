<script setup lang="ts">
defineProps<{ src: string }>()
</script>

<template>
  <audio v-if="src" controls class="w-full">
    <source :src="src" type="audio/mpeg" />
  </audio>
</template>
