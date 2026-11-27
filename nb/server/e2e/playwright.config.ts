import { defineConfig } from '@playwright/test';

export default defineConfig({
  testDir: '.',
  timeout: 30000,
  use: {
    baseURL: 'http://localhost:18080',
  },
  webServer: {
    command: 'go run ../../../cmd/gobook serve fixtures/hello.md --addr localhost:18080',
    url: 'http://localhost:18080/api/status',
    reuseExistingServer: false,
    timeout: 60000,
    cwd: __dirname,
  },
});
