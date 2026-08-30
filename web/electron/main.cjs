const { app, BrowserWindow, Menu, ipcMain, safeStorage } = require("electron");
const http = require("http");
const fs = require("fs");
const path = require("path");
const { createMpvMain } = require("electron-mpv-video/main");
const StoreModule = require('electron-store');
const Store = typeof StoreModule === 'function' ? StoreModule : (StoreModule.default || StoreModule);

let mainWindow;
let server;
const store = new Store();

const mpv = createMpvMain();

if (process.env.ELECTRON_DEV !== "true") {
  Menu.setApplicationMenu(null);
}

// Helper to encrypt and save the token
function saveToken(token) {
  if (safeStorage && safeStorage.isEncryptionAvailable && safeStorage.isEncryptionAvailable()) {
    const encryptedBuffer = safeStorage.encryptString(token);
    store.set('auth_token', encryptedBuffer.toString('hex'));
    store.set('auth_token_encrypted', true);
  } else {
    store.set('auth_token', token);
    store.set('auth_token_encrypted', false);
  }
}

function getToken() {
  const storedVal = store.get('auth_token');
  if (!storedVal) return null;
  const isEncrypted = store.get('auth_token_encrypted');
  if (isEncrypted && safeStorage && safeStorage.isEncryptionAvailable && safeStorage.isEncryptionAvailable()) {
    try {
      const encryptedBuffer = Buffer.from(storedVal, 'hex');
      return safeStorage.decryptString(encryptedBuffer);
    } catch (e) {
      return storedVal;
    }
  }
  return storedVal;
}

ipcMain.handle('store-token', async (event, token) => {
  try {
    saveToken(token);
    return { success: true };
  } catch (error) {
    return { success: false, error: error.message };
  }
});

ipcMain.handle('get-token', async () => {
  try {
    return getToken();
  } catch (error) {
    return null; 
  }
});

ipcMain.handle('delete-token', async () => {
  store.delete('auth_token');
  store.delete('auth_token_encrypted');
  return true;
});

ipcMain.on('toggle-fullscreen', (event) => {
  const win = BrowserWindow.fromWebContents(event.sender);
  if (win) {
    win.setFullScreen(!win.isFullScreen());
  }
});

ipcMain.handle('is-fullscreen', (event) => {
  const win = BrowserWindow.fromWebContents(event.sender);
  if (win) {
    return win.isFullScreen();
  }
  return false;
});

function startStaticServer() {
  const buildPath = app.isPackaged
    ? path.join(process.resourcesPath, "build")
    : path.join(__dirname, "../build");

  server = http.createServer((req, res) => {
    let requestPath = decodeURIComponent(req.url.split("?")[0]);

    // React Router: serve index.html for routes
    if (requestPath === "/") {
      requestPath = "/index.html";
    }

    const filePath = path.join(buildPath, requestPath);

    // Prevent escaping the build directory
    if (!filePath.startsWith(buildPath)) {
      res.writeHead(403);
      res.end("Forbidden");
      return;
    }

    fs.readFile(filePath, (error, data) => {
      if (error) {
        // BrowserRouter routes need to fall back to index.html
        fs.readFile(path.join(buildPath, "index.html"), (indexError, indexData) => {
          if (indexError) {
            res.writeHead(500);
            res.end("Failed to load application");
            return;
          }

          res.writeHead(200, {
            "Content-Type": "text/html",
          });
          res.end(indexData);
        });

        return;
      }

      const extension = path.extname(filePath);

      const contentTypes = {
        ".html": "text/html",
        ".js": "text/javascript",
        ".css": "text/css",
        ".json": "application/json",
        ".png": "image/png",
        ".jpg": "image/jpeg",
        ".jpeg": "image/jpeg",
        ".svg": "image/svg+xml",
        ".ico": "image/x-icon",
        ".woff": "font/woff",
        ".woff2": "font/woff2",
        ".ttf": "font/ttf",
      };

      res.writeHead(200, {
        "Content-Type": contentTypes[extension] || "application/octet-stream",
      });

      res.end(data);
    });
  });

  return new Promise((resolve, reject) => {
    server.once("error", reject);

    server.listen(0, "localhost", () => {
      const address = server.address();

      if (typeof address === "object" && address !== null) {
        resolve(address.port);
      } else {
        reject(new Error("Failed to determine server port"));
      }
    });
  });
}

async function createWindow() {
  mainWindow = new BrowserWindow({
    width: 1400,
    height: 900,
    webPreferences: {
        preload: path.join(__dirname, "preload.cjs"),
        contextIsolation: true,
        nodeIntegration: false,
        sandbox: false,
    },
  });

  mpv.attachWindow(mainWindow);

  if (process.env.ELECTRON_DEV === "true") {
    await mainWindow.loadURL("http://localhost:3000");
  } else {
    const port = await startStaticServer();
    await mainWindow.loadURL(`http://localhost:${port}`);
  }
}

app.whenReady().then(createWindow);

app.on("before-quit", () => {
  void mpv.dispose();
});

app.on("window-all-closed", () => {
  if (server) {
    server.close();
  }

  if (process.platform !== "darwin") {
    app.quit();
  }
});

