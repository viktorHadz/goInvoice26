<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, ref, watch } from 'vue'

type Side = 'top' | 'bottom' | 'left' | 'right'
type Align = 'start' | 'center' | 'end'

const props = withDefaults(
    defineProps<{
        text?: string
        side?: Side
        align?: Align
        offset?: number
        disabled?: boolean
        lines?: { id: number; text: string }[]
        icon?: any
        maxWidthClass?: string
    }>(),
    {
        side: 'top',
        align: 'center',
        offset: 10,
        disabled: false,
        icon: null,
        maxWidthClass: 'max-w-[280px]',
    },
)

const open = ref(false)
const triggerEl = ref<HTMLElement | null>(null)
const tipEl = ref<HTMLElement | null>(null)
const coords = ref({ top: 0, left: 0 })

let raf = 0

function clamp(n: number, min: number, max: number) {
    return Math.max(min, Math.min(max, n))
}

function calcPosition() {
    if (!triggerEl.value || !tipEl.value) return
    const t = triggerEl.value.getBoundingClientRect()
    const p = tipEl.value.getBoundingClientRect()

    const vw = window.innerWidth
    const vh = window.innerHeight
    const gap = props.offset

    const alignX =
        props.align === 'start'
            ? t.left
            : props.align === 'end'
              ? t.right - p.width
              : t.left + (t.width - p.width) / 2

    const alignY =
        props.align === 'start'
            ? t.top
            : props.align === 'end'
              ? t.bottom - p.height
              : t.top + (t.height - p.height) / 2

    let left = alignX
    let top = alignY

    if (props.side === 'top') {
        top = t.top - p.height - gap
        left = alignX
    } else if (props.side === 'bottom') {
        top = t.bottom + gap
        left = alignX
    } else if (props.side === 'left') {
        top = alignY
        left = t.left - p.width - gap
    } else {
        top = alignY
        left = t.right + gap
    }

    const pad = 8
    left = clamp(left, pad, vw - p.width - pad)
    top = clamp(top, pad, vh - p.height - pad)

    coords.value = { top, left }
}

function scheduleCalc() {
    cancelAnimationFrame(raf)
    raf = requestAnimationFrame(calcPosition)
}

function show() {
    if (props.disabled) return
    open.value = true
    nextTick(scheduleCalc)
}
function hide() {
    open.value = false
}

function onKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') hide()
}

watch(open, (v) => {
    if (v) {
        window.addEventListener('scroll', scheduleCalc, true)
        window.addEventListener('resize', scheduleCalc)
        window.addEventListener('keydown', onKeydown)
    } else {
        window.removeEventListener('scroll', scheduleCalc, true)
        window.removeEventListener('resize', scheduleCalc)
        window.removeEventListener('keydown', onKeydown)
    }
})

onBeforeUnmount(() => {
    window.removeEventListener('scroll', scheduleCalc, true)
    window.removeEventListener('resize', scheduleCalc)
    window.removeEventListener('keydown', onKeydown)
    cancelAnimationFrame(raf)
})

const triggerClass =
    'inline-flex items-center justify-center rounded-lg p-1 text-zinc-500 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-emerald-400'
</script>

<template>
    <span
        ref="triggerEl"
        class="inline-flex"
        @mouseenter="show"
        @mouseleave="hide"
        @focusin="show"
        @focusout="hide"
    >
        <span
            :class="[triggerClass, disabled ? 'pointer-events-none opacity-60' : '']"
            tabindex="0"
        >
            <slot name="trigger">
                <component
                    v-if="icon"
                    :is="icon"
                    class="size-5"
                    aria-hidden="true"
                />
                <span
                    v-else
                    class="text-tiny inline-flex h-5 min-w-5 items-center justify-center rounded-full border border-zinc-200 bg-white font-semibold text-zinc-600 shadow-sm dark:border-zinc-700 dark:bg-zinc-900/70 dark:text-zinc-300"
                    aria-hidden="true"
                >
                    i
                </span>
            </slot>
        </span>

        <Teleport to="body">
            <Transition
                enter-active-class="transition duration-120 ease-out"
                enter-from-class="opacity-0 translate-y-0.5 scale-[0.98]"
                enter-to-class="opacity-100 translate-y-0 scale-100"
                leave-active-class="transition duration-100 ease-in"
                leave-from-class="opacity-100 translate-y-0 scale-100"
                leave-to-class="opacity-0 translate-y-0.5 scale-[0.98]"
            >
                <div
                    v-if="open && !disabled && (text || (lines && lines.length))"
                    ref="tipEl"
                    class="pointer-events-none fixed z-101"
                    :style="{ top: coords.top + 'px', left: coords.left + 'px' }"
                    role="tooltip"
                >
                    <div
                        class="rounded-xl border border-zinc-200 bg-white/95 px-2.5 py-2.5 text-center text-xs font-medium text-zinc-700 shadow-sm dark:border-zinc-700 dark:bg-zinc-950/90 dark:text-zinc-200"
                        :class="maxWidthClass"
                    >
                        <div v-if="text">{{ text }}</div>

                        <ul
                            v-if="lines?.length"
                            class="list-inside list-disc space-y-2 text-left"
                        >
                            <li
                                v-for="line in lines"
                                :key="line.id"
                            >
                                {{ line.text }}
                            </li>
                        </ul>
                    </div>
                </div>
            </Transition>
        </Teleport>
    </span>
</template>
