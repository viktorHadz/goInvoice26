<script setup lang="ts">
import DatePick from '@/components/invoice/DatePick.vue'
import { useEditorStore } from '@/stores/editor'
import { useSettingsStore } from '@/stores/settings'
import { computed } from 'vue'
import { formatActiveEditorNodeLabel } from '@/utils/invoiceLabels'

const editStore = useEditorStore()
const setsStore = useSettingsStore()

const inv = computed(() => editStore.draftInvoice)

const invoicePrefix = computed(() => setsStore.settings?.invoicePrefix ?? '')

const invoiceDisplayLabel = computed(() => {
  const i = inv.value
  const node = editStore.activeNode
  if (!i || !node) return ''
  return formatActiveEditorNodeLabel(invoicePrefix.value, node)
})
</script>
<template>
  <header>
    <section class="relative overflow-hidden rounded-2xl bg-white shadow-sm dark:bg-zinc-950/30">
      <div class="relative z-10 space-y-4 p-3 md:p-4">
        <div
          class="grid grid-cols-1 gap-4 lg:grid-cols-2 lg:items-start"
          v-if="inv && editStore.activeNode"
        >
          <div class="min-w-0">
            <div class="mb-2 flex gap-x-4 font-medium text-zinc-700 dark:text-zinc-300">
              <span>Invoice number:</span>
              <span class="font-bold text-sky-600 dark:text-emerald-400">{{
                invoiceDisplayLabel
              }}</span>
            </div>
            <div class="grid grid-cols-1 items-start gap-3 sm:grid-cols-2">
              <div>
                <div class="mb-1 text-xs font-medium text-zinc-700 dark:text-zinc-300">
                  Issue date
                </div>
                <DatePick
                  v-model="inv.issueDate"
                  placeholder="Select issue date"
                  :error="editStore.getFieldError('issueDate')"
                  :forceShowError="editStore.showAllValidation"
                />
              </div>

              <div>
                <div class="mb-1 text-xs font-medium text-zinc-700 dark:text-zinc-300">Due by</div>
                <DatePick
                  v-model="inv.dueByDate"
                  placeholder="Select due date"
                  :error="editStore.getFieldError('dueByDate')"
                  :forceShowError="editStore.showAllValidation"
                />
              </div>
            </div>
          </div>

          <div
            class="min-w-0 rounded-2xl border border-zinc-200 bg-zinc-50/40 p-3 dark:border-zinc-800 dark:bg-zinc-900/40"
          >
            <div class="mb-2 flex items-center justify-between">
              <div class="font-semibold text-sky-600 dark:text-emerald-400">To</div>
              <div
                class="hidden rounded-full border border-zinc-200 bg-white/90 px-2 py-0.5 text-xs font-medium text-zinc-600 backdrop-blur-sm sm:inline-flex dark:border-zinc-700 dark:bg-zinc-900/70 dark:text-zinc-400"
              >
                client details
              </div>
            </div>

            <div class="space-y-2 text-sm">
              <div class="grid grid-cols-[84px_1fr] items-start gap-2">
                <div class="text-zinc-500 dark:text-zinc-400">Name</div>
                <div class="truncate font-medium text-zinc-900 dark:text-zinc-100">
                  {{ inv.clientSnapshot.name || '—' }}
                </div>
              </div>

              <div class="grid grid-cols-[84px_1fr] items-start gap-2">
                <div class="text-zinc-500 dark:text-zinc-400">Company</div>
                <div class="truncate font-medium text-zinc-900 dark:text-zinc-100">
                  {{ inv.clientSnapshot.companyName || '—' }}
                </div>
              </div>

              <div class="grid grid-cols-[84px_1fr] items-start gap-2">
                <div class="text-zinc-500 dark:text-zinc-400">Address</div>
                <div class="line-clamp-2 font-medium text-zinc-900 dark:text-zinc-100">
                  {{ inv.clientSnapshot.address || '—' }}
                </div>
              </div>
              <div class="grid grid-cols-[84px_1fr] items-start gap-2">
                <div class="text-zinc-500 dark:text-zinc-400">Email</div>
                <div class="line-clamp-2 font-medium text-zinc-900 dark:text-zinc-100">
                  {{ inv.clientSnapshot.email || '—' }}
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>
  </header>
</template>
