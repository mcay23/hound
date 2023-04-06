import "./App.css";
import Login from "./pages/Login/Login";
import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import "bootstrap/dist/css/bootstrap.min.css";
import Home from "./pages/Home/Home";
import Logout from "./pages/Logout";
import houndConfig from "./config.json";
import axios from "axios";
import MediaPageLanding from "./pages/MediaPage/MediaPageLanding";
import SearchPage from "./pages/Search/SearchPage";
import Library from "./pages/Library/Library";
import Collection from "./pages/Collection/Collection";
import { LocalizationProvider } from "@mui/x-date-pickers";
import { AdapterDayjs } from "@mui/x-date-pickers/AdapterDayjs";

function App() {
  var isAuthenticated = localStorage.getItem("isAuthenticated");
  // axios defaults
  axios.defaults.withCredentials = true;
  axios.defaults.baseURL = houndConfig.server_host;
  // TODO REVISE LATER
  axios.defaults.headers.common["Content-Type"] =
    houndConfig.axios_config.headers["Content-Type"];
  axios.defaults.headers.common["X-Client"] =
    houndConfig.axios_config.headers["X-Client"];
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
  return (
    <LocalizationProvider dateAdapter={AdapterDayjs}>
      <BrowserRouter>
        <Routes>
          <Route path="/" element={<ProtectedRoute component={<Home />} />} />
          <Route path="login" element={<Login />} />
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
    </LocalizationProvider>
  );
}

export default App;
