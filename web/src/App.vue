<script setup lang="ts">
import { computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { tokenRef, removeToken } from './api'
import { useToast } from './composables/useToast'
import Toast from './components/Toast.vue'

const router = useRouter()
const route = useRoute()
const { toasts, remove } = useToast()

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
  <div :class="isHomePage || isDocsPage ? '' : 'min-h-screen bg-gray-100'">
    <!-- Admin Nav (logged in, on admin pages) -->
    <nav v-if="showAdminNav" class="bg-white shadow">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div class="flex justify-between h-16">
          <div class="flex">
            <router-link to="/" class="flex-shrink-0 flex items-center">
              <span class="text-xl font-bold text-gray-900">go2short</span>
            </router-link>
            <div class="ml-10 flex items-center space-x-1">
              <router-link
                to="/admin"
                class="relative px-3 py-2 text-sm font-medium transition-colors"
                :class="route.name === 'dashboard' ? 'text-blue-600' : 'text-gray-600 hover:text-gray-900'"
              >
                Dashboard
                <span v-if="route.name === 'dashboard'" class="absolute bottom-0 left-3 right-3 h-0.5 bg-blue-600 rounded-full"></span>
              </router-link>
              <router-link
                to="/admin/links"
                class="relative px-3 py-2 text-sm font-medium transition-colors"
                :class="route.name === 'links' || route.name === 'linkStats' ? 'text-blue-600' : 'text-gray-600 hover:text-gray-900'"
              >
                Links
                <span v-if="route.name === 'links' || route.name === 'linkStats'" class="absolute bottom-0 left-3 right-3 h-0.5 bg-blue-600 rounded-full"></span>
              </router-link>
              <router-link
                to="/admin/tokens"
                class="relative px-3 py-2 text-sm font-medium transition-colors"
                :class="route.name === 'tokens' ? 'text-blue-600' : 'text-gray-600 hover:text-gray-900'"
              >
                API Tokens
                <span v-if="route.name === 'tokens'" class="absolute bottom-0 left-3 right-3 h-0.5 bg-blue-600 rounded-full"></span>
              </router-link>
              <router-link
                to="/docs"
                class="relative px-3 py-2 text-sm font-medium transition-colors"
                :class="route.name === 'docs' ? 'text-blue-600' : 'text-gray-600 hover:text-gray-900'"
              >
                API Docs
                <span v-if="route.name === 'docs'" class="absolute bottom-0 left-3 right-3 h-0.5 bg-blue-600 rounded-full"></span>
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
    <nav v-if="showDocsNav" class="bg-gray-900/50 backdrop-blur-sm border-b border-white/10">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div class="flex justify-between h-16">
          <div class="flex items-center">
            <router-link to="/" class="text-xl font-bold text-white hover:text-blue-300">
              go2short
            </router-link>
          </div>
          <div class="flex items-center space-x-4">
            <router-link
              v-if="isLoggedIn"
              to="/admin"
              class="px-3 py-2 rounded-md text-sm font-medium text-blue-100 hover:text-white"
            >
              Dashboard
            </router-link>
            <router-link
              v-else
              to="/admin/login"
              class="px-3 py-2 rounded-md text-sm font-medium text-blue-100 hover:text-white"
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

    <!-- Global Toast Container -->
    <Teleport to="body">
      <div class="fixed top-4 right-4 z-50 space-y-2">
        <TransitionGroup name="toast">
          <Toast
            v-for="toast in toasts"
            :key="toast.id"
            :message="toast.message"
            :type="toast.type"
            @close="remove(toast.id)"
          />
        </TransitionGroup>
      </div>
    </Teleport>
  </div>
</template>
