<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { useClientStore } from '@/stores/clients'
import type { Client } from '@/utils/clientHttpHandler'
import { useEnter, useEscape } from '@/composables/keyHandlers'
import TheInput from '../UI/TheInput.vue'
import TheButton from '../UI/TheButton.vue'
import { validateClientForm } from '@/utils/frontendValidation'

import {
  PencilIcon,
  TrashIcon,
  CheckCircleIcon,
  XCircleIcon,
  UserPlusIcon,
  MagnifyingGlassIcon,
  ChevronDownIcon,
  ArrowPathIcon,
  UsersIcon,
} from '@heroicons/vue/24/outline'
import DecorGradient from '../UI/DecorGradient.vue'
import TheTooltip from '../UI/TheTooltip.vue'
import { emitToastSuccess } from '@/utils/toast'
import { handleActionError } from '@/utils/errors/handleActionError'
import { requestConfirmation } from '@/utils/confirm'

const clientStore = useClientStore()

// Search and order
const searchQuery = ref('')

const filteredClients = computed(() => {
  const q = searchQuery.value.trim().toLowerCase()

  const list = !q
    ? clientStore.clients
    : clientStore.clients.filter((c: Client) => (c.name ?? '').toLowerCase().includes(q))

  // Newest first
  return [...list].sort((a, b) => (b.id ?? 0) - (a.id ?? 0))
})

// Add new client
const createForm = reactive({
  name: '',
  companyName: '',
  email: '',
  address: '',
})
const createFieldErrors = ref<Record<string, string>>({})
const liveCreateErrors = computed(() => validateClientForm(createForm))
const displayCreateErrors = computed(() => ({
  ...createFieldErrors.value,
  ...liveCreateErrors.value,
}))

const canCreate = computed(() => Object.keys(liveCreateErrors.value).length === 0)

watch(
  createForm,
  () => {
    createFieldErrors.value = {}
  },
  { deep: true },
)

function resetCreate() {
  createForm.name = ''
  createForm.companyName = ''
  createForm.email = ''
  createForm.address = ''
  createFieldErrors.value = {}
}

async function addClient() {
  if (!canCreate.value) return

  createFieldErrors.value = {}

  try {
    await clientStore.createNew(createForm)

    resetCreate()

    emitToastSuccess('Client created')
  } catch (err) {
    handleActionError(err, {
      fieldErrors: createFieldErrors,
      toastTitle: 'Create client failed',
    })
  }
}

// Expand / Edit
const openId = ref<number | null>(null)
const editingId = ref<number | null>(null)

const editForm = reactive({
  id: null as number | null,
  name: '',
  companyName: '',
  email: '',
  address: '',
})
const editFieldErrors = ref<Record<string, string>>({})
const liveEditErrors = computed(() => validateClientForm(editForm))
const displayEditErrors = computed(() => ({
  ...editFieldErrors.value,
  ...liveEditErrors.value,
}))

watch(
  editForm,
  () => {
    editFieldErrors.value = {}
  },
  { deep: true },
)

function toggleOpen(id: number) {
  const isClosing = openId.value === id

  if (isClosing && editingId.value === id) cancelEdit()

  // Stop edit when switching rows
  if (!isClosing && editingId.value != null) cancelEdit()

  openId.value = isClosing ? null : id
}

function startEdit(c: Client) {
  openId.value = c.id // expanded editing
  editingId.value = c.id

  editForm.id = c.id
  editForm.name = c.name ?? ''
  editForm.companyName = c.companyName ?? ''
  editForm.email = c.email ?? ''
  editForm.address = c.address ?? ''
  editFieldErrors.value = {}
}

function cancelEdit() {
  editingId.value = null
  editForm.id = null
  editForm.name = ''
  editForm.companyName = ''
  editForm.email = ''
  editForm.address = ''
  editFieldErrors.value = {}
}

async function saveEdit() {
  if (editForm.id === null) return
  editFieldErrors.value = {}
  try {
    console.log('sending to server: ', editForm)
    await clientStore.edit(editForm.id, {
      name: editForm.name,
      companyName: editForm.companyName,
      email: editForm.email,
      address: editForm.address,
    })
    cancelEdit()
  } catch (err) {
    handleActionError(err, {
      fieldErrors: editFieldErrors,
      toastTitle: 'Update client failed',
    })
  }
}

async function removeClient(client: Client) {
  const confirmed = await requestConfirmation({
    title: 'Delete client?',
    message: `Delete ${client.name || 'this client'} from your clients list?`,
    details:
      'This action cannot be undone. If the client is linked to saved invoices, deletion may be blocked.',
    confirmLabel: 'Delete client',
    cancelLabel: 'Keep client',
    confirmVariant: 'danger',
  })

  if (!confirmed) return

  try {
    await clientStore.remove(client.id)
    emitToastSuccess('Client removed')
    if (openId.value === client.id) openId.value = null
    if (editingId.value === client.id) cancelEdit()
  } catch (err) {
    handleActionError(err, {
      toastTitle: 'Delete client failed',
      mapFields: false,
    })
  }
}

// Field Schema
type ClientFieldKey = keyof typeof createForm

type FieldDef = {
  key: ClientFieldKey
  label: string
  placeholder?: string
  autocomplete?: string
}

const clientFields: FieldDef[] = [
  {
    key: 'name',
    label: 'Name',
    placeholder: 'Client name',
    autocomplete: 'name',
  },
  {
    key: 'companyName',
    label: 'Company',
    placeholder: 'Company name',
    autocomplete: 'organization',
  },
  {
    key: 'email',
    label: 'Email',
    placeholder: 'Email',
    autocomplete: 'email',
  },
  {
    key: 'address',
    label: 'Address',
    placeholder: 'Address',
    autocomplete: 'street-address',
  },
]

const displayFields: ClientFieldKey[] = ['email', 'address']
// Keyboard events
useEscape(cancelEdit, { enabled: openId !== null })
useEnter(saveEdit, { enabled: editForm.id === null })
</script>

<template>
  <section class="mx-auto w-full max-w-4xl 2xl:max-w-5xl">
    <!-- Header -->
    <div class="mb-4 flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
      <div class="flex items-center gap-2">
        <div
          class="grid size-12 shrink-0 place-items-center rounded-2xl border border-zinc-300 bg-white shadow-sm dark:border-zinc-700 dark:bg-zinc-900"
        >
          <UsersIcon class="stroke-1.5 size-7 text-sky-600 dark:text-emerald-400" />
        </div>
        <div>
          <h2 class="text-2xl font-semibold tracking-tight text-zinc-900 dark:text-zinc-200">
            Clients
          </h2>
          <p class="text-sm tracking-wide text-zinc-600 dark:text-zinc-400">
            Add, search, and edit clients
          </p>
        </div>
      </div>

      <!-- Search -->
      <div class="w-full sm:max-w-sm">
        <div class="relative shadow-md">
          <MagnifyingGlassIcon
            class="pointer-events-none absolute top-1/2 left-2 size-5 -translate-y-1/2 text-zinc-600 dark:text-zinc-400"
          />
          <input
            id="srchQry-clients-1"
            v-model="searchQuery"
            type="text"
            placeholder="Search by name…"
            class="input input-accent py-1.5 pl-9"
          />
        </div>
      </div>
    </div>

    <!-- Add panel -->
    <div
      class="relative mb-4 overflow-hidden rounded-2xl border border-zinc-300 bg-white shadow-md dark:border-zinc-800 dark:bg-zinc-950/30"
    >
      <!-- border glow/texture  -->
      <DecorGradient />

      <div class="relative p-4">
        <!-- title row -->
        <div class="flex flex-col items-start justify-between gap-3 sm:flex-row">
          <div class="flex items-center gap-3">
            <div class="min-w-0">
              <div class="flex items-center gap-2">
                <h3 class="text-lg font-semibold text-zinc-900 dark:text-zinc-100">Add client</h3>
              </div>

              <p class="mt-0.5 text-xs font-bold text-sky-600 dark:text-emerald-400">
                Create a client, to use in invoices and items
              </p>
            </div>
          </div>

          <div class="flex shrink-0 items-center gap-2">
            <TheTooltip text="Clear inputs">
              <TheButton
                type="button"
                variant="secondary"
                @click="resetCreate"
              >
                <ArrowPathIcon class="size-4" />
                Clear
              </TheButton>
            </TheTooltip>

            <TheTooltip :text="canCreate ? 'Add new' : 'Adjust inputs before creating'">
              <TheButton
                type="button"
                :disabled="!canCreate"
                variant="primary"
                @click="addClient"
                :class="canCreate ? 'cursor-pointer' : ''"
              >
                <UserPlusIcon class="size-5" />
                Create
              </TheButton>
            </TheTooltip>
          </div>
        </div>

        <!-- fields -->
        <div class="mt-4 grid grid-cols-1 gap-2 sm:grid-cols-2 lg:grid-cols-4">
          <TheInput
            v-for="field in clientFields"
            :key="field.key"
            :id="`new-client-${field.key}`"
            :label="field.label"
            :placeholder="field.placeholder"
            :autocomplete="field.autocomplete"
            v-model="createForm[field.key]"
            :error="displayCreateErrors[field.key]"
          />
        </div>

        <div
          class="mt-3 flex items-center justify-between text-xs text-zinc-700 dark:text-zinc-200"
        >
          <div class="hidden sm:block">
            Tip: Company,email and address are optional, but useful for your invoice
          </div>
        </div>
      </div>
    </div>

    <!-- List -->
    <div class="space-y-2">
      <div
        v-if="filteredClients.length === 0"
        class="rounded-xl border border-zinc-300 bg-white p-6 text-center shadow-sm dark:border-zinc-800 dark:bg-zinc-900"
      >
        <p class="font-medium text-zinc-900 dark:text-zinc-100">No clients found</p>
        <p class="mt-1 text-sm text-zinc-600 dark:text-zinc-400">
          Try a different search, or add a new client above.
        </p>
      </div>

      <article
        v-for="c in filteredClients"
        :key="c.id"
        class="rounded-xl border border-zinc-300 bg-white shadow-sm dark:border-zinc-800 dark:bg-zinc-900"
      >
        <button
          type="button"
          class="group flex w-full items-center justify-between gap-3 rounded-xl px-3 py-2 text-left hover:bg-zinc-50/50 dark:hover:bg-zinc-800/40"
          @click="toggleOpen(c.id)"
        >
          <div class="min-w-0">
            <p class="truncate text-sm font-semibold text-zinc-900 dark:text-zinc-100">
              {{ c.name || 'Unnamed client' }}
            </p>
            <p
              class="truncate text-xs"
              :class="
                c.id === openId
                  ? 'text-sky-600 dark:text-emerald-400'
                  : 'text-zinc-600 dark:text-zinc-400'
              "
            >
              {{ c.companyName || '—' }}
            </p>
          </div>

          <div class="flex items-center gap-2">
            <template v-if="editingId === c.id">
              <TheTooltip>
                <template #content>
                  Confirm edit.
                  <span class="text-sky-600 dark:text-emerald-400">Shortcut:</span>
                  <kbd>↵ Enter</kbd>
                </template>
                <button
                  type="button"
                  class="cursor-pointer rounded-md border border-transparent p-1 text-zinc-600 hover:border-sky-900/20 hover:bg-sky-100 hover:text-sky-600 dark:text-zinc-300 dark:hover:border-emerald-900/50 dark:hover:bg-emerald-900/20 dark:hover:text-emerald-400"
                  @click.stop="saveEdit"
                >
                  <CheckCircleIcon class="size-5" />
                </button>
              </TheTooltip>
              <TheTooltip>
                <template #content>
                  Cancel edit.
                  <span class="text-sky-600 dark:text-emerald-400">Shortcut:</span>
                  <kbd>Esc</kbd>
                </template>
                <button
                  type="button"
                  class="cursor-pointer rounded-md border border-transparent p-1 text-zinc-600 hover:border-rose-600/20 hover:bg-rose-50 hover:text-rose-600 dark:text-zinc-300 dark:hover:border-rose-300/20 dark:hover:bg-rose-900/20 dark:hover:text-rose-300"
                  @click.stop="cancelEdit"
                >
                  <XCircleIcon class="size-5" />
                </button>
              </TheTooltip>
            </template>
            <template v-else>
              <TheTooltip
                text="Edit client details"
                side="top"
                max-width-class="w-42"
                align="center"
              >
                <button
                  type="button"
                  class="cursor-pointer rounded-md border border-transparent p-1 text-zinc-600 hover:border-sky-900/30 hover:bg-sky-100 hover:text-sky-600 dark:text-zinc-300 dark:hover:border-emerald-900/50 dark:hover:bg-emerald-900/20 dark:hover:text-emerald-400"
                  @click.stop="startEdit(c)"
                >
                  <PencilIcon class="size-5" />
                </button>
              </TheTooltip>
              <TheTooltip
                text="Delete client"
                side="top"
                max-width-class="w-42"
                align="center"
              >
                <button
                  type="button"
                  class="cursor-pointer rounded-md border border-transparent p-1 text-zinc-600 hover:border-rose-600/20 hover:bg-rose-50 hover:text-rose-500 dark:text-zinc-300 dark:hover:border-rose-300/20 dark:hover:bg-rose-900/20 dark:hover:text-rose-300"
                  @click.stop="removeClient(c)"
                >
                  <TrashIcon class="size-5" />
                </button>
              </TheTooltip>
            </template>

            <ChevronDownIcon
              class="size-5 text-zinc-600 transition-transform group-hover:text-sky-600 dark:text-zinc-400 group-hover:dark:text-emerald-400"
              :class="openId === c.id ? 'rotate-180' : ''"
            />
          </div>
        </button>

        <!-- expandable content -->
        <div
          v-if="openId === c.id"
          class="border-t border-zinc-300 p-3 dark:border-zinc-800"
        >
          <template v-if="editingId === c.id">
            <div class="grid min-w-0 grid-cols-1 gap-2 sm:grid-cols-2">
              <TheInput
                v-for="field in clientFields"
                :key="field.key"
                :id="`edit-${field.key}`"
                :label="field.label"
                v-model="editForm[field.key]"
                :error="displayEditErrors[field.key]"
              />
            </div>
          </template>

          <template v-else>
            <div class="grid grid-cols-1 gap-2 sm:grid-cols-2">
              <div
                v-for="key in displayFields"
                :key="key"
                class="rounded-lg border border-zinc-300 bg-zinc-50 p-3 text-sm dark:border-zinc-800 dark:bg-zinc-800/40"
              >
                <p class="text-xs text-zinc-600 capitalize dark:text-zinc-400">
                  {{ key === 'companyName' ? 'Company' : key }}
                </p>

                <p class="mt-1 wrap-break-word text-zinc-900 dark:text-zinc-100">
                  {{ c[key] || '—' }}
                </p>
              </div>
            </div>
          </template>
        </div>
      </article>
    </div>
  </section>
</template>
