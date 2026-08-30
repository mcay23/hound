const { exposeMpvApi } = require("electron-mpv-video/preload");
const { contextBridge, ipcRenderer } = require('electron');

exposeMpvApi();

contextBridge.exposeInMainWorld('secureAuth', {
  saveToken: (token) => ipcRenderer.invoke('store-token', token),
  getToken: () => ipcRenderer.invoke('get-token'),
  clearToken: () => ipcRenderer.invoke('delete-token')
});

contextBridge.exposeInMainWorld('electron', {
  toggleFullscreen: () => ipcRenderer.send('toggle-fullscreen'),
  isFullscreen: () => ipcRenderer.invoke('is-fullscreen')
});