import { Container, Nav, Navbar } from "react-bootstrap";
import "./Topnav.css";

function Topnav() {
  return (
    <Navbar id="top-navbar" sticky="top" variant="dark" expand="sm">
      <Container fluid>
        <Navbar.Brand id="top-navbar-brand" href="/">
          HOUND
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
            <Nav.Link className="top-navbar-item" href="/tvshows">
              TV Shows
            </Nav.Link>
            <Nav.Link className="top-navbar-item" href="/movies">
              Movies
            </Nav.Link>
            <Nav.Link className="top-navbar-item" href="/books">
              Books
            </Nav.Link>
            <Nav.Link className="top-navbar-item" href="/games">
              Games
            </Nav.Link>
          </Nav>
          <Nav.Link className="top-navbar-item me-3 mt-2 mb-2" href="/logout">
            Logout
          </Nav.Link>
        </Navbar.Collapse>
      </Container>
    </Navbar>
  );
}

export default Topnav;
