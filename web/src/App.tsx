import { useEffect, useState } from "react";
import "./App.css";
import Login from "./pages/Login/Login";
import ServerUnreachableScreen from "./pages/ServerUnreachable/ServerUnreachableScreen";
import { checkServerReachable } from "./utils/serverHealth";
import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import "bootstrap/dist/css/bootstrap.min.css";
import Home from "./pages/Home/Home";
import Logout from "./pages/Logout";
import axios from "axios";
import MediaPageLanding from "./pages/MediaPage/MediaPageLanding";
import SearchPage from "./pages/Search/SearchPage";
import Library from "./pages/Library/Library";
import Collection from "./pages/Collection/Collection";
import { LocalizationProvider } from "@mui/x-date-pickers";
import { AdapterDayjs } from "@mui/x-date-pickers/AdapterDayjs";
import { AllCommunityModule, ModuleRegistry } from "ag-grid-community";
import { createTheme, ThemeProvider, CssBaseline } from "@mui/material";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { getBaseUrl, AXIOS_CONFIG } from "./config/axios_config";
import { getSecureToken, clearSecureToken } from "./utils/secureStore";
import { Toaster } from "react-hot-toast";
import Topnav from "./pages/Topnav/Topnav";
import Admin from "./pages/Admin/Admin";
import Activity from "./pages/Activity/Activity";
import Settings from "./pages/Settings/Settings";
import LiveTV from "./pages/LiveTV/LiveTV";
import { isPlatformElectron } from "./utils/platform";
import { Spinner } from "react-bootstrap";
const queryClient = new QueryClient();

// axios defaults
axios.defaults.withCredentials = true;
axios.defaults.baseURL = getBaseUrl();
// TODO REVISE LATER
axios.defaults.headers.common["Content-Type"] =
  AXIOS_CONFIG.headers["Content-Type"];
axios.defaults.headers.common["X-Client-Id"] =
  AXIOS_CONFIG.headers["X-Client-Id"];
axios.defaults.headers.common["X-Client-Platform"] =
  AXIOS_CONFIG.headers["X-Client-Platform"];
axios.defaults.headers.common["X-Device-Id"] =
  AXIOS_CONFIG.headers["X-Device-Id"];

axios.interceptors.request.use(
  async function (config) {
    const token = await getSecureToken();
    if (token) {
      if (config.headers && typeof config.headers.set === "function") {
        config.headers.set("Authorization", `Bearer ${token}`);
      } else {
        config.headers = config.headers || {};
        config.headers["Authorization"] = `Bearer ${token}`;
      }
    }
    return config;
  },
  function (error) {
    return Promise.reject(error);
  },
);

axios.interceptors.response.use(
  function (response) {
    if (
      response.data &&
      response.data.status === "success" &&
      response.data.data !== undefined
    ) {
      return { ...response, data: response.data.data };
    }
    return response;
  },
  async function (error) {
    console.log(error);
    const statusCode = error.response?.status;
    if (statusCode === 401) {
      await clearSecureToken();
      localStorage.removeItem("isAuthenticated");
      localStorage.removeItem("username");
      localStorage.removeItem("role");
      localStorage.removeItem("displayName");
      if (
        window.location.pathname !== "/logout" &&
        window.location.pathname !== "/login"
      ) {
        window.location.href = "/login";
      }
    }
    return Promise.reject(error);
  },
);
ModuleRegistry.registerModules([AllCommunityModule]);

function App() {
  const [serverStatus, setServerStatus] = useState<
    "checking" | "ok" | "unreachable"
  >(
    isPlatformElectron &&
      !!localStorage.getItem("isAuthenticated") &&
      !sessionStorage.getItem("server_reachable")
      ? "checking"
      : "ok",
  );

  const isAuthenticated = localStorage.getItem("isAuthenticated");

  useEffect(() => {
    getSecureToken();
    if (isPlatformElectron) {
      import("electron-mpv-video/renderer")
        .then(({ defineMpvVideoElement }) => defineMpvVideoElement())
        .catch((err) =>
          console.warn("Electron MPV element not initialized:", err),
        );
      if (
        localStorage.getItem("isAuthenticated") &&
        !sessionStorage.getItem("server_reachable")
      ) {
        getSecureToken().then((token) => {
          checkServerReachable(token ?? undefined).then((ok) => {
            if (ok) sessionStorage.setItem("server_reachable", "true");
            setServerStatus(ok ? "ok" : "unreachable");
          });
        });
      }
    }
  }, []);

  type ProtectedRouteProps = {
    component: JSX.Element;
  };
  function ProtectedRoute({ component }: ProtectedRouteProps) {
    if (!!isAuthenticated) {
      return component;
    } else {
      return <Navigate to={{ pathname: "/login" }} />;
    }
  }
  const theme = createTheme({
    typography: {
      fontFamily: '"Cabin", "Roboto", "Helvetica", "Arial", sans-serif',
    },
  });

  if (serverStatus === "checking")
    return (
      <div className="d-flex justify-content-center align-items-center vh-100">
        <Spinner />
      </div>
    );
  if (serverStatus === "unreachable")
    return <ServerUnreachableScreen onResolved={() => setServerStatus("ok")} />;

  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <Toaster
        toastOptions={{
          duration: 5000,
        }}
      />
      <LocalizationProvider dateAdapter={AdapterDayjs}>
        <QueryClientProvider client={queryClient}>
          <BrowserRouter>
            {!!isAuthenticated && <Topnav />}
            <Routes>
              <Route
                path="/"
                element={<ProtectedRoute component={<Home />} />}
              />
              <Route path="login" element={<Login />} />
              {/* <Route path="register" element={<Register />} /> */}
              <Route path="logout" element={<Logout />} />
              {localStorage.getItem("role") === "admin" && (
                <Route
                  path="admin"
                  element={<ProtectedRoute component={<Admin />} />}
                />
              )}
              <Route
                path="settings"
                element={<ProtectedRoute component={<Settings />} />}
              />
              <Route
                path="library"
                element={<ProtectedRoute component={<Library />} />}
              />
              <Route
                path="live-tv"
                element={<ProtectedRoute component={<LiveTV />} />}
              />
              <Route
                path="activity"
                element={<ProtectedRoute component={<Activity />} />}
              />
              <Route
                path="/tv/:id"
                element={<ProtectedRoute component={<MediaPageLanding />} />}
              />
              <Route
                path="/movie/:id"
                element={<ProtectedRoute component={<MediaPageLanding />} />}
              />
              <Route
                path="/game/:id"
                element={<ProtectedRoute component={<MediaPageLanding />} />}
              />
              <Route
                path="/search"
                element={<ProtectedRoute component={<SearchPage />} />}
              />
              <Route
                path="/collection/:id"
                element={<ProtectedRoute component={<Collection />} />}
              />
            </Routes>
          </BrowserRouter>
        </QueryClientProvider>
      </LocalizationProvider>
    </ThemeProvider>
  );
}

export default App;
