import React from "react";
import "../styles/container.css";

export default function Container({ children }) {
  return (
    <div className="main-container">
      <div className="content">{children}</div>
    </div>
  );
}
