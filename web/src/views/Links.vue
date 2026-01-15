<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { getLinks, createLink, updateLink, deleteLink, setLinkDisabled, type Link } from '../api'
import { useToast } from '../composables/useToast'
import ConfirmDialog from '../components/ConfirmDialog.vue'

const router = useRouter()
const { success, error: showError } = useToast()

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

// Delete confirmation
const showDeleteConfirm = ref(false)
const deletingCode = ref('')

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
      success('Link updated successfully')
    } else {
      await createLink(formUrl.value, formCode.value || undefined, formExpires.value || undefined)
      success('Link created successfully')
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

function confirmDelete(code: string) {
  deletingCode.value = code
  showDeleteConfirm.value = true
}

async function handleDelete() {
  try {
    await deleteLink(deletingCode.value)
    showDeleteConfirm.value = false
    await loadLinks()
    success('Link deleted successfully')
  } catch {
    showError('Failed to delete link')
  }
}

async function handleToggleDisable(link: Link) {
  try {
    await setLinkDisabled(link.code, !link.is_disabled)
    await loadLinks()
    success(link.is_disabled ? 'Link enabled' : 'Link disabled')
  } catch {
    showError('Failed to update link')
  }
}

function viewStats(code: string) {
  router.push(`/admin/links/${code}/stats`)
}

function copyToClipboard(text: string) {
  navigator.clipboard.writeText(text)
  success('Copied to clipboard')
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
        class="mt-4 sm:mt-0 inline-flex items-center px-4 py-2.5 border border-transparent rounded-lg shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 transition-all transform hover:scale-[1.02] active:scale-[0.98]"
      >
        <svg class="w-5 h-5 mr-1.5 -ml-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
        </svg>
        Create Link
      </button>
    </div>

    <div class="mb-4">
      <div class="relative">
        <svg class="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
        </svg>
        <input
          v-model="search"
          type="text"
          placeholder="Search by code or URL..."
          class="block w-full sm:w-80 pl-10 pr-4 py-2.5 border border-gray-300 rounded-lg shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500/50 focus:border-blue-500 transition-all"
          @input="page = 1"
        />
      </div>
    </div>

    <div class="bg-white shadow-sm rounded-xl overflow-hidden">
      <table class="min-w-full divide-y divide-gray-200">
        <thead class="bg-gray-50">
          <tr>
            <th class="px-6 py-3.5 text-left text-xs font-semibold text-gray-500 uppercase tracking-wider">Code</th>
            <th class="px-6 py-3.5 text-left text-xs font-semibold text-gray-500 uppercase tracking-wider">Long URL</th>
            <th class="px-6 py-3.5 text-left text-xs font-semibold text-gray-500 uppercase tracking-wider">Created</th>
            <th class="px-6 py-3.5 text-left text-xs font-semibold text-gray-500 uppercase tracking-wider">Status</th>
            <th class="px-6 py-3.5 text-right text-xs font-semibold text-gray-500 uppercase tracking-wider">Actions</th>
          </tr>
        </thead>
        <tbody class="bg-white divide-y divide-gray-200">
          <tr v-if="loading">
            <td colspan="5" class="px-6 py-12 text-center">
              <svg class="animate-spin h-8 w-8 text-blue-500 mx-auto" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
              </svg>
              <p class="mt-2 text-gray-500">Loading...</p>
            </td>
          </tr>
          <tr v-else-if="links.length === 0">
            <td colspan="5" class="px-6 py-12 text-center">
              <svg class="mx-auto h-12 w-12 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
              </svg>
              <h3 class="mt-2 text-sm font-medium text-gray-900">No links found</h3>
              <p class="mt-1 text-sm text-gray-500">Get started by creating a new short link.</p>
              <button
                @click="openCreate"
                class="mt-4 inline-flex items-center px-4 py-2 border border-transparent rounded-lg shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700"
              >
                <svg class="w-5 h-5 mr-1.5 -ml-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
                </svg>
                Create Link
              </button>
            </td>
          </tr>
          <tr v-for="link in links" :key="link.code" class="hover:bg-gray-50 transition-colors">
            <td class="px-6 py-4 whitespace-nowrap">
              <div class="flex items-center gap-2">
                <a :href="link.short_url" target="_blank" class="text-blue-600 hover:text-blue-800 font-medium">{{ link.code }}</a>
                <button
                  @click="copyToClipboard(link.short_url)"
                  class="text-gray-400 hover:text-gray-600 p-1 hover:bg-gray-100 rounded transition-colors"
                  title="Copy short URL"
                >
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 5H6a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2v-1M8 5a2 2 0 002 2h2a2 2 0 002-2M8 5a2 2 0 012-2h2a2 2 0 012 2m0 0h2a2 2 0 012 2v3m2 4H10m0 0l3-3m-3 3l3 3" />
                  </svg>
                </button>
              </div>
            </td>
            <td class="px-6 py-4">
              <div class="max-w-xs truncate text-sm text-gray-600">{{ link.long_url }}</div>
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{{ formatDate(link.created_at) }}</td>
            <td class="px-6 py-4 whitespace-nowrap">
              <span
                :class="link.is_disabled ? 'bg-red-100 text-red-700' : 'bg-green-100 text-green-700'"
                class="px-2.5 py-0.5 inline-flex text-xs leading-5 font-semibold rounded-full"
              >
                {{ link.is_disabled ? 'Disabled' : 'Active' }}
              </span>
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
              <div class="flex items-center justify-end gap-1">
                <button
                  @click="showQR(link)"
                  class="p-1.5 text-purple-600 hover:bg-purple-50 rounded-lg transition-colors"
                  title="QR Code"
                >
                  <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v1m6 11h2m-6 0h-2v4m0-11v3m0 0h.01M12 12h4.01M16 20h4M4 12h4m12 0h.01M5 8h2a1 1 0 001-1V5a1 1 0 00-1-1H5a1 1 0 00-1 1v2a1 1 0 001 1zm12 0h2a1 1 0 001-1V5a1 1 0 00-1-1h-2a1 1 0 00-1 1v2a1 1 0 001 1zM5 20h2a1 1 0 001-1v-2a1 1 0 00-1-1H5a1 1 0 00-1 1v2a1 1 0 001 1z" />
                  </svg>
                </button>
                <button
                  @click="viewStats(link.code)"
                  class="p-1.5 text-blue-600 hover:bg-blue-50 rounded-lg transition-colors"
                  title="Statistics"
                >
                  <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
                  </svg>
                </button>
                <button
                  @click="openEdit(link)"
                  class="p-1.5 text-gray-600 hover:bg-gray-100 rounded-lg transition-colors"
                  title="Edit"
                >
                  <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                  </svg>
                </button>
                <button
                  @click="handleToggleDisable(link)"
                  class="p-1.5 text-yellow-600 hover:bg-yellow-50 rounded-lg transition-colors"
                  :title="link.is_disabled ? 'Enable' : 'Disable'"
                >
                  <svg v-if="link.is_disabled" class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                  </svg>
                  <svg v-else class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21" />
                  </svg>
                </button>
                <button
                  @click="confirmDelete(link.code)"
                  class="p-1.5 text-red-600 hover:bg-red-50 rounded-lg transition-colors"
                  title="Delete"
                >
                  <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                  </svg>
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div v-if="totalPages() > 1" class="mt-4 flex items-center justify-between">
      <div class="text-sm text-gray-600">
        Showing {{ (page - 1) * limit + 1 }} to {{ Math.min(page * limit, total) }} of {{ total }} results
      </div>
      <div class="flex gap-2">
        <button
          :disabled="page === 1"
          @click="page--"
          class="px-4 py-2 border border-gray-300 rounded-lg text-sm font-medium text-gray-700 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
        >
          Previous
        </button>
        <button
          :disabled="page >= totalPages()"
          @click="page++"
          class="px-4 py-2 border border-gray-300 rounded-lg text-sm font-medium text-gray-700 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
        >
          Next
        </button>
      </div>
    </div>

    <!-- Create/Edit Modal -->
    <Teleport to="body">
      <Transition name="fade">
        <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
          <div class="absolute inset-0 bg-black/50 backdrop-blur-sm" @click="showModal = false"></div>
          <Transition name="scale" appear>
            <div class="relative bg-white rounded-xl shadow-2xl max-w-md w-full">
              <div class="px-6 py-4 border-b">
                <h3 class="text-lg font-semibold text-gray-900">{{ editingLink ? 'Edit Link' : 'Create Link' }}</h3>
              </div>
              <form @submit.prevent="handleSubmit" class="px-6 py-4 space-y-4">
                <div>
                  <label class="block text-sm font-medium text-gray-700 mb-1.5">Long URL</label>
                  <input
                    v-model="formUrl"
                    type="url"
                    required
                    placeholder="https://example.com/your-long-url"
                    class="w-full px-4 py-2.5 border border-gray-300 rounded-lg shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500/50 focus:border-blue-500 transition-all"
                  />
                </div>
                <div v-if="!editingLink">
                  <label class="block text-sm font-medium text-gray-700 mb-1.5">Custom Code (optional)</label>
                  <input
                    v-model="formCode"
                    type="text"
                    placeholder="my-custom-code"
                    class="w-full px-4 py-2.5 border border-gray-300 rounded-lg shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500/50 focus:border-blue-500 transition-all"
                  />
                </div>
                <div>
                  <label class="block text-sm font-medium text-gray-700 mb-1.5">Expires At (optional)</label>
                  <input
                    v-model="formExpires"
                    type="datetime-local"
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
                      Saving...
                    </span>
                    <span v-else>{{ editingLink ? 'Update' : 'Create' }}</span>
                  </button>
                </div>
              </form>
            </div>
          </Transition>
        </div>
      </Transition>
    </Teleport>

    <!-- QR Code Modal -->
    <Teleport to="body">
      <Transition name="fade">
        <div v-if="showQRModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
          <div class="absolute inset-0 bg-black/50 backdrop-blur-sm" @click="showQRModal = false"></div>
          <Transition name="scale" appear>
            <div class="relative bg-white rounded-xl shadow-2xl max-w-sm w-full">
              <div class="px-6 py-4 border-b flex justify-between items-center">
                <h3 class="text-lg font-semibold text-gray-900">QR Code</h3>
                <button @click="showQRModal = false" class="text-gray-400 hover:text-gray-600 p-1 hover:bg-gray-100 rounded-lg transition-colors">
                  <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
              </div>
              <div class="px-6 py-6 flex flex-col items-center">
                <div class="bg-white p-4 rounded-xl shadow-inner border">
                  <img :src="`/${qrCode}/qr?size=256`" :alt="`QR code for ${qrCode}`" class="w-56 h-56" />
                </div>
                <p class="mt-4 text-sm text-gray-600 break-all text-center font-mono bg-gray-50 px-3 py-2 rounded-lg">{{ qrUrl }}</p>
                <button
                  @click="downloadQR"
                  class="mt-4 w-full inline-flex items-center justify-center px-4 py-2.5 border border-transparent rounded-lg shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 transition-colors"
                >
                  <svg class="w-5 h-5 mr-1.5 -ml-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
                  </svg>
                  Download PNG
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
      title="Delete Link"
      message="Are you sure you want to delete this link? This action cannot be undone."
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
