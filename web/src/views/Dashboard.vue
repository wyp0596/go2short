<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { getOverviewStats, getTopLinks, getClickTrend, createLink, type OverviewStats, type TopLink, type DayClick } from '../api'

const stats = ref<OverviewStats | null>(null)
const topLinks = ref<TopLink[]>([])
const trend = ref<DayClick[]>([])
const loading = ref(true)

// Quick create form
const quickUrl = ref('')
const quickLoading = ref(false)
const quickResult = ref('')
const quickError = ref('')

// Trend days selector
const trendDays = ref(7)

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
  quickError.value = ''
  quickResult.value = ''
  try {
    const result = await createLink(quickUrl.value)
    quickResult.value = result.short_url
    quickUrl.value = ''
    // Refresh stats
    stats.value = await getOverviewStats()
  } catch (e: unknown) {
    const err = e as { response?: { data?: { error?: string } } }
    quickError.value = err.response?.data?.error || 'Failed to create link'
  } finally {
    quickLoading.value = false
  }
}

function copyToClipboard(text: string) {
  navigator.clipboard.writeText(text)
}

// Simple bar chart computed
const maxClicks = computed(() => Math.max(...trend.value.map(d => d.clicks), 1))
</script>

<template>
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <h1 class="text-2xl font-bold text-gray-900 mb-8">Dashboard</h1>

    <!-- Quick Create -->
    <div class="bg-white shadow rounded-lg p-6 mb-8">
      <h2 class="text-lg font-medium text-gray-900 mb-4">Quick Create</h2>
      <form @submit.prevent="handleQuickCreate" class="flex gap-4">
        <input
          v-model="quickUrl"
          type="url"
          placeholder="Enter URL to shorten..."
          required
          class="flex-1 px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500"
        />
        <button
          type="submit"
          :disabled="quickLoading"
          class="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50"
        >
          {{ quickLoading ? 'Creating...' : 'Shorten' }}
        </button>
      </form>
      <div v-if="quickResult" class="mt-3 flex items-center gap-2">
        <span class="text-green-600">Created:</span>
        <a :href="quickResult" target="_blank" class="text-blue-600 hover:underline">{{ quickResult }}</a>
        <button @click="copyToClipboard(quickResult)" class="text-gray-500 hover:text-gray-700">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 5H6a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2v-1M8 5a2 2 0 002 2h2a2 2 0 002-2M8 5a2 2 0 012-2h2a2 2 0 012 2m0 0h2a2 2 0 012 2v3m2 4H10m0 0l3-3m-3 3l3 3" />
          </svg>
        </button>
      </div>
      <div v-if="quickError" class="mt-3 text-red-600">{{ quickError }}</div>
    </div>

    <div v-if="loading" class="text-gray-500">Loading...</div>

    <template v-else>
      <!-- Stats Cards -->
      <div v-if="stats" class="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-4 mb-8">
        <div class="bg-white overflow-hidden shadow rounded-lg">
          <div class="p-5">
            <div class="flex items-center">
              <div class="flex-shrink-0">
                <svg class="h-6 w-6 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
                </svg>
              </div>
              <div class="ml-5 w-0 flex-1">
                <dl>
                  <dt class="text-sm font-medium text-gray-500 truncate">Total Links</dt>
                  <dd class="text-lg font-semibold text-gray-900">{{ stats.total_links }}</dd>
                </dl>
              </div>
            </div>
          </div>
        </div>

        <div class="bg-white overflow-hidden shadow rounded-lg">
          <div class="p-5">
            <div class="flex items-center">
              <div class="flex-shrink-0">
                <svg class="h-6 w-6 text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
              </div>
              <div class="ml-5 w-0 flex-1">
                <dl>
                  <dt class="text-sm font-medium text-gray-500 truncate">Active Links</dt>
                  <dd class="text-lg font-semibold text-gray-900">{{ stats.active_links }}</dd>
                </dl>
              </div>
            </div>
          </div>
        </div>

        <div class="bg-white overflow-hidden shadow rounded-lg">
          <div class="p-5">
            <div class="flex items-center">
              <div class="flex-shrink-0">
                <svg class="h-6 w-6 text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 15l-2 5L9 9l11 4-5 2zm0 0l5 5M7.188 2.239l.777 2.897M5.136 7.965l-2.898-.777M13.95 4.05l-2.122 2.122m-5.657 5.656l-2.12 2.122" />
                </svg>
              </div>
              <div class="ml-5 w-0 flex-1">
                <dl>
                  <dt class="text-sm font-medium text-gray-500 truncate">Total Clicks</dt>
                  <dd class="text-lg font-semibold text-gray-900">{{ stats.total_clicks }}</dd>
                </dl>
              </div>
            </div>
          </div>
        </div>

        <div class="bg-white overflow-hidden shadow rounded-lg">
          <div class="p-5">
            <div class="flex items-center">
              <div class="flex-shrink-0">
                <svg class="h-6 w-6 text-yellow-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
              </div>
              <div class="ml-5 w-0 flex-1">
                <dl>
                  <dt class="text-sm font-medium text-gray-500 truncate">Today's Clicks</dt>
                  <dd class="text-lg font-semibold text-gray-900">{{ stats.today_clicks }}</dd>
                </dl>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="grid grid-cols-1 lg:grid-cols-2 gap-8">
        <!-- Click Trend Chart -->
        <div class="bg-white shadow rounded-lg p-6">
          <div class="flex items-center justify-between mb-4">
            <h2 class="text-lg font-medium text-gray-900">Click Trend</h2>
            <select v-model="trendDays" @change="loadTrend" class="border border-gray-300 rounded-md px-2 py-1 text-sm">
              <option :value="7">7 days</option>
              <option :value="30">30 days</option>
              <option :value="90">90 days</option>
            </select>
          </div>
          <div v-if="trend.length === 0" class="text-gray-500 text-center py-8">No data</div>
          <div v-else class="flex items-end gap-1 h-40">
            <div
              v-for="day in trend"
              :key="day.date"
              class="flex-1 bg-blue-500 rounded-t hover:bg-blue-600 relative group"
              :style="{ height: (day.clicks / maxClicks * 100) + '%', minHeight: day.clicks > 0 ? '4px' : '0' }"
            >
              <div class="absolute bottom-full left-1/2 -translate-x-1/2 mb-1 hidden group-hover:block bg-gray-800 text-white text-xs px-2 py-1 rounded whitespace-nowrap">
                {{ day.date }}: {{ day.clicks }}
              </div>
            </div>
          </div>
          <div v-if="trend.length > 0" class="flex justify-between text-xs text-gray-500 mt-2">
            <span>{{ trend[0]?.date }}</span>
            <span>{{ trend[trend.length - 1]?.date }}</span>
          </div>
        </div>

        <!-- Top Links -->
        <div class="bg-white shadow rounded-lg p-6">
          <h2 class="text-lg font-medium text-gray-900 mb-4">Top Links (30 days)</h2>
          <div v-if="topLinks.length === 0" class="text-gray-500 text-center py-8">No data</div>
          <div v-else class="space-y-3">
            <div v-for="(link, index) in topLinks" :key="link.code" class="flex items-center gap-3">
              <span class="text-gray-400 w-6 text-right">{{ index + 1 }}.</span>
              <div class="flex-1 min-w-0">
                <a :href="link.short_url" target="_blank" class="text-blue-600 hover:underline text-sm font-medium">{{ link.code }}</a>
                <div class="text-xs text-gray-500 truncate">{{ link.long_url }}</div>
              </div>
              <div class="flex items-center gap-2">
                <span class="text-sm font-semibold text-gray-900">{{ link.click_count }}</span>
                <button @click="copyToClipboard(link.short_url)" class="text-gray-400 hover:text-gray-600">
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
