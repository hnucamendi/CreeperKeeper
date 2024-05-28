import React from "react";
import "../../styles/form.css";

export default function Select({ id, name, value, onChange, children }) {
  return (
    <div className="select-container">
      <select
        className="custom-select"
        id={id}
        name={name}
        value={value}
        onChange={onChange}
      >
        {children}
      </select>
    </div>
  );
}
