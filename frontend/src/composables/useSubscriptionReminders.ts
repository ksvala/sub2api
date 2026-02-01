import type { UserSubscription } from '@/types'

type ToastFn = (message: string, duration?: number) => string

export function showSubscriptionRemindersOncePerDay(
  subscriptions: UserSubscription[],
  showWarning: ToastFn,
  options?: { isAdminRoute?: boolean }
): void {
  if (options?.isAdminRoute) return
  if (!subscriptions || subscriptions.length === 0) return

  const todayKey = getLocalDateKey(new Date())

  const usage95: Array<{ sub: UserSubscription; pct: number; used: number; limit: number }> = []
  const usage80: Array<{ sub: UserSubscription; pct: number; used: number; limit: number }> = []

  const expiry0: Array<{ sub: UserSubscription; days: number }> = []
  const expiry1: Array<{ sub: UserSubscription; days: number }> = []
  const expiry3: Array<{ sub: UserSubscription; days: number }> = []

  for (const sub of subscriptions) {
    const dailyLimit = sub.group?.daily_limit_usd
    if (dailyLimit != null && dailyLimit > 0) {
      const used = sub.daily_usage_usd
      const pct = (used / dailyLimit) * 100

      if (pct >= 95) {
        if (!getOnceKey(`reminder:daily95:${sub.id}:${todayKey}`)) {
          usage95.push({ sub, pct, used, limit: dailyLimit })
        }
      } else if (pct >= 80) {
        if (!getOnceKey(`reminder:daily80:${sub.id}:${todayKey}`)) {
          usage80.push({ sub, pct, used, limit: dailyLimit })
        }
      }
    }

    if (sub.expires_at) {
      const days = calcDaysRemaining(sub.expires_at)
      if (days === 0 || days === 1 || days === 3) {
        if (!getOnceKey(`reminder:expiry:${sub.id}:${days}:${todayKey}`)) {
          if (days === 0) expiry0.push({ sub, days })
          if (days === 1) expiry1.push({ sub, days })
          if (days === 3) expiry3.push({ sub, days })
        }
      }
    }
  }

  // 1) Usage reminder: prefer 95% then 80%.
  if (usage95.length > 0) {
    markUsageOnce(usage95, todayKey, 95)
    showWarning(buildUsageToastMessage(usage95, 95), 6000)
  } else if (usage80.length > 0) {
    markUsageOnce(usage80, todayKey, 80)
    showWarning(buildUsageToastMessage(usage80, 80), 6000)
  }

  // 2) Expiry reminder: prefer 0 then 1 then 3.
  if (expiry0.length > 0) {
    markExpiryOnce(expiry0, todayKey, 0)
    showWarning(buildExpiryToastMessage(expiry0, 0), 7000)
  } else if (expiry1.length > 0) {
    markExpiryOnce(expiry1, todayKey, 1)
    showWarning(buildExpiryToastMessage(expiry1, 1), 7000)
  } else if (expiry3.length > 0) {
    markExpiryOnce(expiry3, todayKey, 3)
    showWarning(buildExpiryToastMessage(expiry3, 3), 7000)
  }
}

function buildUsageToastMessage(
  items: Array<{ sub: UserSubscription; pct: number; used: number; limit: number }>,
  threshold: 80 | 95
): string {
  const names = summarizeGroups(items.map((i) => i.sub))
  return `用量提醒：你的订阅今日用量已达 ${threshold}%（${names}）`
}

function buildExpiryToastMessage(items: Array<{ sub: UserSubscription }>, days: 0 | 1 | 3): string {
  const names = summarizeGroups(items.map((i) => i.sub))
  if (days === 0) return `到期提醒：你的订阅今天到期（${names}）`
  return `到期提醒：你的订阅剩余 ${days} 天到期（${names}）`
}

function summarizeGroups(subs: UserSubscription[]): string {
  const names = Array.from(
    new Set(
      subs
        .map((s) => s.group?.name)
        .filter((v): v is string => typeof v === 'string' && v.trim().length > 0)
    )
  )
  if (names.length === 0) return '未命名订阅'
  if (names.length <= 3) return names.join('、')
  return `${names.slice(0, 3).join('、')} 等 ${names.length} 个`
}

function getLocalDateKey(d: Date): string {
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${y}-${m}-${day}`
}

function calcDaysRemaining(expiresAt: string): number | null {
  const expires = new Date(expiresAt)
  if (Number.isNaN(expires.getTime())) return null
  const now = Date.now()
  const diffMs = expires.getTime() - now
  const dayMs = 24 * 60 * 60 * 1000
  const days = Math.ceil(diffMs / dayMs)
  if (days < 0) return null
  return days
}

function getOnceKey(key: string): boolean {
  try {
    return localStorage.getItem(key) === '1'
  } catch {
    return false
  }
}

function setOnceKey(key: string): void {
  try {
    localStorage.setItem(key, '1')
  } catch {
    // ignore
  }
}

function markUsageOnce(
  items: Array<{ sub: UserSubscription }>,
  todayKey: string,
  threshold: 80 | 95
): void {
  for (const item of items) {
    const key = threshold === 95 ? `reminder:daily95:${item.sub.id}:${todayKey}` : `reminder:daily80:${item.sub.id}:${todayKey}`
    setOnceKey(key)
  }
}

function markExpiryOnce(items: Array<{ sub: UserSubscription }>, todayKey: string, days: 0 | 1 | 3): void {
  for (const item of items) {
    setOnceKey(`reminder:expiry:${item.sub.id}:${days}:${todayKey}`)
  }
}
