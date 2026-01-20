<template>
  <AppLayout>
    <div class="mx-auto max-w-2xl space-y-6">
      <div class="card">
        <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
            {{ t('invites.admin.settingsTitle') }}
          </h2>
          <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
            {{ t('invites.admin.settingsDesc') }}
          </p>
        </div>
        <div class="p-6">
          <form @submit.prevent="saveSettings" class="space-y-6">
            <div>
              <label class="input-label">{{ t('invites.admin.rewardAmount') }}</label>
              <input
                v-model.number="settings.reward_amount"
                type="number"
                min="0"
                step="0.01"
                required
                class="input"
                :disabled="loading"
              />
              <p class="input-hint">{{ t('invites.admin.rewardAmountHint') }}</p>
            </div>

            <div class="flex justify-end">
              <button type="submit" class="btn btn-primary" :disabled="loading">
                <span v-if="loading">{{ t('common.saving') }}</span>
                <span v-else>{{ t('common.save') }}</span>
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { AppLayout } from '@/components/layout'
import { invitesAdminAPI } from '@/api/admin/invites'
import { useAppStore } from '@/stores/app'

const { t } = useI18n()
const appStore = useAppStore()

const loading = ref(false)
const settings = ref({ reward_amount: 0 })

const loadSettings = async () => {
  loading.value = true
  try {
    const data = await invitesAdminAPI.getInviteSettings()
    settings.value = data
  } catch (error) {
    console.error('Failed to load invite settings:', error)
    appStore.showError(t('common.error'))
  } finally {
    loading.value = false
  }
}

const saveSettings = async () => {
  loading.value = true
  try {
    await invitesAdminAPI.updateInviteSettings(settings.value)
    appStore.showSuccess(t('common.success'))
  } catch (error) {
    console.error('Failed to save invite settings:', error)
    appStore.showError(t('common.error'))
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadSettings()
})
</script>