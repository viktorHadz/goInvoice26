<script setup lang="ts">
withDefaults(
  defineProps<{
    variant?: 'gradientOnly' | 'gridOnly' | 'gradientAndGrid' | 'bottomGlow'
  }>(),
  {
    variant: 'gradientOnly',
  },
)
</script>
<template>
  <div
    v-if="variant === 'gradientOnly'"
    class="pointer-events-none absolute inset-0"
  >
    <div
      class="absolute inset-0 bg-[radial-gradient(900px_circle_at_15%_0%,rgba(56,189,248,0.10),transparent_55%)] dark:bg-[radial-gradient(900px_circle_at_15%_0%,rgba(16,185,129,0.18),transparent_55%)]"
    />
  </div>
  <div
    v-if="variant === 'gridOnly'"
    class="pointer-events-none absolute inset-0"
  >
    <div
      class="absolute inset-0 bg-[linear-gradient(to_right,rgba(56,189,248)_1px,transparent_1px),linear-gradient(to_bottom,rgba(56,189,248)_1px,transparent_1px)] bg-size-[36px_36px] opacity-[0.1] dark:bg-[linear-gradient(to_right,rgba(255,255,255,0.06)_1px,transparent_1px),linear-gradient(to_bottom,rgba(255,255,255,0.06)_1px,transparent_1px)] dark:opacity-[0.2]"
    />
  </div>
  <div
    v-if="variant === 'gradientAndGrid'"
    class="pointer-events-none absolute inset-0"
  >
    <div
      class="absolute inset-0 bg-[radial-gradient(900px_circle_at_15%_0%,rgba(56,189,248,0.2),transparent_55%)] dark:bg-[radial-gradient(900px_circle_at_15%_0%,rgba(16,185,129,0.18),transparent_55%)]"
    />
    <div
      class="absolute inset-0 bg-[linear-gradient(to_right,rgba(255,255,255,0.06)_1px,transparent_1px),linear-gradient(to_bottom,rgba(255,255,255,0.06)_1px,transparent_1px)] bg-size-[36px_36px] opacity-[0.55] dark:opacity-[0.35]"
    />
  </div>
  <div
    v-if="variant === 'bottomGlow'"
    class="pointer-events-none absolute inset-0"
  >
    <div
      class="absolute inset-0 bg-[radial-gradient(900px_circle_at_15%_0%,rgba(56,189,248,0.05),transparent_30%)] dark:bg-[radial-gradient(900px_circle_at_15%_0%,rgba(16,185,129,0.03),transparent_60%)]"
    />

    <div
      class="absolute inset-0 bg-[radial-gradient(900px_circle_at_85%_85%,rgba(2,132,199,0.2),transparent_60%)] dark:bg-[radial-gradient(900px_circle_at_95%_100%,rgba(52,211,153,0.02),transparent_95%)]"
    />
  </div>
</template>
