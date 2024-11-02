"use client"

import { useState, useEffect } from "react";

export default function Home() {
  const ck_url = "https://app.creeperkeeper.com";
  const [currentInstance, setCurrentInstance] = useState("");
  const [start, setStart] = useState(null)
  const [instances, setInstances] = useState([]);
  const [authToken, setAuthToken] = useState("");

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

    console.log({ instances });
  }, [authToken, instances]);

  const getAuth = () => {
    return "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6IkpPcW9POTU4MDFzUmVyTnBza19lSyJ9.eyJpc3MiOiJodHRwczovL2Rldi1ieG4yNDVsNmJlMnl6aGlsLnVzLmF1dGgwLmNvbS8iLCJzdWIiOiJIdWd0eFBkQ01kaThQbXZVWEM2bHc4bEVtNnU1SmFleEBjbGllbnRzIiwiYXVkIjoiY3JlZXBlci1rZWVwZXItcmVzb3VyY2UiLCJpYXQiOjE3MzA1NTk0MDAsImV4cCI6MTczMDY0NTgwMCwic2NvcGUiOiJyZWFkOmFsbCB3cml0ZTphbGwiLCJndHkiOiJjbGllbnQtY3JlZGVudGlhbHMiLCJhenAiOiJIdWd0eFBkQ01kaThQbXZVWEM2bHc4bEVtNnU1SmFleCJ9.gAvRdrFjeP26pUna7-MkcbUa-MR1iE6arP8f2D_yXHSrz4jqgdgeJFhyTVUP__EbrT5UIG8KbOlLyLkaYeB2vkgpsCprUX0RniG7UVR3ZxBAZQU-Po-qyWZjZL4Q_vwY4oiVYnWwLkLGjBBPVES8P7VDlfy_F3MnVLZyM-scs3ElIzMGNC63zbbpLO_xNTA8sV-2mjjnK1TH0ovL7HN8GZWML9y7-9bfTtt1va4_rVn8cFblsJIEM2VSs39b-o42on1MZ00U-pmEIThGNRrf3akt6E0uOvHT-ERlEhb3F_rDlslL2e2soDuZp3du6mVl374y2WjwrQVYG_DBrEiygQ"
  }

  const handleSetInstance = (e) => {
    setCurrentInstance(e.target.value)
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

  // {/* <option key={i} value={instance[i]}>
  //               {instance[i]}
  //             </option> */}

  return (
    <div className="grid grid-rows-[20px_1fr_20px] items-center justify-items-center min-h-screen p-8 pb-20 gap-16 sm:p-20 font-[family-name:var(--font-geist-sans)]">
      <main className="flex gap-8 row-start-2 items-center sm:items-start">
        <div>
          <h1>Select Instance</h1>
          <select onChange={handleSetInstance}>
            <option value="">None</option>
            {/* {
              instances.map((instance, i) => (
                <h1>
                  {instance} {i}
                </h1>
              ))
            } */}
          </select>
        </div>
        <div>
          <h1>Manage Server</h1>
          <h2>{start}</h2>
          <button className={`${currentInstance ? "bg-blue-500 hover:bg-blue-700 text-white" : "bg-gray-500 text-gray-300 cursor-not-allowed"
            } py-2 px-4 rounded`} onClick={handleStartMCServer}>Start</button>
          <button className={`${currentInstance ? "bg-blue-500 hover:bg-blue-700 text-white" : "bg-gray-500 text-gray-300 cursor-not-allowed"
            } py-2 px-4 rounded`} onClick={handleStopMCServer}>Stop</button>
        </div>
      </main >
    </div >
  );
}
