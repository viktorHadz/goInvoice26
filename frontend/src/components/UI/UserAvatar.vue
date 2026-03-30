<script setup lang="ts">
import { computed, ref, watch } from 'vue'

const props = withDefaults(
    defineProps<{
        name?: string | null
        email?: string | null
        avatarUrl?: string | null
    }>(),
    {
        name: '',
        email: '',
        avatarUrl: '',
    },
)

const imageFailed = ref(false)

const normalizedAvatarUrl = computed(() => (props.avatarUrl ?? '').trim())

watch(
    () => props.avatarUrl,
    () => {
        imageFailed.value = false
    },
)

const hasImage = computed(() => {
    return !imageFailed.value && normalizedAvatarUrl.value.length > 0
})

const initials = computed(() => {
    const source = props.name?.trim() || props.email?.trim() || 'A'
    const parts = source.split(/\s+/).filter(Boolean)

    if (parts.length >= 2) {
        return `${parts[0]?.charAt(0) ?? ''}${parts[1]?.charAt(0) ?? ''}`.toUpperCase()
    }

    return source.slice(0, 2).toUpperCase()
})
</script>

<template>
    <div
        class="grid shrink-0 place-items-center overflow-hidden border border-zinc-300 bg-white text-sm font-semibold text-zinc-700 dark:border-zinc-700 dark:bg-zinc-900 dark:text-zinc-200"
    >
        <img
            v-if="hasImage"
            :src="normalizedAvatarUrl"
            :alt="name ?? email ?? 'User avatar'"
            class="h-full w-full object-cover"
            loading="lazy"
            decoding="async"
            referrerpolicy="no-referrer"
            @error="imageFailed = true"
        />
        <span v-else>{{ initials }}</span>
    </div>
</template>
