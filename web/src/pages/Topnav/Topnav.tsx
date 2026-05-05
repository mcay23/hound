import { Container, Nav, Navbar } from "react-bootstrap";
import "./Topnav.css";
import ProfileButton from "./ProfileButton";
import { useServerInfo } from "../../api/hooks/general";
import { InfoRounded } from "@mui/icons-material";
import { Tooltip } from "@mui/material";

function Topnav() {
  const { data: serverInfo, isLoading: isServerInfoLoading } = useServerInfo();
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
            <Nav.Link className="top-navbar-item" href="/">
              Home
            </Nav.Link>
            <Nav.Link className="top-navbar-item" href="/library">
              Library
            </Nav.Link>
            <Nav.Link className="top-navbar-item" href="/activity">
              Activity
            </Nav.Link>
          </Nav>
          {localStorage.getItem("role") !== "admin" || isServerInfoLoading ? (
            <></>
          ) : serverInfo?.latest_version === serverInfo?.version ? (
            <></>
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
