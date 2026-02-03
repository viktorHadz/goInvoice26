<script setup>
import { SunIcon, MoonIcon } from '@heroicons/vue/24/outline'
import { useColorMode } from '@vueuse/core'

const mode = useColorMode({
  attribute: 'data-theme',
  modes: {
    light: 'light',
    dark: 'dark',
  },
  disableTransition: false,
})
</script>
<template>
  <div v-if="mode === 'light'">
    <button
      title="light mode"
      class="flex cursor-pointer flex-col items-center"
      @click="((mode = 'dark'), console.log(mode))"
    >
      <SunIcon class="size-8 stroke-1 transition-all duration-75 hover:text-[#F07B07]"></SunIcon>
    </button>
  </div>
  <div v-if="mode === 'dark'">
    <button
      title="dark mode"
      class="flex cursor-pointer flex-col items-center align-text-top"
      @click="((mode = 'light'), console.log(mode))"
    >
      <MoonIcon class="size-8 stroke-1 hover:text-[#EFCA08]"></MoonIcon>
    </button>
  </div>
</template>

<style scoped></style>
