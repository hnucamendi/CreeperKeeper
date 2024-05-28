import React from "react";
import "../../styles/form.css";

export default function Button({ type, children }) {
  return (
    <button className="custom-button" type={type}>
      {children}
    </button>
  );
}
