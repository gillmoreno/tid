import '@testing-library/jest-dom/vitest'
import 'fake-indexeddb/auto'
import { webcrypto } from 'node:crypto'

Object.defineProperty(globalThis, 'crypto', {
  configurable: true,
  value: webcrypto,
})

Object.defineProperty(navigator, 'clipboard', {
  configurable: true,
  value: { writeText: async () => undefined },
})
