<script setup lang="ts">
defineProps<{
  title?: string
  message: string
  confirmText?: string
  cancelText?: string
  danger?: boolean
}>()

const emit = defineEmits<{ confirm: []; cancel: [] }>()
</script>

<template>
  <Teleport to="body">
    <Transition name="fade">
      <div class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <!-- Backdrop -->
        <div
          class="absolute inset-0 bg-black/50 backdrop-blur-sm"
          @click="emit('cancel')"
        ></div>

        <!-- Dialog -->
        <Transition name="scale" appear>
          <div class="relative bg-white rounded-xl shadow-2xl max-w-md w-full p-6">
            <!-- Title -->
            <h3 class="text-lg font-semibold text-gray-900">
              {{ title || 'Confirm' }}
            </h3>

            <!-- Message -->
            <p class="mt-3 text-gray-600">{{ message }}</p>

            <!-- Actions -->
            <div class="mt-6 flex justify-end gap-3">
              <button
                @click="emit('cancel')"
                class="px-4 py-2 text-gray-700 hover:bg-gray-100 rounded-lg transition-colors"
              >
                {{ cancelText || 'Cancel' }}
              </button>
              <button
                @click="emit('confirm')"
                :class="danger ? 'bg-red-600 hover:bg-red-700' : 'bg-blue-600 hover:bg-blue-700'"
                class="px-4 py-2 text-white rounded-lg transition-colors"
              >
                {{ confirmText || 'Confirm' }}
              </button>
            </div>
          </div>
        </Transition>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
.scale-enter-active,
.scale-leave-active {
  transition: all 0.2s ease;
}
.scale-enter-from,
.scale-leave-to {
  transform: scale(0.95);
  opacity: 0;
}
</style>
