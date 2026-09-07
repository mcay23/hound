import { Container, Nav, Navbar } from "react-bootstrap";
import "./Topnav.css";
import ProfileButton from "./ProfileButton";
import { useServerInfo } from "../../api/hooks/general";
import { GitHub, InfoRounded } from "@mui/icons-material";
import { Tooltip } from "@mui/material";
import { useIPTVProviders } from "../../api/hooks/live_tv";
import { Link, NavLink } from "react-router-dom";

function Topnav() {
  const { data: serverInfo, isLoading: isServerInfoLoading } = useServerInfo();
  const { data: iptvProviders } = useIPTVProviders();
  return (
    <Navbar id="top-navbar" sticky="top" variant="dark" expand="sm">
      <Container fluid>
        <Navbar.Brand id="top-navbar-brand" href="/">
          <img
            src={`${import.meta.env.BASE_URL}hound-logo.png`}
            alt="Hound Logo"
            height="40"
          />
        </Navbar.Brand>
        <Navbar.Toggle
          id="top-navbar-toggle"
          aria-controls="basic-navbar-nav"
        />
        <Navbar.Collapse id="basic-navbar-nav">
          <Nav className="me-auto my-2 my-lg-0 text-light">
            <NavLink
              to="/"
              end
              className={({ isActive }) =>
                `top-navbar-item ${isActive ? "active" : ""}`
              }
            >
              Home
            </NavLink>
            <NavLink
              to="/library"
              className={({ isActive }) =>
                `top-navbar-item ${isActive ? "active" : ""}`
              }
            >
              Library
            </NavLink>
            {((iptvProviders?.length && iptvProviders.length > 0) ||
              localStorage.getItem("role") === "admin") && (
              <NavLink
                to="/live-tv"
                className={({ isActive }) =>
                  `top-navbar-item ${isActive ? "active" : ""}`
                }
              >
                Live TV
              </NavLink>
            )}
            <NavLink
              to="/activity"
              className={({ isActive }) =>
                `top-navbar-item ${isActive ? "active" : ""}`
              }
            >
              Activity
            </NavLink>
          </Nav>
          {localStorage.getItem("role") !== "admin" || isServerInfoLoading ? (
            <></>
          ) : serverInfo?.latest_version === serverInfo?.version ? (
            <a
              target="_blank"
              rel="noopener noreferrer"
              href="https://github.com/Hound-Media-Server/hound"
            >
              <GitHub sx={{ color: "#FFFFFF" }} className="mx-3" />
            </a>
          ) : (
            <a
              target="_blank"
              rel="noopener noreferrer"
              href="https://github.com/Hound-Media-Server/hound/releases"
            >
              <Tooltip
                title={
                  <p style={{ fontSize: "14px" }}>
                    Newer Version Available:{serverInfo?.latest_version}
                  </p>
                }
              >
                <InfoRounded sx={{ color: "#FFFF00" }} className="mx-3" />
              </Tooltip>
            </a>
          )}
          <ProfileButton />
        </Navbar.Collapse>
      </Container>
    </Navbar>
  );
}

export default Topnav;
