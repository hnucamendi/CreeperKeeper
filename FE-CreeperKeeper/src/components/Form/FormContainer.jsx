import React from "react";
import "../../styles/form.css";

export default function FormContainer({ children, onSubmit }) {
  return (
    <form className="form-container" onSubmit={onSubmit}>
      {children}
    </form>
  );
}
