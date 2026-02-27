<script setup lang="ts">
import { computed, ref, watch, useId } from 'vue'

type InputValue = string | number | null | undefined

const props = withDefaults(
    defineProps<{
        label?: string
        id?: string
        name?: string
        modelValue?: InputValue
        labelHidden?: boolean

        /** wrapper classes */
        classNames?: string
        /** extra classes on the input itself */
        inputClass?: string

        type?: string
        error?: string | null

        // reserve one line even when there's no error - prevents layout jump
        reserveErrorSpace?: boolean
    }>(),
    {
        type: 'text',
        labelHidden: false,
        classNames: '',
        inputClass: '',
        error: null,
        reserveErrorSpace: true,
    },
)

const emit = defineEmits<{
    (e: 'update:modelValue', value: InputValue): void
    (e: 'touched', value: boolean): void
    (e: 'dirty', value: boolean): void
}>()

// id auto if not provided
const auto = useId()
const inputId = computed(() => (props.id?.trim() ? props.id : `in_${auto}`))
const errId = computed(() => `${inputId.value}_err`)

// touched
const isTouched = ref(false)
function touch() {
    if (isTouched.value) return
    isTouched.value = true
    emit('touched', true)
}

// dirty vs initial
const initialValue = ref<InputValue>(props.modelValue)
const isDirty = ref(false)

function norm(v: InputValue) {
    return v === undefined || v === null ? '' : String(v)
}

watch(
    () => props.modelValue,
    (v) => {
        const nextDirty = norm(v) !== norm(initialValue.value)
        if (nextDirty !== isDirty.value) {
            isDirty.value = nextDirty
            emit('dirty', nextDirty)
        }
    },
    { immediate: true },
)

// if parent resets modelValue mark clean and untouched
watch(
    () => props.modelValue,
    (v) => {
        if (v === null || v === undefined || v === '') {
            initialValue.value = v
            if (isDirty.value) emit('dirty', false)
            isDirty.value = false
            isTouched.value = false
        }
    },
)

// DOM model bridge
const valueProxy = computed<string>({
    get() {
        const v = props.modelValue
        return v === null || v === undefined ? '' : String(v)
    },
    set(v) {
        if (v === '') {
            emit('update:modelValue', null)
            return
        }
        if (props.type === 'number') {
            const n = Number(v)
            emit('update:modelValue', Number.isFinite(n) ? n : null)
            return
        }
        emit('update:modelValue', v)
    },
})

const showError = computed(() => isTouched.value && !!props.error)

function onFocus(e: FocusEvent) {
    touch()
    ;(e.target as HTMLInputElement | null)?.select()
}

function onBlur() {
    touch()
}
</script>

<template>
    <div
        class="flex min-w-0 flex-col gap-1"
        :class="props.classNames"
    >
        <label
            v-if="!props.labelHidden && props.label"
            :for="inputId"
            class="input-label"
        >
            {{ props.label }}
        </label>

        <input
            v-bind="$attrs"
            :id="inputId"
            :name="props.name"
            :type="props.type"
            v-model="valueProxy"
            class="input"
            :class="[props.inputClass, showError ? 'input-error' : 'input-accent']"
            :aria-invalid="showError ? 'true' : 'false'"
            :aria-describedby="showError ? errId : undefined"
            @focus="onFocus"
            @blur="onBlur"
            @change="touch"
        />

        <p
            :id="errId"
            class="text-xs"
            :class="showError ? 'text-rose-600 dark:text-rose-300' : 'text-transparent'"
            :style="props.reserveErrorSpace ? 'min-height: 1.25rem' : ''"
            v-if="props.reserveErrorSpace || showError"
        >
            {{ showError ? props.error : '•' }}
        </p>
    </div>
</template>
