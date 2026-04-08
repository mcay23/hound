import { Container, Nav, Navbar } from "react-bootstrap";
import "./Topnav.css";
import ProfileButton from "./ProfileButton";

function Topnav() {
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
              Collections
            </Nav.Link>
            <Nav.Link className="top-navbar-item" href="/activity">
              Activity
            </Nav.Link>
          </Nav>
          <ProfileButton />
        </Navbar.Collapse>
      </Container>
    </Navbar>
  );
}

export default Topnav;
