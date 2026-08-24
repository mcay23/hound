import axios from "axios";

declare global {
  interface Window {
    secureAuth?: {
      saveToken: (token: string) => Promise<{ success: boolean; error?: string }>;
      getToken: () => Promise<string | null>;
      clearToken: () => Promise<boolean>;
    };
  }
}

let memoryToken: string | null = null;

export const saveSecureToken = async (token: string): Promise<void> => {
  memoryToken = token;
  try {
    sessionStorage.setItem("auth_token", token);
  } catch (e) {
    console.error("Failed to save to sessionStorage:", e);
  }
  axios.defaults.headers.common["Authorization"] = `Bearer ${token}`;

  if (window.secureAuth?.saveToken) {
    try {
      const res = await window.secureAuth.saveToken(token);
      if (res && res.success === false) {
        console.warn("Electron safeStorage saveToken failed, using sessionStorage fallback:", res.error);
      }
    } catch (err) {
      console.error("Failed to save token to Electron secure store:", err);
    }
  }
};

export const getSecureToken = async (): Promise<string | null> => {
  let token: string | null = null;

  if (window.secureAuth?.getToken) {
    try {
      token = await window.secureAuth.getToken();
    } catch (err) {
      console.error("Failed to get token from Electron secure store:", err);
    }
  }

  if (!token) {
    token = memoryToken || sessionStorage.getItem("auth_token");
  }

  if (token) {
    memoryToken = token;
    axios.defaults.headers.common["Authorization"] = `Bearer ${token}`;
  }

  return token;
};

export const clearSecureToken = async (): Promise<void> => {
  memoryToken = null;
  try {
    sessionStorage.removeItem("auth_token");
  } catch (e) {}
  delete axios.defaults.headers.common["Authorization"];

  if (window.secureAuth?.clearToken) {
    try {
      await window.secureAuth.clearToken();
    } catch (err) {
      console.error("Failed to clear token from Electron secure store:", err);
    }
  }
};

