<script setup>
import { NewspaperIcon } from '@heroicons/vue/24/outline'
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
  <div>
    <div class="text-xs/tight">
      <p>Invoice <br />Book</p>
    </div>
    <button class="cursor-pointer" @click="((mode = 'dark'), console.log(mode))">
      <NewspaperIcon class="hover:text-acc size-8 stroke-1 transition-colors"></NewspaperIcon>
    </button>
  </div>
</template>

<style scoped></style>
