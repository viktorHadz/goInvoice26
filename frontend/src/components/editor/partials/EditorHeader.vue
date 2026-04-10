<script setup lang="ts">
import DateField from '@/components/invoice/DateField.vue'
import TheDropdown from '@/components/UI/TheDropdown.vue'
import { useEditorStore } from '@/stores/editor'
import { computed } from 'vue'
import type { InvoiceStatus } from '@/components/invoice/invoiceTypes'
import { buildInvoiceStatusContext, reachableStatuses } from '@/utils/invoiceStatusOptions'
import { ChevronUpDownIcon, InformationCircleIcon } from '@heroicons/vue/24/outline'
import InvoiceStatusTooltip from './InvoiceStatusTooltip.vue'
import TheTooltip from '@/components/UI/TheTooltip.vue'

const editStore = useEditorStore()

const inv = computed(() => editStore.draftInvoice)

const lifecycleStatus = computed(
  () => (editStore.activeInvoice?.status ?? inv.value?.status ?? 'draft') as InvoiceStatus,
)

const revisionCount = computed(() => {
  const baseNumber = editStore.activeInvoice?.baseNumber ?? inv.value?.baseNumber
  if (!baseNumber) return 1
  return editStore.invoiceBook.find((entry) => entry.baseNo === baseNumber)?.latestRevisionNo ?? 1
})

const statusSelectOptions = computed(() =>
  reachableStatuses(
    lifecycleStatus.value,
    buildInvoiceStatusContext(editStore.activeInvoice, revisionCount.value),
  ),
)

const selectedInvoiceStatus = computed({
  get(): InvoiceStatus {
    return lifecycleStatus.value
  },
  set(next: InvoiceStatus | null) {
    if (next == null || next === lifecycleStatus.value) return
    void editStore.requestInvoiceLifecycleStatusChange(next)
  },
})

const issueDate = computed<string | null>({
  get: () => inv.value?.issueDate ?? null,
  set: (v) => editStore.setIssueDate(v ?? ''),
})

const supplyDate = computed<string | null>({
  get: () => inv.value?.supplyDate ?? null,
  set: (v) => editStore.setSupplyDate(v ?? ''),
})

const dueByDate = computed<string | null>({
  get: () => inv.value?.dueByDate ?? null,
  set: (v) => editStore.setDueByDate(v ?? ''),
})
</script>

<template>
  <div
    v-if="inv && editStore.activeNode"
    class="px-3 py-4 md:px-4 md:py-5"
  >
    <div class="grid grid-cols-1 items-start gap-x-6 gap-y-2 sm:grid-cols-2">
      <!-- Left: invoice details -->
      <section>
        <div class="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
          <div>
            <div class="mb-1 text-xs font-medium text-zinc-700 dark:text-zinc-300">Issue date</div>
            <DateField
              v-model="issueDate"
              placeholder="Select issue date"
              :error="editStore.getFieldError('issueDate')"
              :forceShowError="editStore.showAllValidation"
            />
          </div>

          <div>
            <div class="text-xs font-medium tracking-wide text-zinc-600 dark:text-zinc-400">
              Due by
            </div>

            <div class="mt-1.5">
              <DateField
                v-model="dueByDate"
                placeholder="Select due date"
                :error="editStore.getFieldError('dueByDate')"
                :forceShowError="editStore.showAllValidation"
              />
            </div>
          </div>

          <div>
            <div
              class="mb-1 flex justify-between text-xs font-medium text-zinc-700 dark:text-zinc-300"
            >
              Supply date
              <TheTooltip text="Optional date. Used when the goods are issued to a client.">
                <InformationCircleIcon
                  class="relative inline-flex size-4 cursor-help hover:text-sky-600 dark:hover:text-emerald-400"
                />
              </TheTooltip>
            </div>
            <DateField
              v-model="supplyDate"
              placeholder="Only if different"
              :error="editStore.getFieldError('supplyDate')"
              :forceShowError="editStore.showAllValidation"
            />
          </div>

          <div>
            <div
              class="mb-1 flex justify-between text-xs font-medium text-zinc-700 dark:text-zinc-300"
            >
              <span>Status</span>
              <InvoiceStatusTooltip icon-size="size-4" />
            </div>
            <TheDropdown
              v-model="selectedInvoiceStatus"
              :right-icon="ChevronUpDownIcon"
              :options="statusSelectOptions"
              input-class="py-1 capitalize"
              placeholder="Status"
              :disabled="statusSelectOptions.length <= 1"
            />
          </div>
        </div>
      </section>

      <!-- Right: client card -->
      <section
        class="min-w-0 rounded-2xl border border-zinc-300 bg-white p-3 sm:mt-2 dark:border-zinc-800 dark:bg-zinc-900/40"
      >
        <div class="mb-2 flex items-center justify-between">
          <div class="font-semibold">To</div>
          <div
            class="hidden rounded-full border border-zinc-300 bg-zinc-50 px-2 py-0.5 text-xs font-medium text-zinc-600 backdrop-blur-sm sm:inline-flex dark:border-zinc-700 dark:bg-zinc-900/70 dark:text-zinc-400"
          >
            client details
          </div>
        </div>

        <div class="space-y-4 text-sm">
          <div class="grid grid-cols-[84px_1fr] items-start gap-2">
            <div class="text-zinc-500 dark:text-zinc-400">Name</div>
            <div class="truncate font-medium text-zinc-900 dark:text-zinc-100">
              {{ inv.clientSnapshot.name || '—' }}
            </div>
          </div>

          <div
            v-if="inv.clientSnapshot.companyName"
            class="grid grid-cols-[84px_1fr] items-start gap-2"
          >
            <div class="text-zinc-500 dark:text-zinc-400">Company</div>
            <div class="truncate font-medium text-zinc-900 dark:text-zinc-100">
              {{ inv.clientSnapshot.companyName }}
            </div>
          </div>

          <div
            v-if="inv.clientSnapshot.address"
            class="grid grid-cols-[84px_1fr] items-start gap-2"
          >
            <div class="text-zinc-500 dark:text-zinc-400">Address</div>
            <div class="line-clamp-2 font-medium text-zinc-900 dark:text-zinc-100">
              {{ inv.clientSnapshot.address }}
            </div>
          </div>

          <div
            v-if="inv.clientSnapshot.email"
            class="grid grid-cols-[84px_1fr] items-start gap-2"
          >
            <div class="text-zinc-500 dark:text-zinc-400">Email</div>
            <div class="line-clamp-2 font-medium text-zinc-900 dark:text-zinc-100">
              {{ inv.clientSnapshot.email }}
            </div>
          </div>
        </div>
      </section>
    </div>
  </div>
</template>
