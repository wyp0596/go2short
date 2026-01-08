<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { getLinks, createLink, updateLink, deleteLink, setLinkDisabled, type Link } from '../api'

const router = useRouter()

const links = ref<Link[]>([])
const total = ref(0)
const page = ref(1)
const limit = ref(20)
const search = ref('')
const loading = ref(true)

const showModal = ref(false)
const editingLink = ref<Link | null>(null)
const formUrl = ref('')
const formCode = ref('')
const formExpires = ref('')
const formError = ref('')
const formLoading = ref(false)

const showQRModal = ref(false)
const qrCode = ref('')
const qrUrl = ref('')

async function loadLinks() {
  loading.value = true
  try {
    const data = await getLinks(page.value, limit.value, search.value)
    links.value = data.links
    total.value = data.total
  } finally {
    loading.value = false
  }
}

onMounted(loadLinks)

watch([page, search], loadLinks)

function openCreate() {
  editingLink.value = null
  formUrl.value = ''
  formCode.value = ''
  formExpires.value = ''
  formError.value = ''
  showModal.value = true
}

function openEdit(link: Link) {
  editingLink.value = link
  formUrl.value = link.long_url
  formCode.value = link.code
  formExpires.value = link.expires_at ? link.expires_at.slice(0, 16) : ''
  formError.value = ''
  showModal.value = true
}

async function handleSubmit() {
  formError.value = ''
  formLoading.value = true
  try {
    if (editingLink.value) {
      await updateLink(editingLink.value.code, formUrl.value, formExpires.value || undefined)
    } else {
      await createLink(formUrl.value, formCode.value || undefined, formExpires.value || undefined)
    }
    showModal.value = false
    await loadLinks()
  } catch (e: unknown) {
    const err = e as { response?: { data?: { error?: string } } }
    formError.value = err.response?.data?.error || 'Failed to save link'
  } finally {
    formLoading.value = false
  }
}

async function handleDelete(code: string) {
  if (!confirm('Are you sure you want to delete this link?')) return
  try {
    await deleteLink(code)
    await loadLinks()
  } catch {
    alert('Failed to delete link')
  }
}

async function handleToggleDisable(link: Link) {
  try {
    await setLinkDisabled(link.code, !link.is_disabled)
    await loadLinks()
  } catch {
    alert('Failed to update link')
  }
}

function viewStats(code: string) {
  router.push(`/links/${code}/stats`)
}

function copyToClipboard(text: string) {
  navigator.clipboard.writeText(text)
}

const totalPages = () => Math.ceil(total.value / limit.value)

function formatDate(date: string) {
  return new Date(date).toLocaleString()
}

function showQR(link: Link) {
  qrCode.value = link.code
  qrUrl.value = link.short_url
  showQRModal.value = true
}

function downloadQR() {
  const link = document.createElement('a')
  link.href = `/${qrCode.value}/qr?size=512`
  link.download = `${qrCode.value}-qr.png`
  link.click()
}
</script>

<template>
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <div class="sm:flex sm:items-center sm:justify-between mb-6">
      <h1 class="text-2xl font-bold text-gray-900">Links</h1>
      <button
        @click="openCreate"
        class="mt-4 sm:mt-0 inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700"
      >
        Create Link
      </button>
    </div>

    <div class="mb-4">
      <input
        v-model="search"
        type="text"
        placeholder="Search by code or URL..."
        class="block w-full sm:w-64 px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500"
        @input="page = 1"
      />
    </div>

    <div class="bg-white shadow overflow-hidden sm:rounded-lg">
      <table class="min-w-full divide-y divide-gray-200">
        <thead class="bg-gray-50">
          <tr>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Code</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Long URL</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Created</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
            <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Actions</th>
          </tr>
        </thead>
        <tbody class="bg-white divide-y divide-gray-200">
          <tr v-if="loading">
            <td colspan="5" class="px-6 py-4 text-center text-gray-500">Loading...</td>
          </tr>
          <tr v-else-if="links.length === 0">
            <td colspan="5" class="px-6 py-4 text-center text-gray-500">No links found</td>
          </tr>
          <tr v-for="link in links" :key="link.code">
            <td class="px-6 py-4 whitespace-nowrap">
              <div class="flex items-center gap-2">
                <a :href="link.short_url" target="_blank" class="text-blue-600 hover:text-blue-800">{{ link.code }}</a>
                <button @click="copyToClipboard(link.short_url)" class="text-gray-400 hover:text-gray-600" title="Copy short URL">
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 5H6a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2v-1M8 5a2 2 0 002 2h2a2 2 0 002-2M8 5a2 2 0 012-2h2a2 2 0 012 2m0 0h2a2 2 0 012 2v3m2 4H10m0 0l3-3m-3 3l3 3" />
                  </svg>
                </button>
              </div>
            </td>
            <td class="px-6 py-4">
              <div class="max-w-xs truncate text-sm text-gray-900">{{ link.long_url }}</div>
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{{ formatDate(link.created_at) }}</td>
            <td class="px-6 py-4 whitespace-nowrap">
              <span
                :class="link.is_disabled ? 'bg-red-100 text-red-800' : 'bg-green-100 text-green-800'"
                class="px-2 inline-flex text-xs leading-5 font-semibold rounded-full"
              >
                {{ link.is_disabled ? 'Disabled' : 'Active' }}
              </span>
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium space-x-2">
              <button @click="showQR(link)" class="text-purple-600 hover:text-purple-900">QR</button>
              <button @click="viewStats(link.code)" class="text-blue-600 hover:text-blue-900">Stats</button>
              <button @click="openEdit(link)" class="text-gray-600 hover:text-gray-900">Edit</button>
              <button @click="handleToggleDisable(link)" class="text-yellow-600 hover:text-yellow-900">
                {{ link.is_disabled ? 'Enable' : 'Disable' }}
              </button>
              <button @click="handleDelete(link.code)" class="text-red-600 hover:text-red-900">Delete</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div v-if="totalPages() > 1" class="mt-4 flex items-center justify-between">
      <div class="text-sm text-gray-700">
        Showing {{ (page - 1) * limit + 1 }} to {{ Math.min(page * limit, total) }} of {{ total }} results
      </div>
      <div class="flex space-x-2">
        <button
          :disabled="page === 1"
          @click="page--"
          class="px-3 py-1 border rounded text-sm disabled:opacity-50"
        >
          Previous
        </button>
        <button
          :disabled="page >= totalPages()"
          @click="page++"
          class="px-3 py-1 border rounded text-sm disabled:opacity-50"
        >
          Next
        </button>
      </div>
    </div>

    <!-- Modal -->
    <div v-if="showModal" class="fixed inset-0 bg-gray-500 bg-opacity-75 flex items-center justify-center z-50">
      <div class="bg-white rounded-lg shadow-xl max-w-md w-full mx-4">
        <div class="px-6 py-4 border-b">
          <h3 class="text-lg font-medium text-gray-900">{{ editingLink ? 'Edit Link' : 'Create Link' }}</h3>
        </div>
        <form @submit.prevent="handleSubmit" class="px-6 py-4 space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700">Long URL</label>
            <input
              v-model="formUrl"
              type="url"
              required
              class="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500"
            />
          </div>
          <div v-if="!editingLink">
            <label class="block text-sm font-medium text-gray-700">Custom Code (optional)</label>
            <input
              v-model="formCode"
              type="text"
              class="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700">Expires At (optional)</label>
            <input
              v-model="formExpires"
              type="datetime-local"
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
              {{ formLoading ? 'Saving...' : 'Save' }}
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- QR Code Modal -->
    <div v-if="showQRModal" class="fixed inset-0 bg-gray-500 bg-opacity-75 flex items-center justify-center z-50">
      <div class="bg-white rounded-lg shadow-xl max-w-sm w-full mx-4">
        <div class="px-6 py-4 border-b flex justify-between items-center">
          <h3 class="text-lg font-medium text-gray-900">QR Code</h3>
          <button @click="showQRModal = false" class="text-gray-400 hover:text-gray-600">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
        <div class="px-6 py-6 flex flex-col items-center">
          <img :src="`/${qrCode}/qr?size=256`" :alt="`QR code for ${qrCode}`" class="w-64 h-64" />
          <p class="mt-4 text-sm text-gray-600 break-all text-center">{{ qrUrl }}</p>
          <button
            @click="downloadQR"
            class="mt-4 inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700"
          >
            Download
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
