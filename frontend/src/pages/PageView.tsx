import React from "react";
import type { Location } from "react-router-dom";

import type { Page } from "../types";
import { WidgetCard } from "../components/WidgetCard";

export const PageView: React.FC<{ page: Page; location: Location }> = ({ page, location }) => {
  return (
    <div className="page">
      <div className="page-header">
        <h1>{page.title}</h1>
      </div>
      <div className="widgets">
        {page.widgets.map((widget) => (
          <WidgetCard key={widget.id} widget={widget} location={location} />
        ))}
      </div>
    </div>
  );
};
