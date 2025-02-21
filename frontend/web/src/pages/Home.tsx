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

  const [server, setServer] = useState<Array<Server> | null>(null);
  const [token, setToken] = useState<string | null>(null);
  const [loading, setLoading] = useState<boolean>(false);
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

  const listServers = async () => {
    setLoading(true);
    const url = new URL(baseURL + "/server/list");
    const req = await buildRequest(url, "GET");
    try {
      const res = await fetch(req);
      if (!res.ok)
        throw new Error(`Error fetching list of servers ${res.status}`);
      const resJson: Array<Server> = await res.json();
      setServer(resJson);
    } catch (error: unknown) {
      console.error((error as Error).message);
    } finally {
      setLoading(false);
    }
  };

  const startServer = async (serverID: string) => {
    const url = new URL(baseURL + "/server/start");
    const req = await buildRequest(
      url,
      "POST",
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
    }
  };

  const stopServer = async (serverID: string) => {
    const url = new URL(baseURL + "/server/stop");
    const req = await buildRequest(
      url,
      "POST",
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
    }
  };

  const buildRequest = async (
    url: URL,
    method: HTMLFormMethod,
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
      },
      ...(body != null && method !== "GET" ? { body } : {}),
    });
  };

  if (!isAuthenticated) {
    <Login />;
    return <h1>✋ Please log in ✋</h1>;
  }

  if (loading || server === null) {
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
      {server.map((v: Server, i: number) => (
        <div key={i}>
          <span>{v.row}</span>
          <span>Last updated: {v.lastUpdated}</span>
          <h2>{v.serverName}</h2>
          <p>{v.serverID}</p>
          <p>{v.isRunning ? v.serverIP : `Last server IP: ${v.serverIP}`}</p>
          <p>Status: {v.isRunning ? "RUNNING" : "STOPPED"}</p>
          <button onClick={() => startServer(v.serverID)}>Start</button>
          <button onClick={() => stopServer(v.serverID)}>Stop</button>
        </div>
      ))}
    </main>
  );
}
