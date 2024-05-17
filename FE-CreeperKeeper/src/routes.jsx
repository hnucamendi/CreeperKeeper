import { Route, Routes } from "react-router-dom";
import Console from "./pages/console.jsx";

export default function UserRouter() {
  return (
    <Routes>
      <Route path="/" element={<Console />} />
    </Routes>
  );
}
