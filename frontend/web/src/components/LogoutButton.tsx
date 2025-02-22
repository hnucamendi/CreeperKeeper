import { useAuth0 } from "@auth0/auth0-react";
import React from "react";
import "../styles/components/logout-btn.css";
import "../styles/components/landing-base.css";

export default function LogoutButton(): React.ReactNode {
  const { logout } = useAuth0();

  return (
    <button
      className="landing-base-btn landing-logout"
      onClick={() =>
        logout({ logoutParams: { returnTo: window.location.origin } })
      }
    >
      Logout
    </button>
  );
}
