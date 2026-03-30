<script setup lang="ts">
import { ref } from 'vue'
import { UsersIcon } from '@heroicons/vue/24/outline'
import TheTooltip from '@/components/UI/TheTooltip.vue'
import QuickMenuSidebar from '@/components/UI/QuickMenuSidebar.vue'
import TeamSection from '@/components/team/TeamSection.vue'
import { useEscape } from '@/composables/keyHandlers'

withDefaults(
  defineProps<{
    showTrigger?: boolean
  }>(),
  {
    showTrigger: true,
  },
)

const open = ref(false)

function openMenu() {
  open.value = true
}

defineExpose({
  openMenu,
})

useEscape(() => (open.value = false))
</script>

<template>
  <TheTooltip
    v-if="showTrigger"
    side="bottom"
    align="end"
  >
    <template #content>Team</template>
    <button
      type="button"
      class="flex cursor-pointer rounded-lg border border-zinc-300 p-1 text-zinc-600 hover:text-sky-600 dark:border-transparent dark:text-zinc-300 dark:hover:bg-zinc-800 dark:hover:text-emerald-400"
      @click="openMenu"
    >
      <UsersIcon class="size-6 stroke-1" />
    </button>
  </TheTooltip>

  <QuickMenuSidebar
    :open="open"
    title="Team"
    description="Manage workspace access and invites."
    panel-class="w-[94vw] max-w-4xl"
    :icon="UsersIcon"
    @close="open = false"
  >
    <TeamSection />
  </QuickMenuSidebar>
</template>
