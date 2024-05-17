import { BrowserRouter, Route, Routes } from "react-router-dom";
import { Authenticator } from "@aws-amplify/ui-react";
import { Amplify } from "aws-amplify";
import UserRouter from "./routes.jsx";
import "./App.css";

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route
          path="/login"
          element={
            <>
              <Authenticator loginMechanism="email" />
              <UserRouter />
            </>
          }
        />
      </Routes>
    </BrowserRouter>
  );
}

export default App;
