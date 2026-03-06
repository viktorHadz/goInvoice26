<script setup lang="ts">
import { computed } from 'vue'
import { useStorage } from '@vueuse/core'
import { CalendarIcon, XMarkIcon } from '@heroicons/vue/24/outline'
import { VueDatePicker } from '@vuepic/vue-datepicker'
import '@vuepic/vue-datepicker/dist/main.css'

const props = withDefaults(
    defineProps<{
        modelValue?: string | null
        placeholder?: string
    }>(),
    {
        modelValue: null,
        placeholder: 'Select date',
    },
)

const emit = defineEmits<{
    (e: 'update:modelValue', v: string | null): void
}>()

function toISODate(d: Date) {
    const y = d.getFullYear()
    const m = String(d.getMonth() + 1).padStart(2, '0')
    const day = String(d.getDate()).padStart(2, '0')
    return `${y}-${m}-${day}`
}

function fromISODate(v: string | null | undefined) {
    if (!v) return null

    const [y, m, d] = v.split('-').map(Number)
    if (!y || !m || !d) return null

    return new Date(y, m - 1, d)
}

function formatDisplay(d: Date) {
    const day = String(d.getDate()).padStart(2, '0')
    const month = String(d.getMonth() + 1).padStart(2, '0')
    const year = d.getFullYear()
    return `${day}/${month}/${year}`
}

const pickerValue = computed({
    get: () => fromISODate(props.modelValue),
    set: (v: Date | null) => emit('update:modelValue', v ? toISODate(v) : null),
})

const clear = () => {
    pickerValue.value = null
}

const mode = useStorage('vueuse-color-scheme', 'light')
const isDark = computed(() => mode.value === 'dark')
</script>

<template>
    <VueDatePicker
        v-model="pickerValue"
        :dark="isDark"
        :auto-apply="false"
        :teleport="true"
        :format="formatDisplay"
        :time-config="{ enableTimePicker: false }"
    >
        <template #trigger>
            <div class="input input-accent flex items-center justify-between px-3">
                <div class="flex gap-2">
                    <CalendarIcon class="size-5 text-zinc-500 dark:text-zinc-400" />
                    <p
                        class="max-w-36 truncate text-sm font-medium text-zinc-900 dark:text-zinc-100"
                    >
                        {{ pickerValue ? formatDisplay(pickerValue) : placeholder }}
                    </p>
                </div>
                <XMarkIcon
                    v-if="pickerValue"
                    @click.stop="clear"
                    class="size-4 cursor-pointer text-zinc-500 hover:text-sky-600 dark:text-zinc-400 dark:hover:text-emerald-400"
                />
            </div>
        </template>
    </VueDatePicker>
</template>
