<template>
  <AppLayout>
    <TablePageLayout>
      <template #filters>
        <div class="flex flex-wrap items-center gap-4">
          <!-- Date Range Picker -->
          <DateRangePicker
            v-model:start-date="filters.start_time"
            v-model:end-date="filters.end_time"
            @change="loadLogs(1)"
          />
          <!-- Action Filter -->
          <div class="w-40">
            <Select
              v-model="filters.action"
              :options="[
                { value: '', label: t('common.all') },
                { value: 'bind', label: t('invites.actionBind') },
                { value: 'confirm', label: t('invites.actionConfirm') }
              ]"
              @change="loadLogs(1)"
            />
          </div>
          <!-- Inviter Email Search -->
          <div class="relative w-64">
            <Icon name="search" size="md" class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
            <input
              v-model="filters.inviter_email"
              type="text"
              :placeholder="t('invites.admin.inviter')"
              class="input pl-10"
              @input="handleSearch"
            />
          </div>
          <!-- Invitee Email Search -->
          <div class="relative w-64">
            <Icon name="search" size="md" class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
            <input
              v-model="filters.invitee_email"
              type="text"
              :placeholder="t('invites.admin.invitee')"
              class="input pl-10"
              @input="handleSearch"
            />
          </div>
          
          <button @click="loadLogs(1)" class="btn btn-secondary" :title="t('common.refresh')">
            <Icon name="refresh" size="md" :class="{ 'animate-spin': loading }" />
          </button>
        </div>
      </template>

      <template #table>
        <DataTable :columns="columns" :data="logs" :loading="loading">
          <template #cell-created_at="{ value }">
            <span class="text-sm text-gray-500">{{ formatDateTime(value) }}</span>
          </template>
          <template #cell-reward_amount="{ value }">
            <span class="font-medium text-emerald-600 dark:text-emerald-400">${{ value.toFixed(2) }}</span>
          </template>
          <template #empty>
            <EmptyState :title="t('common.noData')" />
          </template>
        </DataTable>
      </template>

      <template #pagination>
        <Pagination
          v-if="pagination.total > 0"
          :page="pagination.page"
          :total="pagination.total"
          :page-size="pagination.page_size"
          @update:page="loadLogs"
        />
      </template>
    </TablePageLayout>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { AppLayout } from '@/components/layout'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import Pagination from '@/components/common/Pagination.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import Select from '@/components/common/Select.vue'
import DateRangePicker from '@/components/common/DateRangePicker.vue'
import Icon from '@/components/icons/Icon.vue'
import { invitesAdminAPI } from '@/api/admin/invites'
import { formatDateTime } from '@/utils/format'
import type { InviteLog } from '@/types'
import type { Column } from '@/components/common/types'

const { t } = useI18n()

const loading = ref(false)
const logs = ref<InviteLog[]>([])
const pagination = reactive({ page: 1, page_size: 20, total: 0, pages: 0 })
const filters = reactive({
  action: '',
  inviter_email: '',
  invitee_email: '',
  start_time: '',
  end_time: ''
})

const columns: Column[] = [
  { key: 'id', label: 'ID', sortable: false },
  { key: 'action', label: t('invites.admin.action'), sortable: false },
  { key: 'inviter_email', label: t('invites.admin.inviter'), sortable: false },
  { key: 'invitee_email', label: t('invites.admin.invitee'), sortable: false },
  { key: 'reward_amount', label: t('invites.reward'), sortable: false },
  { key: 'created_at', label: t('invites.time'), sortable: false }
]

const loadLogs = async (page = 1) => {
  loading.value = true
  try {
    const res = await invitesAdminAPI.listInviteLogs({
      page,
      page_size: pagination.page_size,
      action: filters.action || undefined,
      inviter_email: filters.inviter_email || undefined,
      invitee_email: filters.invitee_email || undefined,
      start_time: filters.start_time ? new Date(filters.start_time + 'T00:00:00').toISOString() : undefined,
      end_time: filters.end_time ? new Date(filters.end_time + 'T23:59:59.999').toISOString() : undefined
    })
    logs.value = res.items
    pagination.page = page
    pagination.total = res.total
    pagination.pages = res.pages
  } catch (error) {
    console.error('Failed to load invite logs:', error)
  } finally {
    loading.value = false
  }
}

let searchTimer: ReturnType<typeof setTimeout> | null = null
const handleSearch = () => {
  if (searchTimer) clearTimeout(searchTimer)
  searchTimer = setTimeout(() => {
    loadLogs(1)
  }, 500)
}

onMounted(() => {
  loadLogs()
})
</script>