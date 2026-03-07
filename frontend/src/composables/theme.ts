import { useColorMode } from '@vueuse/core'

export function useTheme() {
    const mode = useColorMode({
        attribute: 'data-theme',
        modes: {
            light: 'light',
            dark: 'dark',
        },
        disableTransition: false,
    })

    return { mode }
}
