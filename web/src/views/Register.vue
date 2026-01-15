<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { register, setToken } from '../api'

const router = useRouter()
const email = ref('')
const password = ref('')
const confirmPassword = ref('')
const error = ref('')
const loading = ref(false)
const shakeError = ref(false)

async function handleRegister() {
  error.value = ''

  // Validate passwords match
  if (password.value !== confirmPassword.value) {
    error.value = 'Passwords do not match'
    shakeError.value = true
    setTimeout(() => shakeError.value = false, 300)
    return
  }

  // Validate password length
  if (password.value.length < 6) {
    error.value = 'Password must be at least 6 characters'
    shakeError.value = true
    setTimeout(() => shakeError.value = false, 300)
    return
  }

  loading.value = true
  try {
    const { token } = await register(email.value, password.value)
    setToken(token)
    router.push('/admin')
  } catch (e: any) {
    error.value = e.response?.data?.error || 'Registration failed'
    shakeError.value = true
    setTimeout(() => shakeError.value = false, 300)
  } finally {
    loading.value = false
  }
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
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M18 9v3m0 0v3m0-3h3m-3 0h-3m-2-5a4 4 0 11-8 0 4 4 0 018 0zM3 20a6 6 0 0112 0v1H3v-1z" />
          </svg>
        </div>
        <h2 class="text-2xl font-bold text-white">Create Account</h2>
        <p class="mt-1 text-blue-200 text-sm">Sign up to start shortening URLs</p>
      </div>

      <!-- Register Form -->
      <form @submit.prevent="handleRegister" class="space-y-4">
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
          <label for="password" class="block text-sm font-medium text-blue-200 mb-1.5">Password</label>
          <input
            id="password"
            v-model="password"
            type="password"
            required
            autocomplete="new-password"
            class="w-full px-4 py-3 bg-white/10 border border-white/20 rounded-xl text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500/50 focus:border-transparent transition-all"
            placeholder="At least 6 characters"
          />
        </div>
        <div>
          <label for="confirm-password" class="block text-sm font-medium text-blue-200 mb-1.5">Confirm Password</label>
          <input
            id="confirm-password"
            v-model="confirmPassword"
            type="password"
            required
            autocomplete="new-password"
            class="w-full px-4 py-3 bg-white/10 border border-white/20 rounded-xl text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500/50 focus:border-transparent transition-all"
            placeholder="Confirm your password"
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
            Creating account...
          </span>
          <span v-else>Create Account</span>
        </button>

        <!-- Login Link -->
        <p class="text-center text-blue-200 text-sm">
          Already have an account?
          <router-link to="/admin/login" class="text-blue-400 hover:text-blue-300 font-medium">
            Sign in
          </router-link>
        </p>
      </form>
    </div>
  </div>
</template>
