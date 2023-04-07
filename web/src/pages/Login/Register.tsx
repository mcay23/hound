import React, { useState } from "react";
import { Card, Button, FormGroup, FormControl } from "react-bootstrap";
import "./Login.css";
import axios from "axios";
import { Navigate } from "react-router-dom";

function Register() {
  const [data, setData] = useState({
    username: "",
    password: "",
    first_name: "",
    last_name: "",
  });

  const [alertVisible, setAlertVisible] = useState(false);

  if (!!localStorage.getItem("isAuthenticated")) {
    return <Navigate to="/" />;
  }

  const submitHandler = (event: React.FormEvent<HTMLButtonElement>) => {
    event.preventDefault();
    if (data.password.length < 8) {
      alert("Password >8 chars required");
    }
    axios
      .post("/api/v1/auth/register", data)
      .then((res) => {
        localStorage.setItem("username", res.data.username);
        localStorage.setItem("isAuthenticated", "true");
        window.location.reload();
        setAlertVisible(false);
      })
      .catch((err) => {
        if (err.response.status === 400) {
          setAlertVisible(true);
        }
        console.log("AXIOS ERROR: ", err);
      });
  };

  const handleChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    setData({ ...data, [event.target.name]: event.target.value });
  };

  return (
    <div className="full-screen bg-home">
      <div className="login-main">
        <Card className="login-card shadow p-3 mb-5 bg-white rounded">
          <div className="login-card">
            <h2 className="mb-4">Register</h2>
            <form>
              <FormControl
                type="first_name"
                name="first_name"
                placeholder="First Name"
                className="mt-4"
                value={data.first_name}
                onChange={handleChange}
              />
              <FormControl
                type="last_name"
                name="last_name"
                placeholder="Last Name"
                className="mt-4"
                value={data.last_name}
                onChange={handleChange}
              />
              <FormGroup controlId="username" className="mt-4">
                <FormControl
                  name="username"
                  placeholder="username"
                  type="username"
                  value={data.username}
                  autoComplete="off"
                  onChange={handleChange}
                />
              </FormGroup>
              <FormGroup className="mt-4" controlId="password">
                <FormControl
                  name="password"
                  placeholder="password"
                  type="password"
                  autoComplete="new-password"
                  value={data.password}
                  onChange={handleChange}
                />
              </FormGroup>
              <br />
              {alertVisible ? (
                <div className="d-flex mx-auto">
                  <p className="mx-auto mt-1 alert-incorrect-password">
                    Registration is disabled by the administrator.
                  </p>
                </div>
              ) : null}
              <div className="d-flex flex-row-reverse">
                <Button type="submit" onClick={submitHandler}>
                  Register
                </Button>
              </div>
            </form>
            <div className="d-flex mx-auto mt-4">
              <a
                href="../login"
                className="mx-auto"
                style={{ textDecoration: "underline" }}
              >
                Already have an account? Login here!
              </a>
            </div>
          </div>
        </Card>
      </div>
    </div>
  );
}

export default Register;
