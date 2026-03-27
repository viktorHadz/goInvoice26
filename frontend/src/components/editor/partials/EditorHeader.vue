<script setup lang="ts">
import DateField from '@/components/invoice/DateField.vue'
import TheDropdown from '@/components/UI/TheDropdown.vue'
import { useEditorStore } from '@/stores/editor'
import { computed } from 'vue'
import type { InvoiceStatus } from '@/components/invoice/invoiceTypes'
import { buildInvoiceStatusContext, reachableStatuses } from '@/utils/invoiceStatusOptions'
import { ChevronUpDownIcon } from '@heroicons/vue/24/outline'
import InvoiceStatusTooltip from './InvoiceStatusTooltip.vue'

const editStore = useEditorStore()

const inv = computed(() => editStore.draftInvoice)

const lifecycleStatus = computed(
  () => (editStore.activeInvoice?.status ?? inv.value?.status ?? 'draft') as InvoiceStatus,
)

const revisionCount = computed(() => {
  const baseNumber = editStore.activeInvoice?.baseNumber ?? inv.value?.baseNumber
  if (!baseNumber) return 1
  return editStore.invoiceBook.find((entry) => entry.baseNo === baseNumber)?.revisions.length ?? 1
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
    <div class="grid gap-4 lg:grid-cols-[minmax(0,1fr)_minmax(320px,420px)] lg:items-start">
      <!-- Left: invoice details -->
      <section class="min-w-0">
        <div class="grid gap-3 sm:grid-cols-2 xl:grid-cols-3">
          <div class="min-w-0">
            <div class="text-xs font-medium tracking-wide text-zinc-600 dark:text-zinc-400">
              Issue date
            </div>

            <div class="mt-1.5">
              <DateField
                v-model="issueDate"
                placeholder="Select issue date"
                :error="editStore.getFieldError('issueDate')"
                :forceShowError="editStore.showAllValidation"
              />
            </div>
          </div>

          <div class="min-w-0">
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

          <div class="col-span-2 min-w-0">
            <div
              class="mb-1.5 flex items-center gap-1 text-xs font-medium tracking-wide text-zinc-600 dark:text-zinc-400"
            >
              <span>Status</span>
              <InvoiceStatusTooltip />
            </div>
            <TheDropdown
              v-model="selectedInvoiceStatus"
              :right-icon="ChevronUpDownIcon"
              :options="statusSelectOptions"
              input-class="py-1.5 capitalize"
              placeholder="Status"
              :disabled="statusSelectOptions.length <= 1"
            />
          </div>
        </div>
      </section>

      <!-- Right: client card -->
      <section
        class="min-w-0 rounded-2xl border border-zinc-200 bg-zinc-50/40 p-3 dark:border-zinc-800 dark:bg-zinc-900/40"
      >
        <div class="mb-2 flex items-start justify-between gap-3">
          <div>
            <div class="text-base font-semibold text-zinc-800 dark:text-zinc-100">To</div>
          </div>

          <span
            class="inline-flex rounded-full border border-zinc-200 bg-zinc-50 px-2 py-0.5 text-xs font-medium text-zinc-600 dark:border-zinc-700 dark:bg-zinc-900/60 dark:text-zinc-400"
          >
            Client details
          </span>
        </div>

        <div class="space-y-2 text-sm">
          <div class="grid grid-cols-[88px_minmax(0,1fr)] items-start gap-3">
            <div class="text-zinc-500 dark:text-zinc-400">Name</div>
            <div class="min-w-0 font-medium text-zinc-800 dark:text-zinc-100">
              {{ inv.clientSnapshot.name || '—' }}
            </div>
          </div>

          <div
            v-if="inv.clientSnapshot.companyName"
            class="grid grid-cols-[88px_minmax(0,1fr)] items-start gap-3"
          >
            <div class="text-zinc-500 dark:text-zinc-400">Company</div>
            <div class="min-w-0 font-medium text-zinc-800 dark:text-zinc-100">
              {{ inv.clientSnapshot.companyName }}
            </div>
          </div>

          <div
            v-if="inv.clientSnapshot.address"
            class="grid grid-cols-[88px_minmax(0,1fr)] items-start gap-3"
          >
            <div class="text-zinc-500 dark:text-zinc-400">Address</div>
            <div class="min-w-0 font-medium text-zinc-800 dark:text-zinc-100">
              {{ inv.clientSnapshot.address }}
            </div>
          </div>

          <div
            v-if="inv.clientSnapshot.email"
            class="grid grid-cols-[88px_minmax(0,1fr)] items-start gap-3"
          >
            <div class="text-zinc-500 dark:text-zinc-400">Email</div>
            <div class="min-w-0 font-medium wrap-break-word text-zinc-800 dark:text-zinc-100">
              {{ inv.clientSnapshot.email }}
            </div>
          </div>
        </div>
      </section>
    </div>
  </div>
</template>
