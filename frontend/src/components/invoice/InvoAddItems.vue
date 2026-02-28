<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { ChevronUpDownIcon, PlusIcon, SquaresPlusIcon } from '@heroicons/vue/24/outline'
import TheButton from '@/components/UI/TheButton.vue'
import TheInput from '@/components/UI/TheInput.vue'
import { useProductStore } from '@/stores/products'
import { useInvoiceDraftStore } from '@/stores/invoiceDraft'
import { fmtGBPMinor, toMinor } from '@/utils/money'
import type { Product, ProductType } from '@/utils/productHttpHandler'

const prod = useProductStore()
const inv = useInvoiceDraftStore()

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

const list = computed<Product[]>(() => prod.byType[itemType.value] ?? [])
const filtered = computed(() => {
  const s = q.value.trim().toLowerCase()
  if (!s) return list.value
  return list.value.filter((p) => p.productName.toLowerCase().includes(s))
})

function priceLabel(p: Product) {
  if (p.pricingMode === 'hourly') return `${fmtGBPMinor(p.hourlyRateMinor ?? 0)}/hr`
  return fmtGBPMinor(p.flatPriceMinor ?? 0)
}

function addFromProduct(p: Product) {
  const qty = Number(form.qty) > 0 ? Number(form.qty) : 1

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
    return
  }

  if (p.pricingMode === 'hourly') {
    inv.addLine({
      productId: p.id,
      name: p.productName,
      lineType: 'sample',
      pricingMode: 'hourly',
      quantity: qty,
      unitPriceMinor: p.hourlyRateMinor ?? 0,
      minutesWorked: p.minutesWorked ?? Number(form.minutes) ?? 60,
    })
    return
  }

  inv.addLine({
    productId: p.id,
    name: p.productName,
    lineType: 'sample',
    pricingMode: 'flat',
    quantity: qty,
    unitPriceMinor: p.flatPriceMinor ?? 0,
    minutesWorked: null,
  })
}
</script>

<template>
  <div class="space-y-3">
    <!-- Header row - Toggle -->
    <div class="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
      <div class="text-sm font-semibold text-zinc-700 dark:text-zinc-200">
        Product picker
        <span class="ml-2 text-xs font-medium text-zinc-500 dark:text-zinc-400">
          ({{ itemType }})
        </span>
      </div>

      <div
        class="flex shrink-0 rounded-full border border-zinc-200 bg-white p-1 shadow-sm dark:border-zinc-800 dark:bg-zinc-900"
      >
        <button
          class="rounded-full px-3 py-1.5 text-sm font-medium transition"
          :class="
            itemType === 'style'
              ? 'bg-sky-600 text-white shadow-sm dark:bg-emerald-600'
              : 'text-zinc-600 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-100'
          "
          @click="itemType = 'style'"
        >
          Styles
        </button>

        <button
          class="rounded-full px-3 py-1.5 text-sm font-medium transition"
          :class="
            itemType === 'sample'
              ? 'bg-sky-600 text-white shadow-sm dark:bg-emerald-600'
              : 'text-zinc-600 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-100'
          "
          @click="itemType = 'sample'"
        >
          Samples
        </button>
      </div>
    </div>

    <!-- Controls row -->
    <div class="flex flex-col gap-3 xl:flex-row xl:items-center">
      <!-- Search -->
      <div class="relative min-w-0 flex-1">
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
            title="Toggle results"
          >
            <ChevronUpDownIcon class="size-5" />
          </button>
        </div>

        <transition
          enter-active-class="transition duration-150 origin-top ease-out"
          enter-from-class="opacity-0 scale-y-90"
          enter-to-class="opacity-100 scale-y-100"
          leave-active-class="transition duration-100 ease-in"
          leave-from-class="opacity-100 scale-y-100"
          leave-to-class="opacity-0 scale-y-90"
        >
          <div
            v-if="open && filtered.length"
            class="absolute z-50 mt-2 w-full overflow-hidden rounded-xl border border-zinc-200 bg-white shadow-lg dark:border-zinc-800 dark:bg-zinc-900"
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
                  @click="addFromProduct(p)"
                >
                  <PlusIcon class="size-4" />
                  Add
                </TheButton>
              </div>
            </div>
          </div>
        </transition>
      </div>

      <!-- Qty -->
      <div class="w-full sm:max-w-28 xl:w-24">
        <div class="mb-1 text-sm font-medium text-zinc-700 dark:text-zinc-300">Qty</div>
        <TheInput
          v-model="form.qty"
          type="number"
          placeholder="1"
        />
      </div>

      <!-- Minutes -->
      <div class="w-full sm:max-w-32 xl:w-28">
        <div class="mb-1 text-sm font-medium text-zinc-700 dark:text-zinc-300">Minutes</div>
        <TheInput
          v-model="form.minutes"
          type="number"
          :disabled="itemType === 'style'"
          :title="itemType === 'style' ? 'Styles do not use minutes' : 'Used for hourly samples'"
        />
      </div>

      <!-- Custom line -->
      <div class="w-full xl:w-auto xl:shrink-0">
        <TheButton
          class="w-full py-2.5 xl:w-auto"
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
          <SquaresPlusIcon class="size-5" />
          Custom item
        </TheButton>
      </div>
    </div>
  </div>
</template>
