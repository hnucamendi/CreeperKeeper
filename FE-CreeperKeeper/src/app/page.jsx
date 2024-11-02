"use client"

import { useState, useEffect } from "react";

export default function Home() {
  const ck_url = "https://app.creeperkeeper.com";
  const [currentInstance, setCurrentInstance] = useState("");
  const [start, setStart] = useState(null)
  const [instances, setInstances] = useState([]);
  const [authToken, setAuthToken] = useState("");

  useEffect(() => {
    const fetchInstances = async () => {
      const path = "getInstances";
      const url = `${ck_url}/${path}`;

      setAuthToken(`Bearer ${getAuth()}`);

      const body = {
        method: "GET",
        headers: {
          "Content-Type": "application/json",
          "Authorization": `Bearer ${getAuth()}`,
        },
      };

      try {
        const req = await fetch(url, body);
        if (!req.ok) {
          throw new Error(`HTTP error! status: ${req.status}`);
        }
        const res = await req.json(); // Parse the response as JSON
        setInstances(res.message);
      } catch (error) {
        console.error("Error getting instances:", error);
      }
    };

    fetchInstances();
  }, []);

  const getAuth = () => {
    return "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6IkpPcW9POTU4MDFzUmVyTnBza19lSyJ9.eyJpc3MiOiJodHRwczovL2Rldi1ieG4yNDVsNmJlMnl6aGlsLnVzLmF1dGgwLmNvbS8iLCJzdWIiOiJIdWd0eFBkQ01kaThQbXZVWEM2bHc4bEVtNnU1SmFleEBjbGllbnRzIiwiYXVkIjoiY3JlZXBlci1rZWVwZXItcmVzb3VyY2UiLCJpYXQiOjE3MzA1NTE3MDIsImV4cCI6MTczMDYzODEwMiwic2NvcGUiOiJyZWFkOmFsbCB3cml0ZTphbGwiLCJndHkiOiJjbGllbnQtY3JlZGVudGlhbHMiLCJhenAiOiJIdWd0eFBkQ01kaThQbXZVWEM2bHc4bEVtNnU1SmFleCJ9.X_bF3FaRePz1Lmghjp8QiUGte66pxRECIEA6nRCjMpnbLB_ur7rhlsmuqWqPzWofQlfPezns-SizbDAw04T9wCBEYvum4ynurfs0LxutYPSzfXlzb3ukyF3xaNp-uBFAgCm_GfojBR1vtU6WrTwm6AjSvBc7Ww5mS7838yIWA_VV50jaD1lBvZxSrU7rqFgcefQm4a5Vm8901XvQLxvqnxDWdPZJbhO6hAiIS_gb1_V1wcBz1D2C02nyk3wk-g1JGtucOYfPFF8sz-k1zAqjyKVVD2COrAFJwA0CCJM1xgExz1geEA0D08sMOUEb-VWis7TaXe6GJ2D527XEC8GqJw"
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
            {instances.map((instance, i) => (
              <h1>
                {instance} {i}
              </h1>
            ))}
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
