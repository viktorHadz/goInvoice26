<script setup lang="ts">
import { SunIcon, MoonIcon } from '@heroicons/vue/24/outline'
import TheTooltip from './TheTooltip.vue'
import { useTheme } from '@/composables/theme'

const { mode } = useTheme()
</script>
<template>
    <TheTooltip
        side="bottom"
        class=""
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
        <div v-if="mode === 'light'">
            <button
                type="button"
                class="flex cursor-pointer rounded-lg border border-zinc-300 p-1 hover:text-sky-600"
                @click="mode = 'dark'"
            >
                <SunIcon class="size-6 stroke-1"></SunIcon>
            </button>
        </div>
        <div v-if="mode === 'dark'">
            <button
                type="button"
                class="flex cursor-pointer rounded-lg p-1 dark:hover:bg-zinc-800 dark:hover:text-emerald-400"
                @click="((mode = 'light'), console.log(mode))"
            >
                <MoonIcon class="size-6 stroke-1"></MoonIcon>
            </button>
        </div>
    </TheTooltip>
</template>
