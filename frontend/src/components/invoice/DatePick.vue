<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useStorage } from '@vueuse/core'
import { CalendarIcon } from '@heroicons/vue/24/outline'

import { VueDatePicker } from '@vuepic/vue-datepicker'
import '@vuepic/vue-datepicker/dist/main.css'

const date = ref(new Date())
const emit = defineEmits(['update-date'])
const format = (date) => {
  const day = date.getDate()
  const month = date.getMonth() + 1
  const year = date.getFullYear()

  return `${day}/${month}/${year}`
}
const handleDate = (modelData) => {
  console.log('Unformated date: ', modelData)
  const formatedDate = format(modelData)

  console.log('Formated date: ', formatedDate)
  emit('update-date', formatedDate)
}
watch(
  date,
  (newVal) => {
    emit('update-date', format(newVal))
  },
  { immediate: true },
)

// Color theme
const mode = useStorage('vueuse-color-scheme')
const datePickerMode = computed(() => {
  return mode.value === 'dark' ? true : false
})
</script>
<template>
  <VueDatePicker
    v-model="date"
    :format="format"
    @update:model-value="handleDate(date)"
    :enable-time-picker="false"
    :dark="datePickerMode"
  >
    <template #trigger>
      <div
        class="text-fg hover:text-acc border-fg/40 input flex cursor-pointer items-center gap-2 p-1.5"
      >
        <CalendarIcon class="size-4.5"></CalendarIcon>
        <p class="max-w-28 truncate text-sm">
          {{ format(date) }}
        </p>
      </div>
    </template>
  </VueDatePicker>
</template>
