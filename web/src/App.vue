<script setup lang="ts">
import { computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { tokenRef, removeToken } from './api'

const router = useRouter()
const route = useRoute()

const isLoggedIn = computed(() => !!tokenRef.value)
const isHomePage = computed(() => route.name === 'home')
const isDocsPage = computed(() => route.name === 'docs')
const isAdminPage = computed(() => route.path.startsWith('/admin'))
const showAdminNav = computed(() => isLoggedIn.value && isAdminPage.value && route.name !== 'login')
const showDocsNav = computed(() => isDocsPage.value)

function logout() {
  removeToken()
  router.push('/admin/login')
}
</script>

<template>
  <div :class="isHomePage ? '' : 'min-h-screen bg-gray-100'">
    <!-- Admin Nav (logged in, on admin pages) -->
    <nav v-if="showAdminNav" class="bg-white shadow">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div class="flex justify-between h-16">
          <div class="flex">
            <router-link to="/" class="flex-shrink-0 flex items-center">
              <span class="text-xl font-bold text-gray-900">go2short</span>
            </router-link>
            <div class="ml-10 flex items-center space-x-4">
              <router-link
                to="/admin"
                class="px-3 py-2 rounded-md text-sm font-medium"
                :class="route.name === 'dashboard' ? 'bg-gray-100 text-gray-900' : 'text-gray-600 hover:text-gray-900'"
              >
                Dashboard
              </router-link>
              <router-link
                to="/admin/links"
                class="px-3 py-2 rounded-md text-sm font-medium"
                :class="route.name === 'links' ? 'bg-gray-100 text-gray-900' : 'text-gray-600 hover:text-gray-900'"
              >
                Links
              </router-link>
              <router-link
                to="/admin/tokens"
                class="px-3 py-2 rounded-md text-sm font-medium"
                :class="route.name === 'tokens' ? 'bg-gray-100 text-gray-900' : 'text-gray-600 hover:text-gray-900'"
              >
                API Tokens
              </router-link>
              <router-link
                to="/docs"
                class="px-3 py-2 rounded-md text-sm font-medium"
                :class="route.name === 'docs' ? 'bg-gray-100 text-gray-900' : 'text-gray-600 hover:text-gray-900'"
              >
                API Docs
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

    <!-- Docs Nav -->
    <nav v-if="showDocsNav" class="bg-white shadow">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div class="flex justify-between h-16">
          <div class="flex items-center">
            <router-link to="/" class="text-xl font-bold text-gray-900 hover:text-blue-600">
              go2short
            </router-link>
          </div>
          <div class="flex items-center space-x-4">
            <router-link
              v-if="isLoggedIn"
              to="/admin"
              class="px-3 py-2 rounded-md text-sm font-medium text-gray-600 hover:text-gray-900"
            >
              Dashboard
            </router-link>
            <router-link
              v-else
              to="/admin/login"
              class="px-3 py-2 rounded-md text-sm font-medium text-white bg-blue-600 hover:bg-blue-700"
            >
              Login
            </router-link>
          </div>
        </div>
      </div>
    </nav>

    <main>
      <router-view />
    </main>
  </div>
</template>
