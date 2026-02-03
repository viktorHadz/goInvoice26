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
    class="hover:text-acc cursor-pointer text-neutral-900 decoration-2 underline-offset-2 hover:underline"
    :class="{ 'cursor-not-allowed opacity-70': disabled }"
  >
    {{ buttonText }}
    <slot></slot>
  </button>
</template>
