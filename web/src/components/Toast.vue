<script setup lang="ts">
defineProps<{
  message: string
  type?: 'success' | 'error' | 'info'
}>()

const emit = defineEmits<{ close: [] }>()
</script>

<template>
  <div
    class="flex items-center gap-3 px-4 py-3 rounded-lg shadow-lg text-white min-w-[280px] max-w-md"
    :class="{
      'bg-green-500': type === 'success',
      'bg-red-500': type === 'error',
      'bg-blue-500': type === 'info' || !type
    }"
  >
    <!-- Icon -->
    <svg v-if="type === 'success'" class="w-5 h-5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
    </svg>
    <svg v-else-if="type === 'error'" class="w-5 h-5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
    </svg>
    <svg v-else class="w-5 h-5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
    </svg>

    <!-- Message -->
    <span class="flex-1 text-sm font-medium">{{ message }}</span>

    <!-- Close button -->
    <button @click="emit('close')" class="flex-shrink-0 hover:opacity-80 transition-opacity">
      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
      </svg>
    </button>
  </div>
</template>
