import React, { useState, useEffect } from "react";
import { useAuth0 } from "@auth0/auth0-react";
import Logout from "../components/Logout";
import Login from "../components/Login";
import { HTMLFormMethod } from "react-router-dom";

interface Server {
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
  const [pageLoading, setPageLoading] = useState<boolean>(false);
  const [startLoading, setStartLoading] = useState<boolean>(false);
  const [stopLoading, setStopLoading] = useState<boolean>(false);
  const [serverStateChange, setServerStateChange] = useState<number>(0);
  const THREE_MINUTES: number = 60 * 3000;
  const ONE_MINUTE: number = 60 * 1000;

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

  const sleep = async (time: number): Promise<void> =>
    new Promise((resolve) => setTimeout(resolve, time));

  const listServers = async (): Promise<void> => {
    setPageLoading(true);
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
    } finally {
      setPageLoading(false);
    }
  };

  const startServer = async (serverID: string) => {
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
      await sleep(THREE_MINUTES);
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
      await sleep(ONE_MINUTE);
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
    <Login />;
    return <h1>✋ Please log in ✋</h1>;
  }

  if (pageLoading || servers === null) {
    return (
      <main>
        <h1>Available servers are loading... 🐌</h1>
      </main>
    );
  }

  return (
    <main>
      <Logout />
      <h1> 🌚 Welcome 🎃</h1>
      <button onClick={listServers}>trigger list servers</button>
      {servers.map((v: Server) => (
        <div key={v.serverID}>
          <span>{v.row}</span>
          <span>Last updated: {v.lastUpdated}</span>
          <h2>{v.serverName}</h2>
          <p>{v.serverID}</p>
          <p>{v.isRunning ? v.serverIP : `Last server IP: ${v.serverIP}`}</p>
          <p>Status: {v.isRunning ? "RUNNING" : "STOPPED"}</p>
          <button
            onClick={() => startServer(v.serverID)}
            disabled={startLoading || stopLoading}
          >
            {startLoading ? "Starting..." : "Start"}
          </button>
          <button
            onClick={() => stopServer(v.serverID)}
            disabled={startLoading || stopLoading}
          >
            {stopLoading ? "Stopping..." : "Stop"}
          </button>
        </div>
      ))}
    </main>
  );
}
