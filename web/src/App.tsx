import "./App.css";
import Login from "./pages/Login/Login";
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
import Register from "./pages/Login/Register";
import { AllCommunityModule, ModuleRegistry } from "ag-grid-community";
import { createTheme, ThemeProvider, CssBaseline } from "@mui/material";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { SERVER_URL, AXIOS_CONFIG } from "./config/axios_config";

const queryClient = new QueryClient();

// axios defaults
axios.defaults.withCredentials = true;
axios.defaults.baseURL = SERVER_URL;
// TODO REVISE LATER
axios.defaults.headers.common["Content-Type"] =
  AXIOS_CONFIG.headers["Content-Type"];
axios.defaults.headers.common["X-Client"] = AXIOS_CONFIG.headers["X-Client"];
// Add a request interceptor
axios.interceptors.request.use(
  function (config) {
    return config;
  },
  function (error) {
    return Promise.reject(error);
  }
);
// Add a response interceptor
axios.interceptors.response.use(
  function (response) {
    return response;
  },
  function (error) {
    console.log(error);
    const statusCode = error.response.status;
    if (statusCode === 401) {
      console.log("logging out");
      const win: Window = window;
      win.location = "/logout";
    }
    return Promise.reject(error);
  }
);
ModuleRegistry.registerModules([AllCommunityModule]);

function App() {
  var isAuthenticated = localStorage.getItem("isAuthenticated");
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

  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <LocalizationProvider dateAdapter={AdapterDayjs}>
        <QueryClientProvider client={queryClient}>
          <BrowserRouter>
            <Routes>
              <Route
                path="/"
                element={<ProtectedRoute component={<Home />} />}
              />
              <Route path="login" element={<Login />} />
              <Route path="register" element={<Register />} />
              <Route
                path="logout"
                element={<ProtectedRoute component={<Logout />} />}
              />
              <Route
                path="library"
                element={<ProtectedRoute component={<Library />} />}
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
