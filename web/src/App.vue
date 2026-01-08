<script setup lang="ts">
import { computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { getToken, removeToken } from './api'

const router = useRouter()
const route = useRoute()

const isLoggedIn = computed(() => !!getToken())
const showNav = computed(() => isLoggedIn.value && route.name !== 'login')

function logout() {
  removeToken()
  router.push('/login')
}
</script>

<template>
  <div class="min-h-screen bg-gray-100">
    <nav v-if="showNav" class="bg-white shadow">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div class="flex justify-between h-16">
          <div class="flex">
            <div class="flex-shrink-0 flex items-center">
              <span class="text-xl font-bold text-gray-900">go2short</span>
            </div>
            <div class="ml-10 flex items-center space-x-4">
              <router-link
                to="/"
                class="px-3 py-2 rounded-md text-sm font-medium"
                :class="route.name === 'dashboard' ? 'bg-gray-100 text-gray-900' : 'text-gray-600 hover:text-gray-900'"
              >
                Dashboard
              </router-link>
              <router-link
                to="/links"
                class="px-3 py-2 rounded-md text-sm font-medium"
                :class="route.name === 'links' ? 'bg-gray-100 text-gray-900' : 'text-gray-600 hover:text-gray-900'"
              >
                Links
              </router-link>
              <router-link
                to="/tokens"
                class="px-3 py-2 rounded-md text-sm font-medium"
                :class="route.name === 'tokens' ? 'bg-gray-100 text-gray-900' : 'text-gray-600 hover:text-gray-900'"
              >
                API Tokens
              </router-link>
            </div>
          </div>
          <div class="flex items-center">
            <button
              @click="logout"
              class="px-3 py-2 rounded-md text-sm font-medium text-gray-600 hover:text-gray-900"
            >
              Logout
            </button>
          </div>
        </div>
      </div>
    </nav>
    <main>
      <router-view />
    </main>
  </div>
</template>
