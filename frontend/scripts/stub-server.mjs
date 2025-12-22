import http from "node:http";
import { readFile } from "node:fs/promises";
import { fileURLToPath } from "node:url";
import path from "node:path";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const configPath = path.join(__dirname, "stub-config.json");

const config = JSON.parse(await readFile(configPath, "utf-8"));

const users = Array.from({ length: 120 }, (_, index) => {
  const id = index + 1;
  return {
    id,
    name: `User ${id}`,
    email: `user${id}@example.com`,
  };
});

const reports = Array.from({ length: 30 }, (_, index) => {
  const week = index + 1;
  return {
    week: `2024-W${String(week).padStart(2, "0")}`,
    status: week % 3 === 0 ? "delayed" : "ok",
    owner: week % 2 === 0 ? "Ops" : "Finance",
  };
});

const server = http.createServer((req, res) => {
  if (!req.url) {
    res.writeHead(400);
    res.end();
    return;
  }

  if (req.url.startsWith("/api/config")) {
    res.setHeader("Content-Type", "application/json");
    res.end(JSON.stringify(config));
    return;
  }

  if (req.url.startsWith("/api/widgets/")) {
    const requestUrl = new URL(req.url, "http://localhost");
    const limit = Number(requestUrl.searchParams.get("limit") ?? "50");
    const cursor = requestUrl.searchParams.get("offset");
    const widgetId = requestUrl.pathname.replace("/api/widgets/", "");
    const dataset = widgetId === "reports_table" ? reports : users;
    let startIndex = 0;
    if (cursor) {
      const cursorId = Number(cursor);
      const foundIndex = dataset.findIndex((item) =>
        widgetId === "reports_table" ? item.week === cursor : item.id === cursorId,
      );
      if (foundIndex !== -1) {
        startIndex = foundIndex + 1;
      }
    }

    const slice = dataset.slice(startIndex, startIndex + limit + 1);
    const hasMore = slice.length > limit;
    const data = hasMore ? slice.slice(0, limit) : slice;
    const nextCursor =
      data.length > 0
        ? widgetId === "reports_table"
          ? String(data[data.length - 1].week)
          : String(data[data.length - 1].id)
        : "";
    res.setHeader("Content-Type", "application/json");
    res.end(
      JSON.stringify({
        data,
        total: data.length,
        next_cursor: nextCursor,
        has_more: hasMore,
      }),
    );
    return;
  }

  res.writeHead(404);
  res.end();
});

server.listen(4173, () => {
  console.log("Stub server listening on http://localhost:4173");
});
