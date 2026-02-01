import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import i18n from './i18n'
import './style.css'

const app = createApp(App)
const pinia = createPinia()
app.use(pinia)

const GA_MEASUREMENT_ID = 'G-13DCM6B9KD'
let gtagLoaded = false

function loadGtag(): void {
  if (gtagLoaded) return
  gtagLoaded = true

  const script = document.createElement('script')
  script.async = true
  script.src = `https://www.googletagmanager.com/gtag/js?id=${GA_MEASUREMENT_ID}`
  document.head.appendChild(script)

  window.dataLayer = window.dataLayer || []
  function gtag(...args: any[]) {
    window.dataLayer?.push(args)
  }
  gtag('js', new Date())
  gtag('config', GA_MEASUREMENT_ID)

  window.gtag = gtag
}

function shouldTrackRoute(): boolean {
  return router.currentRoute.value.meta?.requiresAdmin !== true
}

// Initialize settings from injected config BEFORE mounting (prevents flash)
// This must happen after pinia is installed but before router and i18n
import { useAppStore } from '@/stores/app'
const appStore = useAppStore()
appStore.initFromInjectedConfig()

// Set document title immediately after config is loaded
if (appStore.siteName && appStore.siteName !== 'Sub2API') {
  document.title = `${appStore.siteName} - AI API Gateway`
}

app.use(router)
app.use(i18n)

// 等待路由器完成初始导航后再挂载，避免竞态条件导致的空白渲染
router.isReady().then(() => {
  if (shouldTrackRoute()) {
    loadGtag()
  }
  router.afterEach((to) => {
    if (to.meta?.requiresAdmin === true) {
      return
    }
    if (!gtagLoaded) {
      loadGtag()
      return
    }
    window.gtag?.('config', GA_MEASUREMENT_ID, { page_path: to.fullPath })
  })
  app.mount('#app')
})
