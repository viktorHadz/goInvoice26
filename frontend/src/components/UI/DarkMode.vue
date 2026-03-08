<script setup lang="ts">
import { SunIcon, MoonIcon } from '@heroicons/vue/24/outline'
import TheTooltip from './TheTooltip.vue'
import { useTheme } from '@/composables/theme'

const { mode } = useTheme()
</script>
<template>
  <TheTooltip side="bottom">
    <template #content>
      <div class="hidden sm:block">
        <span class="mr-1 text-sky-600 dark:text-emerald-400">
          Switch to {{ mode === 'dark' ? 'light' : 'dark' }}
        </span>
        <div class="flex items-center text-start">
          <kbd>Ctrl</kbd>
          +
          <kbd>Shift</kbd>
          +
          <kbd>m</kbd>
        </div>
      </div>
    </template>
    <div v-if="mode === 'light'">
      <button
        type="button"
        class="flex cursor-pointer flex-col items-center"
        @click="mode = 'dark'"
      >
        <SunIcon
          class="size-8 stroke-1 transition-all duration-300 hover:text-sky-600 dark:hover:text-emerald-400"
        ></SunIcon>
      </button>
    </div>
    <div v-if="mode === 'dark'">
      <button
        type="button"
        class="flex cursor-pointer flex-col items-center align-text-top"
        @click="((mode = 'light'), console.log(mode))"
      >
        <MoonIcon class="hover:text-acc size-8 stroke-1"></MoonIcon>
      </button>
    </div>
  </TheTooltip>
</template>

<style scoped></style>
