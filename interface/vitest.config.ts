import { defineConfig } from 'vitest/config'
import react from '@vitejs/plugin-react'
import path from 'path'

export default defineConfig({
    plugins: [react()],
    test: {
        environment: 'jsdom',
        globals: true,
        setupFiles: ['./__tests__/setup.ts'],
        exclude: ['e2e/**', 'node_modules/**'],
        alias: {
            '@': path.resolve(__dirname, './')
        },
        coverage: {
            provider: 'v8',
            reporter: ['text', 'lcov', 'json-summary'],
            include: ['components/**', 'store/**', 'lib/**', 'app/**'],
            exclude: ['e2e/**', 'node_modules/**', '**/*.d.ts', '**/*.config.*'],
            thresholds: {
                statements: 30,
                branches: 25,
                functions: 30,
                lines: 30,
            },
        },
    },
})
