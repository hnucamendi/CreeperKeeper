import React from "react";
import "../../styles/form.css";

export default function Option({ value, children }) {
  return (
    <option className="custom-option" value={value}>
      {children}
    </option>
  );
}
