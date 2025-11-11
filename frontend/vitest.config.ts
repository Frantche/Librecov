import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [vue()],
  test: {
    globals: true,
    environment: 'happy-dom',
    coverage: {
      provider: 'v8',
      reporter: ['text', 'json', 'html'],
      reportsDirectory: './coverage',
      include: ['src/**/*.{js,ts,vue}'],
      exclude: [
        'node_modules',
        'dist',
        '**/*.spec.ts',
        '**/*.test.ts',
        'src/main.ts',
        'src/**/*.d.ts'
      ]
    }
  }
})
