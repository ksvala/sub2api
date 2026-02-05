<template>
  <BaseDialog :show="show" :title="t('admin.users.bulkConcurrencyTitle')" width="narrow" @close="$emit('close')">
    <form id="bulk-concurrency-form" @submit.prevent="handleSubmit" class="space-y-5">
      <div>
        <label class="input-label">{{ t('admin.users.bulkConcurrencyLabel') }}</label>
        <input v-model.number="form.concurrency" type="number" min="1" class="input" />
        <p class="input-hint">{{ t('admin.users.bulkConcurrencyHint') }}</p>
      </div>
    </form>
    <template #footer>
      <div class="flex justify-end gap-3">
        <button @click="$emit('close')" type="button" class="btn btn-secondary">{{ t('common.cancel') }}</button>
        <button type="submit" form="bulk-concurrency-form" :disabled="submitting" class="btn btn-primary">
          {{ submitting ? t('common.processing') : t('common.confirm') }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { adminAPI } from '@/api/admin'
import BaseDialog from '@/components/common/BaseDialog.vue'

const props = defineProps<{ show: boolean }>()
const emit = defineEmits(['close', 'success'])

const { t } = useI18n()
const appStore = useAppStore()

const submitting = ref(false)
const form = reactive({ concurrency: 1 })

watch(() => props.show, (val) => {
  if (val) {
    form.concurrency = 1
  }
})

const handleSubmit = async () => {
  if (!form.concurrency || form.concurrency < 1) {
    appStore.showError(t('admin.users.concurrencyMin'))
    return
  }
  submitting.value = true
  try {
    const result = await adminAPI.users.bulkUpdateConcurrency(form.concurrency)
    appStore.showSuccess(t('admin.users.bulkConcurrencySuccess', { updated: result.updated, skipped: result.skipped }))
    emit('success')
    emit('close')
  } catch (e: any) {
    appStore.showError(e.response?.data?.detail || t('admin.users.bulkConcurrencyFailed'))
  } finally {
    submitting.value = false
  }
}
</script>
