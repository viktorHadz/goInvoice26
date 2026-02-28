<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useStorage } from '@vueuse/core'
import { CalendarIcon } from '@heroicons/vue/24/outline'

import { VueDatePicker } from '@vuepic/vue-datepicker'
import '@vuepic/vue-datepicker/dist/main.css'

const emit = defineEmits<{
  (e: 'update-date', v: string): void
}>()

const date = ref<Date>(new Date())
// TODO: Big one!
// For backend processing
//   function toISODate(d: Date) {
//   const y = d.getFullYear()
//   const m = String(d.getMonth() + 1).padStart(2, '0')
//   const day = String(d.getDate()).padStart(2, '0')
//   return `${y}-${m}-${day}` // 2026-03-01
// }

// emit('update-date', toISODate(v))

// For client side display
const format = (d: Date) => {
  const day = d.getDate()
  const month = d.getMonth() + 1
  const year = d.getFullYear()
  return `${day}/${month}/${year}`
}

function handleModelValue(v: Date | null) {
  if (!v) return
  date.value = v
  emit('update-date', format(v))
}

// always emit initial
watch(date, (v) => emit('update-date', format(v)), { immediate: true })

// Color theme
const mode = useStorage('vueuse-color-scheme', 'light')
const datePickerMode = computed(() => mode.value === 'dark')
</script>

<template>
  <VueDatePicker
    v-model="date"
    :format="format"
    :time-config="{ enableTimePicker: false }"
    :dark="datePickerMode"
    :auto-apply="true"
    :teleport="true"
    @update:model-value="handleModelValue"
  >
    <template #trigger>
      <div class="input flex cursor-pointer items-center gap-2 px-3">
        <CalendarIcon class="size-5 text-zinc-500 dark:text-zinc-400" />
        <p class="max-w-36 truncate text-sm font-medium text-zinc-900 dark:text-zinc-100">
          {{ format(date) }}
        </p>
      </div>
    </template>
  </VueDatePicker>
</template>
