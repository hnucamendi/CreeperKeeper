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
  const [addInstance, setAddInstance] = useState("");
  const [newInstanceID, setNewInstanceID] = useState("");
  const [authToken, setAuthToken] = useState("");


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
        const res = await req.json() // Parse the response as JSON
        setInstances(res.message);
      } catch (error) {
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

  const handleAddInstance = async (e: any) => {
    e.preventDefault();
    setAddInstance("");
    const path = "addInstance";
    const url = `${ck_url}/${path}`;

    const body = {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Authorization": authToken,
      },
      body: JSON.stringify({
        instanceID: newInstanceID,
      }),
    };

    try {
      const req = await fetch(url, body);
      if (!req.ok) {
        throw new Error(`HTTP error! status: ${req.status}`);
      }
      const res = await req.json() // Parse the response as JSON
      setAddInstance(res.message);
    } catch (error) {
      setAddInstance("Error adding the instance")
      console.error("Error getting instances:", error);
    }
  }

  const handleStartMCServer = async () => {
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
      const req = await fetch(url, body);
      if (!req.ok) {
        throw new Error(`HTTP error! status: ${req.status}`);
      }
      const res = await req.json(); // Parse the response as JSON
      console.log("Start response:", res)
      setStart({ ip: res.message.ip, success: res.message.success })
    } catch (error) {
      console.error("Error starting the server:", error);
      setStart({ ip: "", success: "Error starting the server" })
    }
  };

  const handleStopMCServer = async () => {
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
      const req = await fetch(url, body);
      if (!req.ok) {
        throw new Error(`HTTP error! status: ${req.status}`);
      }
      const res = await req.json(); // Parse the response as JSON
      setStop(res.message)
    } catch (error) {
      console.error("Error stopping the server:", error);
      setStop("Error stopping the server")
    }
  };

  return (
    isAuthenticated && (
      <div className="grid grid-rows-[20px_1fr_20px] items-center justify-items-center min-h-screen p-4 sm:p-8 pb-20 gap-8 sm:gap-16 font-[family-name:var(--font-geist-sans)]">
        <main className="flex flex-col sm:flex-row gap-8 row-start-2 items-center sm:items-start w-full max-w-4xl">
          <Logout />
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
          <div className="w-full p-4 border rounded shadow">
            <h1 className="text-xl font-bold mb-4">Add an Instance</h1>
            <h3>{addInstance}</h3>
            <form onSubmit={handleAddInstance} className="flex flex-col gap-4">
              <input
                type="text"
                placeholder="Instance Name"
                value={newInstanceID}
                onChange={(e) => setNewInstanceID(e.target.value)}
                className="p-2 border rounded"
              />
              <button
                type="submit"
                className="bg-blue-500 hover:bg-blue-700 text-white py-2 px-4 rounded"
              >
                Add
              </button>
            </form>
          </div>
        </main>
      </div>
    )
  );
}

