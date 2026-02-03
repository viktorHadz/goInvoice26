<script setup></script>
<template>
  <div class="relative col-span-8 grid grid-cols-subgrid items-center gap-6 py-5">
    <!-- Toggle Item Type -->
    <div class="absolute -bottom-1 w-full max-w-36">
      <span
        class="absolute top-0 left-1.5 h-4 w-2 rounded-bl-sm border-b border-l mask-t-from-0"
      ></span>
      <div class="relative ml-4">
        <span
          class="bg-acc/20 dark:bg-acc/20 absolute bottom-0 left-0 z-[1] h-5.5 w-1/2 rounded transition-transform duration-300"
          :class="itemType === 'style' ? 'translate-x-0' : 'translate-x-full'"
        ></span>
        <button
          @click="toggleItemType()"
          :class="
            itemType === 'style'
              ? 'text-acc dark:text-acc'
              : 'text-fg hover:text-acc cursor-pointer dark:hover:text-white'
          "
          class="z-10 w-1/2"
        >
          style
        </button>
        <button
          @click="toggleItemType()"
          :class="
            itemType === 'sample'
              ? 'text-acc dark:text-acc'
              : 'text-fg hover:text-acc cursor-pointer dark:hover:text-white'
          "
          class="z-10 w-1/2"
        >
          sample
        </button>
      </div>
      <span
        class="absolute top-0 -right-3 h-4 w-2 rounded-br-sm border-r border-b mask-t-from-0%"
      ></span>
    </div>
    <div class="relative col-span-4 block py-1">
      <!-- Input -->
      <input
        id="combo-item-add-1"
        type="text"
        class="input input hover:drop-shadow-acc/25 focus:drop-shadow-acc/25 text-base drop-shadow-md transition-shadow duration-75 sm:text-sm/6"
        v-model="selectedItem.name"
        @input="searchQueries[itemType] = $event.target.value"
        :placeholder="`Type to search by ${itemType} name`"
      />
      <div class="absolute top-2 right-0">
        <button
          class="relative flex cursor-pointer items-center rounded-r-md px-2 focus:outline-hidden"
          @click="(toggleDropdown(), console.log(dropDown))"
        >
          <ChevronUpDownIcon class="hover:text-acc size-6 text-gray-400" aria-hidden="true" />
        </button>
      </div>

      <!-- Dropdown -->
      <transition
        enter-active-class="transition duration-300 origin-top ease-out "
        enter-from-class=" opacity-0 scale-y-25  "
        enter-to-class="  opacity-100 scale-y-100 "
        leave-active-class="transition duration-150  ease-in"
        leave-from-class=" scale-y-100 origin-bottom opacity-100"
        leave-to-class=" scale-y-0 origin-top opacity-0"
      >
        <div
          v-if="filteredItems.length > 0 && dropDown"
          class="input-dropdown absolute z-10 mt-2 max-h-60 max-w-full min-w-full overflow-auto rounded-md py-1 text-base shadow-md focus:outline-hidden sm:text-sm"
        >
          <div v-for="(item, index) in filteredItems" :key="item.id" :value="item">
            <li
              class="relative grid w-full grid-cols-[minmax(0,1fr)_auto_auto_auto] items-center gap-2 px-2 py-2 font-normal"
            >
              <button
                class="hover:text-acc dark:hover:text-acc flex cursor-pointer justify-between"
                @click="selectItem(item)"
              >
                <div class="truncate">{{ item.name }}</div>
                <div class="truncate">{{ item.price }}</div>
              </button>
              <LeInput
                :id="`${item.id}-${index}`"
                class="placeholder:text-acc text-acc w-12"
                v-model="selectedItem.qty"
                placeholder="1"
              ></LeInput>
              <LeBtn class="mr-1">add</LeBtn>
              <div title="quick add item without selecting">
                <QuestionMarkCircleIcon
                  class="hover:text-acc relative z-20 size-5 cursor-pointer"
                ></QuestionMarkCircleIcon>
              </div>
            </li>
          </div>
        </div>
      </transition>
    </div>
    <!-- Numerics -->
    <LeInput
      class="col-span-1"
      placeholder="x1"
      v-model="selectedItem.qty"
      :id="'invo-qty-1'"
      :type="'number'"
    ></LeInput>
    <LeInput
      class="col-span-1"
      v-model="selectedItem.time"
      :disabled="itemType === 'style' ? true : false"
      :id="'invo-time-1'"
      :type="'number'"
      :placeholder="itemType === 'style' ? 'N/A' : 'min'"
      :title="itemType === 'style' ? 'No time property for styles' : 'min'"
    ></LeInput>
    <LeInput
      class="col-span-1"
      v-model="selectedItem.price"
      :id="'invo-price-1'"
      :type="'number'"
      placeholder="item price"
    ></LeInput>
    <input
      id="invo-total-1"
      class="text-acc placeholder:text-acc input hover:!ring-brdr col-span-1 cursor-default select-none"
      v-model="selectedItem.itemTotal"
      disabled="true"
      placeholder="0"
      title="Cannot edit item total price"
    />
  </div>
</template>
