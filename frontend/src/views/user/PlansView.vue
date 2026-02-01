<template>
  <AppLayout>
    <div class="space-y-8">
      <!-- Loading State -->
      <div v-if="loading" class="flex items-center justify-center py-12">
        <div class="h-8 w-8 animate-spin rounded-full border-b-2 border-primary-600"></div>
      </div>

      <!-- Empty State -->
      <div v-else-if="planGroups.length === 0" class="flex flex-col items-center justify-center py-12">
        <div class="rounded-full bg-gray-100 p-4 dark:bg-dark-800">
          <Icon name="creditCard" class="h-8 w-8 text-gray-400" />
        </div>
        <h3 class="mt-4 text-lg font-medium text-gray-900 dark:text-white">
          {{ t('plans.noPlans') }}
        </h3>
      </div>

      <!-- Plan Groups -->
      <div v-else class="space-y-12">
        <div v-for="group in planGroups" :key="group.name" class="space-y-6">
          <div v-if="group.name" class="flex items-center gap-4">
            <h2 class="text-xl font-bold text-gray-900 dark:text-white">
              {{ group.name }}
            </h2>
            <div class="h-px flex-1 bg-gray-200 dark:bg-dark-700"></div>
          </div>

          <div class="grid gap-6 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
            <div
              v-for="plan in group.plans"
              :key="plan.id"
              class="card group relative flex flex-col overflow-hidden transition-all hover:scale-[1.02] hover:shadow-xl dark:hover:shadow-primary-900/10"
            >
              <!-- Card Header -->
              <div class="border-b border-gray-100 bg-gray-50 p-6 dark:border-dark-700 dark:bg-dark-800/50">
                <h3 class="text-lg font-bold text-gray-900 dark:text-white">{{ plan.title }}</h3>
                <div class="mt-4 flex items-baseline">
                  <span class="text-3xl font-extrabold tracking-tight text-primary-600">¥{{ plan.price }}</span>
                </div>
              </div>

              <!-- Card Body -->
              <div class="flex flex-1 flex-col p-6">
                <div class="mb-6 flex-1">
                  <p class="text-sm text-gray-500 dark:text-gray-400">
                    {{ plan.description }}
                  </p>
                </div>

                <div class="mb-6 space-y-3">
                  <div class="flex items-center justify-between text-sm">
                    <span class="text-gray-500 dark:text-gray-400">{{ t('plans.dailyQuota') }}</span>
                    <span class="font-medium text-gray-900 dark:text-white">${{ plan.daily_quota }}</span>
                  </div>
                  <div class="flex items-center justify-between text-sm">
                    <span class="text-gray-500 dark:text-gray-400">{{ t('plans.totalQuota') }}</span>
                    <span class="font-medium text-gray-900 dark:text-white">${{ plan.total_quota }}</span>
                  </div>
                </div>

                <button
                  @click="openPurchaseModal(plan)"
                  class="btn btn-primary w-full shadow-lg shadow-primary-600/20"
                >
                  {{ t('plans.purchase') }}
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Purchase Modal -->
      <BaseDialog :show="showModal" :title="selectedPlan?.title || ''" @close="showModal = false">
        <div class="flex flex-col items-center space-y-6 p-4">
          <div class="text-center">
            <p class="text-sm text-gray-500 dark:text-gray-400">
              {{ t('plans.scanToPurchase') }}
            </p>
            <div class="mt-2 text-2xl font-bold text-primary-600">
              ¥{{ selectedPlan?.price }}
            </div>
          </div>

          <div class="h-64 w-64 overflow-hidden rounded-xl border-2 border-dashed border-gray-200 bg-gray-50 p-2 dark:border-dark-700 dark:bg-dark-800">
            <img 
              v-if="selectedPlan?.purchase_qr_url" 
              :src="selectedPlan.purchase_qr_url" 
              class="h-full w-full object-contain"
            />
            <div v-else class="flex h-full w-full items-center justify-center text-gray-400">
              {{ t('plans.noQrCode') }}
            </div>
          </div>

          <p class="text-center text-xs text-gray-400">
            {{ t('plans.purchaseNote') }}
          </p>

          <div class="flex w-full gap-3">
            <button @click="showModal = false" class="btn btn-secondary flex-1">
              {{ t('common.close') }}
            </button>
            <button @click="showModal = false" class="btn btn-primary flex-1">
              {{ t('plans.contacted') }}
            </button>
          </div>
        </div>
      </BaseDialog>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { plansAPI } from '@/api/plans'
import { useAppStore } from '@/stores'
import type { Plan, PlanGroup } from '@/types'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'

const { t } = useI18n()
const appStore = useAppStore()

const plans = ref<Plan[]>([])
const loading = ref(true)
const showModal = ref(false)
const selectedPlan = ref<Plan | null>(null)

const planGroups = computed<PlanGroup[]>(() => {
  const groups: Record<string, Plan[]> = {}
  
  // Group plans
  plans.value.forEach(plan => {
    if (plan.enabled === false) return
    const groupName = plan.group_name || t('plans.defaultGroup')
    if (!groups[groupName]) {
      groups[groupName] = []
    }
    groups[groupName].push(plan)
  })

  // Convert to array and sort
  return Object.entries(groups)
    .map(([name, groupPlans]) => {
      // Sort plans within group
      groupPlans.sort((a, b) => a.sort_order - b.sort_order)
      return {
        name,
        // Use the group_sort of the first plan in the group (assuming consistent)
        sort: groupPlans[0]?.group_sort || 0,
        plans: groupPlans
      }
    })
    .sort((a, b) => a.sort - b.sort)
})

async function fetchPlans() {
  loading.value = true
  try {
    plans.value = await plansAPI.getPlans()
  } catch (error: any) {
    appStore.showError(error.message)
  } finally {
    loading.value = false
  }
}

function openPurchaseModal(plan: Plan) {
  selectedPlan.value = plan
  showModal.value = true
}

onMounted(() => {
  fetchPlans()
})
</script>
