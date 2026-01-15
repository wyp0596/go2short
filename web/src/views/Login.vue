<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { login, userLogin, setToken, getOAuthProviders } from '../api'

const router = useRouter()
const route = useRoute()
const email = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)
const shakeError = ref(false)
const mode = ref<'user' | 'admin'>('user')
const oauthProviders = ref<string[]>([])

// Admin login fields
const adminUsername = ref('')
const adminPassword = ref('')

onMounted(async () => {
  // Check for OAuth callback token
  const token = route.query.token as string
  if (token) {
    setToken(token)
    router.replace('/admin')
    return
  }

  // Load OAuth providers
  try {
    const { providers } = await getOAuthProviders()
    oauthProviders.value = providers || []
  } catch {
    // OAuth not configured
  }
})

async function handleUserLogin() {
  error.value = ''
  loading.value = true
  try {
    const { token } = await userLogin(email.value, password.value)
    setToken(token)
    router.push('/admin')
  } catch (e: any) {
    error.value = e.response?.data?.error || 'Invalid email or password'
    shakeError.value = true
    setTimeout(() => shakeError.value = false, 300)
  } finally {
    loading.value = false
  }
}

async function handleAdminLogin() {
  error.value = ''
  loading.value = true
  try {
    const { token } = await login(adminUsername.value, adminPassword.value)
    setToken(token)
    router.push('/admin')
  } catch (e: any) {
    error.value = e.response?.data?.error || 'Invalid credentials'
    shakeError.value = true
    setTimeout(() => shakeError.value = false, 300)
  } finally {
    loading.value = false
  }
}

function googleLogin() {
  window.location.href = '/api/auth/google'
}

function githubLogin() {
  window.location.href = '/api/auth/github'
}
</script>

<template>
  <div class="min-h-screen flex items-center justify-center bg-gradient-to-br from-gray-900 via-blue-900 to-gray-900 px-4">
    <!-- Back to Home -->
    <router-link
      to="/"
      class="absolute top-6 left-6 inline-flex items-center gap-2 text-blue-200 hover:text-white transition-colors group"
    >
      <svg class="w-5 h-5 transform group-hover:-translate-x-1 transition-transform" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
      </svg>
      Back to Home
    </router-link>

    <div class="max-w-md w-full bg-white/10 backdrop-blur-md border border-white/20 rounded-2xl shadow-2xl p-8">
      <!-- Logo -->
      <div class="text-center mb-6">
        <div class="inline-flex items-center justify-center w-16 h-16 bg-blue-500/20 rounded-xl mb-4">
          <svg class="w-8 h-8 text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
          </svg>
        </div>
        <h2 class="text-2xl font-bold text-white">go2short</h2>
        <p class="mt-1 text-blue-200 text-sm">Sign in to continue</p>
      </div>

      <!-- Mode Toggle -->
      <div class="flex gap-2 mb-6 p-1 bg-white/5 rounded-xl">
        <button
          @click="mode = 'user'"
          :class="mode === 'user' ? 'bg-blue-600 text-white' : 'text-blue-200 hover:text-white'"
          class="flex-1 py-2 rounded-lg text-sm font-medium transition-all"
        >
          User Login
        </button>
        <button
          @click="mode = 'admin'"
          :class="mode === 'admin' ? 'bg-blue-600 text-white' : 'text-blue-200 hover:text-white'"
          class="flex-1 py-2 rounded-lg text-sm font-medium transition-all"
        >
          Admin Login
        </button>
      </div>

      <!-- User Login Form -->
      <form v-if="mode === 'user'" @submit.prevent="handleUserLogin" class="space-y-4">
        <div>
          <label for="email" class="block text-sm font-medium text-blue-200 mb-1.5">Email</label>
          <input
            id="email"
            v-model="email"
            type="email"
            required
            autocomplete="email"
            class="w-full px-4 py-3 bg-white/10 border border-white/20 rounded-xl text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500/50 focus:border-transparent transition-all"
            placeholder="Enter your email"
          />
        </div>
        <div>
          <label for="user-password" class="block text-sm font-medium text-blue-200 mb-1.5">Password</label>
          <input
            id="user-password"
            v-model="password"
            type="password"
            required
            autocomplete="current-password"
            class="w-full px-4 py-3 bg-white/10 border border-white/20 rounded-xl text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500/50 focus:border-transparent transition-all"
            placeholder="Enter password"
          />
        </div>

        <!-- Error message -->
        <div
          v-if="error"
          :class="{ 'animate-shake': shakeError }"
          class="flex items-center gap-2 p-3 bg-red-500/20 border border-red-500/30 rounded-lg text-red-300"
        >
          <svg class="w-5 h-5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <span class="text-sm">{{ error }}</span>
        </div>

        <button
          type="submit"
          :disabled="loading"
          class="w-full py-3 bg-blue-600 hover:bg-blue-700 text-white rounded-xl font-medium transition-all transform hover:scale-[1.02] active:scale-[0.98] disabled:opacity-50 disabled:hover:scale-100"
        >
          <span v-if="loading" class="inline-flex items-center gap-2">
            <svg class="animate-spin h-5 w-5" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
            Signing in...
          </span>
          <span v-else>Sign in</span>
        </button>

        <!-- OAuth Buttons -->
        <div v-if="oauthProviders.length > 0" class="space-y-3">
          <div class="relative">
            <div class="absolute inset-0 flex items-center">
              <div class="w-full border-t border-white/20"></div>
            </div>
            <div class="relative flex justify-center text-sm">
              <span class="px-2 bg-transparent text-blue-200">Or continue with</span>
            </div>
          </div>

          <div class="flex gap-3">
            <button
              v-if="oauthProviders.includes('google')"
              type="button"
              @click="googleLogin"
              class="flex-1 flex items-center justify-center gap-2 py-3 bg-white/10 hover:bg-white/20 border border-white/20 rounded-xl text-white transition-all"
            >
              <svg class="w-5 h-5" viewBox="0 0 24 24">
                <path fill="currentColor" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"/>
                <path fill="currentColor" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"/>
                <path fill="currentColor" d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"/>
                <path fill="currentColor" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"/>
              </svg>
              Google
            </button>
            <button
              v-if="oauthProviders.includes('github')"
              type="button"
              @click="githubLogin"
              class="flex-1 flex items-center justify-center gap-2 py-3 bg-white/10 hover:bg-white/20 border border-white/20 rounded-xl text-white transition-all"
            >
              <svg class="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
                <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/>
              </svg>
              GitHub
            </button>
          </div>
        </div>

        <!-- Register Link -->
        <p class="text-center text-blue-200 text-sm">
          Don't have an account?
          <router-link to="/admin/register" class="text-blue-400 hover:text-blue-300 font-medium">
            Sign up
          </router-link>
        </p>
      </form>

      <!-- Admin Login Form -->
      <form v-else @submit.prevent="handleAdminLogin" class="space-y-4">
        <div>
          <label for="admin-username" class="block text-sm font-medium text-blue-200 mb-1.5">Username</label>
          <input
            id="admin-username"
            v-model="adminUsername"
            type="text"
            required
            autocomplete="username"
            class="w-full px-4 py-3 bg-white/10 border border-white/20 rounded-xl text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500/50 focus:border-transparent transition-all"
            placeholder="Enter admin username"
          />
        </div>
        <div>
          <label for="admin-password" class="block text-sm font-medium text-blue-200 mb-1.5">Password</label>
          <input
            id="admin-password"
            v-model="adminPassword"
            type="password"
            required
            autocomplete="current-password"
            class="w-full px-4 py-3 bg-white/10 border border-white/20 rounded-xl text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500/50 focus:border-transparent transition-all"
            placeholder="Enter admin password"
          />
        </div>

        <!-- Error message -->
        <div
          v-if="error"
          :class="{ 'animate-shake': shakeError }"
          class="flex items-center gap-2 p-3 bg-red-500/20 border border-red-500/30 rounded-lg text-red-300"
        >
          <svg class="w-5 h-5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <span class="text-sm">{{ error }}</span>
        </div>

        <button
          type="submit"
          :disabled="loading"
          class="w-full py-3 bg-blue-600 hover:bg-blue-700 text-white rounded-xl font-medium transition-all transform hover:scale-[1.02] active:scale-[0.98] disabled:opacity-50 disabled:hover:scale-100"
        >
          <span v-if="loading" class="inline-flex items-center gap-2">
            <svg class="animate-spin h-5 w-5" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
            Signing in...
          </span>
          <span v-else>Sign in as Admin</span>
        </button>

        <p class="text-center text-blue-200/60 text-xs">
          Super admin login (configured via environment variables)
        </p>
      </form>
    </div>
  </div>
</template>
