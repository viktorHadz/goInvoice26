<script setup lang="ts">
import DatePick from '@/components/invoice/DatePick.vue'
import { useEditorStore } from '@/stores/editor'
import { computed } from 'vue'

const editStore = useEditorStore()

const inv = computed(() => editStore.draftInvoice)
</script>
<template>
  <header>
    <div
      v-if="inv && editStore.activeNode"
      class="px-3 pb-4 pt-1 md:px-4"
    >
      <div
        class="grid grid-cols-1 gap-6 lg:grid-cols-2 lg:items-start lg:gap-10"
      >
        <div class="min-w-0">
          <div class="mb-2 text-xs font-medium text-zinc-500 dark:text-zinc-400">
            Dates
          </div>
          <div class="grid grid-cols-1 items-start gap-3 sm:grid-cols-2">
            <div>
              <div class="mb-1 text-xs font-medium text-zinc-600 dark:text-zinc-400">
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
              <div class="mb-1 text-xs font-medium text-zinc-600 dark:text-zinc-400">
                Due by
              </div>
              <DatePick
                v-model="inv.dueByDate"
                placeholder="Select due date"
                :error="editStore.getFieldError('dueByDate')"
                :forceShowError="editStore.showAllValidation"
              />
            </div>
          </div>
        </div>

        <div class="min-w-0">
          <div class="text-xs font-medium text-zinc-500 dark:text-zinc-400">
            Bill to
          </div>
          <div class="mt-1.5 space-y-1 text-sm text-zinc-900 dark:text-zinc-100">
            <div class="font-medium">
              {{ inv.clientSnapshot.name || '—' }}
              <span
                v-if="inv.clientSnapshot.companyName"
                class="font-normal text-zinc-600 dark:text-zinc-300"
              >
                · {{ inv.clientSnapshot.companyName }}
              </span>
            </div>
            <div
              v-if="inv.clientSnapshot.email"
              class="text-zinc-600 dark:text-zinc-300"
            >
              {{ inv.clientSnapshot.email }}
            </div>
            <div
              v-if="inv.clientSnapshot.address"
              class="line-clamp-2 text-zinc-600 dark:text-zinc-300"
            >
              {{ inv.clientSnapshot.address }}
            </div>
          </div>
        </div>
      </div>
    </div>
  </header>
</template>
