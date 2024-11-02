import { useState, useEffect } from "react";

export default function App() {
  const ck_url = "https://app.creeperkeeper.com";
  const [currentInstance, setCurrentInstance] = useState("");
  const [start, setStart] = useState("")
  const [instances, setInstances] = useState([]);
  const [authToken, setAuthToken] = useState("");

  const [addInstance, setAddInstance] = useState("");
  const [newInstanceID, setNewInstanceID] = useState("");

  useEffect(() => {
    const token = getAuth();
    setAuthToken(`Bearer ${token}`);
  }, []);

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

  const getAuth = () => {
    return "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6IkpPcW9POTU4MDFzUmVyTnBza19lSyJ9.eyJpc3MiOiJodHRwczovL2Rldi1ieG4yNDVsNmJlMnl6aGlsLnVzLmF1dGgwLmNvbS8iLCJzdWIiOiJIdWd0eFBkQ01kaThQbXZVWEM2bHc4bEVtNnU1SmFleEBjbGllbnRzIiwiYXVkIjoiY3JlZXBlci1rZWVwZXItcmVzb3VyY2UiLCJpYXQiOjE3MzA1NTk0MDAsImV4cCI6MTczMDY0NTgwMCwic2NvcGUiOiJyZWFkOmFsbCB3cml0ZTphbGwiLCJndHkiOiJjbGllbnQtY3JlZGVudGlhbHMiLCJhenAiOiJIdWd0eFBkQ01kaThQbXZVWEM2bHc4bEVtNnU1SmFleCJ9.gAvRdrFjeP26pUna7-MkcbUa-MR1iE6arP8f2D_yXHSrz4jqgdgeJFhyTVUP__EbrT5UIG8KbOlLyLkaYeB2vkgpsCprUX0RniG7UVR3ZxBAZQU-Po-qyWZjZL4Q_vwY4oiVYnWwLkLGjBBPVES8P7VDlfy_F3MnVLZyM-scs3ElIzMGNC63zbbpLO_xNTA8sV-2mjjnK1TH0ovL7HN8GZWML9y7-9bfTtt1va4_rVn8cFblsJIEM2VSs39b-o42on1MZ00U-pmEIThGNRrf3akt6E0uOvHT-ERlEhb3F_rDlslL2e2soDuZp3du6mVl374y2WjwrQVYG_DBrEiygQ"
  }

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
      setStart(res.message)
    } catch (error) {
      console.error("Error starting the server:", error);
      setStart("Error starting the server")
    }
  };

  const handleStopMCServer = async () => {
    const path = "stop";

    const url = `${ck_url}/${path}`;

    const body = {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Authorization": `Bearer ${getAuth()}`,
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
      setStart(res.message)
    } catch (error) {
      console.error("Error stopping the server:", error);
      setStart("Error stopping the server")
    }
  };

  return (
    <div className="grid grid-rows-[20px_1fr_20px] items-center justify-items-center min-h-screen p-4 sm:p-8 pb-20 gap-8 sm:gap-16 font-[family-name:var(--font-geist-sans)]">
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
          <h2 className="text-lg mb-4">{start}</h2>
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
  );
}

