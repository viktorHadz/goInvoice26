<script setup lang="ts">
import {
  Listbox,
  ListboxButton,
  ListboxLabel,
  ListboxOption,
  ListboxOptions,
} from '@headlessui/vue'
import { ChevronUpDownIcon, CheckIcon, UserIcon } from '@heroicons/vue/24/outline'
import { useClientStore } from '@/stores/clients'

const clientStore = useClientStore()
const props = defineProps({
  selectTitle: String,
  selectTitleClass: String,
})
</script>

<template>
  <Listbox as="div" v-model="clientStore.selectedClient">
    <div class="relative min-w-30">
      <ListboxLabel for="top-bar-select-1" class="block pl-1 font-medium tracking-tight"
        :class="[`${props.selectTitleClass}`]">
        {{ props.selectTitle }}
      </ListboxLabel>
      <ListboxButton :id="'top-bar-select-1'" class="input flex items-center justify-between"
        @click="clientStore.openToggle">
        <UserIcon class="fixed mt-0.5 -ml-0.5 size-4"></UserIcon>
        <span class="col-start-1 row-start-1 ml-5 truncate">{{
          clientStore.selectedClient?.name || 'Select a client to continue'
        }}</span>
        <ChevronUpDownIcon class="col-start-1 row-start-1 size-5 self-center justify-self-end sm:size-4"
          aria-hidden="true" />
      </ListboxButton>

      <transition leave-active-class="transition ease-in duration-100" leave-from-class="opacity-100"
        leave-to-class="opacity-0">
        <ListboxOptions
          class="bg-primary absolute z-[100] mt-1 max-h-26 w-full overflow-auto rounded-md py-1 text-base shadow-lg ring-1 ring-neutral-400 focus:outline-hidden sm:text-sm dark:bg-neutral-800 dark:ring-neutral-600">
          <ListboxOption as="template" v-for="client in clientStore.clients" :key="client.id" :value="client"
            v-slot="{ active, selected }" @click="clientStore.openToggle">
            <li :class="[
              active
                ? 'text-fg dark:text-acc bg-sec/30 hover:text-fg dark:hover:text-acc outline-hidden dark:bg-zinc-800'
                : 'text-neutral-700 dark:text-neutral-400',
              'relative block cursor-default truncate select-none',

              selected
                ? 'text-fg dark:text-acc hover:text-fg dark:hover:text-acc font-semibold'
                : '',
            ]">
              <div class="flex items-center justify-between px-2 py-1">
                {{ client.name }}
                <CheckIcon v-if="selected" class="size-4 p-0.5"></CheckIcon>
              </div>
            </li>
          </ListboxOption>
        </ListboxOptions>
      </transition>
    </div>
  </Listbox>
</template>
