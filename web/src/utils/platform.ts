export const isPlatformElectron = typeof window !== "undefined" &&
  ("_electronMpvVideo" in window ||
    "electron" in window ||
    navigator.userAgent.includes("Electron"));