import { copyFile } from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const frontendRoot = path.resolve(__dirname, "..");
const distIndex = path.join(frontendRoot, "dist", "index.html");
const backendIndex = path.resolve(frontendRoot, "..", "server", "web", "index.html");

await copyFile(distIndex, backendIndex);
console.log(`Copied ${distIndex} to ${backendIndex}`);
