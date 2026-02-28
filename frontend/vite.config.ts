import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import vueDevTools from 'vite-plugin-vue-devtools'
import tailwindcss from '@tailwindcss/vite'
import checker from 'vite-plugin-checker'
// https://vite.dev/config/
export default defineConfig({
    plugins: [
        vue(),
        tailwindcss(),
        vueDevTools(),
        checker({
            overlay: true,
            vueTsc: { tsconfigPath: './tsconfig.app.json' },
            typescript: { tsconfigPath: './tsconfig.app.json' },
        }),
    ],
    resolve: {
        alias: {
            '@': fileURLToPath(new URL('./src', import.meta.url)),
        },
    },
    server: {
        proxy: {
            '/api': {
                target: 'http://localhost:4206',
                changeOrigin: true,
            },
        },
    },
})
