<script setup lang="ts">
import {
  CheckCircleIcon,
  ExclamationCircleIcon,
  InformationCircleIcon,
  XMarkIcon,
} from '@heroicons/vue/20/solid'
import { onMounted, onUnmounted, ref } from 'vue'
import { onToast } from '@/utils/toast'

export type ToastNotice = {
  id: string
  level: 'error' | 'success' | 'info'
  code?: string
  title?: string
  message: string
  durationMs?: number
}

type ToneConfig = {
  wrapper: string
  iconColor: string
  titleColor: string
  codeColor: string
  messageColor: string
  button: string
  closeButton: string
  closeIcon: string
  progressTrack: string
  progressBar: string
  defaultTitle: string
  icon: typeof InformationCircleIcon
}

type ToastTimer = {
  startedAt: number
  remainingMs: number
  timeoutId: ReturnType<typeof setTimeout> | null
  paused: boolean
}

const emit = defineEmits<{
  toast: [toast: ToastNotice]
  dismiss: [id: string]
  clear: []
}>()

const MAX_TOASTS = 4
const DEFAULT_DURATION_MS = 5000

const toasts = ref<ToastNotice[]>([])
const timers = new Map<string, ToastTimer>()

const baseButton =
  'rounded-md px-2 py-1.5 text-sm font-medium focus:ring-2 focus:ring-offset-2 focus:outline-hidden'

const toneMap: Record<ToastNotice['level'], ToneConfig> = {
  success: {
    wrapper: 'border-green-200 bg-green-50 dark:border-emerald-400/30 dark:bg-emerald-900/20',
    iconColor: 'text-green-500 dark:text-emerald-400',
    titleColor: 'text-green-900 dark:text-emerald-200',
    codeColor: 'text-green-700 dark:text-emerald-300/90',
    messageColor: 'text-green-800 dark:text-emerald-300',
    button: `${baseButton} bg-green-50 text-green-800 hover:bg-green-100 focus:ring-green-600 focus:ring-offset-green-50 dark:bg-emerald-900/30 dark:text-emerald-200 dark:hover:bg-emerald-900/40 dark:focus:ring-emerald-400 dark:focus:ring-offset-zinc-950`,
    closeButton:
      'rounded-md p-1.5 hover:bg-black/5 focus:outline-hidden focus:ring-2 focus:ring-green-600 dark:hover:bg-white/5 dark:focus:ring-emerald-400',
    closeIcon: 'text-green-700 dark:text-emerald-200',
    progressTrack: 'bg-green-200/70 dark:bg-emerald-400/15',
    progressBar: 'bg-green-500 dark:bg-emerald-400',
    defaultTitle: 'Success',
    icon: CheckCircleIcon,
  },

  info: {
    wrapper: 'border-sky-200 bg-sky-50 dark:border-sky-400/30 dark:bg-sky-900/20',
    iconColor: 'text-sky-500 dark:text-sky-300',
    titleColor: 'text-sky-900 dark:text-sky-200',
    codeColor: 'text-sky-700 dark:text-sky-300/90',
    messageColor: 'text-sky-800 dark:text-sky-300',
    button: `${baseButton} bg-sky-50 text-sky-800 hover:bg-sky-100 focus:ring-sky-600 focus:ring-offset-sky-50 dark:bg-sky-900/30 dark:text-sky-200 dark:hover:bg-sky-900/40 dark:focus:ring-offset-zinc-950`,
    closeButton:
      'rounded-md p-1.5 hover:bg-black/5 focus:outline-hidden focus:ring-2 focus:ring-sky-600 dark:hover:bg-white/5 dark:focus:ring-sky-400',
    closeIcon: 'text-sky-700 dark:text-sky-200',
    progressTrack: 'bg-sky-200/70 dark:bg-sky-400/15',
    progressBar: 'bg-sky-500 dark:bg-sky-400',
    defaultTitle: 'Information',
    icon: InformationCircleIcon,
  },

  error: {
    wrapper: 'border-red-200 bg-red-50 dark:border-red-400/30 dark:bg-red-900/20',
    iconColor: 'text-red-500 dark:text-red-300',
    titleColor: 'text-red-900 dark:text-red-200',
    codeColor: 'text-red-700 dark:text-red-300/90',
    messageColor: 'text-red-800 dark:text-red-300',
    button: `${baseButton} bg-red-50 text-red-800 hover:bg-red-100 focus:ring-red-600 focus:ring-offset-red-50 dark:bg-red-900/30 dark:text-red-200 dark:hover:bg-red-900/40 dark:focus:ring-offset-zinc-950`,
    closeButton:
      'rounded-md p-1.5 hover:bg-black/5 focus:outline-hidden focus:ring-2 focus:ring-red-600 dark:hover:bg-white/5 dark:focus:ring-red-400',
    closeIcon: 'text-red-700 dark:text-red-200',
    progressTrack: 'bg-red-200/70 dark:bg-red-400/15',
    progressBar: 'bg-red-500 dark:bg-red-400',
    defaultTitle: 'Something went wrong',
    icon: ExclamationCircleIcon,
  },
}

function getTone(level: ToastNotice['level']) {
  return toneMap[level]
}

function isAutoDismissible(toast: ToastNotice) {
  return toast.level === 'success' || toast.level === 'info'
}

function getToastDuration(toast: ToastNotice) {
  if (!isAutoDismissible(toast)) return 0
  return toast.durationMs ?? DEFAULT_DURATION_MS
}

function clearToastTimer(id: string) {
  const timer = timers.get(id)
  if (!timer) return

  if (timer.timeoutId) {
    clearTimeout(timer.timeoutId)
    timer.timeoutId = null
  }
}

function removeToastTimer(id: string) {
  clearToastTimer(id)
  timers.delete(id)
}

function startToastTimer(toast: ToastNotice, remainingMs?: number) {
  const duration = remainingMs ?? getToastDuration(toast)
  if (duration <= 0) return

  clearToastTimer(toast.id)

  const timer: ToastTimer = {
    startedAt: Date.now(),
    remainingMs: duration,
    timeoutId: setTimeout(() => {
      dismissToast(toast.id)
    }, duration),
    paused: false,
  }

  timers.set(toast.id, timer)
}

function pauseToastTimer(id: string) {
  const timer = timers.get(id)
  if (!timer || timer.paused) return

  const elapsed = Date.now() - timer.startedAt
  timer.remainingMs = Math.max(0, timer.remainingMs - elapsed)
  timer.paused = true

  if (timer.timeoutId) {
    clearTimeout(timer.timeoutId)
    timer.timeoutId = null
  }
}

function resumeToastTimer(id: string) {
  const toast = toasts.value.find((t) => t.id === id)
  const timer = timers.get(id)

  if (!toast || !timer || !timer.paused) return
  if (timer.remainingMs <= 0) {
    dismissToast(id)
    return
  }

  timer.startedAt = Date.now()
  timer.paused = false
  timer.timeoutId = setTimeout(() => {
    dismissToast(id)
  }, timer.remainingMs)
}

function getProgressStyle(toast: ToastNotice) {
  const duration = getToastDuration(toast)
  if (!duration) return { width: '0%' }

  const timer = timers.get(toast.id)
  if (!timer) return { width: '100%' }

  let remaining = timer.remainingMs

  if (!timer.paused) {
    const elapsed = Date.now() - timer.startedAt
    remaining = Math.max(0, timer.remainingMs - elapsed)
  }

  const pct = Math.max(0, Math.min(100, (remaining / duration) * 100))
  return { width: `${pct}%` }
}

function pushToast(toast: ToastNotice) {
  const existingIndex = toasts.value.findIndex((t) => t.id === toast.id)

  if (existingIndex !== -1) {
    toasts.value[existingIndex] = toast
    removeToastTimer(toast.id)
  } else {
    toasts.value.unshift(toast)
  }

  if (toasts.value.length > MAX_TOASTS) {
    const removed = toasts.value.slice(MAX_TOASTS)
    for (const toast of removed) removeToastTimer(toast.id)
    toasts.value = toasts.value.slice(0, MAX_TOASTS)
  }

  if (isAutoDismissible(toast)) {
    startToastTimer(toast)
  }

  emit('toast', toast)
}

function dismissToast(id: string) {
  const index = toasts.value.findIndex((toast) => toast.id === id)
  if (index === -1) return

  toasts.value.splice(index, 1)
  removeToastTimer(id)
  emit('dismiss', id)
}

function clearToasts() {
  for (const toast of toasts.value) {
    removeToastTimer(toast.id)
  }

  toasts.value = []
  emit('clear')
}

let removeToastListener = () => {}

onMounted(() => {
  removeToastListener = onToast((toast) => {
    pushToast(toast)
  })
})

onUnmounted(() => {
  removeToastListener()
  clearToasts()
})

defineExpose({
  onToast: pushToast,
  onError: pushToast,
  dismissToast,
  clearErrors: clearToasts,
  clearToasts,
})
</script>

<template>
  <Teleport to="body">
    <div
      v-if="toasts.length"
      class="fixed top-0 left-1/2 z-50 mx-auto mt-4 flex w-full max-w-md -translate-x-1/2 flex-col gap-3 px-4"
    >
      <div
        v-for="toast in toasts"
        :key="toast.id"
        class="relative overflow-hidden rounded-md border p-4 shadow-sm"
        :class="getTone(toast.level).wrapper"
        @mouseenter="pauseToastTimer(toast.id)"
        @mouseleave="resumeToastTimer(toast.id)"
      >
        <div class="flex items-start">
          <div class="shrink-0">
            <component
              :is="getTone(toast.level).icon"
              class="size-5"
              :class="getTone(toast.level).iconColor"
              aria-hidden="true"
            />
          </div>

          <div class="ml-3 min-w-0 flex-1 pr-2">
            <h3
              class="text-sm font-medium"
              :class="getTone(toast.level).titleColor"
            >
              {{ toast.title || getTone(toast.level).defaultTitle }}
            </h3>

            <p
              v-if="toast.code"
              class="mt-1 text-xs font-medium"
              :class="getTone(toast.level).codeColor"
            >
              {{ toast.code }}
            </p>

            <div
              class="mt-2 text-sm"
              :class="getTone(toast.level).messageColor"
            >
              <p>{{ toast.message }}</p>
            </div>

            <div class="mt-4 flex items-center gap-3">
              <button
                type="button"
                :class="getTone(toast.level).button"
                @click="dismissToast(toast.id)"
              >
                Dismiss
              </button>
            </div>
          </div>

          <button
            type="button"
            class="shrink-0"
            :class="getTone(toast.level).closeButton"
            @click="dismissToast(toast.id)"
            aria-label="Dismiss notification"
          >
            <XMarkIcon
              class="size-4"
              :class="getTone(toast.level).closeIcon"
            />
          </button>
        </div>

        <div
          v-if="toast.level !== 'error'"
          class="mt-3 h-1 w-full rounded-full"
          :class="getTone(toast.level).progressTrack"
        >
          <div
            class="linear h-1 rounded-full transition-[width] duration-100"
            :class="getTone(toast.level).progressBar"
            :style="getProgressStyle(toast)"
          />
        </div>
      </div>
    </div>
  </Teleport>
</template>
