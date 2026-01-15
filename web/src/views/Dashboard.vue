<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { getOverviewStats, getTopLinks, getClickTrend, createLink, type OverviewStats, type TopLink, type DayClick } from '../api'
import { useToast } from '../composables/useToast'
import StatCardSkeleton from '../components/StatCardSkeleton.vue'

const { success, error: showError } = useToast()

const stats = ref<OverviewStats | null>(null)
const topLinks = ref<TopLink[]>([])
const trend = ref<DayClick[]>([])
const loading = ref(true)

// Quick create form
const quickUrl = ref('')
const quickLoading = ref(false)
const quickResult = ref('')

// Trend days selector
const trendDays = ref(7)

// Copy feedback
const copied = ref(false)

async function loadData() {
  loading.value = true
  try {
    const [statsData, topData, trendData] = await Promise.all([
      getOverviewStats(),
      getTopLinks(10, 30),
      getClickTrend(trendDays.value)
    ])
    stats.value = statsData
    topLinks.value = topData.links
    trend.value = trendData.trend
  } finally {
    loading.value = false
  }
}

async function loadTrend() {
  const data = await getClickTrend(trendDays.value)
  trend.value = data.trend
}

onMounted(loadData)

async function handleQuickCreate() {
  if (!quickUrl.value) return
  quickLoading.value = true
  quickResult.value = ''
  try {
    const result = await createLink(quickUrl.value)
    quickResult.value = result.short_url
    quickUrl.value = ''
    stats.value = await getOverviewStats()
    success('Link created successfully')
  } catch (e: unknown) {
    const err = e as { response?: { data?: { error?: string } } }
    showError(err.response?.data?.error || 'Failed to create link')
  } finally {
    quickLoading.value = false
  }
}

function copyToClipboard(text: string) {
  navigator.clipboard.writeText(text)
  copied.value = true
  success('Copied to clipboard')
  setTimeout(() => copied.value = false, 2000)
}

// Simple bar chart computed
const maxClicks = computed(() => Math.max(...trend.value.map(d => d.clicks), 1))
</script>

<template>
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <h1 class="text-2xl font-bold text-gray-900 mb-8">Dashboard</h1>

    <!-- Quick Create -->
    <div class="bg-white shadow-sm rounded-xl p-6 mb-8">
      <h2 class="text-lg font-semibold text-gray-900 mb-4">Quick Create</h2>
      <form @submit.prevent="handleQuickCreate" class="flex gap-4">
        <input
          v-model="quickUrl"
          type="url"
          placeholder="Enter URL to shorten..."
          required
          class="flex-1 px-4 py-2.5 border border-gray-300 rounded-lg shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500/50 focus:border-blue-500 transition-all"
        />
        <button
          type="submit"
          :disabled="quickLoading"
          class="px-5 py-2.5 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-all transform hover:scale-[1.02] active:scale-[0.98] disabled:opacity-50 disabled:hover:scale-100 font-medium"
        >
          <span v-if="quickLoading" class="inline-flex items-center gap-2">
            <svg class="animate-spin h-4 w-4" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
            Creating...
          </span>
          <span v-else>Shorten</span>
        </button>
      </form>
      <div v-if="quickResult" class="mt-4 flex items-center gap-3 p-3 bg-green-50 border border-green-200 rounded-lg">
        <svg class="w-5 h-5 text-green-500 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
        </svg>
        <a :href="quickResult" target="_blank" class="text-blue-600 hover:underline font-medium">{{ quickResult }}</a>
        <button
          @click="copyToClipboard(quickResult)"
          class="ml-auto px-3 py-1.5 text-sm bg-white border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors inline-flex items-center gap-1.5"
        >
          <svg v-if="!copied" class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 5H6a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2v-1M8 5a2 2 0 002 2h2a2 2 0 002-2M8 5a2 2 0 012-2h2a2 2 0 012 2m0 0h2a2 2 0 012 2v3m2 4H10m0 0l3-3m-3 3l3 3" />
          </svg>
          <svg v-else class="w-4 h-4 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
          </svg>
          {{ copied ? 'Copied!' : 'Copy' }}
        </button>
      </div>
    </div>

    <!-- Loading Skeleton -->
    <div v-if="loading" class="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-4 mb-8">
      <StatCardSkeleton v-for="i in 4" :key="i" />
    </div>

    <template v-else>
      <!-- Stats Cards -->
      <div v-if="stats" class="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-4 mb-8">
        <div class="bg-white rounded-xl shadow-sm hover:shadow-md transition-all duration-300 hover:-translate-y-1 p-6">
          <div class="flex items-center gap-4">
            <div class="w-12 h-12 bg-blue-100 rounded-xl flex items-center justify-center">
              <svg class="h-6 w-6 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
              </svg>
            </div>
            <div>
              <p class="text-sm font-medium text-gray-500">Total Links</p>
              <p class="text-2xl font-bold text-gray-900">{{ stats.total_links.toLocaleString() }}</p>
            </div>
          </div>
        </div>

        <div class="bg-white rounded-xl shadow-sm hover:shadow-md transition-all duration-300 hover:-translate-y-1 p-6">
          <div class="flex items-center gap-4">
            <div class="w-12 h-12 bg-green-100 rounded-xl flex items-center justify-center">
              <svg class="h-6 w-6 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            </div>
            <div>
              <p class="text-sm font-medium text-gray-500">Active Links</p>
              <p class="text-2xl font-bold text-gray-900">{{ stats.active_links.toLocaleString() }}</p>
            </div>
          </div>
        </div>

        <div class="bg-white rounded-xl shadow-sm hover:shadow-md transition-all duration-300 hover:-translate-y-1 p-6">
          <div class="flex items-center gap-4">
            <div class="w-12 h-12 bg-purple-100 rounded-xl flex items-center justify-center">
              <svg class="h-6 w-6 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 15l-2 5L9 9l11 4-5 2zm0 0l5 5M7.188 2.239l.777 2.897M5.136 7.965l-2.898-.777M13.95 4.05l-2.122 2.122m-5.657 5.656l-2.12 2.122" />
              </svg>
            </div>
            <div>
              <p class="text-sm font-medium text-gray-500">Total Clicks</p>
              <p class="text-2xl font-bold text-gray-900">{{ stats.total_clicks.toLocaleString() }}</p>
            </div>
          </div>
        </div>

        <div class="bg-white rounded-xl shadow-sm hover:shadow-md transition-all duration-300 hover:-translate-y-1 p-6">
          <div class="flex items-center gap-4">
            <div class="w-12 h-12 bg-yellow-100 rounded-xl flex items-center justify-center">
              <svg class="h-6 w-6 text-yellow-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            </div>
            <div>
              <p class="text-sm font-medium text-gray-500">Today's Clicks</p>
              <p class="text-2xl font-bold text-gray-900">{{ stats.today_clicks.toLocaleString() }}</p>
            </div>
          </div>
        </div>
      </div>

      <div class="grid grid-cols-1 lg:grid-cols-2 gap-8">
        <!-- Click Trend Chart -->
        <div class="bg-white shadow-sm rounded-xl p-6">
          <div class="flex items-center justify-between mb-6">
            <h2 class="text-lg font-semibold text-gray-900">Click Trend</h2>
            <select
              v-model="trendDays"
              @change="loadTrend"
              class="border border-gray-300 rounded-lg px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500/50"
            >
              <option :value="7">7 days</option>
              <option :value="30">30 days</option>
              <option :value="90">90 days</option>
            </select>
          </div>
          <div v-if="trend.length === 0" class="text-gray-500 text-center py-12">
            <svg class="mx-auto h-12 w-12 text-gray-400 mb-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
            </svg>
            No data available
          </div>
          <div v-else class="flex items-end gap-1 h-44">
            <div
              v-for="day in trend"
              :key="day.date"
              class="flex-1 bg-blue-500 rounded-t cursor-pointer hover:bg-blue-600 transition-colors relative group"
              :style="{ height: (day.clicks / maxClicks * 100) + '%', minHeight: day.clicks > 0 ? '4px' : '0' }"
            >
              <!-- Tooltip -->
              <div class="absolute bottom-full left-1/2 -translate-x-1/2 mb-2 hidden group-hover:block bg-gray-900 text-white text-xs px-3 py-2 rounded-lg shadow-lg whitespace-nowrap z-10">
                <div class="font-medium">{{ day.date }}</div>
                <div class="text-gray-300">{{ day.clicks }} clicks</div>
                <div class="absolute top-full left-1/2 -translate-x-1/2 border-4 border-transparent border-t-gray-900"></div>
              </div>
            </div>
          </div>
          <div v-if="trend.length > 0" class="flex justify-between text-xs text-gray-500 mt-3 px-1">
            <span>{{ trend[0]?.date }}</span>
            <span>{{ trend[trend.length - 1]?.date }}</span>
          </div>
        </div>

        <!-- Top Links -->
        <div class="bg-white shadow-sm rounded-xl p-6">
          <h2 class="text-lg font-semibold text-gray-900 mb-6">Top Links (30 days)</h2>
          <div v-if="topLinks.length === 0" class="text-gray-500 text-center py-12">
            <svg class="mx-auto h-12 w-12 text-gray-400 mb-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
            </svg>
            No links yet
          </div>
          <div v-else class="space-y-3">
            <div
              v-for="(link, index) in topLinks"
              :key="link.code"
              class="flex items-center gap-3 p-2 -mx-2 rounded-lg hover:bg-gray-50 transition-colors"
            >
              <span class="text-gray-400 w-6 text-right font-medium">{{ index + 1 }}.</span>
              <div class="flex-1 min-w-0">
                <a :href="link.short_url" target="_blank" class="text-blue-600 hover:underline text-sm font-medium">{{ link.code }}</a>
                <div class="text-xs text-gray-500 truncate">{{ link.long_url }}</div>
              </div>
              <div class="flex items-center gap-2">
                <span class="text-sm font-semibold text-gray-900 bg-gray-100 px-2 py-0.5 rounded">{{ link.click_count.toLocaleString() }}</span>
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
            </div>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>
