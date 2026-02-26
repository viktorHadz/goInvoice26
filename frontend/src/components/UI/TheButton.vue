<script setup lang="ts">
import { computed } from 'vue'

type Variant = 'primary' | 'secondary' | 'success' | 'danger'

const props = withDefaults(
  defineProps<{
    variant?: Variant
    type?: 'button' | 'submit' | 'reset'
  }>(),
  {
    variant: 'primary',
    type: 'button',
  },
)

const base =
  'inline-flex items-center justify-center gap-2 rounded-lg border px-3 py-2 text-sm font-medium select-none transition-colors'

const variants: Record<Variant, string> = {
  primary:
    'border-zinc-200 bg-white text-zinc-800 hover:border-sky-300 hover:text-sky-700 hover:bg-sky-50 ' +
    'dark:border-zinc-800 dark:bg-zinc-900/60 dark:text-zinc-100 dark:hover:border-emerald-400/20 dark:hover:text-emerald-200 dark:hover:bg-emerald-950/40',

  secondary:
    'border-zinc-300 bg-white text-zinc-800 hover:border-zinc-400 hover:bg-zinc-50 ' +
    'dark:border-zinc-800 dark:bg-zinc-900/50 dark:text-zinc-100 dark:hover:border-zinc-600 dark:hover:bg-zinc-800/20',

  success:
    'border-emerald-200 bg-emerald-50 text-emerald-700 hover:bg-emerald-100 hover:border-emerald-300 ' +
    'dark:border-emerald-400/20 dark:bg-emerald-950/25 dark:text-emerald-200 dark:hover:bg-emerald-950/40 dark:hover:border-emerald-400/35',

  danger:
    'border-red-200 bg-red-50 text-red-700 hover:bg-red-100 hover:border-red-300 ' +
    'dark:border-red-400/20 dark:bg-red-950/25 dark:text-red-200 dark:hover:bg-red-950/40 dark:hover:border-red-400/35',
}

const cls = computed(() => [base, variants[props.variant]])
</script>

<template>
  <button
    :type="type"
    :class="cls"
  >
    <slot />
  </button>
</template>
