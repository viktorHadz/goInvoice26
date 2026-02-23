<script setup lang="ts">
const navLinks = [
  { name: 'home', to: '/' },
  { name: 'clients', to: '/clients' },
  { name: 'invoice', to: '/invoice' },
  { name: 'editor', to: '/editor' },
]
</script>
<template>
  <nav class="navigation">
    <ul
      v-for="link in navLinks"
      :key="link.name"
    >
      <li>
        <RouterLink
          :to="link.to"
          class="link"
        >
          {{ link.name }}
        </RouterLink>
      </li>
    </ul>
  </nav>
</template>
