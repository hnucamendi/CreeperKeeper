// import React, { useState } from "react";
// import "../styles/app.css";
//
// function App() {
//   const [isLoggedIn, setIsLoggedIn] = useState(true);
//   const [username, setUsername] = useState("");
//   const [password, setPassword] = useState("");
//
//   if (isLoggedIn) {
//     return (
//       <div className="server-menu">
//         <div>
//           <form>
//             <label htmlFor="server-dropdown">Select Version</label>
//             <select id="server-dropdown"></select>
//             <button>Create Server</button>
//           </form>
//         </div>
//
//         <div>
//           <form>
//             <label htmlFor="server-dropdown">Select Saved Instance</label>
//             <select id="server-dropdown"></select>
//             <button>Startup Server</button>
//           </form>
//         </div>
//       </div>
//     );
//   }
//
//   if (!isLoggedIn) {
//     return (
//       <div className="login-form">
//         <h1>Please login to begin</h1>
//         <form>
//           <label htmlFor="username">Username</label>
//           <input id="username"></input>
//           <label htmlFor="password">Password</label>
//           <input id="password"></input>
//         </form>
//       </div>
//     );
//   }
//
//   return <h1>Please login to begin</h1>;
// }
//
// export default App;