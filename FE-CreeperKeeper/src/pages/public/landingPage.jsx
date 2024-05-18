import React, { useState } from "react";
import { Authenticator } from "@aws-amplify/ui-react";

export default function LandingPage() {
  const [isAuthorized, setIsAuthorized] = useState(false);

  if (isAuthorized) {
    return (
      <div className="server-menu">
        <div>
          <form>
            <label htmlFor="server-dropdown">Select Version</label>
            <select id="server-dropdown"></select>
            <button>Create Server</button>
          </form>
        </div>

        <div>
          <form>
            <label htmlFor="server-dropdown">Select Saved Instance</label>
            <select id="server-dropdown"></select>
            <button>Startup Server</button>
          </form>
        </div>
      </div>
    );
  }

  return <Authenticator loginMechanism={["email"]} />;
}
