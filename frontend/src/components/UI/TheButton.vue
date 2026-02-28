<script setup lang="ts">
import { computed } from 'vue'

type Variant = 'primary' | 'secondary' | 'success' | 'danger'

const props = withDefaults(
  defineProps<{
    variant?: Variant
    type?: 'button' | 'submit' | 'reset'
    disabled?: boolean
  }>(),
  {
    variant: 'primary',
    type: 'button',
    disabled: false,
  },
)

const base = [
  'inline-flex items-center justify-center gap-2 rounded-lg border px-3 py-2',
  'text-sm font-medium select-none',
  'transition-[transform,background-color,border-color,color,box-shadow] duration-150 ease-out',
  // keyboard accessibility
  'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-sky-500/30 focus-visible:ring-offset-2 focus-visible:ring-offset-white',
  'dark:focus-visible:ring-emerald-400/25 dark:focus-visible:ring-offset-zinc-950',
  // click feedback
  'active:translate-y-px active:scale-[0.99]',
  // don't allow weird double-tap highlight on mobile
  'touch-manipulation',
].join(' ')

const disabledCls = 'opacity-60 cursor-not-allowed pointer-events-none'

const variants: Record<Variant, string> = {
  primary: [
    'border-zinc-200 bg-white text-zinc-800',
    'hover:border-sky-300 hover:text-sky-700 hover:bg-sky-50',
    // pressed feel
    'active:bg-sky-100 active:border-sky-300',
    'dark:border-zinc-800 dark:bg-zinc-900/60 dark:text-zinc-100',
    'dark:hover:border-emerald-400/20 dark:hover:text-emerald-200 dark:hover:bg-emerald-950/40',
    'dark:active:bg-emerald-950/55 dark:active:border-emerald-400/30',
  ].join(' '),

  secondary: [
    'border-zinc-300 bg-white text-zinc-800',
    'hover:border-zinc-400 hover:bg-zinc-50',
    'active:bg-zinc-100 active:border-zinc-400',
    'dark:border-zinc-800 dark:bg-zinc-900/50 dark:text-zinc-100',
    'dark:hover:border-zinc-600 dark:hover:bg-zinc-800/20',
    'dark:active:bg-zinc-800/30 dark:active:border-zinc-500/70',
  ].join(' '),

  success: [
    'border-emerald-200 bg-emerald-50 text-emerald-700',
    'hover:bg-emerald-100 hover:border-emerald-300',
    'active:bg-emerald-200/60 active:border-emerald-300',
    'dark:border-emerald-400/20 dark:bg-emerald-950/25 dark:text-emerald-200',
    'dark:hover:bg-emerald-950/40 dark:hover:border-emerald-400/35',
    'dark:active:bg-emerald-950/55 dark:active:border-emerald-400/45',
  ].join(' '),

  danger: [
    'border-red-200 bg-red-50 text-red-700',
    'hover:bg-red-100 hover:border-red-300',
    'active:bg-red-200/60 active:border-red-300',
    'dark:border-red-400/20 dark:bg-red-950/25 dark:text-red-200',
    'dark:hover:bg-red-950/40 dark:hover:border-red-400/35',
    'dark:active:bg-red-950/55 dark:active:border-red-400/45',
  ].join(' '),
}

const cls = computed(() => [base, variants[props.variant], props.disabled ? disabledCls : ''])
</script>

<template>
  <button
    :type="type"
    :class="cls"
    :disabled="disabled"
  >
    <slot />
  </button>
</template>
