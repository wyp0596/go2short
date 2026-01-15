<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { getAPITokens, createAPIToken, deleteAPIToken, type APIToken } from '../api'
import { useToast } from '../composables/useToast'
import ConfirmDialog from '../components/ConfirmDialog.vue'

const { success, error: showError } = useToast()

const tokens = ref<APIToken[]>([])
const loading = ref(true)
const origin = window.location.origin
const curlCommand = computed(() => `curl -X POST ${origin}/api/links -H 'Authorization: Bearer YOUR_TOKEN' -H 'Content-Type: application/json' -d '{"long_url":"https://example.com"}'`)

const showModal = ref(false)
const formName = ref('')
const formError = ref('')
const formLoading = ref(false)

const showTokenModal = ref(false)
const newToken = ref('')

// Delete confirmation
const showDeleteConfirm = ref(false)
const deletingId = ref<number | null>(null)

async function loadTokens() {
  loading.value = true
  try {
    const data = await getAPITokens()
    tokens.value = data.tokens || []
  } finally {
    loading.value = false
  }
}

onMounted(loadTokens)

function openCreate() {
  formName.value = ''
  formError.value = ''
  showModal.value = true
}

async function handleSubmit() {
  formError.value = ''
  formLoading.value = true
  try {
    const result = await createAPIToken(formName.value)
    showModal.value = false
    newToken.value = result.token
    showTokenModal.value = true
    await loadTokens()
    success('Token created successfully')
  } catch (e: unknown) {
    const err = e as { response?: { data?: { error?: string } } }
    formError.value = err.response?.data?.error || 'Failed to create token'
  } finally {
    formLoading.value = false
  }
}

function confirmDelete(id: number) {
  deletingId.value = id
  showDeleteConfirm.value = true
}

async function handleDelete() {
  if (deletingId.value === null) return
  try {
    await deleteAPIToken(deletingId.value)
    showDeleteConfirm.value = false
    await loadTokens()
    success('Token deleted successfully')
  } catch {
    showError('Failed to delete token')
  }
}

function copyToClipboard(text: string) {
  navigator.clipboard.writeText(text)
  success('Copied to clipboard')
}

function formatDate(date: string) {
  return new Date(date).toLocaleString()
}
</script>

<template>
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <div class="sm:flex sm:items-center sm:justify-between mb-6">
      <h1 class="text-2xl font-bold text-gray-900">API Tokens</h1>
      <button
        @click="openCreate"
        class="mt-4 sm:mt-0 inline-flex items-center px-4 py-2.5 border border-transparent rounded-lg shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 transition-all transform hover:scale-[1.02] active:scale-[0.98]"
      >
        <svg class="w-5 h-5 mr-1.5 -ml-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
        </svg>
        Create Token
      </button>
    </div>

    <div class="bg-white shadow-sm rounded-xl overflow-hidden">
      <table class="min-w-full divide-y divide-gray-200">
        <thead class="bg-gray-50">
          <tr>
            <th class="px-6 py-3.5 text-left text-xs font-semibold text-gray-500 uppercase tracking-wider">Name</th>
            <th class="px-6 py-3.5 text-left text-xs font-semibold text-gray-500 uppercase tracking-wider">Created</th>
            <th class="px-6 py-3.5 text-left text-xs font-semibold text-gray-500 uppercase tracking-wider">Last Used</th>
            <th class="px-6 py-3.5 text-right text-xs font-semibold text-gray-500 uppercase tracking-wider">Actions</th>
          </tr>
        </thead>
        <tbody class="bg-white divide-y divide-gray-200">
          <tr v-if="loading">
            <td colspan="4" class="px-6 py-12 text-center">
              <svg class="animate-spin h-8 w-8 text-blue-500 mx-auto" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
              </svg>
              <p class="mt-2 text-gray-500">Loading...</p>
            </td>
          </tr>
          <tr v-else-if="tokens.length === 0">
            <td colspan="4" class="px-6 py-12 text-center">
              <svg class="mx-auto h-12 w-12 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z" />
              </svg>
              <h3 class="mt-2 text-sm font-medium text-gray-900">No tokens found</h3>
              <p class="mt-1 text-sm text-gray-500">Create a token to use the API.</p>
              <button
                @click="openCreate"
                class="mt-4 inline-flex items-center px-4 py-2 border border-transparent rounded-lg shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700"
              >
                <svg class="w-5 h-5 mr-1.5 -ml-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
                </svg>
                Create Token
              </button>
            </td>
          </tr>
          <tr v-for="token in tokens" :key="token.id" class="hover:bg-gray-50 transition-colors">
            <td class="px-6 py-4 whitespace-nowrap">
              <div class="flex items-center gap-2">
                <div class="w-8 h-8 bg-blue-100 rounded-lg flex items-center justify-center">
                  <svg class="w-4 h-4 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z" />
                  </svg>
                </div>
                <span class="text-sm font-medium text-gray-900">{{ token.name }}</span>
              </div>
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{{ formatDate(token.created_at) }}</td>
            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
              {{ token.last_used_at ? formatDate(token.last_used_at) : 'Never' }}
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
              <button
                @click="confirmDelete(token.id)"
                class="p-1.5 text-red-600 hover:bg-red-50 rounded-lg transition-colors"
                title="Delete"
              >
                <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                </svg>
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div class="mt-6 bg-blue-50 border border-blue-200 rounded-xl p-5">
      <div class="flex items-start gap-3">
        <div class="w-8 h-8 bg-blue-100 rounded-lg flex items-center justify-center flex-shrink-0">
          <svg class="w-4 h-4 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
        </div>
        <div class="flex-1">
          <h3 class="text-sm font-semibold text-blue-900">API Usage</h3>
          <p class="mt-1 text-sm text-blue-700">
            Use your token to create short links via the API:
          </p>
          <div class="mt-3 flex items-start gap-2">
            <pre class="flex-1 text-xs bg-blue-100 p-3 rounded-lg overflow-x-auto font-mono text-blue-900">curl -X POST {{ origin }}/api/links \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"long_url":"https://example.com"}'</pre>
            <button
              @click="copyToClipboard(curlCommand)"
              class="px-3 py-2 text-xs border border-blue-300 rounded-lg text-blue-700 hover:bg-blue-100 transition-colors font-medium"
            >
              Copy
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Create Modal -->
    <Teleport to="body">
      <Transition name="fade">
        <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
          <div class="absolute inset-0 bg-black/50 backdrop-blur-sm" @click="showModal = false"></div>
          <Transition name="scale" appear>
            <div class="relative bg-white rounded-xl shadow-2xl max-w-md w-full">
              <div class="px-6 py-4 border-b">
                <h3 class="text-lg font-semibold text-gray-900">Create API Token</h3>
              </div>
              <form @submit.prevent="handleSubmit" class="px-6 py-4 space-y-4">
                <div>
                  <label class="block text-sm font-medium text-gray-700 mb-1.5">Token Name</label>
                  <input
                    v-model="formName"
                    type="text"
                    required
                    placeholder="e.g., my-app, production"
                    class="w-full px-4 py-2.5 border border-gray-300 rounded-lg shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500/50 focus:border-blue-500 transition-all"
                  />
                </div>
                <div v-if="formError" class="flex items-center gap-2 p-3 bg-red-50 border border-red-200 rounded-lg text-red-700 text-sm">
                  <svg class="w-5 h-5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                  {{ formError }}
                </div>
                <div class="flex justify-end gap-3 pt-2">
                  <button
                    type="button"
                    @click="showModal = false"
                    class="px-4 py-2.5 border border-gray-300 rounded-lg text-sm font-medium text-gray-700 hover:bg-gray-50 transition-colors"
                  >
                    Cancel
                  </button>
                  <button
                    type="submit"
                    :disabled="formLoading"
                    class="px-4 py-2.5 border border-transparent rounded-lg shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 disabled:opacity-50 transition-colors"
                  >
                    <span v-if="formLoading" class="inline-flex items-center gap-2">
                      <svg class="animate-spin h-4 w-4" fill="none" viewBox="0 0 24 24">
                        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                        <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                      </svg>
                      Creating...
                    </span>
                    <span v-else>Create</span>
                  </button>
                </div>
              </form>
            </div>
          </Transition>
        </div>
      </Transition>
    </Teleport>

    <!-- Token Display Modal -->
    <Teleport to="body">
      <Transition name="fade">
        <div v-if="showTokenModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
          <div class="absolute inset-0 bg-black/50 backdrop-blur-sm"></div>
          <Transition name="scale" appear>
            <div class="relative bg-white rounded-xl shadow-2xl max-w-lg w-full">
              <div class="px-6 py-4 border-b">
                <h3 class="text-lg font-semibold text-gray-900">Token Created</h3>
              </div>
              <div class="px-6 py-4">
                <div class="flex items-start gap-3 p-4 bg-yellow-50 border border-yellow-200 rounded-lg mb-4">
                  <svg class="w-5 h-5 text-yellow-600 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                  </svg>
                  <p class="text-sm text-yellow-800">
                    Copy this token now. You won't be able to see it again!
                  </p>
                </div>
                <div class="flex items-center gap-2">
                  <input
                    :value="newToken"
                    readonly
                    class="flex-1 px-4 py-2.5 bg-gray-100 border border-gray-300 rounded-lg font-mono text-sm"
                  />
                  <button
                    @click="copyToClipboard(newToken)"
                    class="px-4 py-2.5 border border-gray-300 rounded-lg text-sm font-medium text-gray-700 hover:bg-gray-50 transition-colors"
                  >
                    Copy
                  </button>
                </div>
              </div>
              <div class="px-6 py-4 border-t flex justify-end">
                <button
                  @click="showTokenModal = false"
                  class="px-4 py-2.5 border border-transparent rounded-lg shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 transition-colors"
                >
                  Done
                </button>
              </div>
            </div>
          </Transition>
        </div>
      </Transition>
    </Teleport>

    <!-- Delete Confirmation -->
    <ConfirmDialog
      v-if="showDeleteConfirm"
      title="Delete Token"
      message="Are you sure you want to delete this token? Any applications using this token will no longer be able to access the API."
      confirm-text="Delete"
      :danger="true"
      @confirm="handleDelete"
      @cancel="showDeleteConfirm = false"
    />
  </div>
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
