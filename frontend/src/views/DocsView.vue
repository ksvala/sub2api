<template>
  <div class="relative min-h-screen overflow-hidden bg-gradient-to-br from-gray-50 via-primary-50/30 to-gray-100 dark:from-dark-950 dark:via-dark-900 dark:to-dark-950">
    <!-- Background Decorations -->
    <div class="pointer-events-none absolute inset-0 overflow-hidden">
      <div class="absolute -right-40 -top-40 h-96 w-96 rounded-full bg-primary-400/20 blur-3xl"></div>
      <div class="absolute -bottom-40 -left-40 h-96 w-96 rounded-full bg-primary-500/15 blur-3xl"></div>
      <div class="absolute left-1/3 top-1/4 h-72 w-72 rounded-full bg-primary-300/10 blur-3xl"></div>
      <div class="absolute inset-0 bg-[linear-gradient(rgba(20,184,166,0.03)_1px,transparent_1px),linear-gradient(90deg,rgba(20,184,166,0.03)_1px,transparent_1px)] bg-[size:64px_64px]"></div>
    </div>

    <!-- Header -->
    <header class="relative z-10 px-6 py-4">
      <nav class="mx-auto flex max-w-6xl flex-wrap items-center justify-between gap-4">
        <router-link to="/home" class="flex items-center gap-3">
          <div class="h-10 w-10 overflow-hidden rounded-xl shadow-md">
            <img :src="siteLogo || '/logo.png'" alt="Logo" class="h-full w-full object-contain" />
          </div>
          <span class="text-lg font-bold text-gray-900 dark:text-white">{{ siteName }}</span>
        </router-link>

        <div class="flex items-center gap-3">
          <router-link
            to="/home"
            class="text-sm font-medium text-gray-600 hover:text-primary-600 dark:text-gray-300 dark:hover:text-primary-400"
          >
            {{ t('nav.home') }}
          </router-link>
          <router-link
            v-if="isAuthenticated"
            to="/plans"
            class="text-sm font-medium text-gray-600 hover:text-primary-600 dark:text-gray-300 dark:hover:text-primary-400"
          >
            {{ t('nav.plans') }}
          </router-link>
          <LocaleSwitcher />
          <router-link
            v-if="isAuthenticated"
            :to="dashboardPath"
            class="inline-flex items-center gap-2 rounded-full bg-gray-900 px-4 py-1.5 text-xs font-medium text-white transition-colors hover:bg-gray-800 dark:bg-gray-800 dark:hover:bg-gray-700"
          >
            {{ t('home.dashboard') }}
          </router-link>
          <router-link
            v-else
            to="/login"
            class="inline-flex items-center gap-2 rounded-full bg-gray-900 px-4 py-1.5 text-xs font-medium text-white transition-colors hover:bg-gray-800 dark:bg-gray-800 dark:hover:bg-gray-700"
          >
            {{ t('home.login') }}
          </router-link>
        </div>
      </nav>
    </header>

    <!-- Content -->
    <main class="relative z-10 px-6 pb-16 pt-8">
      <div class="mx-auto max-w-4xl">
        <div class="rounded-3xl border border-gray-200/70 bg-white/90 p-6 shadow-lg shadow-gray-200/40 backdrop-blur dark:border-dark-700/70 dark:bg-dark-900/70 dark:shadow-dark-900/40 md:p-10">
          <div class="mb-6 flex items-center justify-between">
            <h1 class="text-2xl font-bold text-gray-900 dark:text-white md:text-3xl">
              {{ t('nav.docs') }}
            </h1>
          </div>

          <div v-if="!docMarkdown" class="rounded-2xl border border-dashed border-gray-200 bg-gray-50 p-6 text-center text-sm text-gray-500 dark:border-dark-700 dark:bg-dark-900/40 dark:text-dark-400">
            {{ t('common.noData') }}
          </div>
          <div v-else class="docs-content" v-html="renderedMarkdown"></div>
        </div>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { marked } from 'marked'
import { useAppStore, useAuthStore } from '@/stores'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'

const { t } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()

const isAuthenticated = computed(() => authStore.isAuthenticated)
const isAdmin = computed(() => authStore.isAdmin)
const dashboardPath = computed(() => (isAdmin.value ? '/admin/dashboard' : '/dashboard'))

const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'Sub2API')
const siteLogo = computed(() => appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '')
const docMarkdown = computed(
  () => appStore.cachedPublicSettings?.doc_markdown || appStore.docMarkdown || ''
)

marked.setOptions({
  gfm: true,
  breaks: true
})

const renderedMarkdown = computed(() => {
  if (!docMarkdown.value) return ''
  return marked.parse(docMarkdown.value)
})

onMounted(() => {
  if (!appStore.publicSettingsLoaded) {
    appStore.fetchPublicSettings()
  }
})
</script>

<style scoped>
.docs-content :deep(h1) {
  @apply mb-4 text-3xl font-bold text-gray-900 dark:text-white;
}

.docs-content :deep(h2) {
  @apply mt-8 text-2xl font-semibold text-gray-900 dark:text-white;
}

.docs-content :deep(h3) {
  @apply mt-6 text-xl font-semibold text-gray-900 dark:text-white;
}

.docs-content :deep(p) {
  @apply mt-3 text-sm leading-7 text-gray-700 dark:text-dark-200;
}

.docs-content :deep(ul),
.docs-content :deep(ol) {
  @apply mt-3 list-disc space-y-2 pl-6 text-sm text-gray-700 dark:text-dark-200;
}

.docs-content :deep(ol) {
  @apply list-decimal;
}

.docs-content :deep(a) {
  @apply text-primary-600 underline decoration-primary-300 underline-offset-4 hover:text-primary-700 dark:text-primary-400;
}

.docs-content :deep(code) {
  @apply rounded bg-gray-100 px-1.5 py-0.5 font-mono text-xs text-gray-800 dark:bg-dark-800 dark:text-dark-200;
}

.docs-content :deep(pre) {
  @apply mt-4 overflow-x-auto rounded-xl bg-gray-900 p-4 text-xs text-gray-100;
}

.docs-content :deep(pre code) {
  @apply bg-transparent p-0 text-gray-100;
}

.docs-content :deep(blockquote) {
  @apply mt-4 border-l-4 border-primary-400/60 bg-primary-50/60 px-4 py-3 text-sm text-gray-700 dark:border-primary-400/40 dark:bg-primary-900/20 dark:text-dark-200;
}

.docs-content :deep(hr) {
  @apply my-6 border-gray-200 dark:border-dark-700;
}
</style>
