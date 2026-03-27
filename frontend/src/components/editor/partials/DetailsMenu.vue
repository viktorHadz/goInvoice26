<script setup lang="ts">
import TheTooltip from '@/components/UI/TheTooltip.vue'
import { emitToastInfo } from '@/utils/toast'
import { Menu, MenuButton, MenuItem, MenuItems } from '@headlessui/vue'
import { DocumentArrowDownIcon, EllipsisVerticalIcon } from '@heroicons/vue/24/outline'
import type { Component } from 'vue'

export type MenuOption = {
  id: number
  name: string
  effect: () => void | Promise<void>
  icon?: Component
  disabled?: boolean
  disabledReason?: string
}
function onOptionClick(option: MenuOption) {
  if (option.disabled) {
    emitToastInfo(option.disabledReason ?? 'You cannot do that right now.')
    return
  }

  void option.effect()
}

defineProps<{
  pdfDisabled?: boolean
  options?: MenuOption[]
}>()

const emit = defineEmits<{
  pdf: []
  option: any[]
}>()
</script>

<template>
  <Menu
    as="div"
    class="relative inline-block text-left"
  >
    <MenuButton
      type="button"
      class="inline-flex h-10 items-center justify-center rounded-xl border border-zinc-300 bg-white px-2.5 font-medium text-zinc-700 transition-colors hover:bg-zinc-50 focus-visible:ring-2 focus-visible:ring-sky-500/30 focus-visible:outline-none dark:border-zinc-700 dark:bg-zinc-900 dark:text-zinc-200 dark:hover:bg-zinc-800/80 dark:focus-visible:ring-emerald-400/25"
      aria-label="More actions"
    >
      <TheTooltip text="Invoice operations">
        <EllipsisVerticalIcon
          class="size-5 text-zinc-600 dark:text-zinc-300"
          aria-hidden="true"
        />
      </TheTooltip>
    </MenuButton>

    <transition
      enter-active-class="transition ease-out duration-100"
      enter-from-class="transform opacity-0 scale-95"
      enter-to-class="transform opacity-100 scale-100"
      leave-active-class="transition ease-in duration-75"
      leave-from-class="transform opacity-100 scale-100"
      leave-to-class="transform opacity-0 scale-95"
    >
      <MenuItems
        class="absolute right-0 z-50 mt-1 w-48 origin-top-right rounded-lg border border-zinc-200 bg-white py-1 shadow-lg focus:outline-hidden dark:border-zinc-700 dark:bg-zinc-900"
      >
        <MenuItem
          v-for="option in options"
          :key="option.id"
          as="template"
          v-slot="{ active }"
        >
          <button
            type="button"
            @click="onOptionClick(option)"
            :aria-disabled="option.disabled ? 'true' : 'false'"
            class="flex w-full items-center gap-2 px-3 py-2 text-left text-zinc-700 dark:text-zinc-100"
            :class="[
              active
                ? 'cursor-pointer bg-sky-50 text-zinc-900 dark:bg-emerald-950/25 dark:text-zinc-100'
                : '',
              option.disabled ? 'cursor-not-allowed line-through opacity-50' : '',
            ]"
          >
            <component
              v-if="option.icon"
              :is="option.icon"
              class="size-5"
            />
            {{ option.name }}
          </button>
        </MenuItem>
        <MenuItem v-slot="{ active }">
          <button
            type="button"
            class="flex w-full items-center gap-2 px-3 py-2 text-left text-zinc-700 disabled:opacity-50 dark:text-zinc-100"
            :class="
              active
                ? 'cursor-pointer bg-sky-50 text-zinc-900 dark:bg-emerald-950/25 dark:text-zinc-100'
                : ''
            "
            :disabled="pdfDisabled"
            @click="emit('pdf')"
          >
            <DocumentArrowDownIcon class="size-5" />
            Generate PDF
          </button>
        </MenuItem>
        <MenuItem v-slot="{ active }">
          <button
            type="button"
            class="flex w-full items-center px-3 py-2 text-left text-zinc-700 disabled:opacity-50 dark:text-zinc-100"
            :class="
              active
                ? 'cursor-pointer bg-sky-50 text-zinc-900 dark:bg-emerald-950/25 dark:text-zinc-100'
                : ''
            "
            :disabled="pdfDisabled"
            @click="emit('pdf')"
          >
            <TheTooltip
              text="Google Docs and MS Word readable format"
              class="gap-2"
            >
              <DocumentArrowDownIcon class="size-5" />
              Generate Docx
            </TheTooltip>
          </button>
        </MenuItem>
      </MenuItems>
    </transition>
  </Menu>
</template>
