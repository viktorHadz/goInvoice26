<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { ChevronUpDownIcon, MagnifyingGlassIcon, SquaresPlusIcon } from '@heroicons/vue/24/outline'
import { onClickOutside } from '@vueuse/core'

import TheButton from '@/components/UI/TheButton.vue'
import TheInput from '@/components/UI/TheInput.vue'

import { useProductStore } from '@/stores/products'
import { useInvoiceStore } from '@/stores/invoice'
import type { Product, ProductType } from '@/utils/productHttpHandler'

const prod = useProductStore()
const invStore = useInvoiceStore()

const itemType = ref<ProductType>('style')
const q = ref('')
const open = ref(false)

const form = reactive({
  qty: 1,
  minutes: 60,
})

watch(itemType, () => {
  q.value = ''
  open.value = false
})

const pickerRef = ref<HTMLElement | null>(null)
onClickOutside(pickerRef, () => (open.value = false))

const list = computed<Product[]>(() => prod.byType[itemType.value] ?? [])
const filtered = computed(() => {
  const s = q.value.trim().toLowerCase()
  if (!s) return list.value
  return list.value.filter((p) => (p.productName ?? '').toLowerCase().includes(s))
})

function priceLabel(p: Product) {
  if (p.pricingMode === 'hourly') return `${invStore.fmtGBPMinor(p.hourlyRateMinor ?? 0)}/hr`
  return invStore.fmtGBPMinor(p.flatPriceMinor ?? 0)
}

function safeQty(): number {
  const n = Number(form.qty)
  if (!Number.isFinite(n) || n <= 0) return 1
  return Math.floor(n)
}

function safeMinutes(defaultMinutes = 60): number {
  const n = Number(form.minutes)
  if (!Number.isFinite(n) || n <= 0) return defaultMinutes
  return Math.floor(n)
}

function addFromProduct(p: Product) {
  const qty = safeQty()

  if (p.productType === 'style') {
    invStore.addLine({
      productId: p.id,
      name: p.productName,
      lineType: 'style',
      pricingMode: 'flat',
      quantity: qty,
      unitPriceMinor: p.flatPriceMinor ?? 0,
      minutesWorked: null,
    })
    return
  }

  if (p.pricingMode === 'hourly') {
    invStore.addLine({
      productId: p.id,
      name: p.productName,
      lineType: 'sample',
      pricingMode: 'hourly',
      quantity: qty,
      unitPriceMinor: p.hourlyRateMinor ?? 0,
      minutesWorked: safeMinutes(p.minutesWorked ?? 60),
    })
    return
  }

  invStore.addLine({
    productId: p.id,
    name: p.productName,
    lineType: 'sample',
    pricingMode: 'flat',
    quantity: qty,
    unitPriceMinor: p.flatPriceMinor ?? 0,
    minutesWorked: null,
  })
}

function addCustomItem() {
  invStore.addLine({
    productId: null,
    name: 'Custom item',
    lineType: 'custom',
    pricingMode: 'flat',
    quantity: 1,
    unitPriceMinor: invStore.toMinor(0),
    minutesWorked: null,
  })
  open.value = false
}
</script>

<template>
  <div class="space-y-3">
    <!-- Header row -->
    <div class="flex flex-col font-medium sm:flex-row sm:items-center sm:justify-between">
      <div class="text-zinc-700 capitalize dark:text-zinc-200">{{ itemType }} picker</div>

      <!-- Toggle -->
      <div
        class="flex shrink-0 rounded-full border border-zinc-200 bg-white p-1 shadow-sm dark:border-zinc-700 dark:bg-zinc-900/60"
      >
        <button
          type="button"
          class="transform-gpu rounded-full px-3 py-1.5 text-xs font-medium transition will-change-transform"
          :class="
            itemType === 'style'
              ? 'bg-sky-100 text-sky-700 shadow-sm outline-none focus:outline-none focus-visible:ring-1 focus-visible:ring-sky-300 focus-visible:ring-inset dark:bg-emerald-950/60 dark:text-emerald-200 dark:focus-visible:ring-emerald-400/30'
              : 'text-zinc-600 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-100'
          "
          @click="itemType = 'style'"
        >
          Styles
        </button>

        <button
          type="button"
          class="transform-gpu rounded-full px-3 py-1.5 text-xs font-medium transition will-change-transform"
          :class="
            itemType === 'sample'
              ? 'bg-sky-100 text-sky-700 shadow-sm outline-none focus:outline-none focus-visible:ring-1 focus-visible:ring-sky-300 focus-visible:ring-inset dark:bg-emerald-950/60 dark:text-emerald-200 dark:focus-visible:ring-emerald-400/30'
              : 'text-zinc-600 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-100'
          "
          @click="itemType = 'sample'"
        >
          Samples
        </button>
      </div>
    </div>

    <!-- Product Picker  -->
    <div class="flex flex-col gap-3 md:flex-row md:items-center">
      <!-- Search -->
      <div class="relative min-w-0 flex-1">
        <div
          class="group relative text-zinc-500 hover:text-sky-600 dark:text-zinc-400 hover:dark:text-emerald-400"
        >
          <label
            :for="itemType + '-picker'"
            class="absolute -top-6 text-sm font-medium text-zinc-700 dark:text-zinc-300"
          >
            Product Menu
          </label>

          <input
            v-model="q"
            :id="itemType + '-picker'"
            class="input input-accent w-full px-10 py-1.5 text-sm group-hover:placeholder:text-sky-600 dark:group-hover:placeholder:text-emerald-400"
            placeholder="Search..."
            @focus="open = true"
            @input="open = true"
          />
          <MagnifyingGlassIcon class="pointer-events-none absolute top-2 left-2 size-4" />

          <ChevronUpDownIcon
            class="pointer-events-none absolute top-1/2 right-2 size-5 -translate-y-1/2 rounded-lg"
          />
        </div>

        <transition
          enter-active-class="transition duration-150 origin-top ease-out"
          enter-from-class="opacity-0 scale-y-50"
          enter-to-class="opacity-100 scale-y-100"
          leave-active-class="transition duration-100 origin-top ease-in"
          leave-from-class="opacity-100 scale-y-100"
          leave-to-class="opacity-0 scale-y-50"
        >
          <div
            v-if="open && filtered.length"
            class="absolute z-50 mt-2 w-full overflow-hidden rounded-xl border border-zinc-200 bg-white shadow-lg dark:border-zinc-800 dark:bg-zinc-900"
            ref="pickerRef"
          >
            <div class="max-h-72 overflow-auto">
              <div
                v-for="p in filtered"
                :key="p.id"
                class="flex items-center justify-between gap-3 px-3 py-2.5 hover:bg-zinc-50 dark:hover:bg-zinc-800/50"
              >
                <div class="min-w-0">
                  <div class="truncate text-sm font-semibold text-zinc-900 dark:text-zinc-100">
                    {{ p.productName }}
                  </div>
                  <div class="text-sm text-zinc-500 dark:text-zinc-400">
                    {{ p.productType }} · {{ priceLabel(p) }}
                  </div>
                </div>

                <TheButton
                  class="shrink-0"
                  @click.stop="addFromProduct(p)"
                >
                  <SquaresPlusIcon class="size-4" />
                  Add
                </TheButton>
              </div>
            </div>
          </div>
        </transition>
      </div>

      <!-- Qty -->
      <div class="w-full md:w-12">
        <div class="mb-1 text-sm font-medium text-zinc-700 dark:text-zinc-300">Qty</div>
        <TheInput
          v-model="form.qty"
          input-class="text-right py-1"
          type="number"
          placeholder="1"
        />
      </div>

      <!-- Minutes -->
      <div class="w-full md:w-13">
        <div class="mb-1 text-sm font-medium text-zinc-700 dark:text-zinc-300">Mins</div>
        <TheInput
          v-model="form.minutes"
          type="number"
          input-class="text-right py-1"
          :disabled="itemType === 'style'"
          :title="itemType === 'style' ? 'Styles do not use minutes' : 'Used for hourly samples'"
        />
      </div>

      <!-- Custom item -->
      <div class="w-full md:w-auto md:shrink-0">
        <TheButton
          class="w-full py-2 text-sm md:w-auto"
          @click="addCustomItem"
        >
          <SquaresPlusIcon class="size-5" />
          Add Custom Item
        </TheButton>
      </div>
    </div>
  </div>
</template>
