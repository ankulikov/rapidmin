import React, { useEffect, useMemo, useState } from "react";
import { Link, NavLink, Route, Routes, useLocation } from "react-router-dom";

import { fetchConfig } from "../api";
import { setPathPrefix } from "../pathPrefix";
import type { AppConfig, MenuItem, Page } from "../types";
import { PageView } from "./PageView";

const defaultConfig: AppConfig = {
  title: "Rapidmin",
  menu: [],
  pages: [],
};

export const App: React.FC = () => {
  const [config, setConfig] = useState<AppConfig>(defaultConfig);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const location = useLocation();

  useEffect(() => {
    let active = true;
    setLoading(true);
    fetchConfig()
      .then((cfg) => {
        if (!active) return;
        setConfig(cfg);
        setPathPrefix(cfg.path_prefix);
        setError(null);
      })
      .catch((err) => {
        if (!active) return;
        setError(err instanceof Error ? err.message : String(err));
      })
      .finally(() => {
        if (!active) return;
        setLoading(false);
      });
    return () => {
      active = false;
    };
  }, []);

  const routes = useMemo(() => buildRoutes(config.pages), [config.pages]);

  return (
    <div className="app">
      <aside className="sidebar">
        <div className="brand">{config.title}</div>
        {renderMenu(config.menu)}
      </aside>
      <main className="content">
        {loading ? (
          <div className="state">Loading config...</div>
        ) : error ? (
          <div className="state error">{error}</div>
        ) : (
          <Routes>
            {routes.map((page) => (
              <Route
                key={page.slug}
                path={`/${page.slug}`}
                element={<PageView page={page} location={location} />}
              />
            ))}
            <Route
              path="*"
              element={<div className="state">Select a page from the menu.</div>}
            />
          </Routes>
        )}
      </main>
    </div>
  );
};

function buildRoutes(pages: Page[]): Page[] {
  return pages;
}

function renderMenu(items: MenuItem[]): React.ReactElement | null {
  if (items.length === 0) return null;
  return (
    <nav className="menu">
      {items.map((item) => (
        <div key={`${item.title}-${item.page ?? item.href ?? ""}`} className="menu-item">
          {item.page ? (
            <NavLink
              to={`/${item.page}`}
              className={({ isActive }) => `menu-link${isActive ? " is-active" : ""}`}
            >
              {item.title}
            </NavLink>
          ) : item.href ? (
            <a className="menu-link" href={item.href}>
              {item.title}
            </a>
          ) : (
            <span className="menu-label">{item.title}</span>
          )}
          {item.children && <div className="menu-children">{renderMenu(item.children)}</div>}
        </div>
      ))}
    </nav>
  );
}
