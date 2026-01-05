import type { AppConfig, DataResponse } from "./types";
import { withPathPrefix } from "./pathPrefix";

export async function fetchConfig(): Promise<AppConfig> {
  const res = await fetch(withPathPrefix("/api/config"));
  if (!res.ok) {
    throw new Error(`Config request failed: ${res.status}`);
  }
  return res.json();
}

export async function fetchWidgetData(
  widgetId: string,
  params: URLSearchParams,
): Promise<DataResponse> {
  const query = params.toString();
  const url = withPathPrefix(
    query ? `/api/widgets/${widgetId}?${query}` : `/api/widgets/${widgetId}`,
  );
  const res = await fetch(url);
  if (!res.ok) {
    throw new Error(`Widget request failed: ${res.status}`);
  }
  return res.json();
}
