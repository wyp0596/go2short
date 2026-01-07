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
      <button @click="router.push('/links')" class="text-blue-600 hover:text-blue-800 text-sm">
        &larr; Back to Links
      </button>
    </div>

    <div class="sm:flex sm:items-center sm:justify-between mb-6">
      <h1 class="text-2xl font-bold text-gray-900">Stats for {{ code }}</h1>
      <select
        v-model="days"
        @change="loadStats"
        class="mt-4 sm:mt-0 px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500"
      >
        <option :value="7">Last 7 days</option>
        <option :value="30">Last 30 days</option>
        <option :value="90">Last 90 days</option>
      </select>
    </div>

    <div v-if="loading" class="text-gray-500">Loading...</div>

    <div v-else-if="stats">
      <div class="bg-white overflow-hidden shadow rounded-lg mb-8">
        <div class="p-5">
          <div class="text-sm font-medium text-gray-500">Total Clicks</div>
          <div class="text-3xl font-bold text-gray-900">{{ stats.total_clicks }}</div>
        </div>
      </div>

      <div class="bg-white overflow-hidden shadow rounded-lg">
        <div class="p-5">
          <h3 class="text-lg font-medium text-gray-900 mb-4">Daily Clicks</h3>
          <div v-if="stats.daily_clicks.length === 0" class="text-gray-500">No click data available</div>
          <div v-else class="flex items-end space-x-1 h-48">
            <div
              v-for="day in stats.daily_clicks"
              :key="day.date"
              class="flex-1 flex flex-col items-center"
            >
              <div
                class="w-full bg-blue-500 rounded-t"
                :style="{ height: barHeight(day.clicks) + '%' }"
                :title="`${day.date}: ${day.clicks} clicks`"
              ></div>
              <div class="text-xs text-gray-500 mt-1 transform -rotate-45 origin-top-left whitespace-nowrap">
                {{ day.date.slice(5) }}
              </div>
            </div>
          </div>
        </div>
      </div>

      <div v-if="stats.daily_clicks.length > 0" class="mt-8 bg-white overflow-hidden shadow rounded-lg">
        <div class="p-5">
          <h3 class="text-lg font-medium text-gray-900 mb-4">Daily Breakdown</h3>
          <table class="min-w-full divide-y divide-gray-200">
            <thead>
              <tr>
                <th class="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">Date</th>
                <th class="px-4 py-2 text-right text-xs font-medium text-gray-500 uppercase">Clicks</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-200">
              <tr v-for="day in stats.daily_clicks" :key="day.date">
                <td class="px-4 py-2 text-sm text-gray-900">{{ day.date }}</td>
                <td class="px-4 py-2 text-sm text-gray-900 text-right">{{ day.clicks }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  </div>
</template>
