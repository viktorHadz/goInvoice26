<script setup lang="ts">
import { computed } from 'vue'
import { DocumentArrowDownIcon, DocumentTextIcon, PrinterIcon } from '@heroicons/vue/24/outline'
import TheButton from '@/components/UI/TheButton.vue'
import TheInput from '@/components/UI/TheInput.vue'
import DatePick from '@/components/invoice/DatePick.vue'
import { useClientStore } from '@/stores/clients'
import { useInvoiceDraftStore } from '@/stores/invoiceDraft'
import DecorGradient from '../UI/DecorGradient.vue'

const clients = useClientStore()
const inv = useInvoiceDraftStore()

const client = computed(() => clients.selectedClient)
const draft = computed(() => inv.draft)

const emailProxy = computed<string>({
  get() {
    return inv.draft?.clientSnapshot.email ?? ''
  },
  set(v) {
    if (!inv.draft) return
    inv.draft.clientSnapshot.email = String(v ?? '')
  },
})
</script>

<template>
  <header>
    <div class="mb-4 flex items-center gap-3">
      <div
        class="grid size-12 shrink-0 place-items-center rounded-2xl border border-zinc-200 bg-white shadow-sm dark:border-zinc-700 dark:bg-zinc-900"
      >
        <DocumentTextIcon class="size-7 text-sky-700 dark:text-emerald-400" />
      </div>

      <div class="min-w-0">
        <div class="text-2xl font-semibold tracking-tight text-zinc-900 dark:text-zinc-200">
          Invoice
        </div>
        <div class="text-sm text-zinc-500 dark:text-zinc-400">Create and export invoices</div>
      </div>
    </div>

    <section
      class="relative overflow-hidden rounded-2xl border border-zinc-200 bg-white shadow-sm dark:border-zinc-800 dark:bg-zinc-950/30"
    >
      <div class="relative z-10 space-y-4 p-3 md:p-4">
        <div class="grid grid-cols-1 gap-4 lg:grid-cols-2 lg:items-start">
          <div class="min-w-0">
            <!-- removed useless flex wrapper -->
            <div class="mb-2 text-base font-semibold text-zinc-800 dark:text-zinc-100">
              Invoice details
            </div>

            <div class="grid grid-cols-1 items-start gap-3 sm:grid-cols-2">
              <div>
                <div class="mb-1 text-xs font-medium text-zinc-700 dark:text-zinc-300">
                  Issue date
                </div>
                <DatePick @update-date="(v) => inv.draft && (inv.draft.issueDate = v)" />
              </div>

              <div>
                <div class="mb-1 text-xs font-medium text-zinc-700 dark:text-zinc-300">Due by</div>
                <DatePick @update-date="(v) => inv.draft && (inv.draft.dueByDate = v)" />
              </div>
            </div>
          </div>

          <div
            class="min-w-0 rounded-2xl border border-zinc-200 bg-zinc-50/40 p-3 dark:border-zinc-800 dark:bg-zinc-900/40"
          >
            <div class="mb-2 flex items-center justify-between">
              <div class="text-sm font-semibold text-zinc-800 dark:text-zinc-100">To</div>
              <div class="text-xs text-zinc-500 dark:text-zinc-400">client snapshot</div>
            </div>

            <div class="space-y-2 text-sm">
              <div class="grid grid-cols-[84px_1fr] items-start gap-2">
                <div class="text-zinc-500 dark:text-zinc-400">Name</div>
                <div class="truncate font-medium text-zinc-900 dark:text-zinc-100">
                  {{ inv.draft?.clientSnapshot.name || client?.name || '—' }}
                </div>
              </div>

              <div class="grid grid-cols-[84px_1fr] items-start gap-2">
                <div class="text-zinc-500 dark:text-zinc-400">Company</div>
                <div class="truncate font-medium text-zinc-900 dark:text-zinc-100">
                  {{ inv.draft?.clientSnapshot.companyName || client?.companyName || '—' }}
                </div>
              </div>

              <div class="grid grid-cols-[84px_1fr] items-start gap-2">
                <div class="text-zinc-500 dark:text-zinc-400">Address</div>
                <div class="line-clamp-2 font-medium text-zinc-900 dark:text-zinc-100">
                  {{ inv.draft?.clientSnapshot.address || client?.address || '—' }}
                </div>
              </div>
            </div>

            <div class="mt-3">
              <div class="mb-1 text-xs font-medium text-zinc-700 dark:text-zinc-300">
                Email (optional)
              </div>
              <TheInput
                :modelValue="emailProxy"
                @update:modelValue="(v) => (emailProxy = String(v ?? ''))"
                :disabled="!inv.draft"
                placeholder="client@email.com"
              />
            </div>
          </div>
        </div>
      </div>
      <DecorGradient />
    </section>
  </header>
</template>
