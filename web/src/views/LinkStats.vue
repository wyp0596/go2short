<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getLinkStats, type LinkStats } from '../api'

const route = useRoute()
const router = useRouter()

const code = route.params.code as string
const stats = ref<LinkStats | null>(null)
const loading = ref(true)
const days = ref(30)

async function loadStats() {
  loading.value = true
  try {
    stats.value = await getLinkStats(code, days.value)
  } finally {
    loading.value = false
  }
}

onMounted(loadStats)

const maxClicks = computed(() => {
  if (!stats.value?.daily_clicks.length) return 1
  return Math.max(...stats.value.daily_clicks.map(d => d.clicks), 1)
})

function barHeight(clicks: number) {
  return Math.max((clicks / maxClicks.value) * 100, 2)
}
</script>

<template>
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <div class="mb-6">
      <button
        @click="router.push('/admin/links')"
        class="inline-flex items-center gap-2 text-gray-600 hover:text-gray-900 transition-colors group"
      >
        <svg class="w-5 h-5 transform group-hover:-translate-x-1 transition-transform" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
        </svg>
        Back to Links
      </button>
    </div>

    <div class="sm:flex sm:items-center sm:justify-between mb-8">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">Statistics</h1>
        <p class="mt-1 text-gray-500">Link: <span class="font-mono text-blue-600">{{ code }}</span></p>
      </div>
      <select
        v-model="days"
        @change="loadStats"
        class="mt-4 sm:mt-0 px-4 py-2.5 border border-gray-300 rounded-lg shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500/50 focus:border-blue-500 transition-all"
      >
        <option :value="7">Last 7 days</option>
        <option :value="30">Last 30 days</option>
        <option :value="90">Last 90 days</option>
      </select>
    </div>

    <div v-if="loading" class="space-y-6">
      <!-- Skeleton for total clicks card -->
      <div class="bg-white rounded-xl shadow-sm p-6 animate-pulse">
        <div class="h-4 bg-gray-200 rounded w-24 mb-2"></div>
        <div class="h-8 bg-gray-200 rounded w-20"></div>
      </div>
      <!-- Skeleton for chart -->
      <div class="bg-white rounded-xl shadow-sm p-6 animate-pulse">
        <div class="h-5 bg-gray-200 rounded w-32 mb-6"></div>
        <div class="h-48 bg-gray-200 rounded"></div>
      </div>
    </div>

    <div v-else-if="stats" class="space-y-6">
      <!-- Total Clicks Card -->
      <div class="bg-white rounded-xl shadow-sm hover:shadow-md transition-all duration-300 p-6">
        <div class="flex items-center gap-4">
          <div class="w-14 h-14 bg-purple-100 rounded-xl flex items-center justify-center">
            <svg class="w-7 h-7 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 15l-2 5L9 9l11 4-5 2zm0 0l5 5M7.188 2.239l.777 2.897M5.136 7.965l-2.898-.777M13.95 4.05l-2.122 2.122m-5.657 5.656l-2.12 2.122" />
            </svg>
          </div>
          <div>
            <p class="text-sm font-medium text-gray-500">Total Clicks</p>
            <p class="text-3xl font-bold text-gray-900">{{ stats.total_clicks.toLocaleString() }}</p>
          </div>
        </div>
      </div>

      <!-- Daily Clicks Chart -->
      <div class="bg-white rounded-xl shadow-sm p-6">
        <h3 class="text-lg font-semibold text-gray-900 mb-6">Daily Clicks</h3>
        <div v-if="stats.daily_clicks.length === 0" class="text-center py-12">
          <svg class="mx-auto h-12 w-12 text-gray-400 mb-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
          </svg>
          <p class="text-gray-500">No click data available</p>
        </div>
        <div v-else class="flex items-end gap-1 h-52">
          <div
            v-for="day in stats.daily_clicks"
            :key="day.date"
            class="flex-1 bg-purple-500 rounded-t cursor-pointer hover:bg-purple-600 transition-colors relative group"
            :style="{ height: barHeight(day.clicks) + '%' }"
          >
            <!-- Tooltip -->
            <div class="absolute bottom-full left-1/2 -translate-x-1/2 mb-2 hidden group-hover:block bg-gray-900 text-white text-xs px-3 py-2 rounded-lg shadow-lg whitespace-nowrap z-10">
              <div class="font-medium">{{ day.date }}</div>
              <div class="text-gray-300">{{ day.clicks }} clicks</div>
              <div class="absolute top-full left-1/2 -translate-x-1/2 border-4 border-transparent border-t-gray-900"></div>
            </div>
          </div>
        </div>
        <div v-if="stats.daily_clicks.length > 0" class="flex justify-between text-xs text-gray-500 mt-3 px-1">
          <span>{{ stats.daily_clicks[0]?.date }}</span>
          <span>{{ stats.daily_clicks[stats.daily_clicks.length - 1]?.date }}</span>
        </div>
      </div>

      <!-- Daily Breakdown Table -->
      <div v-if="stats.daily_clicks.length > 0" class="bg-white rounded-xl shadow-sm overflow-hidden">
        <div class="px-6 py-4 border-b">
          <h3 class="text-lg font-semibold text-gray-900">Daily Breakdown</h3>
        </div>
        <div class="overflow-x-auto">
          <table class="min-w-full divide-y divide-gray-200">
            <thead class="bg-gray-50">
              <tr>
                <th class="px-6 py-3.5 text-left text-xs font-semibold text-gray-500 uppercase tracking-wider">Date</th>
                <th class="px-6 py-3.5 text-right text-xs font-semibold text-gray-500 uppercase tracking-wider">Clicks</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-200">
              <tr v-for="day in stats.daily_clicks" :key="day.date" class="hover:bg-gray-50 transition-colors">
                <td class="px-6 py-4 text-sm text-gray-900">{{ day.date }}</td>
                <td class="px-6 py-4 text-sm text-gray-900 text-right font-medium">{{ day.clicks.toLocaleString() }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- Device Analytics -->
      <div v-if="stats.device_stats" class="bg-white rounded-xl shadow-sm p-6">
        <h3 class="text-lg font-semibold text-gray-900 mb-6">Device Analytics</h3>
        <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
          <!-- Device Type -->
          <div>
            <h4 class="text-sm font-medium text-gray-500 mb-3">Device Type</h4>
            <div v-if="stats.device_stats.device_type.length === 0" class="text-gray-400 text-sm">No data</div>
            <div v-else class="space-y-2">
              <div v-for="item in stats.device_stats.device_type" :key="item.name" class="flex items-center gap-2">
                <span class="w-20 text-sm text-gray-600 truncate" :title="item.name">{{ item.name }}</span>
                <div class="flex-1 h-5 bg-gray-100 rounded overflow-hidden">
                  <div
                    class="h-full bg-blue-500 rounded"
                    :style="{ width: Math.max((item.count / (stats.device_stats.device_type[0]?.count || 1)) * 100, 2) + '%' }"
                  ></div>
                </div>
                <span class="w-12 text-right text-sm font-medium text-gray-700">{{ item.count }}</span>
              </div>
            </div>
          </div>
          <!-- Browser -->
          <div>
            <h4 class="text-sm font-medium text-gray-500 mb-3">Browser</h4>
            <div v-if="stats.device_stats.browser.length === 0" class="text-gray-400 text-sm">No data</div>
            <div v-else class="space-y-2">
              <div v-for="item in stats.device_stats.browser" :key="item.name" class="flex items-center gap-2">
                <span class="w-20 text-sm text-gray-600 truncate" :title="item.name">{{ item.name }}</span>
                <div class="flex-1 h-5 bg-gray-100 rounded overflow-hidden">
                  <div
                    class="h-full bg-green-500 rounded"
                    :style="{ width: Math.max((item.count / (stats.device_stats.browser[0]?.count || 1)) * 100, 2) + '%' }"
                  ></div>
                </div>
                <span class="w-12 text-right text-sm font-medium text-gray-700">{{ item.count }}</span>
              </div>
            </div>
          </div>
          <!-- OS -->
          <div>
            <h4 class="text-sm font-medium text-gray-500 mb-3">Operating System</h4>
            <div v-if="stats.device_stats.os.length === 0" class="text-gray-400 text-sm">No data</div>
            <div v-else class="space-y-2">
              <div v-for="item in stats.device_stats.os" :key="item.name" class="flex items-center gap-2">
                <span class="w-20 text-sm text-gray-600 truncate" :title="item.name">{{ item.name }}</span>
                <div class="flex-1 h-5 bg-gray-100 rounded overflow-hidden">
                  <div
                    class="h-full bg-purple-500 rounded"
                    :style="{ width: Math.max((item.count / (stats.device_stats.os[0]?.count || 1)) * 100, 2) + '%' }"
                  ></div>
                </div>
                <span class="w-12 text-right text-sm font-medium text-gray-700">{{ item.count }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
