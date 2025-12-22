import React, { useEffect, useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import type { Location } from "react-router-dom";

import { fetchWidgetData } from "../api";
import type { DataResponse, Widget } from "../types";
import { appendFilters, emptyFilterState, parseFilterParams, type FilterState } from "./FilterPanel";
import { TableWidget } from "./TableWidget";

type Props = {
  widget: Widget;
  location: Location;
};

export const WidgetCard: React.FC<Props> = ({ widget, location }) => {
  const navigate = useNavigate();
  const [data, setData] = useState<DataResponse | null>(null);
  const [rows, setRows] = useState<Record<string, unknown>[]>([]);
  const [nextCursor, setNextCursor] = useState<string | undefined>();
  const [hasMore, setHasMore] = useState(false);
  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const baseParams = useMemo(() => {
    const params = new URLSearchParams(location.search);
    params.delete("offset");
    params.delete("limit");
    const filterIds = widget.table?.filters?.map((filter) => filter.id) ?? [];
    if (filterIds.length > 0) {
      const keys = Array.from(params.keys());
      keys.forEach((key) => {
        for (const id of filterIds) {
          if (key === id || key.startsWith(`${id}.`)) {
            params.delete(key);
            break;
          }
        }
      });
    }
    return params;
  }, [location.search, widget.table?.filters]);

  const [filtersDraft, setFiltersDraft] = useState<FilterState>({});
  const [filtersApplied, setFiltersApplied] = useState<FilterState>({});

  useEffect(() => {
    const parsed = parseFilterParams(widget.table?.filters ?? [], location.search);
    setFiltersDraft(parsed);
    setFiltersApplied(parsed);
  }, [widget.id, widget.table?.filters, location.search]);

  const params = useMemo(() => {
    const combined = new URLSearchParams(baseParams);
    appendFilters(combined, filtersApplied, widget.table?.filters ?? []);
    return combined;
  }, [baseParams, filtersApplied, widget.table?.filters]);

  useEffect(() => {
    let active = true;
    setLoading(true);
    setRows([]);
    setNextCursor(undefined);
    setHasMore(false);
    fetchWidgetData(widget.id, params)
      .then((payload) => {
        if (!active) return;
        setData(payload);
        setRows(payload.data);
        setNextCursor(payload.next_cursor);
        setHasMore(Boolean(payload.has_more));
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
  }, [widget.id, params]);

  const handleLoadMore = async () => {
    if (!nextCursor || loadingMore) return;
    setLoadingMore(true);
    try {
      const nextParams = new URLSearchParams(params);
      nextParams.set("offset", nextCursor);
      const payload = await fetchWidgetData(widget.id, nextParams);
      setRows((prev) => [...prev, ...payload.data]);
      setNextCursor(payload.next_cursor);
      setHasMore(Boolean(payload.has_more));
      setData(payload);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err));
    } finally {
      setLoadingMore(false);
    }
  };

  const applyFiltersToUrl = (nextFilters: FilterState) => {
    const nextParams = new URLSearchParams(baseParams);
    appendFilters(nextParams, nextFilters, widget.table?.filters ?? []);
    navigate(
      {
        search: nextParams.toString() ? `?${nextParams.toString()}` : "",
      },
      { replace: false },
    );
  };

  const handleResetFilters = () => {
    const cleared = emptyFilterState(widget.table?.filters ?? []);
    setFiltersDraft(cleared);
    setFiltersApplied(cleared);
    applyFiltersToUrl(cleared);
  };

  return (
    <section className="widget">
      <header className="widget-header">
        <h2>{widget.title}</h2>
      </header>
      {loading ? (
        <div className="state">Loading data...</div>
      ) : error ? (
        <div className="state error">{error}</div>
      ) : !data ? (
        <div className="state">No data.</div>
      ) : widget.type === "table" ? (
        <TableWidget
          widget={widget}
          data={{
            data: rows,
            total: rows.length,
            next_cursor: nextCursor,
            has_more: hasMore,
          }}
          onLoadMore={handleLoadMore}
          loadingMore={loadingMore}
          filters={widget.table?.filters ?? []}
          filtersDraft={filtersDraft}
          onFilterChange={setFiltersDraft}
          onApplyFilters={() => {
            setFiltersApplied(filtersDraft);
            applyFiltersToUrl(filtersDraft);
          }}
          onResetFilters={handleResetFilters}
        />
      ) : (
        <div className="state">Unsupported widget type: {widget.type}</div>
      )}
    </section>
  );
};
