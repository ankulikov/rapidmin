import React from "react";
import { Link } from "react-router-dom";

import type { ColumnSpec, DataResponse, FilterSpec, Widget } from "../types";
import { FilterPanel, type FilterState } from "./FilterPanel";

type Props = {
  widget: Widget;
  data: DataResponse;
  onLoadMore: () => void;
  loadingMore: boolean;
  filters: FilterSpec[];
  filtersDraft: FilterState;
  onFilterChange: (next: FilterState) => void;
  onApplyFilters: () => void;
  onResetFilters: () => void;
};

export const TableWidget: React.FC<Props> = ({
  widget,
  data,
  onLoadMore,
  loadingMore,
  filters,
  filtersDraft,
  onFilterChange,
  onApplyFilters,
  onResetFilters,
}) => {
  const columns = widget.table?.columns ?? [];
  return (
    <div className="table">
      {filters.length > 0 && (
        <FilterPanel
          filters={filters}
          state={filtersDraft}
          onChange={onFilterChange}
          onApply={onApplyFilters}
          onReset={onResetFilters}
        />
      )}
      <table>
        <thead>
          <tr>
            {columns.map((col) => (
              <th key={col.id}>{col.title ?? col.id}</th>
            ))}
          </tr>
        </thead>
        <tbody>
          {data.data.length === 0 ? (
            <tr>
              <td colSpan={Math.max(columns.length, 1)}>No rows</td>
            </tr>
          ) : (
            data.data.map((row, idx) => (
              <tr key={idx}>
                {columns.map((col) => (
                  <td key={col.id}>{renderCell(col, row)}</td>
                ))}
              </tr>
            ))
          )}
        </tbody>
      </table>
      {data.has_more && data.next_cursor && (
        <div className="table-footer">
          <button className="load-more" type="button" onClick={onLoadMore} disabled={loadingMore}>
            {loadingMore ? "Loading..." : "Load more"}
          </button>
        </div>
      )}
    </div>
  );
};

function renderCell(column: ColumnSpec, row: Record<string, unknown>) {
  const rawValue = row[column.id];
  const fallbackText = rawValue === null || rawValue === undefined ? "" : String(rawValue);
  const render = column.render;
  if (!render || render.type !== "link") {
    return formatValue(rawValue);
  }
  const url = render.url ? applyTemplate(render.url, row) : "";
  const text = render.text ? applyTemplate(render.text, row) : fallbackText;
  if (!url) {
    return text;
  }
  if (render.external) {
    return (
      <a href={url} target="_blank" rel="noreferrer">
        {text}
      </a>
    );
  }
  return <Link to={url}>{text}</Link>;
}

function applyTemplate(template: string, row: Record<string, unknown>) {
  return template.replace(/{{\s*([^}]+)\s*}}/g, (_, key) => {
    const value = row[String(key).trim()];
    if (value === null || value === undefined) return "";
    return String(value);
  });
}

function formatValue(value: unknown): React.ReactNode {
  if (Array.isArray(value)) {
    if (value.length === 0) return "";
    return (
      <div className="cell-list">
        {value.map((item, index) => (
          <div key={index} className="cell-list-item">
            {formatPrimitive(item)}
          </div>
        ))}
      </div>
    );
  }
  if (value && typeof value === "object") {
    return JSON.stringify(value);
  }
  if (value === null || value === undefined) return "";
  return String(value);
}

function formatPrimitive(value: unknown): string {
  if (value === null || value === undefined) return "";
  if (typeof value === "object") return JSON.stringify(value);
  return String(value);
}
