const { app, BrowserWindow } = require("electron");
const http = require("http");
const fs = require("fs");
const path = require("path");

let mainWindow;
let server;

function startStaticServer() {
  const buildPath = path.join(__dirname, "../build");

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
  });

  if (process.env.ELECTRON_DEV === "true") {
    await mainWindow.loadURL("http://localhost:3000");
  } else {
    const port = await startStaticServer();
    await mainWindow.loadURL(`http://localhost:${port}`);
  }
}

app.whenReady().then(createWindow);

app.on("window-all-closed", () => {
  if (server) {
    server.close();
  }

  if (process.platform !== "darwin") {
    app.quit();
  }
});