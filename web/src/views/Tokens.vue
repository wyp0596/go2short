<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getAPITokens, createAPIToken, deleteAPIToken, type APIToken } from '../api'

const tokens = ref<APIToken[]>([])
const loading = ref(true)
const origin = window.location.origin

const showModal = ref(false)
const formName = ref('')
const formError = ref('')
const formLoading = ref(false)

const showTokenModal = ref(false)
const newToken = ref('')

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
  } catch (e: unknown) {
    const err = e as { response?: { data?: { error?: string } } }
    formError.value = err.response?.data?.error || 'Failed to create token'
  } finally {
    formLoading.value = false
  }
}

async function handleDelete(id: number) {
  if (!confirm('Are you sure you want to delete this token?')) return
  try {
    await deleteAPIToken(id)
    await loadTokens()
  } catch {
    alert('Failed to delete token')
  }
}

function copyToClipboard(text: string) {
  navigator.clipboard.writeText(text)
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
        class="mt-4 sm:mt-0 inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700"
      >
        Create Token
      </button>
    </div>

    <div class="bg-white shadow overflow-hidden sm:rounded-lg">
      <table class="min-w-full divide-y divide-gray-200">
        <thead class="bg-gray-50">
          <tr>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Name</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Created</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Last Used</th>
            <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Actions</th>
          </tr>
        </thead>
        <tbody class="bg-white divide-y divide-gray-200">
          <tr v-if="loading">
            <td colspan="4" class="px-6 py-4 text-center text-gray-500">Loading...</td>
          </tr>
          <tr v-else-if="tokens.length === 0">
            <td colspan="4" class="px-6 py-4 text-center text-gray-500">No tokens found</td>
          </tr>
          <tr v-for="token in tokens" :key="token.id">
            <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">{{ token.name }}</td>
            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{{ formatDate(token.created_at) }}</td>
            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
              {{ token.last_used_at ? formatDate(token.last_used_at) : 'Never' }}
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
              <button @click="handleDelete(token.id)" class="text-red-600 hover:text-red-900">Delete</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div class="mt-6 bg-blue-50 border border-blue-200 rounded-lg p-4">
      <h3 class="text-sm font-medium text-blue-800">Usage</h3>
      <p class="mt-1 text-sm text-blue-700">
        Use the token to create short links via API:
      </p>
      <pre class="mt-2 text-xs bg-blue-100 p-2 rounded overflow-x-auto">curl -X POST {{ origin }}/api/links \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"long_url":"https://example.com"}'</pre>
    </div>

    <!-- Create Modal -->
    <div v-if="showModal" class="fixed inset-0 bg-gray-500 bg-opacity-75 flex items-center justify-center z-50">
      <div class="bg-white rounded-lg shadow-xl max-w-md w-full mx-4">
        <div class="px-6 py-4 border-b">
          <h3 class="text-lg font-medium text-gray-900">Create API Token</h3>
        </div>
        <form @submit.prevent="handleSubmit" class="px-6 py-4 space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700">Token Name</label>
            <input
              v-model="formName"
              type="text"
              required
              placeholder="e.g., my-app, production"
              class="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500"
            />
          </div>
          <div v-if="formError" class="text-red-600 text-sm">{{ formError }}</div>
          <div class="flex justify-end space-x-3 pt-4">
            <button
              type="button"
              @click="showModal = false"
              class="px-4 py-2 border border-gray-300 rounded-md text-sm font-medium text-gray-700 hover:bg-gray-50"
            >
              Cancel
            </button>
            <button
              type="submit"
              :disabled="formLoading"
              class="px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 disabled:opacity-50"
            >
              {{ formLoading ? 'Creating...' : 'Create' }}
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- Token Display Modal -->
    <div v-if="showTokenModal" class="fixed inset-0 bg-gray-500 bg-opacity-75 flex items-center justify-center z-50">
      <div class="bg-white rounded-lg shadow-xl max-w-lg w-full mx-4">
        <div class="px-6 py-4 border-b">
          <h3 class="text-lg font-medium text-gray-900">Token Created</h3>
        </div>
        <div class="px-6 py-4">
          <div class="bg-yellow-50 border border-yellow-200 rounded-lg p-4 mb-4">
            <p class="text-sm text-yellow-800">
              Copy this token now. You won't be able to see it again!
            </p>
          </div>
          <div class="flex items-center gap-2">
            <input
              :value="newToken"
              readonly
              class="flex-1 px-3 py-2 bg-gray-100 border border-gray-300 rounded-md font-mono text-sm"
            />
            <button
              @click="copyToClipboard(newToken)"
              class="px-3 py-2 border border-gray-300 rounded-md text-sm font-medium text-gray-700 hover:bg-gray-50"
            >
              Copy
            </button>
          </div>
        </div>
        <div class="px-6 py-4 border-t flex justify-end">
          <button
            @click="showTokenModal = false"
            class="px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700"
          >
            Done
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
