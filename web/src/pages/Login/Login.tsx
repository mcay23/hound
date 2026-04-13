import React, { useState } from "react";
import { Card, Button, FormGroup, FormControl } from "react-bootstrap";
import "./Login.css";
import axios from "axios";
import { Navigate } from "react-router-dom";
import toast from "react-hot-toast";

function Login() {
  const [data, setData] = useState({
    username: "",
    password: "",
  });

  if (!!localStorage.getItem("isAuthenticated")) {
    return <Navigate to="/" />;
  }

  const submitHandler = (event: React.FormEvent<HTMLButtonElement>) => {
    event.preventDefault();
    axios
      .post("/api/v1/auth/login", data)
      .then((res) => {
        localStorage.setItem("username", res.data.username);
        localStorage.setItem("isAuthenticated", "true");
        localStorage.setItem("role", res.data.role);
        localStorage.setItem("displayName", res.data.display_name);
        window.location.reload();
      })
      .catch((err) => {
        if (err.response?.status === 401 || err.response?.status === 404) {
          toast.error("Incorrect username/password");
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
              <div className="d-flex flex-row-reverse">
                <Button type="submit" onClick={submitHandler}>
                  Login
                </Button>
              </div>
            </form>
          </div>
        </Card>
      </div>
    </div>
  );
}

export default Login;
