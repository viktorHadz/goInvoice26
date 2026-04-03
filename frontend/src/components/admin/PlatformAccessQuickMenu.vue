<script setup lang="ts">
import { ref, watch } from 'vue'
import { KeyIcon } from '@heroicons/vue/24/outline'
import QuickMenuSidebar from '@/components/UI/QuickMenuSidebar.vue'
import PlatformAccessPanel from '@/components/team/PlatformAccessPanel.vue'
import { useEscape } from '@/composables/keyHandlers'
import { useAuthStore } from '@/stores/auth'

withDefaults(
  defineProps<{
    showTrigger?: boolean
  }>(),
  {
    showTrigger: true,
  },
)

const open = ref(false)
const authStore = useAuthStore()

function openMenu() {
  if (!authStore.canManagePlatformAccess) return
  open.value = true
}

defineExpose({
  openMenu,
})

useEscape(() => (open.value = false))

watch(
  () => authStore.canManagePlatformAccess,
  (canManage) => {
    if (!canManage) {
      open.value = false
    }
  },
)
</script>

<template>
  <QuickMenuSidebar
    :open="open"
    title="Platform access"
    description="Manage trusted email grants, promo codes, and team-tier access."
    panel-class="w-[94vw] max-w-4xl"
    :icon="KeyIcon"
    @close="open = false"
  >
    <div class="h-full overflow-y-auto p-4 sm:p-5">
      <section class="grid grid-cols-1 gap-4">
        <PlatformAccessPanel />
      </section>
    </div>
  </QuickMenuSidebar>
</template>
