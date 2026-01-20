<template>
  <AppLayout>
    <div class="space-y-6">
      <!-- Header -->
      <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-end">
        <button @click="openCreateModal" class="btn btn-primary">
          <Icon name="plus" class="mr-2 h-5 w-5" />
          {{ t('admin.plans.create') }}
        </button>
      </div>

      <!-- Loading State -->
      <div v-if="loading" class="flex items-center justify-center py-12">
        <div class="h-8 w-8 animate-spin rounded-full border-b-2 border-primary-600"></div>
      </div>

      <!-- Empty State -->
      <div v-else-if="plans.length === 0" class="flex flex-col items-center justify-center py-12">
        <div class="rounded-full bg-gray-100 p-4 dark:bg-dark-800">
          <Icon name="creditCard" class="h-8 w-8 text-gray-400" />
        </div>
        <h3 class="mt-4 text-lg font-medium text-gray-900 dark:text-white">
          {{ t('admin.plans.noPlans') }}
        </h3>
        <p class="mt-1 text-gray-500 dark:text-gray-400">
          {{ t('admin.plans.createFirst') }}
        </p>
      </div>

      <!-- Plans List -->
      <div v-else class="grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
        <div
          v-for="plan in sortedPlans"
          :key="plan.id"
          class="card flex flex-col overflow-hidden transition-shadow hover:shadow-lg"
        >
          <!-- Card Header -->
          <div class="border-b border-gray-100 bg-gray-50 px-6 py-4 dark:border-dark-700 dark:bg-dark-800/50">
            <div class="flex items-center justify-between">
              <span class="inline-flex items-center rounded-full bg-blue-100 px-2.5 py-0.5 text-xs font-medium text-blue-800 dark:bg-blue-900/30 dark:text-blue-300">
                {{ plan.group_name }}
              </span>
              <div class="flex items-center gap-2">
                <button
                  @click="toggleEnabled(plan)"
                  class="relative inline-flex h-5 w-9 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2"
                  :class="[plan.enabled ? 'bg-primary-600' : 'bg-gray-200 dark:bg-dark-600']"
                >
                  <span
                    class="pointer-events-none inline-block h-4 w-4 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out"
                    :class="[plan.enabled ? 'translate-x-4' : 'translate-x-0']"
                  />
                </button>
                <div class="dropdown dropdown-end">
                  <button class="btn btn-icon btn-sm btn-ghost">
                    <Icon name="more" class="h-5 w-5" />
                  </button>
                  <ul class="dropdown-content menu rounded-box w-52 bg-base-100 p-2 shadow-xl dark:bg-dark-800">
                    <li>
                      <button @click="openEditModal(plan)" class="text-left hover:bg-gray-100 dark:hover:bg-dark-700">
                        <Icon name="edit" class="h-4 w-4" />
                        {{ t('common.edit') }}
                      </button>
                    </li>
                    <li>
                      <button @click="confirmDelete(plan)" class="text-left text-red-600 hover:bg-red-50 dark:text-red-400 dark:hover:bg-red-900/20">
                        <Icon name="trash" class="h-4 w-4" />
                        {{ t('common.delete') }}
                      </button>
                    </li>
                  </ul>
                </div>
              </div>
            </div>
            <h3 class="mt-2 text-xl font-bold text-gray-900 dark:text-white">{{ plan.title }}</h3>
            <div class="mt-1 flex items-baseline text-2xl font-bold text-primary-600">
              ¥{{ plan.price }}
            </div>
          </div>

          <!-- Card Body -->
          <div class="flex-1 p-6">
            <p class="mb-4 text-sm text-gray-500 line-clamp-3 dark:text-gray-400">
              {{ plan.description }}
            </p>
            
            <div class="space-y-3 text-sm">
              <div class="flex justify-between border-b border-gray-100 pb-2 dark:border-dark-700">
                <span class="text-gray-500 dark:text-gray-400">{{ t('admin.plans.dailyQuota') }}</span>
                <span class="font-medium text-gray-900 dark:text-white">${{ plan.daily_quota }}</span>
              </div>
              <div class="flex justify-between border-b border-gray-100 pb-2 dark:border-dark-700">
                <span class="text-gray-500 dark:text-gray-400">{{ t('admin.plans.totalQuota') }}</span>
                <span class="font-medium text-gray-900 dark:text-white">${{ plan.total_quota }}</span>
              </div>
              <div class="flex justify-between">
                <span class="text-gray-500 dark:text-gray-400">{{ t('admin.plans.sortOrder') }}</span>
                <span class="font-medium text-gray-900 dark:text-white">{{ plan.sort_order }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Create/Edit Modal -->
      <BaseDialog :show="showModal" :title="isEditing ? t('admin.plans.editPlan') : t('admin.plans.createPlan')" @close="showModal = false">
        <form @submit.prevent="savePlan" class="space-y-4">
          <!-- Basic Info -->
          <div class="grid grid-cols-2 gap-4">
            <div class="col-span-2">
              <label class="label">{{ t('admin.plans.form.title') }}</label>
              <input v-model="form.title" type="text" class="input w-full" required />
            </div>
            
            <div class="col-span-2">
              <label class="label">{{ t('admin.plans.form.description') }}</label>
              <textarea v-model="form.description" class="input w-full" rows="3"></textarea>
            </div>

            <div>
              <label class="label">{{ t('admin.plans.form.price') }}</label>
              <div class="relative">
                <span class="absolute left-3 top-2 text-gray-500">¥</span>
                <input v-model.number="form.price" type="number" step="0.01" min="0" class="input w-full pl-7" required />
              </div>
            </div>

            <div>
              <label class="label">{{ t('admin.plans.form.sortOrder') }}</label>
              <input v-model.number="form.sort_order" type="number" class="input w-full" required />
            </div>
          </div>

          <!-- Group Info -->
          <div class="grid grid-cols-2 gap-4 rounded-lg bg-gray-50 p-4 dark:bg-dark-800">
            <div class="col-span-2 text-sm font-medium text-gray-700 dark:text-gray-300">
              {{ t('admin.plans.form.groupSettings') }}
            </div>
            <div>
              <label class="label">{{ t('admin.plans.form.groupName') }}</label>
              <input v-model="form.group_name" type="text" class="input w-full" required />
            </div>
            <div>
              <label class="label">{{ t('admin.plans.form.groupSort') }}</label>
              <input v-model.number="form.group_sort" type="number" class="input w-full" required />
            </div>
          </div>

          <!-- Quota Info -->
          <div class="grid grid-cols-2 gap-4 rounded-lg bg-gray-50 p-4 dark:bg-dark-800">
            <div class="col-span-2 text-sm font-medium text-gray-700 dark:text-gray-300">
              {{ t('admin.plans.form.quotaSettings') }}
            </div>
            <div>
              <label class="label">{{ t('admin.plans.form.dailyQuota') }}</label>
              <div class="relative">
                <span class="absolute left-3 top-2 text-gray-500">$</span>
                <input v-model.number="form.daily_quota" type="number" step="0.01" min="0" class="input w-full pl-7" required />
              </div>
            </div>
            <div>
              <label class="label">{{ t('admin.plans.form.totalQuota') }}</label>
              <div class="relative">
                <span class="absolute left-3 top-2 text-gray-500">$</span>
                <input v-model.number="form.total_quota" type="number" step="0.01" min="0" class="input w-full pl-7" required />
              </div>
            </div>
          </div>

          <!-- QR Upload -->
          <div>
            <label class="label">{{ t('admin.plans.form.purchaseQr') }}</label>
            <div class="flex items-center gap-4">
              <div v-if="form.purchase_qr_url" class="relative h-24 w-24 overflow-hidden rounded-lg border border-gray-200 dark:border-dark-700">
                <img :src="form.purchase_qr_url" class="h-full w-full object-contain" />
                <button 
                  type="button" 
                  @click="form.purchase_qr_url = ''"
                  class="absolute right-0 top-0 bg-red-500 p-1 text-white hover:bg-red-600"
                >
                  <Icon name="x" class="h-3 w-3" />
                </button>
              </div>
              <div class="flex-1">
                <label class="btn btn-secondary cursor-pointer">
                  <Icon name="upload" class="mr-2 h-4 w-4" />
                  {{ t('common.upload') }}
                  <input type="file" accept="image/*" class="hidden" @change="handleUpload" />
                </label>
                <div v-if="uploading" class="mt-2 text-sm text-gray-500">
                  {{ t('common.uploading') }}...
                </div>
              </div>
            </div>
          </div>

          <!-- Enabled Toggle -->
          <div class="flex items-center gap-2">
             <Toggle v-model="form.enabled" />
             <span class="text-sm font-medium">{{ t('admin.plans.form.enabled') }}</span>
          </div>

          <!-- Footer -->
          <div class="flex justify-end gap-3 pt-4">
            <button type="button" class="btn btn-ghost" @click="showModal = false">
              {{ t('common.cancel') }}
            </button>
            <button type="submit" class="btn btn-primary" :disabled="saving">
              {{ saving ? t('common.saving') : t('common.save') }}
            </button>
          </div>
        </form>
      </BaseDialog>

      <!-- Delete Confirmation Modal -->
      <ConfirmDialog
        :show="showDeleteModal"
        :title="t('admin.plans.deleteTitle')"
        :message="t('admin.plans.deleteConfirm', { title: selectedPlan?.title })"
        danger
        @confirm="handleDelete"
        @cancel="showDeleteModal = false"
      />
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, reactive } from 'vue'
import { useI18n } from 'vue-i18n'
import { adminAPI } from '@/api/admin'
import { useAppStore } from '@/stores'
import type { Plan } from '@/types'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import Toggle from '@/components/common/Toggle.vue'

const { t } = useI18n()
const appStore = useAppStore()

const plans = ref<Plan[]>([])
const loading = ref(true)
const saving = ref(false)
const uploading = ref(false)
const showModal = ref(false)
const showDeleteModal = ref(false)
const selectedPlan = ref<Plan | null>(null)

const form = reactive({
  title: '',
  description: '',
  price: 0,
  group_name: '',
  group_sort: 0,
  daily_quota: 0,
  total_quota: 0,
  purchase_qr_url: '',
  enabled: true,
  sort_order: 0
})

const isEditing = computed(() => !!selectedPlan.value)

const sortedPlans = computed(() => {
  return [...plans.value].sort((a, b) => {
    // First by group sort
    if (a.group_sort !== b.group_sort) return a.group_sort - b.group_sort
    // Then by sort order
    return a.sort_order - b.sort_order
  })
})

async function fetchPlans() {
  loading.value = true
  try {
    plans.value = await adminAPI.plans.getPlans()
  } catch (error: any) {
    appStore.showError(error.message)
  } finally {
    loading.value = false
  }
}

function openCreateModal() {
  selectedPlan.value = null
  Object.assign(form, {
    title: '',
    description: '',
    price: 0,
    group_name: '',
    group_sort: 0,
    daily_quota: 0,
    total_quota: 0,
    purchase_qr_url: '',
    enabled: true,
    sort_order: 0
  })
  showModal.value = true
}

function openEditModal(plan: Plan) {
  selectedPlan.value = plan
  Object.assign(form, {
    title: plan.title,
    description: plan.description,
    price: plan.price,
    group_name: plan.group_name,
    group_sort: plan.group_sort,
    daily_quota: plan.daily_quota,
    total_quota: plan.total_quota,
    purchase_qr_url: plan.purchase_qr_url,
    enabled: plan.enabled,
    sort_order: plan.sort_order
  })
  showModal.value = true
}

async function handleUpload(event: Event) {
  const input = event.target as HTMLInputElement
  if (!input.files?.length) return

  uploading.value = true
  try {
    const file = input.files[0]
    const { url } = await adminAPI.uploads.uploadImage(file)
    form.purchase_qr_url = url
  } catch (error: any) {
    appStore.showError(error.message)
  } finally {
    uploading.value = false
    input.value = ''
  }
}

async function savePlan() {
  saving.value = true
  try {
    if (isEditing) {
      await adminAPI.plans.updatePlan(selectedPlan.value!.id, form)
      appStore.showSuccess(t('common.updated'))
    } else {
      await adminAPI.plans.createPlan(form)
      appStore.showSuccess(t('common.created'))
    }
    showModal.value = false
    fetchPlans()
  } catch (error: any) {
    appStore.showError(error.message)
  } finally {
    saving.value = false
  }
}

function confirmDelete(plan: Plan) {
  selectedPlan.value = plan
  showDeleteModal.value = true
}

async function handleDelete() {
  if (!selectedPlan.value) return
  try {
    await adminAPI.plans.deletePlan(selectedPlan.value.id)
    appStore.showSuccess(t('common.deleted'))
    showDeleteModal.value = false
    fetchPlans()
  } catch (error: any) {
    appStore.showError(error.message)
  }
}

async function toggleEnabled(plan: Plan) {
  try {
    await adminAPI.plans.updatePlan(plan.id, { enabled: !plan.enabled })
    plan.enabled = !plan.enabled
  } catch (error: any) {
    appStore.showError(error.message)
  }
}

onMounted(() => {
  fetchPlans()
})
</script>
