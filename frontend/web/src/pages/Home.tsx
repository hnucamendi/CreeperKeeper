import React from "react";
import { useState, useEffect } from "react";
import { useAuth0 } from "@auth0/auth0-react";
import Logout from "../components/Logout";
import Login from "../components/Login";

interface Server {
  ID: string;
  SK: string;
  IP: string;
  Name: string;
  LastUpdated: string;
  IsRunning: boolean;
}

export default function Home(): React.ReactNode {
  const { isAuthenticated, getAccessTokenSilently } = useAuth0();
  const baseURL = "https://api.creeperkeeper.com";

  const [server, setServer] = useState<Server | null>(null);
  const [token, setToken] = useState<string | null>(null);
  const [loading, setLoading] = useState<boolean>(false);

  useEffect(() => {
    const getAuthToken = async () => {
      try {
        const accessToken = await getAccessTokenSilently();
        setToken(`Bearer ${accessToken}`);
      } catch (error) {
        console.error("Error getting access token:", error);
      }
    };

    if (isAuthenticated) getAuthToken();
  }, [isAuthenticated, getAccessTokenSilently]);

  useEffect(() => {
    const listServers = async () => {
      const path = "/servers/list";
      const url = baseURL + path;

      const body = {
        method: "GET",
        header: {
          "Content-Type": "application/json",
          Authorization: token || "",
        },
      };

      try {
        setLoading(true);
        const res = await fetch(url, body);
        if (!res.ok)
          throw new Error(`Error fetching list of servers ${res.status}`);
        const resJson: Server = await res.json();
        setServer(resJson);
      } catch (error: unknown) {
        setLoading(false);
        console.error((error as Error).message);
      } finally {
        setLoading(false);
      }
    };
    if (token) listServers();
  }, [token]);

  if (!isAuthenticated) {
    <Login />;
    return <h1>✋ Please log in ✋</h1>;
  }

  if (loading) {
    return (
      <main>
        <h1>Available servers are loading... 🐌</h1>
      </main>
    );
  }

  console.log(server);
  return (
    <main>
      <Logout />
      <h1> 🌚 Welcome 🎃</h1>
    </main>
  );
}
