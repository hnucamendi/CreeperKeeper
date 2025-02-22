import React from "react";
import { useAuth0 } from "@auth0/auth0-react";
import "../styles/components/login-btn.css";
import "../styles/components/landing-base.css";

export default function LoginButton(): React.ReactNode {
  const { loginWithRedirect } = useAuth0();

  return (
    <button
      className="landing-base-btn landing-login"
      onClick={() => loginWithRedirect()}
    >
      Login
    </button>
  );
}
