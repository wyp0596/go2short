<script setup lang="ts">
defineProps<{
  type?: 'line' | 'circle' | 'card'
  width?: string
  height?: string
}>()
</script>

<template>
  <div
    class="animate-pulse bg-gray-200 rounded"
    :class="{
      'h-4': type === 'line' || !type,
      'rounded-full': type === 'circle',
      'h-32': type === 'card'
    }"
    :style="{ width: width || '100%', height }"
  ></div>
</template>
