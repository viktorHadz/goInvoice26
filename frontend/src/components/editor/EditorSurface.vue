<script setup lang="ts">
// NOTE: To whoever might look at this in the future. I have made the conscious decision to duplicate invoice code to avoid complexity.
import { StopCircleIcon } from '@heroicons/vue/24/outline'
import TheButton from '../UI/TheButton.vue'
import { useEditorStore } from '@/stores/editor'

const editStore = useEditorStore()
</script>
<template>
  <div v-if="editStore.activeInvoice">
    Editing invoice
    <TheButton @click="editStore.cancelEdit">
      <StopCircleIcon class="size-5" />
      Cancel Edit
    </TheButton>
  </div>
</template>
