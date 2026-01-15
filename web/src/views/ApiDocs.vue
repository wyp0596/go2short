<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'

const origin = window.location.origin
const activeSection = ref('overview')

const sections = [
  { id: 'overview', title: 'Overview' },
  { id: 'redirect', title: 'Redirect' },
  { id: 'qrcode', title: 'QR Code' },
  { id: 'create', title: 'Create Link' },
  { id: 'batch', title: 'Batch Create' },
  { id: 'preview', title: 'Link Preview' },
  { id: 'ratelimit', title: 'Rate Limiting' },
]

// Curl commands for copy
const curlRedirect = computed(() => `curl -L ${origin}/abc123`)
const curlQR = computed(() => `curl ${origin}/abc123/qr?size=512 -o qr.png`)
const curlCreate = computed(() => `curl -X POST ${origin}/api/links \\\n  -H "Authorization: Bearer YOUR_TOKEN" \\\n  -H "Content-Type: application/json" \\\n  -d '{"long_url": "https://example.com/very/long/path"}'`)
const curlBatch = computed(() => `curl -X POST ${origin}/api/links/batch \\\n  -H "Authorization: Bearer YOUR_TOKEN" \\\n  -H "Content-Type: application/json" \\\n  -d '{"items": [{"long_url": "https://example1.com"}, {"long_url": "https://example2.com", "custom_code": "foo"}]}'`)
const curlPreview = computed(() => `curl ${origin}/api/links/abc123/preview \\\n  -H "Authorization: Bearer YOUR_TOKEN"`)

function copyToClipboard(text: string, event: Event) {
  navigator.clipboard.writeText(text)
  const btn = event.target as HTMLButtonElement
  const original = btn.textContent
  btn.textContent = 'Copied!'
  setTimeout(() => { btn.textContent = original }, 1500)
}

function scrollToSection(id: string) {
  const el = document.getElementById(id)
  if (el) {
    el.scrollIntoView({ behavior: 'smooth', block: 'start' })
  }
}

function handleScroll() {
  const scrollPos = window.scrollY + 100
  for (const section of sections) {
    const el = document.getElementById(section.id)
    if (el && el.offsetTop <= scrollPos) {
      activeSection.value = section.id
    }
  }
}

onMounted(() => window.addEventListener('scroll', handleScroll))
onUnmounted(() => window.removeEventListener('scroll', handleScroll))
</script>

<template>
  <div class="min-h-screen bg-gradient-to-br from-gray-900 via-blue-900 to-gray-900">
    <!-- Header -->
    <div class="border-b border-white/10">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
        <h1 class="text-4xl font-bold text-white">go2short API</h1>
        <p class="mt-2 text-blue-200 text-lg">Simple, fast URL shortening service</p>
      </div>
    </div>

    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <div class="lg:grid lg:grid-cols-12 lg:gap-8">
        <!-- Sidebar Navigation -->
        <aside class="hidden lg:block lg:col-span-3">
          <nav class="sticky top-8 space-y-1 bg-white/10 backdrop-blur-sm rounded-xl p-3 border border-white/20">
            <button
              v-for="section in sections"
              :key="section.id"
              @click="scrollToSection(section.id)"
              class="w-full text-left px-3 py-2 text-sm font-medium rounded-lg transition-colors"
              :class="activeSection === section.id
                ? 'bg-blue-500 text-white'
                : 'text-blue-100 hover:bg-white/10 hover:text-white'"
            >
              {{ section.title }}
            </button>
          </nav>
        </aside>

        <!-- Main Content -->
        <main class="lg:col-span-9 space-y-8 bg-white/10 backdrop-blur-sm rounded-xl p-6 border border-white/20">
          <!-- Overview -->
          <section id="overview" class="scroll-mt-8">
            <h2 class="text-2xl font-bold text-white border-b border-white/20 pb-2">Overview</h2>
            <div class="mt-4 prose prose-blue max-w-none">
              <p class="text-blue-100">
                go2short provides a simple REST API for creating and managing short URLs.
                All API endpoints that create or modify data require an API token for authentication.
              </p>
              <div class="mt-4 bg-amber-500/20 border border-amber-400/30 rounded-lg p-4">
                <p class="text-sm text-amber-200">
                  <strong>Authentication:</strong> Include your API token in the Authorization header:
                  <code class="bg-amber-500/30 px-1 rounded">Authorization: Bearer YOUR_TOKEN</code>
                </p>
              </div>
            </div>
          </section>

          <!-- Redirect -->
          <section id="redirect" class="scroll-mt-8">
            <h2 class="text-2xl font-bold text-white border-b border-white/20 pb-2">Redirect</h2>
            <div class="mt-4">
              <div class="flex items-center gap-2 mb-4">
                <span class="px-2 py-1 bg-green-500/20 text-green-300 text-xs font-semibold rounded">GET</span>
                <code class="text-blue-100 font-mono">/:code</code>
              </div>
              <p class="text-blue-100 mb-4">Redirect to the original URL. No authentication required.</p>
              <div class="bg-gray-900 rounded-lg overflow-hidden">
                <div class="flex items-center justify-between px-4 py-2 bg-gray-800">
                  <span class="text-gray-400 text-sm">Example</span>
                  <button @click="copyToClipboard(curlRedirect, $event)" class="text-xs text-gray-400 hover:text-white">Copy</button>
                </div>
                <pre class="p-4 text-sm text-gray-100 overflow-x-auto"><code>curl -L {{ origin }}/abc123</code></pre>
              </div>
              <div class="mt-4 text-sm text-blue-200">
                <strong>Response:</strong> 302 redirect to the original URL
              </div>
            </div>
          </section>

          <!-- QR Code -->
          <section id="qrcode" class="scroll-mt-8">
            <h2 class="text-2xl font-bold text-white border-b border-white/20 pb-2">QR Code</h2>
            <div class="mt-4">
              <div class="flex items-center gap-2 mb-4">
                <span class="px-2 py-1 bg-green-500/20 text-green-300 text-xs font-semibold rounded">GET</span>
                <code class="text-blue-100 font-mono">/:code/qr</code>
              </div>
              <p class="text-blue-100 mb-4">Get QR code image for a short link. No authentication required.</p>
              <div class="overflow-x-auto">
                <table class="min-w-full text-sm">
                  <thead>
                    <tr class="border-b border-white/20">
                      <th class="text-left py-2 text-blue-200 font-medium">Parameter</th>
                      <th class="text-left py-2 text-blue-200 font-medium">Type</th>
                      <th class="text-left py-2 text-blue-200 font-medium">Description</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr class="border-b border-white/10">
                      <td class="py-2 font-mono text-blue-100">size</td>
                      <td class="py-2 text-blue-100">integer</td>
                      <td class="py-2 text-blue-100">QR code size in pixels (128-1024, default: 256)</td>
                    </tr>
                  </tbody>
                </table>
              </div>
              <div class="mt-4 bg-gray-900 rounded-lg overflow-hidden">
                <div class="flex items-center justify-between px-4 py-2 bg-gray-800">
                  <span class="text-gray-400 text-sm">Example</span>
                  <button @click="copyToClipboard(curlQR, $event)" class="text-xs text-gray-400 hover:text-white">Copy</button>
                </div>
                <pre class="p-4 text-sm text-gray-100 overflow-x-auto"><code>curl {{ origin }}/abc123/qr?size=512 -o qr.png</code></pre>
              </div>
              <div class="mt-4 text-sm text-blue-200">
                <strong>Response:</strong> PNG image
              </div>
            </div>
          </section>

          <!-- Create Link -->
          <section id="create" class="scroll-mt-8">
            <h2 class="text-2xl font-bold text-white border-b border-white/20 pb-2">Create Link</h2>
            <div class="mt-4">
              <div class="flex items-center gap-2 mb-4">
                <span class="px-2 py-1 bg-blue-500/20 text-blue-300 text-xs font-semibold rounded">POST</span>
                <code class="text-blue-100 font-mono">/api/links</code>
                <span class="px-2 py-1 bg-red-500/20 text-red-300 text-xs rounded">Auth Required</span>
              </div>
              <p class="text-blue-100 mb-4">Create a new short link.</p>
              <div class="overflow-x-auto">
                <table class="min-w-full text-sm">
                  <thead>
                    <tr class="border-b border-white/20">
                      <th class="text-left py-2 text-blue-200 font-medium">Field</th>
                      <th class="text-left py-2 text-blue-200 font-medium">Type</th>
                      <th class="text-left py-2 text-blue-200 font-medium">Description</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr class="border-b border-white/10">
                      <td class="py-2 font-mono text-blue-100">long_url</td>
                      <td class="py-2 text-blue-100">string</td>
                      <td class="py-2 text-blue-100">The original URL to shorten (required)</td>
                    </tr>
                    <tr class="border-b border-white/10">
                      <td class="py-2 font-mono text-blue-100">custom_code</td>
                      <td class="py-2 text-blue-100">string</td>
                      <td class="py-2 text-blue-100">Custom short code (optional)</td>
                    </tr>
                    <tr class="border-b border-white/10">
                      <td class="py-2 font-mono text-blue-100">expires_at</td>
                      <td class="py-2 text-blue-100">string</td>
                      <td class="py-2 text-blue-100">Expiration time in ISO 8601 format (optional)</td>
                    </tr>
                  </tbody>
                </table>
              </div>
              <div class="mt-4 bg-gray-900 rounded-lg overflow-hidden">
                <div class="flex items-center justify-between px-4 py-2 bg-gray-800">
                  <span class="text-gray-400 text-sm">Request</span>
                  <button @click="copyToClipboard(curlCreate, $event)" class="text-xs text-gray-400 hover:text-white">Copy</button>
                </div>
                <pre class="p-4 text-sm text-gray-100 overflow-x-auto"><code>curl -X POST {{ origin }}/api/links \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"long_url": "https://example.com/very/long/path"}'</code></pre>
              </div>
              <div class="mt-4 bg-gray-900 rounded-lg overflow-hidden">
                <div class="px-4 py-2 bg-gray-800">
                  <span class="text-gray-400 text-sm">Response</span>
                </div>
                <pre class="p-4 text-sm text-gray-100 overflow-x-auto"><code>{
  "code": "abc123",
  "short_url": "{{ origin }}/abc123",
  "created_at": "2025-01-13T12:00:00Z"
}</code></pre>
              </div>
            </div>
          </section>

          <!-- Batch Create -->
          <section id="batch" class="scroll-mt-8">
            <h2 class="text-2xl font-bold text-white border-b border-white/20 pb-2">Batch Create</h2>
            <div class="mt-4">
              <div class="flex items-center gap-2 mb-4">
                <span class="px-2 py-1 bg-blue-500/20 text-blue-300 text-xs font-semibold rounded">POST</span>
                <code class="text-blue-100 font-mono">/api/links/batch</code>
                <span class="px-2 py-1 bg-red-500/20 text-red-300 text-xs rounded">Auth Required</span>
              </div>
              <p class="text-blue-100 mb-4">Create multiple short links in one request. Maximum 100 items per request.</p>
              <div class="mt-4 bg-gray-900 rounded-lg overflow-hidden">
                <div class="flex items-center justify-between px-4 py-2 bg-gray-800">
                  <span class="text-gray-400 text-sm">Request</span>
                  <button @click="copyToClipboard(curlBatch, $event)" class="text-xs text-gray-400 hover:text-white">Copy</button>
                </div>
                <pre class="p-4 text-sm text-gray-100 overflow-x-auto"><code>curl -X POST {{ origin }}/api/links/batch \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"items": [{"long_url": "https://example1.com"}, {"long_url": "https://example2.com", "custom_code": "foo"}]}'</code></pre>
              </div>
              <div class="mt-4 bg-gray-900 rounded-lg overflow-hidden">
                <div class="px-4 py-2 bg-gray-800">
                  <span class="text-gray-400 text-sm">Response</span>
                </div>
                <pre class="p-4 text-sm text-gray-100 overflow-x-auto"><code>{
  "results": [
    {"index": 0, "code": "abc123", "short_url": "{{ origin }}/abc123"},
    {"index": 1, "code": "foo", "short_url": "{{ origin }}/foo"}
  ]
}</code></pre>
              </div>
            </div>
          </section>

          <!-- Link Preview -->
          <section id="preview" class="scroll-mt-8">
            <h2 class="text-2xl font-bold text-white border-b border-white/20 pb-2">Link Preview</h2>
            <div class="mt-4">
              <div class="flex items-center gap-2 mb-4">
                <span class="px-2 py-1 bg-green-500/20 text-green-300 text-xs font-semibold rounded">GET</span>
                <code class="text-blue-100 font-mono">/api/links/:code/preview</code>
                <span class="px-2 py-1 bg-red-500/20 text-red-300 text-xs rounded">Auth Required</span>
              </div>
              <p class="text-blue-100 mb-4">Get link details without triggering a redirect.</p>
              <div class="mt-4 bg-gray-900 rounded-lg overflow-hidden">
                <div class="flex items-center justify-between px-4 py-2 bg-gray-800">
                  <span class="text-gray-400 text-sm">Request</span>
                  <button @click="copyToClipboard(curlPreview, $event)" class="text-xs text-gray-400 hover:text-white">Copy</button>
                </div>
                <pre class="p-4 text-sm text-gray-100 overflow-x-auto"><code>curl {{ origin }}/api/links/abc123/preview \
  -H "Authorization: Bearer YOUR_TOKEN"</code></pre>
              </div>
              <div class="mt-4 bg-gray-900 rounded-lg overflow-hidden">
                <div class="px-4 py-2 bg-gray-800">
                  <span class="text-gray-400 text-sm">Response</span>
                </div>
                <pre class="p-4 text-sm text-gray-100 overflow-x-auto"><code>{
  "code": "abc123",
  "long_url": "https://example.com/very/long/path"
}</code></pre>
              </div>
            </div>
          </section>

          <!-- Rate Limiting -->
          <section id="ratelimit" class="scroll-mt-8">
            <h2 class="text-2xl font-bold text-white border-b border-white/20 pb-2">Rate Limiting</h2>
            <div class="mt-4">
              <p class="text-blue-100 mb-4">
                Link creation endpoints are rate limited to <strong>60 requests per minute</strong> per IP address.
              </p>
              <div class="overflow-x-auto">
                <table class="min-w-full text-sm">
                  <thead>
                    <tr class="border-b border-white/20">
                      <th class="text-left py-2 text-blue-200 font-medium">Header</th>
                      <th class="text-left py-2 text-blue-200 font-medium">Description</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr class="border-b border-white/10">
                      <td class="py-2 font-mono text-blue-100">X-RateLimit-Limit</td>
                      <td class="py-2 text-blue-100">Maximum requests allowed per window</td>
                    </tr>
                    <tr class="border-b border-white/10">
                      <td class="py-2 font-mono text-blue-100">X-RateLimit-Remaining</td>
                      <td class="py-2 text-blue-100">Remaining requests in current window</td>
                    </tr>
                    <tr class="border-b border-white/10">
                      <td class="py-2 font-mono text-blue-100">X-RateLimit-Reset</td>
                      <td class="py-2 text-blue-100">Unix timestamp when the window resets</td>
                    </tr>
                  </tbody>
                </table>
              </div>
              <div class="mt-4 bg-red-500/20 border border-red-400/30 rounded-lg p-4">
                <p class="text-sm text-red-200">
                  When rate limited, the API returns <code class="bg-red-500/30 px-1 rounded">429 Too Many Requests</code> with a <code class="bg-red-500/30 px-1 rounded">Retry-After</code> header.
                </p>
              </div>
            </div>
          </section>
        </main>
      </div>
    </div>
  </div>
</template>
