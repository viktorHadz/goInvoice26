<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { onClickOutside } from '@vueuse/core'
import { ChevronUpDownIcon, PlusIcon } from '@heroicons/vue/24/outline'

import TheInput from '@/components/UI/TheInput.vue'
import TheButton from '@/components/UI/TheButton.vue'

import { useProductStore } from '@/stores/products'
import { useInvoiceDraftStore } from '@/stores/invoiceDraft'
import type { Product, ProductType } from '@/utils/productHttpHandler'
import { fmtGBPMinor, toMinor } from '@/utils/money'

const prod = useProductStore()
const inv = useInvoiceDraftStore()

const itemType = ref<ProductType>('style')
const q = ref('')
const open = ref(false)

const panelRef = ref<HTMLElement | null>(null)
onClickOutside(panelRef, () => (open.value = false))

const form = reactive({
  qty: 1,
  minutes: 60, // used for hourly samples
})

const list = computed<Product[]>(() => prod.byType[itemType.value] ?? [])
const filteredItems = computed(() => {
  const s = q.value.trim().toLowerCase()
  if (!s) return list.value
  return list.value.filter((p) => p.productName.toLowerCase().includes(s))
})

function toggleItemType() {
  itemType.value = itemType.value === 'style' ? 'sample' : 'style'
  q.value = ''
  open.value = false
}

function priceLabel(p: Product) {
  if (p.pricingMode === 'hourly') return `${fmtGBPMinor(p.hourlyRateMinor ?? 0)}/hr`
  return fmtGBPMinor(p.flatPriceMinor ?? 0)
}

function addItem(p: Product) {
  const qty = Number(form.qty) > 0 ? Number(form.qty) : 1

  // style => always flat
  if (p.productType === 'style') {
    inv.addLine({
      productId: p.id,
      name: p.productName,
      lineType: 'style',
      pricingMode: 'flat',
      quantity: qty,
      unitPriceMinor: p.flatPriceMinor ?? 0,
      minutesWorked: null,
    })
    open.value = false
    return
  }

  // sample hourly
  if (p.pricingMode === 'hourly') {
    const minutes = p.minutesWorked ?? (Number(form.minutes) || 60)

    inv.addLine({
      productId: p.id,
      name: p.productName,
      lineType: 'sample',
      pricingMode: 'hourly',
      quantity: qty,
      unitPriceMinor: p.hourlyRateMinor ?? 0,
      minutesWorked: minutes,
    })
    open.value = false
    return
  }

  // sample flat
  inv.addLine({
    productId: p.id,
    name: p.productName,
    lineType: 'sample',
    pricingMode: 'flat',
    quantity: qty,
    unitPriceMinor: p.flatPriceMinor ?? 0,
    minutesWorked: null,
  })
  open.value = false
}
</script>

<template>
  <div
    ref="panelRef"
    class="relative col-span-8 grid grid-cols-subgrid items-end gap-4 py-4"
  >
    <!-- Toggle Item Type -->
    <div class="relative col-span-4">
      <div class="mb-2 flex items-center justify-between">
        <div class="text-sm font-semibold text-zinc-700 dark:text-zinc-200">
          Add item
          <span class="ml-2 text-xs font-medium text-zinc-500 dark:text-zinc-400">
            ({{ itemType }})
          </span>
        </div>

        <div
          class="relative inline-flex h-8 w-44 rounded-xl border border-zinc-200 bg-white p-0.5 text-sm dark:border-zinc-800 dark:bg-zinc-950/40"
        >
          <span
            class="absolute top-0 left-0 h-full w-1/2 rounded-lg bg-sky-50 transition-transform dark:bg-emerald-500/10"
            :class="itemType === 'style' ? 'translate-x-0' : 'translate-x-full'"
          />
          <button
            class="relative z-10 w-1/2 rounded-lg px-2 py-1"
            :class="
              itemType === 'style'
                ? 'text-sky-700 dark:text-emerald-400'
                : 'text-zinc-600 hover:text-zinc-900 dark:text-zinc-300 dark:hover:text-zinc-100'
            "
            @click="itemType = 'style'"
          >
            style
          </button>
          <button
            class="relative z-10 w-1/2 rounded-lg px-2 py-1"
            :class="
              itemType === 'sample'
                ? 'text-sky-700 dark:text-emerald-400'
                : 'text-zinc-600 hover:text-zinc-900 dark:text-zinc-300 dark:hover:text-zinc-100'
            "
            @click="itemType = 'sample'"
          >
            sample
          </button>
        </div>
      </div>

      <!-- Search input -->
      <div class="relative">
        <input
          class="input h-10.5 w-full px-3 pr-10 text-base"
          v-model="q"
          :placeholder="`Search ${itemType}s…`"
          @focus="open = true"
          @input="open = true"
        />
        <button
          class="absolute top-1/2 right-1 -translate-y-1/2 rounded-lg px-2 py-1 text-zinc-400 hover:text-sky-600 dark:hover:text-emerald-400"
          @click="open = !open"
          type="button"
        >
          <ChevronUpDownIcon class="size-5" />
        </button>
      </div>

      <!-- Dropdown -->
      <transition
        enter-active-class="transition duration-150 origin-top ease-out"
        enter-from-class="opacity-0 scale-y-90"
        enter-to-class="opacity-100 scale-y-100"
        leave-active-class="transition duration-100 ease-in"
        leave-from-class="opacity-100 scale-y-100"
        leave-to-class="opacity-0 scale-y-90"
      >
        <div
          v-if="open && filteredItems.length"
          class="absolute z-50 mt-2 w-full overflow-hidden rounded-xl border border-zinc-200 bg-white shadow-lg dark:border-zinc-800 dark:bg-zinc-900"
        >
          <div class="max-h-72 overflow-auto">
            <div
              v-for="p in filteredItems"
              :key="p.id"
              class="flex items-center justify-between gap-3 px-3 py-2.5 hover:bg-zinc-50 dark:hover:bg-zinc-800/50"
            >
              <div class="min-w-0">
                <div class="truncate text-sm font-semibold text-zinc-900 dark:text-zinc-100">
                  {{ p.productName }}
                </div>
                <div class="text-sm text-zinc-500 dark:text-zinc-400">
                  {{ p.pricingMode }} · {{ priceLabel(p) }}
                </div>
              </div>

              <TheButton
                class="shrink-0"
                @click="addItem(p)"
              >
                <PlusIcon class="size-4" />
                Add
              </TheButton>
            </div>
          </div>
        </div>
      </transition>
    </div>

    <!-- Numerics -->
    <div class="col-span-1">
      <div class="mb-1 text-sm font-medium text-zinc-700 dark:text-zinc-300">Qty</div>
      <TheInput
        v-model="form.qty"
        type="number"
        placeholder="1"
      />
    </div>

    <div class="col-span-1">
      <div class="mb-1 text-sm font-medium text-zinc-700 dark:text-zinc-300">Minutes</div>
      <TheInput
        v-model="form.minutes"
        type="number"
        :disabled="itemType === 'style'"
        :placeholder="itemType === 'style' ? '—' : '60'"
        :title="itemType === 'style' ? 'Styles do not use minutes' : 'Used for hourly samples'"
      />
    </div>

    <div class="col-span-2 flex justify-end">
      <TheButton
        class="h-10.5"
        @click="
          inv.addLine({
            productId: null,
            name: 'Custom item',
            lineType: 'custom',
            pricingMode: 'flat',
            quantity: 1,
            unitPriceMinor: toMinor(0),
            minutesWorked: null,
          })
        "
      >
        <PlusIcon class="size-4" />
        Custom
      </TheButton>
    </div>
  </div>
</template>
