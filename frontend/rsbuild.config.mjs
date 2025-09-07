import { defineConfig } from '@rsbuild/core';
import { pluginReact } from '@rsbuild/plugin-react';

export default defineConfig({
  plugins: [pluginReact()],
  html: {
    title: 'CCS Server Dashboard',
  },
  server: {
    port: 3000,
    proxy: {
      '/api': 'http://localhost:8080', // proxy to your Gin backend
    },
  },
});
