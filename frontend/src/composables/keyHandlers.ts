import { onMounted, onUnmounted, toValue, type MaybeRefOrGetter } from 'vue'

type KeyHandlerOptions = {
    enabled?: MaybeRefOrGetter<boolean>
    target?: MaybeRefOrGetter<HTMLElement | null>
}

type Modifier = 'ctrl' | 'shift' | 'alt' | 'meta'

export type ShortcutDefinition = {
    key: string
    modifiers?: Modifier[]
    action: () => void
}

function isEventInside(target: EventTarget | null, container: HTMLElement | null) {
    return target instanceof Node && !!container && container.contains(target)
}

function isFormControl(target: EventTarget | null) {
    const el = target as HTMLElement | null
    if (!el) return false

    const tag = el.tagName
    return tag === 'INPUT' || tag === 'TEXTAREA' || tag === 'SELECT' || el.isContentEditable
}

export function useEscape(callback: () => void, options: KeyHandlerOptions = {}) {
    function handler(e: KeyboardEvent) {
        if (e.key !== 'Escape') return

        const enabled = options.enabled !== undefined ? toValue(options.enabled) : true
        if (!enabled) return

        const container = options.target !== undefined ? toValue(options.target) : null
        if (options.target !== undefined && !isEventInside(e.target, container)) return

        callback()
    }

    onMounted(() => window.addEventListener('keyup', handler))
    onUnmounted(() => window.removeEventListener('keyup', handler))
}

export function useEnter(callback: () => void, options: KeyHandlerOptions = {}) {
    let countFiredOnInput = 0
    console.log('FIREE', countFiredOnInput)

    function handler(e: KeyboardEvent) {
        if (countFiredOnInput >= 1) return
        if (e.key !== 'Enter') return
        if (e.repeat) return
        console.log(e.repeat)
        console.log('TARGET', e.target)
        const enabled = options.enabled !== undefined ? toValue(options.enabled) : true
        if (!enabled) return

        const container = options.target !== undefined ? toValue(options.target) : null
        if (options.target !== undefined && !isEventInside(e.target, container)) return

        if (!isFormControl(e.target)) return

        callback()
    }

    onMounted(() => window.addEventListener('keyup', handler))
    onUnmounted(() => window.removeEventListener('keyup', handler))
}

const modifierMap: Record<Modifier, keyof KeyboardEvent> = {
    ctrl: 'ctrlKey',
    shift: 'shiftKey',
    alt: 'altKey',
    meta: 'metaKey',
}

export function useShortcuts(shortcuts: ShortcutDefinition[]) {
    function handler(e: KeyboardEvent) {
        for (const shortcut of shortcuts) {
            if (e.key.toLowerCase() !== shortcut.key.toLowerCase()) continue
            if (e.repeat) return
            const modifiers = shortcut.modifiers ?? []
            const allModifiersMatch = modifiers.every((mod) => e[modifierMap[mod]])

            const noExtraModifiers = (['ctrl', 'shift', 'alt', 'meta'] as Modifier[])
                .filter((mod) => !modifiers.includes(mod))
                .every((mod) => !e[modifierMap[mod]])

            if (allModifiersMatch && noExtraModifiers) {
                e.preventDefault()
                shortcut.action()
                return
            }
        }
    }

    onMounted(() => window.addEventListener('keydown', handler))
    onUnmounted(() => window.removeEventListener('keydown', handler))
}
