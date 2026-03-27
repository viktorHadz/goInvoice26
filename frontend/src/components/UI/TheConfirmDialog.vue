<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { ExclamationTriangleIcon, XMarkIcon } from '@heroicons/vue/24/outline'
import TheButton from './TheButton.vue'
import DecorGradient from './DecorGradient.vue'
import { useEscape } from '@/composables/keyHandlers'
import { onConfirmationRequest, type ConfirmationRequest } from '@/utils/confirm'

const queue = ref<ConfirmationRequest[]>([])
const panelEl = ref<HTMLElement | null>(null)

const activeRequest = computed(() => queue.value[0] ?? null)

function settleCurrent(confirmed: boolean) {
  const request = queue.value.shift()
  request?.respond(confirmed)
}

function cancelCurrent() {
  settleCurrent(false)
}

function confirmCurrent() {
  settleCurrent(true)
}

watch(activeRequest, async (request) => {
  if (!request) return

  await nextTick()
  panelEl.value?.focus()
})

let removeConfirmListener = () => {}

onMounted(() => {
  removeConfirmListener = onConfirmationRequest((request) => {
    queue.value.push(request)
  })
})

onUnmounted(() => {
  removeConfirmListener()

  while (queue.value.length > 0) {
    settleCurrent(false)
  }
})

useEscape(cancelCurrent, {
  enabled: () => activeRequest.value !== null,
})
</script>

<template>
  <Teleport to="body">
    <Transition
      enter-active-class="transition duration-150 ease-out"
      enter-from-class="opacity-0"
      enter-to-class="opacity-100"
      leave-active-class="transition duration-100 ease-in"
      leave-from-class="opacity-100"
      leave-to-class="opacity-0"
    >
      <div
        v-if="activeRequest"
        class="fixed inset-0 z-110 bg-zinc-950/50 backdrop-blur-[2px]"
        @click="cancelCurrent"
      />
    </Transition>

    <Transition
      enter-active-class="transition duration-150 ease-out"
      enter-from-class="translate-y-2 opacity-0 scale-[0.98]"
      enter-to-class="translate-y-0 opacity-100 scale-100"
      leave-active-class="transition duration-100 ease-in"
      leave-from-class="translate-y-0 opacity-100 scale-100"
      leave-to-class="translate-y-2 opacity-0 scale-[0.98]"
    >
      <section
        v-if="activeRequest"
        ref="panelEl"
        tabindex="-1"
        role="alertdialog"
        aria-modal="true"
        :aria-labelledby="`confirm-title-${activeRequest.id}`"
        :aria-describedby="`confirm-message-${activeRequest.id}`"
        class="fixed inset-0 z-111 m-auto flex h-fit w-[min(92vw,32rem)] max-w-lg flex-col overflow-hidden rounded-3xl border border-zinc-200 bg-white text-zinc-900 shadow-2xl focus:outline-hidden dark:border-zinc-800 dark:bg-zinc-900 dark:text-zinc-100"
      >
        <header
          class="relative overflow-hidden border-b border-zinc-200 bg-zinc-50 dark:border-zinc-800 dark:bg-zinc-950/80"
        >
          <DecorGradient variant="gradientAndGrid" />

          <div class="relative z-10 flex items-start justify-between gap-4 px-5 py-4">
            <div class="flex min-w-0 items-start gap-4">
              <div
                class="grid size-12 shrink-0 place-items-center rounded-2xl border border-rose-200 bg-white shadow-sm dark:border-rose-400/20 dark:bg-zinc-900"
              >
                <ExclamationTriangleIcon class="size-7 text-rose-500 dark:text-rose-300" />
              </div>

              <div class="min-w-0">
                <h2
                  :id="`confirm-title-${activeRequest.id}`"
                  class="text-2xl font-semibold tracking-tight text-zinc-900 dark:text-zinc-100"
                >
                  {{ activeRequest.title }}
                </h2>
                <p
                  :id="`confirm-message-${activeRequest.id}`"
                  class="mt-1 text-sm leading-6 text-zinc-600 dark:text-zinc-300"
                >
                  {{ activeRequest.message }}
                </p>
              </div>
            </div>

            <button
              type="button"
              class="shrink-0 cursor-pointer rounded-lg p-2 text-zinc-600 transition hover:bg-rose-50 hover:text-rose-500 dark:text-zinc-300 dark:hover:bg-rose-400/15 dark:hover:text-rose-300"
              aria-label="Cancel confirmation"
              @click="cancelCurrent"
            >
              <XMarkIcon class="size-5" />
            </button>
          </div>
        </header>

        <div
          v-if="activeRequest.details"
          class="px-5 py-4"
        >
          <p
            class="rounded-2xl border border-zinc-200 bg-zinc-50 px-4 py-3 text-sm leading-6 text-zinc-700 dark:border-zinc-800 dark:bg-zinc-950/40 dark:text-zinc-300"
          >
            {{ activeRequest.details }}
          </p>
        </div>

        <footer
          class="flex flex-col-reverse gap-2 border-t border-zinc-200 px-5 py-4 sm:flex-row sm:justify-end dark:border-zinc-800"
        >
          <TheButton
            type="button"
            variant="secondary"
            class="cursor-pointer"
            @click="cancelCurrent"
          >
            {{ activeRequest.cancelLabel }}
          </TheButton>
          <TheButton
            type="button"
            :variant="activeRequest.confirmVariant"
            class="cursor-pointer"
            @click="confirmCurrent"
          >
            {{ activeRequest.confirmLabel }}
          </TheButton>
        </footer>
      </section>
    </Transition>
  </Teleport>
</template>
