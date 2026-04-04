<script setup lang="ts">
import { ref } from 'vue'
import DarkMode from '@/components/UI/DarkMode.vue'
import { ArrowRightIcon, Bars3Icon, XMarkIcon } from '@heroicons/vue/24/outline'
import logoDark from '@/assets/logo/logoDark.svg'
import logoLight from '@/assets/logo/logoLight.svg'

const isMobileMenuOpen = ref(false)

function closeMobileMenu() {
  isMobileMenuOpen.value = false
}
</script>
<template>
  <header
    class="border-b border-white/80 bg-white/85 px-4 py-4 shadow-sm backdrop-blur dark:border-white/10 dark:bg-zinc-950/75"
  >
    <div class="flex flex-col gap-4">
      <div class="flex items-center justify-between gap-4">
        <div class="flex items-center gap-4">
          <RouterLink
            to="/"
            aria-label="Invoice and Go"
            class="inline-flex items-center leading-none"
            @click="closeMobileMenu"
          >
            <img
              :src="logoLight"
              alt=""
              aria-hidden="true"
              class="block h-11 w-auto sm:h-12 dark:hidden"
            />
            <img
              :src="logoDark"
              alt=""
              aria-hidden="true"
              class="hidden h-11 w-auto sm:h-12 dark:block"
            />
          </RouterLink>

          <DarkMode variant="pill" />
        </div>

        <div class="hidden items-center gap-2 sm:flex">
          <RouterLink
            to="/login"
            class="inline-flex items-center justify-center rounded-full border border-zinc-300 bg-white px-4 py-2 text-sm font-medium text-zinc-700 transition hover:border-sky-400 hover:text-zinc-950 dark:border-zinc-700 dark:bg-zinc-900 dark:text-zinc-200 dark:hover:border-emerald-400/60 dark:hover:text-white"
          >
            Log in
          </RouterLink>

          <RouterLink
            to="/signup"
            class="inline-flex items-center justify-center gap-2 rounded-full bg-sky-600 px-6 py-2 text-sm font-semibold text-white shadow-sm transition hover:bg-sky-500 dark:bg-emerald-500 dark:text-zinc-950 dark:hover:bg-emerald-400"
          >
            Register
            <ArrowRightIcon class="size-4" />
          </RouterLink>
        </div>

        <button
          type="button"
          class="inline-flex items-center justify-center rounded-2xl border border-zinc-300 bg-white p-2 text-zinc-700 transition hover:border-sky-400 hover:text-zinc-950 sm:hidden dark:border-zinc-700 dark:bg-zinc-900 dark:text-zinc-200 dark:hover:border-emerald-400/60 dark:hover:text-white"
          :aria-expanded="isMobileMenuOpen"
          aria-controls="landing-mobile-menu"
          aria-label="Toggle navigation menu"
          @click="isMobileMenuOpen = !isMobileMenuOpen"
        >
          <Bars3Icon
            v-if="!isMobileMenuOpen"
            class="size-5"
          />
          <XMarkIcon
            v-else
            class="size-5"
          />
        </button>
      </div>

      <Transition
        enter-active-class="transition duration-200 ease-out"
        enter-from-class="opacity-0 -translate-y-2"
        enter-to-class="opacity-100 translate-y-0"
        leave-active-class="transition duration-150 ease-in"
        leave-from-class="opacity-100 translate-y-0"
        leave-to-class="opacity-0 -translate-y-2"
      >
        <div
          v-if="isMobileMenuOpen"
          id="landing-mobile-menu"
          class="border-t border-zinc-200 pt-4 sm:hidden dark:border-zinc-800"
        >
          <div class="flex flex-col gap-2">
            <RouterLink
              to="/login"
              class="inline-flex items-center justify-center rounded-full border border-zinc-300 bg-white px-4 py-3 text-sm font-medium text-zinc-700 transition hover:border-sky-400 hover:text-zinc-950 dark:border-zinc-700 dark:bg-zinc-900 dark:text-zinc-200 dark:hover:border-emerald-400/60 dark:hover:text-white"
              @click="closeMobileMenu"
            >
              Log in
            </RouterLink>

            <RouterLink
              to="/signup"
              class="inline-flex items-center justify-center gap-2 rounded-full bg-sky-600 px-4 py-3 text-sm font-semibold text-white shadow-sm transition hover:bg-sky-500 dark:bg-emerald-500 dark:text-zinc-950 dark:hover:bg-emerald-400"
              @click="closeMobileMenu"
            >
              Register
              <ArrowRightIcon class="size-4" />
            </RouterLink>
          </div>
        </div>
      </Transition>
    </div>
  </header>
</template>
