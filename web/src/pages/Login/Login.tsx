import React, { useState } from "react";
import { Card, Button, FormGroup, FormControl } from "react-bootstrap";
import "./Login.css";
import houndConfig from "../../config.json";
import axios from "axios";
import { Navigate } from "react-router-dom";

function Login() {
  const [data, setData] = useState({
    username: "",
    password: "",
  });

  const [alertVisible, setAlertVisible] = useState(false);

  if (!!localStorage.getItem("isAuthenticated")) {
    return <Navigate to="/" />;
  }

  const submitHandler = (event: React.FormEvent<HTMLButtonElement>) => {
    event.preventDefault();
    axios
      .post("/api/v1/auth/login", data)
      .then((res) => {
        console.log("RESPONSE RECEIVED: ", res.data);
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
            <h2 className="mb-4">Login</h2>
            <form>
              <FormGroup controlId="username" className="mt-4">
                <FormControl
                  autoFocus
                  type="username"
                  name="username"
                  placeholder="username"
                  value={data.username}
                  onChange={handleChange}
                />
              </FormGroup>
              <FormGroup className="mt-4" controlId="password">
                <FormControl
                  type="password"
                  name="password"
                  placeholder="password"
                  value={data.password}
                  onChange={handleChange}
                />
              </FormGroup>
              <br />
              {alertVisible ? (
                <div className="d-flex mx-auto">
                  <p className="mx-auto mt-1 alert-incorrect-password">
                    Incorrect username or password
                  </p>
                </div>
              ) : null}
              <div className="d-flex flex-row-reverse">
                <Button type="submit" onClick={submitHandler}>
                  Login
                </Button>
              </div>
            </form>
            <div className="d-flex mx-auto mt-4">
              <a
                href="../signup"
                className="mx-auto"
                style={{ textDecoration: "underline" }}
              >
                Don't have an account? Sign up here!
              </a>
            </div>
          </div>
        </Card>
      </div>
    </div>
  );
}

export default Login;
