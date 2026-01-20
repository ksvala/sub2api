<template>
  <AppLayout>
    <div class="mx-auto max-w-4xl space-y-6">
      <!-- Title -->
      <div class="flex items-center justify-between">
        <div>
          <h1 class="text-2xl font-bold text-gray-900 dark:text-white">
            {{ t('invites.title') }}
          </h1>
          <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
            {{ t('invites.description') }}
          </p>
        </div>
      </div>

      <!-- Summary Card -->
      <div class="grid gap-6 md:grid-cols-2">
        <!-- Invite Code & Link -->
        <div class="card bg-gradient-to-br from-indigo-500 to-purple-600 text-white">
          <div class="p-6">
            <h2 class="text-lg font-semibold text-indigo-100">{{ t('invites.myInviteCode') }}</h2>
            <div class="mt-4 flex items-center justify-between gap-4 rounded-xl bg-white/10 p-4 backdrop-blur-sm">
              <span class="font-mono text-2xl font-bold tracking-wider">{{ summary?.invite_code || '...' }}</span>
              <button
                @click="copyInviteCode"
                class="rounded-lg bg-white/20 p-2 transition-colors hover:bg-white/30"
                :title="t('common.copyToClipboard')"
              >
                <Icon name="copy" size="md" />
              </button>
            </div>
            <div class="mt-4">
              <button
                @click="copyInviteLink"
                class="flex w-full items-center justify-center gap-2 rounded-xl bg-white/20 py-3 font-medium transition-colors hover:bg-white/30"
              >
                <Icon name="link" size="md" />
                {{ t('invites.copyLink') }}
              </button>
            </div>
          </div>
        </div>

        <!-- Stats -->
        <div class="card p-6">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('common.total') }}</h2>
          <div class="mt-6 grid grid-cols-2 gap-4">
            <div class="rounded-xl bg-gray-50 p-4 dark:bg-dark-800">
              <p class="text-sm text-gray-500 dark:text-gray-400">{{ t('invites.totalInvites') }}</p>
              <p class="mt-1 text-2xl font-bold text-gray-900 dark:text-white">{{ summary?.total_invites || 0 }}</p>
            </div>
            <div class="rounded-xl bg-gray-50 p-4 dark:bg-dark-800">
              <p class="text-sm text-gray-500 dark:text-gray-400">{{ t('invites.totalReward') }}</p>
              <p class="mt-1 text-2xl font-bold text-emerald-600 dark:text-emerald-400">
                ${{ summary?.total_reward_amount?.toFixed(2) || '0.00' }}
              </p>
            </div>
            <div class="rounded-xl bg-gray-50 p-4 dark:bg-dark-800">
              <p class="text-sm text-gray-500 dark:text-gray-400">{{ t('invites.pendingInvites') }}</p>
              <p class="mt-1 text-xl font-semibold text-orange-600 dark:text-orange-400">{{ summary?.pending_invites || 0 }}</p>
            </div>
            <div class="rounded-xl bg-gray-50 p-4 dark:bg-dark-800">
              <p class="text-sm text-gray-500 dark:text-gray-400">{{ t('invites.confirmedInvites') }}</p>
              <p class="mt-1 text-xl font-semibold text-green-600 dark:text-green-400">{{ summary?.confirmed_invites || 0 }}</p>
            </div>
          </div>
        </div>
      </div>

      <!-- Tables -->
      <div class="grid gap-6 lg:grid-cols-2">
        <!-- Invite Records -->
        <div class="card flex flex-col">
          <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
            <h2 class="font-semibold text-gray-900 dark:text-white">{{ t('invites.inviteRecords') }}</h2>
          </div>
          <div class="flex-1 overflow-x-auto p-0">
            <table class="w-full text-left text-sm">
              <thead class="bg-gray-50 text-gray-500 dark:bg-dark-800 dark:text-gray-400">
                <tr>
                  <th class="px-6 py-3 font-medium">{{ t('invites.invitee') }}</th>
                  <th class="px-6 py-3 font-medium">{{ t('invites.status') }}</th>
                  <th class="px-6 py-3 font-medium">{{ t('invites.registeredAt') }}</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
                <tr v-if="inviteRecords.length === 0">
                  <td colspan="3" class="px-6 py-8 text-center text-gray-500">{{ t('common.noData') }}</td>
                </tr>
                <tr v-for="record in inviteRecords" :key="record.id" class="hover:bg-gray-50 dark:hover:bg-dark-800/50">
                  <td class="px-6 py-3 text-gray-900 dark:text-white">{{ record.invitee_email }}</td>
                  <td class="px-6 py-3">
                    <span
                      class="inline-flex items-center rounded-md px-2 py-1 text-xs font-medium"
                      :class="record.status === 'confirmed' ? 'bg-green-50 text-green-700 dark:bg-green-900/30 dark:text-green-400' : 'bg-orange-50 text-orange-700 dark:bg-orange-900/30 dark:text-orange-400'"
                    >
                      {{ record.status === 'confirmed' ? t('invites.statusConfirmed') : t('invites.statusPending') }}
                    </span>
                  </td>
                  <td class="px-6 py-3 text-gray-500">{{ formatDateTime(record.created_at) }}</td>
                </tr>
              </tbody>
            </table>
          </div>
          <div v-if="invitePagination.pages > 1" class="border-t border-gray-100 p-4 dark:border-dark-700">
            <Pagination
              :page="invitePagination.page"
              :total="invitePagination.total"
              :page-size="invitePagination.page_size"
              @update:page="(p) => loadInviteRecords(p)"
            />
          </div>
        </div>

        <!-- Reward Records -->
        <div class="card flex flex-col">
          <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
            <h2 class="font-semibold text-gray-900 dark:text-white">{{ t('invites.rewardRecords') }}</h2>
          </div>
          <div class="flex-1 overflow-x-auto p-0">
            <table class="w-full text-left text-sm">
              <thead class="bg-gray-50 text-gray-500 dark:bg-dark-800 dark:text-gray-400">
                <tr>
                  <th class="px-6 py-3 font-medium">{{ t('invites.reward') }}</th>
                  <th class="px-6 py-3 font-medium">{{ t('invites.time') }}</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
                <tr v-if="rewardRecords.length === 0">
                  <td colspan="2" class="px-6 py-8 text-center text-gray-500">{{ t('common.noData') }}</td>
                </tr>
                <tr v-for="record in rewardRecords" :key="record.id" class="hover:bg-gray-50 dark:hover:bg-dark-800/50">
                  <td class="px-6 py-3 font-medium text-emerald-600 dark:text-emerald-400">+${{ record.amount.toFixed(2) }}</td>
                  <td class="px-6 py-3 text-gray-500">{{ formatDateTime(record.created_at) }}</td>
                </tr>
              </tbody>
            </table>
          </div>
          <div v-if="rewardPagination.pages > 1" class="border-t border-gray-100 p-4 dark:border-dark-700">
            <Pagination
              :page="rewardPagination.page"
              :total="rewardPagination.total"
              :page-size="rewardPagination.page_size"
              @update:page="(p) => loadRewardRecords(p)"
            />
          </div>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { AppLayout } from '@/components/layout'
import Icon from '@/components/icons/Icon.vue'
import Pagination from '@/components/common/Pagination.vue'
import { inviteAPI } from '@/api/invites'
import { useAppStore } from '@/stores/app'
import { formatDateTime } from '@/utils/format'
import type { InviteSummary, InviteRecord, InviteRewardRecord } from '@/types'

const { t } = useI18n()
const appStore = useAppStore()

const summary = ref<InviteSummary | null>(null)
const inviteRecords = ref<InviteRecord[]>([])
const rewardRecords = ref<InviteRewardRecord[]>([])

const invitePagination = ref({ page: 1, page_size: 10, total: 0, pages: 0 })
const rewardPagination = ref({ page: 1, page_size: 10, total: 0, pages: 0 })

const loadSummary = async () => {
  try {
    summary.value = await inviteAPI.getInviteSummary()
  } catch (error) {
    console.error('Failed to load invite summary:', error)
  }
}

const loadInviteRecords = async (page = 1) => {
  try {
    const res = await inviteAPI.listInviteRecords({ page, page_size: invitePagination.value.page_size })
    inviteRecords.value = res.items
    invitePagination.value = { ...invitePagination.value, page, total: res.total, pages: res.pages }
  } catch (error) {
    console.error('Failed to load invite records:', error)
  }
}

const loadRewardRecords = async (page = 1) => {
  try {
    const res = await inviteAPI.listInviteRewards({ page, page_size: rewardPagination.value.page_size })
    rewardRecords.value = res.items
    rewardPagination.value = { ...rewardPagination.value, page, total: res.total, pages: res.pages }
  } catch (error) {
    console.error('Failed to load reward records:', error)
  }
}

const copyInviteCode = () => {
  if (summary.value?.invite_code) {
    copyToClipboard(summary.value.invite_code)
  }
}

const copyInviteLink = () => {
  if (summary.value?.invite_code) {
    const url = `${window.location.origin}/register?invite=${summary.value.invite_code}`
    copyToClipboard(url)
  }
}

const copyToClipboard = async (text: string) => {
  try {
    await navigator.clipboard.writeText(text)
    appStore.showSuccess(t('common.copiedToClipboard'))
  } catch (err) {
    appStore.showError(t('common.copyFailed'))
  }
}

onMounted(() => {
  loadSummary()
  loadInviteRecords()
  loadRewardRecords()
})
</script>