<template>
  <BaseDialog :show="isOpen" :title="t('common.customerService')" width="narrow" @close="isOpen = false">
    <div class="flex flex-col items-center space-y-6 p-6">
      <!-- Customer Service QR -->
      <div v-if="qrCodeUrl" class="flex flex-col items-center space-y-4 w-full">
        <p class="text-sm font-medium text-gray-500 dark:text-gray-400">
          {{ t('common.scanToContact') }}
        </p>
        <div class="relative group">
          <div class="absolute -inset-1 rounded-2xl bg-gradient-to-r from-primary-600 to-primary-400 opacity-20 blur transition duration-1000 group-hover:opacity-40 group-hover:duration-200"></div>
          <div 
            class="relative overflow-hidden rounded-xl border border-gray-100 bg-white p-2 shadow-xl dark:border-white/10 dark:bg-dark-800"
            :class="[afterSalesQrUrl ? 'h-56 w-56' : 'h-64 w-64']"
          >
            <img 
              :src="qrCodeUrl" 
              class="h-full w-full object-contain"
              alt="Customer Service QR"
            />
          </div>
        </div>
      </div>

      <!-- After Sales Group QR -->
      <div v-if="afterSalesQrUrl" class="flex flex-col items-center space-y-4 w-full">
        <p class="text-sm font-medium text-gray-500 dark:text-gray-400">
          {{ t('common.scanToJoin') }}
        </p>
        <div class="relative group">
          <div class="absolute -inset-1 rounded-2xl bg-gradient-to-r from-primary-600 to-primary-400 opacity-20 blur transition duration-1000 group-hover:opacity-40 group-hover:duration-200"></div>
          <div 
            class="relative overflow-hidden rounded-xl border border-gray-100 bg-white p-2 shadow-xl dark:border-white/10 dark:bg-dark-800"
            :class="[qrCodeUrl ? 'h-56 w-56' : 'h-64 w-64']"
          >
            <img 
              :src="afterSalesQrUrl" 
              class="h-full w-full object-contain"
              alt="After Sales Group QR"
            />
          </div>
        </div>
      </div>

      <!-- No QR Code Fallback -->
      <div v-if="!qrCodeUrl && !afterSalesQrUrl" class="flex flex-col items-center justify-center gap-3 py-8 text-gray-400">
        <Icon name="chat" class="h-12 w-12 opacity-20" />
        <span class="text-sm">{{ t('common.noQrCode') || 'No QR Code' }}</span>
      </div>

      <div v-if="contactInfo" class="w-full rounded-lg bg-gray-50 p-4 text-center dark:bg-white/5">
        <p class="text-sm font-medium text-gray-900 dark:text-white">{{ contactInfo }}</p>
      </div>

      <p class="text-center text-xs text-gray-400 max-w-[80%]">
        {{ t('plans.purchaseNote') }}
      </p>

      <button @click="isOpen = false" class="btn btn-primary w-full shadow-lg shadow-primary-600/20 transition-all hover:shadow-primary-600/40 hover:-translate-y-0.5">
        {{ t('common.close') }}
      </button>
    </div>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'

const props = defineProps<{
  modelValue: boolean
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
}>()

const { t } = useI18n()
const appStore = useAppStore()

const isOpen = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

const qrCodeUrl = computed(() => appStore.cachedPublicSettings?.customer_service_qr)
const afterSalesQrUrl = computed(() => appStore.cachedPublicSettings?.after_sales_group_qr)
const contactInfo = computed(() => appStore.cachedPublicSettings?.contact_info)
</script>
