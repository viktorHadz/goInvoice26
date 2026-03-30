<script setup lang="ts">
import TheTooltip from '@/components/UI/TheTooltip.vue'
import { InformationCircleIcon } from '@heroicons/vue/24/outline'
</script>

<template>
    <TheTooltip
        side="top"
        align="center"
        max-width-class="max-w-[340px]"
    >
        <template #content>
            <div class="space-y-2 text-left leading-relaxed">
                <div>
                    <span class="font-semibold text-sky-600 dark:text-emerald-400">Draft</span>
                    : before sending to client. Editable and deletable. No revisions on edit.
                </div>

                <div>
                    <span class="font-semibold text-sky-600 dark:text-emerald-400">Issued</span>
                    : after sending to client. Edits create revisions.
                </div>

                <div>
                    <span class="font-semibold text-sky-600 dark:text-emerald-400">Paid</span>
                    : fully paid. Locked from editing.
                </div>

                <div>
                    <span class="font-semibold text-sky-600 dark:text-emerald-400">Void</span>
                    : cancelled. Kept for record. No changes allowed.
                </div>
            </div>
        </template>

        <button
            type="button"
            class="inline-flex cursor-help text-zinc-700 transition hover:text-sky-600 dark:text-zinc-400 dark:hover:text-emerald-400"
            aria-label="Invoice status help"
        >
            <InformationCircleIcon class="size-5 cursor-help" />
        </button>
    </TheTooltip>
</template>
