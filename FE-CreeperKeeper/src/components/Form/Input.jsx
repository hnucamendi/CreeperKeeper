import React from "react";
import "../../styles/form.css";

export default function Input({ type, id, name, value, onChange }) {
  return (
    <input
      className="custom-input"
      type={type}
      id={id}
      name={name}
      value={value}
      onChange={onChange}
    />
  );
}
