<script setup lang="ts">
import { nextTick, onBeforeUnmount, ref, watch } from 'vue'

type Side = 'top' | 'bottom' | 'left' | 'right'
type Align = 'start' | 'center' | 'end'
type Trigger = 'hover' | 'click' | 'both'

const props = withDefaults(
  defineProps<{
    text?: string
    side?: Side
    align?: Align
    offset?: number
    disabled?: boolean
    maxWidthClass?: string
    trigger?: Trigger
  }>(),
  {
    side: 'top',
    align: 'center',
    offset: 10,
    disabled: false,
    maxWidthClass: 'max-w-[280px]',
    trigger: 'hover',
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

function toggle() {
  if (props.disabled) return
  open.value ? hide() : show()
}

function onClickOutside(e: MouseEvent) {
  if (!triggerEl.value?.contains(e.target as Node)) hide()
}

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape') hide()
}

watch(open, (v) => {
  if (v) {
    window.addEventListener('scroll', scheduleCalc, true)
    window.addEventListener('resize', scheduleCalc)
    window.addEventListener('keydown', onKeydown)
    if (props.trigger === 'click' || props.trigger === 'both') {
      window.addEventListener('click', onClickOutside)
    }
  } else {
    window.removeEventListener('scroll', scheduleCalc, true)
    window.removeEventListener('resize', scheduleCalc)
    window.removeEventListener('keydown', onKeydown)
    window.removeEventListener('click', onClickOutside)
  }
})

onBeforeUnmount(() => {
  window.removeEventListener('scroll', scheduleCalc, true)
  window.removeEventListener('resize', scheduleCalc)
  window.removeEventListener('keydown', onKeydown)
  window.removeEventListener('click', onClickOutside)
  cancelAnimationFrame(raf)
})
</script>

<template>
  <span
    ref="triggerEl"
    class="relative inline-flex"
    @mouseenter="trigger !== 'click' ? show() : undefined"
    @mouseleave="trigger !== 'click' ? hide() : undefined"
    @click="trigger !== 'hover' ? toggle() : undefined"
    @focusin="trigger !== 'click' ? show() : undefined"
    @focusout="trigger !== 'click' ? hide() : undefined"
  >
    <slot
      :show="show"
      :hide="hide"
      :toggle="toggle"
      :open="open"
    />
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
          v-if="open && !disabled"
          ref="tipEl"
          class="pointer-events-none fixed z-110"
          :style="{ top: coords.top + 'px', left: coords.left + 'px' }"
          role="tooltip"
        >
          <div
            class="rounded-xl border border-zinc-200 bg-white/95 px-2.5 py-2.5 text-center text-xs font-medium text-zinc-700 shadow-sm dark:border-zinc-700 dark:bg-zinc-950/90 dark:text-zinc-200"
            :class="maxWidthClass"
          >
            <slot name="content">{{ text }}</slot>
          </div>
        </div>
      </Transition>
    </Teleport>
  </span>
</template>
