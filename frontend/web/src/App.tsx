import React from "react";
import { BrowserRouter, Navigate, Route, Routes } from "react-router-dom";
import { useAuth0 } from "@auth0/auth0-react";
import PropTypes from "prop-types";
import Home from "./pages/Home";
import Login from "./components/Login";
import Logout from "./components/Logout";

function App(): React.ReactNode {
  const { isLoading, error, isAuthenticated } = useAuth0();

  if (error) {
    return (
      <div>
        <h1>Oops... {error.message}</h1>
      </div>
    );
  }

  if (isLoading) {
    return (
      <div>
        <h1>Loading...</h1>
      </div>
    );
  }

  return (
    <BrowserRouter>
      <Routes>
        <Route
          path="/"
          element={
            isAuthenticated ? <Home /> : <Navigate replace to="/login" />
          }
        ></Route>
        <Route path="/login" element={<Login />}></Route>
        <Route path="/logout" element={<Logout />}></Route>
      </Routes>
    </BrowserRouter>
  );
}

App.propTypes = {
  children: PropTypes.node,
};

export default App;
