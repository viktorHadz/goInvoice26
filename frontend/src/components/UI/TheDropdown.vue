<script setup lang="ts" generic="T extends string | number | boolean | Record<string, unknown>">
import { computed } from 'vue'
import type { Component } from 'vue'
import {
    Listbox,
    ListboxButton,
    ListboxLabel,
    ListboxOption,
    ListboxOptions,
} from '@headlessui/vue'
import { ChevronDownIcon, CheckIcon } from '@heroicons/vue/24/outline'

type Obj = Record<string, unknown>

const props = withDefaults(
    defineProps<{
        modelValue: T | null
        options: T[]

        id?: string
        name?: string
        selectTitle?: string
        selectTitleClass?: string
        inputClass?: string
        placeholder?: string

        // only used when T is an object; ignored for primitives
        labelKey?: string
        valueKey?: string

        leftIcon?: Component
        rightIcon?: Component
        disabled?: boolean
    }>(),
    {
        id: '',
        name: '',
        selectTitle: '',
        selectTitleClass: '',
        inputClass: '',
        placeholder: 'Select…',
        labelKey: 'name',
        valueKey: 'id',
        disabled: false,
    },
)

const emit = defineEmits<{
    (e: 'update:modelValue', v: T | null): void
}>()

const selected = computed({
    get: () => props.modelValue,
    set: (v) => emit('update:modelValue', v),
})

function isObject(v: unknown): v is Obj {
    return typeof v === 'object' && v !== null
}

function labelOf(v: T | null): string {
    if (v == null) return ''
    if (!isObject(v)) return String(v)

    const key = props.labelKey
    return String((v as Obj)[key] ?? '')
}

function keyOf(v: T): string | number {
    if (!isObject(v)) return String(v)

    const key = props.valueKey
    const k = (v as Obj)[key]
    return typeof k === 'string' || typeof k === 'number' ? k : JSON.stringify(k)
}
</script>

<template>
    <Listbox
        as="div"
        v-model="selected"
        :disabled="disabled"
    >
        <div class="relative">
            <ListboxLabel
                v-if="selectTitle"
                class="input-label"
                :class="[selectTitleClass]"
            >
                {{ selectTitle }}
            </ListboxLabel>

            <ListboxButton
                v-bind="props.id?.trim() ? { id: props.id } : {}"
                class="input-dropdown input-dropdown-accent flex items-center justify-between gap-2"
                :class="[inputClass, disabled ? 'pointer-events-none opacity-60' : '']"
            >
                <div class="flex w-full items-center gap-2">
                    <component
                        v-if="leftIcon"
                        :is="leftIcon"
                        class="size-4 shrink-0 text-zinc-500 dark:text-zinc-400"
                    />

                    <span class="truncate text-zinc-900 dark:text-zinc-100">
                        {{ labelOf(selected) || placeholder }}
                    </span>
                </div>

                <component
                    v-if="rightIcon"
                    :is="rightIcon"
                    class="size-4 shrink-0 text-zinc-500 dark:text-zinc-400"
                />
                <ChevronDownIcon
                    v-else
                    class="size-4 shrink-0 text-zinc-500 dark:text-zinc-400"
                    aria-hidden="true"
                />
            </ListboxButton>

            <transition
                leave-active-class="transition ease-in duration-100"
                leave-from-class="opacity-100"
                leave-to-class="opacity-0"
            >
                <ListboxOptions
                    class="absolute z-50 mt-1 max-h-60 w-full overflow-auto rounded-lg border border-zinc-200 bg-white p-1 text-sm shadow-lg focus:outline-hidden dark:border-zinc-800 dark:bg-zinc-900"
                >
                    <ListboxOption
                        as="template"
                        v-for="opt in options"
                        :key="keyOf(opt)"
                        :value="opt"
                        v-slot="{ active, selected }"
                    >
                        <li
                            class="relative cursor-pointer rounded-md px-2 py-1.5 select-none"
                            :class="[
                                // active hover state
                                active
                                    ? 'bg-sky-50 text-zinc-900 dark:bg-emerald-950/25 dark:text-zinc-100'
                                    : 'text-zinc-700 dark:text-zinc-300',

                                // selected emphasis (subtle)
                                selected ? 'font-semibold' : '',
                            ]"
                        >
                            <div class="flex items-center justify-between gap-2">
                                <span class="truncate">{{ labelOf(opt) }}</span>

                                <CheckIcon
                                    v-if="selected"
                                    class="size-4 shrink-0 text-sky-600 dark:text-emerald-300"
                                />
                            </div>
                        </li>
                    </ListboxOption>

                    <div
                        v-if="options.length === 0"
                        class="px-2 py-2 text-xs text-zinc-500 dark:text-zinc-400"
                    >
                        No options.
                    </div>
                </ListboxOptions>
            </transition>
        </div>
    </Listbox>
</template>
