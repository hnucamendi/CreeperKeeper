import React from "react";
import { useAuth0 } from "@auth0/auth0-react";
import LogoutButton from "./LogoutButton.tsx";
import "../styles/components/creeperkeeper-nav-bar.css";

export default function CreeperKeeperNavBar(): React.ReactNode {
  const NavWrapper: React.FC<{ children: React.ReactNode }> = ({
    children,
  }) => <nav className="navbar">{children}</nav>;
  const { isAuthenticated } = useAuth0();

  if (isAuthenticated) {
    return (
      <NavWrapper>
        <div className="authenticated-nav-container">
          <h1>CreeperKeeper</h1>
          <div className="authenticated-nav-items">
            <LogoutButton />
            <img src="../assets/userIcon.svg" />
          </div>
        </div>
      </NavWrapper>
    );
  }

  return (
    <NavWrapper>
      <h1>CreeperKeeper</h1>
    </NavWrapper>
  );
}
