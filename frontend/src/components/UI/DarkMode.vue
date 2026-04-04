<script setup lang="ts">
import { computed } from 'vue'
import { SunIcon, MoonIcon } from '@heroicons/vue/24/outline'
import TheTooltip from './TheTooltip.vue'
import { useTheme } from '@/composables/theme'

const props = withDefaults(
  defineProps<{
    variant?: 'icon' | 'pill'
  }>(),
  {
    variant: 'icon',
  },
)

const { mode } = useTheme()
const nextModeLabel = computed(() => (mode.value === 'dark' ? 'Light' : 'Dark'))
</script>
<template>
  <!-- WITH TOOLTIP icon + desktop only -->
  <TheTooltip
    v-if="props.variant !== 'pill'"
    side="bottom"
    class="hidden sm:block"
  >
    <template #content>
      <span class="text-sky-600 dark:text-emerald-400">Toggle Theme:</span>
      <br />
      <div class="mt-1">
        <kbd>Ctrl</kbd>
        +
        <kbd>Shift</kbd>
        +
        <kbd>M</kbd>
      </div>
    </template>

    <template #default>
      <!-- ICON -->
      <button
        v-if="mode === 'light'"
        type="button"
        class="flex cursor-pointer rounded-lg border border-zinc-300 p-1 hover:text-sky-600"
        @click="mode = 'dark'"
      >
        <SunIcon class="size-6 stroke-1" />
      </button>

      <button
        v-else
        type="button"
        class="flex cursor-pointer rounded-lg p-1 dark:hover:bg-zinc-800 dark:hover:text-emerald-400"
        @click="mode = 'light'"
      >
        <MoonIcon class="size-6 stroke-1" />
      </button>
    </template>
  </TheTooltip>

  <!-- NO TOOLTIP (pill OR mobile) -->
  <button
    v-else
    type="button"
    class="inline-flex items-center gap-1.5 rounded-full border border-zinc-200/80 bg-white/70 px-2.5 py-1.5 text-xs font-medium text-zinc-600 shadow-sm backdrop-blur-sm transition hover:border-sky-300 hover:text-zinc-900 focus:ring-2 focus:ring-sky-500/20 focus:outline-none dark:border-zinc-800 dark:bg-zinc-900/60 dark:text-zinc-300 dark:hover:border-emerald-400/40 dark:hover:text-white dark:focus:ring-emerald-400/20"
    @click="mode = mode === 'dark' ? 'light' : 'dark'"
  >
    <SunIcon
      v-if="mode === 'dark'"
      class="size-3.5 stroke-2"
    />
    <MoonIcon
      v-else
      class="size-3.5 stroke-2"
    />
    <span>{{ nextModeLabel }}</span>
  </button>
</template>
