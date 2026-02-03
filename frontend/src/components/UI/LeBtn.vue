<script setup>
defineProps({
  buttonText: String,
  type: String,
  disabled: Boolean,
})
</script>
<template>
  <button
    :disabled="disabled"
    class="ring-brdr hover:text-acc text-fg hover:ring-acc cursor-pointer truncate rounded-md bg-white p-1.5 text-sm shadow-xs ring-1 transition-transform ring-inset dark:bg-white/10 dark:hover:bg-white/20"
    :class="{ 'cursor-not-allowed opacity-70': disabled }"
  >
    {{ buttonText }}
    <slot></slot>
  </button>
</template>
