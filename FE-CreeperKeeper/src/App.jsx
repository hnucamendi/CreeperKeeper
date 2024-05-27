import { BrowserRouter, Navigate, Route, Routes } from "react-router-dom";
import Login from "./pages/public/Login.jsx";
import Home from "./pages/Home.jsx";
import ConfirmUser from "./pages/public/ConfirmUser.jsx";

function App() {
  const isAuthenticated = () => {
    const accessToken = sessionStorage.getItem("accessToken");
    return !!accessToken;
  };

  return (
    <BrowserRouter>
      <Routes>
        <Route
          path="/"
          element={
            isAuthenticated() ? (
              <Navigate replace to="/home" />
            ) : (
              <Navigate replace to="/login" />
            )
          }
        />
        <Route path="/login" element={<Login />} />
        <Route path="/confirm" element={<ConfirmUser />} />
        <Route
          path="/home"
          element={
            isAuthenticated() ? <Home /> : <Navigate replace to="/login" />
          }
        />
      </Routes>
    </BrowserRouter>
  );
}

export default App;
