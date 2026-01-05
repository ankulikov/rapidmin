import React from "react";
import { createRoot } from "react-dom/client";
import { BrowserRouter } from "react-router-dom";

import { getPathPrefix } from "./pathPrefix";
import { App } from "./pages/App";
import "./styles.css";

const container = document.getElementById("root");
if (!container) {
    throw new Error("Missing root element");
}

const basename = getPathPrefix();

createRoot(container).render(
  <React.StrictMode>
    <BrowserRouter basename={basename || undefined}>
      <App />
    </BrowserRouter>
  </React.StrictMode>,
);
