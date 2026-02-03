<template>
  <div class="relative col-span-9 grid grid-cols-subgrid items-center pt-5">
    <Popover v-slot="{ open }" class="col-span-4">
      <PopoverButton
        :class="open ? 'text-white' : 'text-white/90'"
        class="group inline-flex items-center rounded-md px-3 py-2 text-base font-medium hover:text-white focus:outline-none focus-visible:ring-2 focus-visible:ring-white/75"
      >
        <LeBtn class="flex items-center gap-2">
          Add
          <ChevronDownIcon
            :class="open ? 'text-acc' : 'text-acc/70'"
            class="group-hover:acc/80 size-4 transition duration-150 ease-in-out"
            aria-hidden="true"
          />
        </LeBtn>
      </PopoverButton>

      <transition
        enter-active-class="transition duration-200 ease-out"
        enter-from-class="translate-y-1 opacity-0"
        enter-to-class="translate-y-0 opacity-100"
        leave-active-class="transition duration-150 ease-in"
        leave-from-class="translate-y-0 opacity-100"
        leave-to-class="translate-y-1 opacity-0"
      >
        <PopoverPanel class="absolute right-0 left-0 z-50 ml-3 max-w-md focus:outline-none">
          <div class="ring-brdr overflow-hidden rounded-lg ring-1">
            <div class="bg-primary">
              <div
                v-if="filteredItems.length > 0"
                class="max-h-60 w-full overflow-auto rounded-md text-base focus:outline-hidden sm:text-sm"
              >
                <div class="border-brdr sticky top-0 h-12 w-full border-b bg-amber-200">
                  <p class="text-2xl">Hello my friends</p>
                </div>
                <div
                  v-for="item in filteredItems"
                  :key="item.id"
                  :value="item"
                  class="my-2 grid grid-cols-[minmax(0,1fr)_auto_auto_auto] gap-2"
                >
                  <div class="">{{ item.name }}</div>
                  <LeBtn class="text-xs" title="Select to edit item">select</LeBtn>
                  <input class="input h-6.5 max-w-8" v-model="itemForm.qty" />
                  <LeBtn class="text-xs">add</LeBtn>
                </div>
              </div>
            </div>
          </div>
        </PopoverPanel>
      </transition>
    </Popover>
  </div>
</template>
