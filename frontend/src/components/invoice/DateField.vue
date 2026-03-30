<script setup lang="ts">
import { computed, ref, useId, watch } from 'vue'
import { isValidISODate, normalizeISODateOrNull } from '@/utils/dates'

const props = withDefaults(
    defineProps<{
        modelValue?: string | null
        id?: string
        name?: string
        placeholder?: string
        error?: string | null
        forceShowError?: boolean
        min?: string
        max?: string
        disabled?: boolean
    }>(),
    {
        modelValue: null,
        placeholder: 'YYYY-MM-DD',
        error: null,
        forceShowError: false,
        min: undefined,
        max: undefined,
        disabled: false,
    },
)

const emit = defineEmits<{
    (e: 'update:modelValue', value: string | null): void
}>()

const autoId = useId()
const inputId = computed(() => (props.id?.trim() ? props.id : `date_${autoId}`))
const errId = computed(() => `${inputId.value}_err`)

const isTouched = ref(false)
const showError = computed(() => (isTouched.value || props.forceShowError) && !!props.error)

const valueProxy = computed<string>({
    get() {
        return normalizeISODateOrNull(props.modelValue) ?? ''
    },
    set(value: string) {
        emit('update:modelValue', normalizeISODateOrNull(value))
    },
})

watch(
    () => props.modelValue,
    (next) => {
        if (!next) isTouched.value = false
    },
)

const minDate = computed(() => (isValidISODate(props.min) ? props.min : undefined))
const maxDate = computed(() => (isValidISODate(props.max) ? props.max : undefined))

function onBlur() {
    isTouched.value = true
}
</script>

<template>
    <div class="min-w-0">
        <input
            :id="inputId"
            :name="name"
            v-model="valueProxy"
            type="date"
            class="input w-full min-w-0 px-3 py-1"
            :class="showError ? 'input-error' : 'input-accent'"
            :placeholder="placeholder"
            pattern="\d{4}-\d{2}-\d{2}"
            inputmode="numeric"
            :min="minDate"
            :max="maxDate"
            :disabled="disabled"
            :aria-invalid="showError ? 'true' : 'false'"
            :aria-describedby="showError ? errId : undefined"
            @blur="onBlur"
        />
        <p
            :id="errId"
            class="mt-1 min-h-5 text-xs"
            :class="showError ? 'text-rose-600 dark:text-rose-300' : 'text-transparent'"
        >
            {{ showError ? error : '•' }}
        </p>
    </div>
</template>
