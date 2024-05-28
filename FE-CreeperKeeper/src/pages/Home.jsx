//import React, { useState } from "react";
//
//export default function LandingPage() {
//  const [isAuthorized, setIsAuthorized] = useState(true);
//
//  if (isAuthorized) {
//    return (
//      <div className="server-menu">
//        <div>
//          <form>
//            <label htmlFor="server-dropdown">Select Version</label>
//            <select id="server-dropdown"></select>
//            <button>Create Server</button>
//          </form>
//        </div>
//
//        <div>
//          <form>
//            <label htmlFor="server-dropdown">Select Saved Instance</label>
//            <select id="server-dropdown"></select>
//            <button>Startup Server</button>
//          </form>
//        </div>
//      </div>
//    );
//  }
//
//  return <Authenticator loginMechanism={["email"]} />;
//}

import { useNavigate } from "react-router-dom";

const Home = () => {
  const navigate = useNavigate();
  const handleLogout = () => {
    sessionStorage.clear();
    navigate("/login");
  };

  return (
    <div>
      <h1>Hello World</h1>
      <p>See console log for Amazon Cognito user tokens.</p>
      <button onClick={handleLogout}>Logout</button>
    </div>
  );
};

export default Home;
