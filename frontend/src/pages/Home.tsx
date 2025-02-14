/* eslint-disable @typescript-eslint/no-explicit-any */
import { useState, useEffect } from "react";
import { useAuth0 } from "@auth0/auth0-react";
import Logout from "../component/Logout";


export default function Home() {
  const { isAuthenticated, getAccessTokenSilently } = useAuth0();
  const ck_url = "https://app.creeperkeeper.com";
  const [currentInstance, setCurrentInstance] = useState("");
  const [start, setStart] = useState({ ip: "", success: "" })
  const [stop, setStop] = useState("")
  const [instances, setInstances] = useState([]);
  const [authToken, setAuthToken] = useState("");
  const [loading, setLoading] = useState(false);



  useEffect(() => {
    setStart({ ip: "", success: "" });
    setStop("");
  }, [currentInstance])

  useEffect(() => {
    const getAuthToken = async () => {
      try {
        const accessToken = await getAccessTokenSilently();

        setAuthToken(`Bearer ${accessToken}`);
      } catch (e) {
        console.error("Error getting access token:", e);
      }
    };

    if (isAuthenticated) {
      getAuthToken();
    }
  });

  useEffect(() => {
    const fetchInstances = async () => {
      const path = "getInstances";
      const url = `${ck_url}/${path}`;

      const body = {
        method: "GET",
        headers: {
          "Content-Type": "application/json",
          "Authorization": authToken,
        },
      };

      try {
        const req = await fetch(url, body);
        if (!req.ok) {
          throw new Error(`HTTP error! status: ${req.status}`);
        }
        setLoading(true)
        const res = await req.json() // Parse the response as JSON
        setLoading(false)
        setInstances(res.message);
      } catch (error) {
        setLoading(false)
        setInstances([]);
        console.error("Error getting instances:", error);
      }
    };

    if (authToken) {
      fetchInstances();
    }
  }, [authToken]);

  const handleSetInstance = (e: any) => {
    setCurrentInstance(e.target.value)
  }

  const handleStartMCServer = async () => {
    setStart({ ip: "", success: "" });
    const path = "start";

    const url = `${ck_url}/${path}`;

    const body = {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Authorization": authToken,
      },
      body: JSON.stringify({
        instanceID: currentInstance,
      }),
    }

    try {
      setLoading(true)
      const req = await fetch(url, body);
      if (!req.ok) {
        throw new Error(`HTTP error! status: ${req.status}`);
      }
      const res = await req.json(); // Parse the response as JSON
      setLoading(false)
      const resJSON = JSON.parse(res.message).message
      setStart({ ip: resJSON.ip, success: resJSON.success })
    } catch (error) {
      setLoading(false)
      console.error("Error starting the server:", error);
      setStart({ ip: "", success: "Error starting the server" })
    }
  };

  const handleStopMCServer = async () => {
    setStop("");
    const path = "stop";

    const url = `${ck_url}/${path}`;

    const body = {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Authorization": authToken,
      },
      body: JSON.stringify({
        instanceID: currentInstance,
      }),
    }

    try {
      setLoading(true)
      const req = await fetch(url, body);
      if (!req.ok) {
        throw new Error(`HTTP error! status: ${req.status}`);
      }
      const res = await req.json(); // Parse the response as JSON
      setLoading(false)
      setStop(res.message)
    } catch (error) {
      console.error("Error stopping the server:", error);
      setStop("Error server is probably already stopped")
    }
  };

  return (
    isAuthenticated && (
      <div className="grid grid-rows-[20px_1fr_20px] items-center justify-items-center min-h-screen p-4 sm:p-8 pb-20 gap-8 sm:gap-16 font-[family-name:var(--font-geist-sans)]">
        <Logout />
        <main className="flex flex-col sm:flex-row gap-8 row-start-2 items-center sm:items-start w-full max-w-4xl">
          <div className="w-full sm:w-1/2 p-4 border rounded shadow">
            <h1 className="text-xl font-bold mb-4">Select Instance</h1>
            <select
              onChange={handleSetInstance}
              className="w-full p-2 border rounded"
            >
              <option value="">None</option>
              {instances.map((instance, i) => (
                <option key={i} value={instance}>
                  {instance}
                </option>
              ))}
            </select>
          </div>
          <div className="w-full sm:w-1/2 p-4 border rounded shadow">
            <h1 className="text-xl font-bold mb-4">Manage Server</h1>
            {loading ? <h2 className="text-xl font-bold mb-4">Loading...</h2> : null}
            <div>
              <h2 className="text-lg mb-4">{`IP Address: ${start.ip}, ${start.success}`}</h2>
              <button
                className={`${currentInstance
                  ? "bg-blue-500 hover:bg-blue-700 text-white"
                  : "bg-gray-500 text-gray-300 cursor-not-allowed"
                  } py-2 px-4 rounded mb-2 w-full`}
                onClick={handleStartMCServer}
                disabled={!currentInstance}
              >
                Start
              </button>
            </div>

            <div>
              <h2 className="text-lg mb-4">{stop}</h2>
              <button
                className={`${currentInstance
                  ? "bg-blue-500 hover:bg-blue-700 text-white"
                  : "bg-gray-500 text-gray-300 cursor-not-allowed"
                  } py-2 px-4 rounded w-full`}
                onClick={handleStopMCServer}
                disabled={!currentInstance}
              >
                Stop
              </button>
            </div>
          </div>
        </main>
      </div>
    )
  );
}

