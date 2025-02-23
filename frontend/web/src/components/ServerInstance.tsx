import React from "react";
import { Server } from "../pages/Home";
import "../styles/components/server-instance.css";
import "../styles/components/btn-base.css";

interface ServerInstanceProps {
  serverList: Array<Server>;
  startState: boolean;
  stopState: boolean;
  startServer: (serverID: string) => Promise<void>;
  stopServer: (serverID: string) => Promise<void>;
  refreshServer: (serverID: string) => Promise<string>;
}

export default function ServerInstance({
  serverList,
  startState,
  stopState,
  startServer,
  stopServer,
  refreshServer,
}: ServerInstanceProps): React.ReactNode {
  let serverStatus: Promise<string> | null = null;
  return (
    <>
      {serverList.map((v: Server) => (
        <div key={v.serverID}>
          <div className="server-detail-container">
            <div className="server-btn-group">
              <button
                className="btn-base"
                onClick={() => (serverStatus = refreshServer(v.serverID))}
              >
                Refresh
              </button>
              <button
                className="btn-base"
                onClick={() => startServer(v.serverID)}
                disabled={v.isRunning || startState || stopState}
              >
                {(serverStatus ?? startState) ? "Starting..." : "Start"}
              </button>
              <button
                className="btn-base"
                onClick={() => stopServer(v.serverID)}
                disabled={!v.isRunning || startState || stopState}
              >
                {(serverStatus ?? stopState) ? "Stopping..." : "Stop"}
              </button>
            </div>

            <div className="server-detail-group">
              <p>Last updated: {v.lastUpdated}</p>
              <p>Server Name: {v.serverName}</p>
              <p>Server ID: {v.serverID}</p>
              <p>
                {v.isRunning
                  ? `Server IP: ${v.serverIP}`
                  : `Last server IP: ${v.serverIP}`}
              </p>
              <p>Status: {v.isRunning ? "RUNNING" : "STOPPED"}</p>
            </div>
          </div>
        </div>
      ))}
    </>
  );
}
