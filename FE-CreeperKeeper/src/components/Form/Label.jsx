import React from "react";
import "../../styles/form.css";

export default function Label({ htmlFor, children }) {
  return (
    <label className="custom-label" htmlFor={htmlFor}>
      {children}
    </label>
  );
}
