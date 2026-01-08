const metaName = "rapidmin-path-prefix";

function normalizePrefix(prefix: string | null | undefined): string {
    const trimmed = (prefix ?? "").trim();
    if (!trimmed || trimmed === "/") return "";
    const normalized = trimmed.startsWith("/") ? trimmed : `/${trimmed}`;
    return normalized.endsWith("/") ? normalized.slice(0, -1) : normalized;
}

function readMetaPrefix(): string {
    if (typeof document === "undefined") return "";
    const meta = document.querySelector(`meta[name="${metaName}"]`);
    return normalizePrefix(meta?.getAttribute("content"));
}

let currentPrefix = readMetaPrefix();

export function getPathPrefix(): string {
    return currentPrefix;
}

export function setPathPrefix(prefix: string | null | undefined): void {
    currentPrefix = normalizePrefix(prefix);
}

export function withPathPrefix(path: string): string {
    if (!path.startsWith("/")) {
        return `${currentPrefix}/${path}`;
    }
    return `${currentPrefix}${path}`;
}
