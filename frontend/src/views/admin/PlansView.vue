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

      <!-- Group Sorting -->
      <div v-if="groupOrder.length > 1" class="rounded-2xl border border-gray-200/70 bg-white/80 p-5 shadow-sm dark:border-dark-700/70 dark:bg-dark-900/50">
        <div class="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <div class="text-sm font-semibold text-gray-900 dark:text-white">
              {{ t('admin.plans.groupSorting.title') }}
            </div>
            <p class="text-xs text-gray-500 dark:text-dark-400">
              {{ t('admin.plans.groupSorting.description') }}
            </p>
          </div>
          <div v-if="savingGroupSort" class="text-xs text-gray-500 dark:text-dark-400">
            {{ t('common.saving') }}
          </div>
        </div>

        <div class="mt-4 space-y-2">
          <div
            v-for="(group, index) in groupOrder"
            :key="group.name"
            class="flex items-center justify-between rounded-xl border border-gray-200/70 bg-gray-50/60 px-4 py-3 text-sm transition dark:border-dark-700/70 dark:bg-dark-900/40"
            :class="{
              'border-primary-400/70 bg-primary-50/60 dark:border-primary-400/40 dark:bg-primary-900/20': dragOverIndex === index,
              'opacity-60': draggingGroupIndex === index
            }"
            draggable="true"
            @dragstart="onGroupDragStart(index)"
            @dragover.prevent="onGroupDragOver(index)"
            @drop.prevent="onGroupDrop(index)"
            @dragend="onGroupDragEnd"
          >
            <div class="flex items-center gap-3">
              <div class="flex h-8 w-8 items-center justify-center rounded-lg bg-white text-gray-400 shadow-sm dark:bg-dark-800">
                <Icon name="sort" size="sm" />
              </div>
              <div>
                <div class="font-medium text-gray-900 dark:text-white">
                  {{ group.label }}
                </div>
                <div class="text-xs text-gray-500 dark:text-dark-400">
                  {{ group.count }} {{ t('admin.plans.groupSorting.items') }}
                </div>
              </div>
            </div>
            <div class="text-xs font-medium text-gray-500 dark:text-dark-400">
              #{{ index + 1 }}
            </div>
          </div>
        </div>
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
                <div class="relative">
                  <button
                    class="btn btn-icon btn-sm btn-ghost"
                    :aria-expanded="openMenuId === plan.id"
                    @click.stop="toggleMenu(plan.id)"
                  >
                    <Icon name="more" class="h-5 w-5" />
                  </button>
                  <div
                    v-if="openMenuId === plan.id"
                    class="absolute right-0 z-10 mt-2 w-44 rounded-xl border border-gray-200 bg-white p-2 shadow-xl dark:border-dark-700 dark:bg-dark-800"
                  >
                    <button
                      class="flex w-full items-center gap-2 rounded-lg px-3 py-2 text-left text-sm hover:bg-gray-100 dark:hover:bg-dark-700"
                      @click="openEditModal(plan); closeMenu()"
                    >
                      <Icon name="edit" class="h-4 w-4" />
                      {{ t('common.edit') }}
                    </button>
                    <button
                      class="flex w-full items-center gap-2 rounded-lg px-3 py-2 text-left text-sm text-red-600 hover:bg-red-50 dark:text-red-400 dark:hover:bg-red-900/20"
                      @click="confirmDelete(plan); closeMenu()"
                    >
                      <Icon name="trash" class="h-4 w-4" />
                      {{ t('common.delete') }}
                    </button>
                  </div>
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
              <div class="flex justify-between">
                <span class="text-gray-500 dark:text-gray-400">{{ t('admin.plans.dailyQuota') }}</span>
                <span class="font-medium text-gray-900 dark:text-white">${{ plan.daily_quota }}</span>
              </div>
              <div class="flex justify-between">
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
        <form @submit.prevent="savePlan" class="space-y-6">
          <!-- Basic Info -->
          <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
            <div class="md:col-span-2">
              <label class="label mb-2 block" for="plan-title">{{ t('admin.plans.form.title') }}</label>
              <input id="plan-title" v-model="form.title" type="text" class="input w-full" required />
            </div>
            
            <div class="md:col-span-2">
              <label class="label mb-2 block" for="plan-description">{{ t('admin.plans.form.description') }}</label>
              <textarea id="plan-description" v-model="form.description" class="input w-full" rows="3"></textarea>
            </div>

            <div>
              <label class="label mb-2 block" for="plan-price">{{ t('admin.plans.form.price') }}</label>
              <div class="relative">
                <span class="absolute left-3 top-2 text-gray-500">¥</span>
                <input id="plan-price" v-model.number="form.price" type="number" step="0.01" min="0" class="input w-full pl-7" required />
              </div>
            </div>

            <div>
              <label class="label mb-2 block" for="plan-sort-order">{{ t('admin.plans.form.sortOrder') }}</label>
              <input id="plan-sort-order" v-model.number="form.sort_order" type="number" class="input w-full" required />
            </div>
          </div>

          <!-- Group Info -->
          <div class="grid grid-cols-1 gap-4 rounded-xl border border-gray-200/70 bg-gray-50/60 p-4 md:grid-cols-2 dark:border-dark-700/70 dark:bg-dark-900/40">
            <div class="md:col-span-2 text-xs font-semibold uppercase tracking-wide text-gray-500 dark:text-dark-300">
              {{ t('admin.plans.form.groupSettings') }}
            </div>
            <div>
              <label class="label mb-2 block" for="plan-group-name">{{ t('admin.plans.form.groupName') }}</label>
              <input id="plan-group-name" v-model="form.group_name" type="text" class="input w-full" required />
            </div>
            <div>
              <label class="label mb-2 block" for="plan-group-sort">{{ t('admin.plans.form.groupSort') }}</label>
              <input id="plan-group-sort" v-model.number="form.group_sort" type="number" class="input w-full" required />
            </div>
          </div>

          <!-- Quota Info -->
          <div class="grid grid-cols-1 gap-4 rounded-xl border border-gray-200/70 bg-gray-50/60 p-4 md:grid-cols-2 dark:border-dark-700/70 dark:bg-dark-900/40">
            <div class="md:col-span-2 text-xs font-semibold uppercase tracking-wide text-gray-500 dark:text-dark-300">
              {{ t('admin.plans.form.quotaSettings') }}
            </div>
            <div>
              <label class="label mb-2 block" for="plan-daily-quota">{{ t('admin.plans.form.dailyQuota') }}</label>
              <div class="relative">
                <span class="absolute left-3 top-2 text-gray-500">$</span>
                <input id="plan-daily-quota" v-model.number="form.daily_quota" type="number" step="0.01" min="0" class="input w-full pl-7" required />
              </div>
            </div>
            <div>
              <label class="label mb-2 block" for="plan-total-quota">{{ t('admin.plans.form.totalQuota') }}</label>
              <div class="relative">
                <span class="absolute left-3 top-2 text-gray-500">$</span>
                <input id="plan-total-quota" v-model.number="form.total_quota" type="number" step="0.01" min="0" class="input w-full pl-7" required />
              </div>
            </div>
          </div>

          <!-- QR Upload -->
          <div class="rounded-xl border border-gray-200/70 bg-gray-50/60 p-4 dark:border-dark-700/70 dark:bg-dark-900/40">
            <label class="label mb-2 block" for="plan-purchase-qr">{{ t('admin.plans.form.purchaseQr') }}</label>
            <div class="flex flex-col gap-4 sm:flex-row sm:items-center">
              <div v-if="form.purchase_qr_url" class="relative h-24 w-24 overflow-hidden rounded-lg border border-gray-200 dark:border-dark-700">
                <img
                  :src="form.purchase_qr_url"
                  class="h-full w-full cursor-zoom-in object-contain"
                  @click="openQrPreview(form.purchase_qr_url)"
                />
                <button 
                  type="button" 
                  @click="form.purchase_qr_url = ''"
                  class="absolute right-0 top-0 bg-red-500 p-1 text-white hover:bg-red-600"
                >
                  <Icon name="x" class="h-3 w-3" />
                </button>
              </div>
              <div class="flex-1">
                <input
                  id="plan-purchase-qr"
                  ref="purchaseQrInputRef"
                  type="file"
                  accept="image/*"
                  class="sr-only"
                  @change="handleUpload"
                />
                <button type="button" class="btn btn-secondary" @click="triggerPurchaseQrUpload">
                  <Icon name="upload" class="mr-2 h-4 w-4" />
                  {{ t('common.upload') }}
                </button>
                <div v-if="uploading" class="mt-2 text-sm text-gray-500">
                  {{ t('common.uploading') }}...
                </div>
              </div>
            </div>
          </div>

          <!-- Enabled Toggle -->
          <div class="flex items-center justify-between rounded-xl border border-gray-200/70 bg-gray-50/60 px-4 py-3 dark:border-dark-700/70 dark:bg-dark-900/40">
             <span class="text-sm font-medium text-gray-700 dark:text-dark-200">{{ t('admin.plans.form.enabled') }}</span>
             <Toggle v-model="form.enabled" />
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

      <Teleport to="body">
        <div
          v-if="showQrPreview"
          class="fixed inset-0 z-[60] flex items-center justify-center bg-black/70 p-4"
          @click="closeQrPreview"
        >
          <div class="relative max-h-full max-w-3xl" @click.stop>
            <img :src="qrPreviewUrl" class="max-h-[85vh] w-auto rounded-lg bg-white object-contain" />
            <button
              type="button"
              class="absolute right-2 top-2 rounded-full bg-black/70 p-2 text-white hover:bg-black"
              aria-label="Close"
              @click="closeQrPreview"
            >
              <Icon name="x" class="h-4 w-4" />
            </button>
          </div>
        </div>
      </Teleport>

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
import { ref, computed, onMounted, onUnmounted, reactive } from 'vue'
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
const showQrPreview = ref(false)
const qrPreviewUrl = ref('')
const selectedPlan = ref<Plan | null>(null)
const purchaseQrInputRef = ref<HTMLInputElement | null>(null)
const openMenuId = ref<number | null>(null)
const savingGroupSort = ref(false)
const draggingGroupIndex = ref<number | null>(null)
const dragOverIndex = ref<number | null>(null)

type PlanGroupItem = {
  name: string
  label: string
  sort: number
  count: number
}

const groupOrder = ref<PlanGroupItem[]>([])

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

function buildGroupOrder(items: Plan[]) {
  const groups = new Map<string, PlanGroupItem>()
  items.forEach((plan) => {
    const name = plan.group_name || 'default'
    const label = name === 'default' ? t('plans.defaultGroup') : name
    if (!groups.has(name)) {
      groups.set(name, {
        name,
        label,
        sort: plan.group_sort || 0,
        count: 0
      })
    }
    const group = groups.get(name)
    if (!group) return
    group.count += 1
    if ((plan.group_sort || 0) < group.sort) {
      group.sort = plan.group_sort || 0
    }
  })

  groupOrder.value = Array.from(groups.values()).sort((a, b) => {
    if (a.sort !== b.sort) return a.sort - b.sort
    return a.label.localeCompare(b.label)
  })
}

async function fetchPlans() {
  loading.value = true
  try {
    plans.value = await adminAPI.plans.getPlans()
    buildGroupOrder(plans.value)
  } catch (error: any) {
    appStore.showError(error.message)
  } finally {
    loading.value = false
  }
}

function onGroupDragStart(index: number) {
  draggingGroupIndex.value = index
}

function onGroupDragOver(index: number) {
  dragOverIndex.value = index
}

function onGroupDragEnd() {
  draggingGroupIndex.value = null
  dragOverIndex.value = null
}

function moveGroup(from: number, to: number) {
  if (from === to) return
  const next = [...groupOrder.value]
  const [item] = next.splice(from, 1)
  if (!item) return
  next.splice(to, 0, item)
  groupOrder.value = next
}

async function onGroupDrop(index: number) {
  if (draggingGroupIndex.value === null) return
  moveGroup(draggingGroupIndex.value, index)
  onGroupDragEnd()
  await saveGroupSorts()
}

async function saveGroupSorts() {
  if (savingGroupSort.value) return
  if (groupOrder.value.length === 0) return

  savingGroupSort.value = true
  const updates = groupOrder.value.map((group, index) => ({
    group_name: group.name,
    group_sort: index
  }))

  try {
    await adminAPI.plans.updateGroupSorts({ groups: updates })
    const sortMap = new Map(updates.map((item) => [item.group_name, item.group_sort]))
    plans.value = plans.value.map((plan) => ({
      ...plan,
      group_sort: sortMap.get(plan.group_name || 'default') ?? plan.group_sort
    }))
    groupOrder.value = groupOrder.value.map((group, index) => ({
      ...group,
      sort: index
    }))
    appStore.showSuccess(t('admin.plans.groupSorting.saved'))
  } catch (error: any) {
    appStore.showError(error.message)
    fetchPlans()
  } finally {
    savingGroupSort.value = false
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

function triggerPurchaseQrUpload() {
  purchaseQrInputRef.value?.click()
}

function openQrPreview(url: string) {
  if (!url) return
  qrPreviewUrl.value = url
  showQrPreview.value = true
}

function closeQrPreview() {
  showQrPreview.value = false
  qrPreviewUrl.value = ''
}

async function savePlan() {
  saving.value = true
  try {
    if (isEditing.value && selectedPlan.value) {
      await adminAPI.plans.updatePlan(selectedPlan.value.id, form)
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

function toggleMenu(planId: number) {
  openMenuId.value = openMenuId.value === planId ? null : planId
}

function closeMenu() {
  openMenuId.value = null
}

onMounted(() => {
  fetchPlans()
  document.addEventListener('click', closeMenu)
})

onUnmounted(() => {
  document.removeEventListener('click', closeMenu)
})
</script>
