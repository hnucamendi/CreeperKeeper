import React from "react";
import LoginButton from "../components/LoginButton";
import SignupButton from "../components/SignupButton";
import CreeperKeeperNavBar from "../components/CreeperKeeperNavBar";
import "../styles/pages/login.css";

export default function Login(): React.ReactNode {
  return (
    <div className="landing-page-base">
      <CreeperKeeperNavBar />
      <div className="landing-hero-container">
        <h2>
          Spin up Minecraft servers quickly and only pay for what you use!
        </h2>
        <div className="btn-group">
          <SignupButton />
          <LoginButton />
        </div>
      </div>
    </div>
  );
}
