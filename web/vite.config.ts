import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import tsconfigPaths from 'vite-tsconfig-paths';
import path from 'path';
import fs from 'fs';

const electronMpvDir = path.resolve(__dirname, 'electron/node_modules/electron-mpv-video');
const hasElectronMpv = fs.existsSync(electronMpvDir);
const stubPath = path.resolve(__dirname, 'src/utils/mpvStub.ts');

export default defineConfig({
  plugins: [react(), tsconfigPaths()],
  server: {
    port: 3000,
    open: true,
    fs: {
      allow: ['..'],
    },
  },
  resolve: {
    alias: [
      {
        find: 'electron-mpv-video/renderer',
        replacement: hasElectronMpv
          ? path.resolve(electronMpvDir, 'dist/renderer/index.js')
          : stubPath,
      },
      {
        find: 'electron-mpv-video/main',
        replacement: hasElectronMpv
          ? path.resolve(electronMpvDir, 'dist/main/index.js')
          : stubPath,
      },
      {
        find: 'electron-mpv-video',
        replacement: hasElectronMpv
          ? path.resolve(electronMpvDir, 'dist/main/index.js')
          : stubPath,
      },
    ],
  },
  build: {
    outDir: 'build',
    assetsDir: 'static',
  },
});
