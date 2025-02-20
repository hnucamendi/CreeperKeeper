import React from "react";
import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { Auth0Provider } from "@auth0/auth0-react";
import "./index.css";
import App from "./App.tsx";

const root = createRoot(document.getElementById("root")!);
console.log("origin", window.location.origin);
root.render(
  <Auth0Provider
    domain="https://dev-bxn245l6be2yzhil.us.auth0.com"
    clientId="Ne5QmRSrbFuXW9p0ahbQUrIETB6lWhQL"
    authorizationParams={{
      redirect_uri: window.location.origin,
      audience: "creeperkeeper-resource",
      scope: "read:all",
    }}
  >
    <StrictMode>
      <App />
    </StrictMode>
  </Auth0Provider>,
);
