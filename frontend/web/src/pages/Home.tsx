import React, { useState, useEffect } from "react";
import { useAuth0 } from "@auth0/auth0-react";
import LoginButton from "../components/LoginButton";
import CreeperKeeperNavBar from "../components/CreeperKeeperNavBar";
import ServerInstance from "../components/ServerInstance";
import { HTMLFormMethod } from "react-router-dom";
import "../styles/pages/home.css";

export interface Server {
  serverID: string;
  row: string;
  serverIP: string;
  serverName: string;
  lastUpdated: string;
  isRunning: boolean;
}

export default function Home(): React.ReactNode {
  const { isAuthenticated, getAccessTokenSilently } = useAuth0();
  const baseURL = "https://api.creeperkeeper.com";
  const [servers, setServers] = useState<Array<Server> | null>(null);
  const [token, setToken] = useState<string | null>(null);
  const [serverStatus, setServerStatus] = useState<string | null>(null);
  const [startLoading, setStartLoading] = useState<boolean>(false);
  const [stopLoading, setStopLoading] = useState<boolean>(false);
  const [serverStateChange, setServerStateChange] = useState<number>(0);

  useEffect(() => {
    const gat = async () => {
      return await getAuthToken();
    };
    if (isAuthenticated) gat();
  }, [isAuthenticated, getAccessTokenSilently]);

  useEffect(() => {
    const ls = async () => {
      return await listServers();
    };
    if (token) ls();
  }, [token, serverStateChange]);

  const getAuthToken = async () => {
    try {
      const accessToken = await getAccessTokenSilently();
      setToken(`Bearer ${accessToken}`);
    } catch (error) {
      console.error("Error getting access token:", error);
    }
  };

  const refreshServer = async (serverID: string): Promise<void> => {
    const url = new URL(baseURL + `/server/ping/${serverID}`);
    const req = await buildRequest(url, "GET");
    try {
      const res = await fetch(req);
      if (!res.ok)
        throw new Error(
          `Error refreshing sever status; response: ${res.status}`,
        );
      const resJson: string = await res.json();
      setServerStatus(resJson);
    } catch (error: unknown) {
      console.error(error);
    }
  };

  const listServers = async (): Promise<void> => {
    const url = new URL(baseURL + "/server/list");
    const storedETag = localStorage.getItem("servers_etag");
    const req = await buildRequest(url, "GET", storedETag);
    try {
      const res = await fetch(req);
      if (res.status === 304) {
        const cachedData = localStorage.getItem("servers");
        if (cachedData) setServers(JSON.parse(cachedData));
        return;
      }
      if (!res.ok)
        throw new Error(`Error fetching list of servers ${res.status}`);

      const resJson: Array<Server> = await res.json();
      setServers(resJson);

      const newETag = res.headers.get("etag");
      if (newETag) localStorage.setItem("servers_etag", newETag);
      localStorage.setItem("servers", JSON.stringify(resJson));
    } catch (error: unknown) {
      console.error((error as Error).message);
    }
  };

  const startServer = async (serverID: string): Promise<void> => {
    setStartLoading(true);
    const url = new URL(baseURL + "/server/start");
    const req = await buildRequest(
      url,
      "POST",
      null,
      JSON.stringify({
        serverID: serverID,
      }),
    );
    try {
      await fetch(req);
    } catch (error) {
      throw new Error(`Failed to start server ${serverID} Error: ${error} `);
    } finally {
      setServerStateChange((prev) => prev + 1);
      setStartLoading(false);
    }
  };

  const stopServer = async (serverID: string) => {
    setStopLoading(true);
    const url = new URL(baseURL + "/server/stop");
    const req = await buildRequest(
      url,
      "POST",
      null,
      JSON.stringify({
        serverID: serverID,
      }),
    );

    try {
      await fetch(req);
    } catch (error) {
      throw new Error(`Failed to stop server ${serverID} Error: ${error} `);
    } finally {
      setServerStateChange((prev) => prev + 1);
      setStopLoading(false);
    }
  };

  const buildRequest = async (
    url: URL,
    method: HTMLFormMethod,
    etag?: string | null,
    body?: BodyInit,
  ): Promise<Request> => {
    if (!token) {
      try {
        await getAuthToken();
      } catch (error) {
        throw new Error(
          `Failed to get AuthToken when building request ${error}`,
        );
      }
    }

    if (!token) {
      throw new Error(
        "AuthToken is still missing after attempting to fetch new token",
      );
    }

    return new Request(url, {
      method: method,
      headers: {
        "Content-Type": "application/json",
        Authorization: token,
        ...(etag ? { "If-None-Match": etag } : {}),
      },
      ...(body != null && method !== "GET" ? { body } : {}),
    });
  };

  if (!isAuthenticated) {
    <LoginButton />;
    return <h1>✋ Please log in ✋</h1>;
  }

  if (servers === null) {
    return (
      <main>
        <h1>Available servers are loading... 🐌</h1>
      </main>
    );
  }

  return (
    <main>
      <CreeperKeeperNavBar />
      <div className="server-container">
        <ServerInstance
          serverList={servers}
          startState={startLoading}
          stopState={stopLoading}
          serverStatus={serverStatus || ""}
          startServer={startServer}
          stopServer={stopServer}
          refreshServer={refreshServer}
        />
      </div>
    </main>
  );
}
