<script setup lang="ts">
const navLinks = [
  { name: 'home', to: '/' },
  { name: 'clients', to: '/clients' },
  { name: 'invoice', to: '/invoice' },
  { name: 'editor', to: '/editor' },
]
</script>

<template>
  <nav class="absolute top-3 left-3 z-50 sm:top-4 sm:left-4">
    <div
      class="flex items-center gap-1 rounded-2xl border border-zinc-200 bg-white p-3 shadow-lg dark:border-zinc-800 dark:bg-zinc-900"
    >
      <RouterLink
        v-for="link in navLinks"
        :key="link.name"
        :to="link.to"
        v-slot="{ isActive }"
      >
        <span
          :class="[
            'rounded-xl px-3 py-1.5 text-sm font-medium transition-colors',
            isActive
              ? 'bg-zinc-100 text-sky-600 dark:bg-zinc-800 dark:text-emerald-400'
              : 'text-zinc-600 hover:bg-zinc-50 hover:text-zinc-900 dark:text-zinc-300 dark:hover:bg-zinc-800/60 dark:hover:text-zinc-100',
          ]"
        >
          {{ link.name }}
        </span>
      </RouterLink>
    </div>
  </nav>
</template>
