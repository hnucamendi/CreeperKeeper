import { useAuth0 } from "@auth0/auth0-react";
import React from "react";

export default function Logout(): React.ReactNode {
  const { logout } = useAuth0();

  return (
    <button
      onClick={() =>
        logout({ logoutParams: { returnTo: window.location.origin } })
      }
    >
      Logout
    </button>
  );
}
