import {defineConfig, PluginOption} from "vite";
import react from "@vitejs/plugin-react";
import {viteSingleFile} from "vite-plugin-singlefile";

const transformIndexHtml = () => (
    {
        name: "dev-html-replacement",
        apply: "serve",
        async transformIndexHtml(html, config) {
            // resolve route 'pathPrefix' to root path (empty value) on dev (stub) mode
            return html.replace(
                '{{pathPrefix}}',
                ''
            );
        }
    }
) as PluginOption

export default defineConfig({
    plugins: [react(), viteSingleFile(), transformIndexHtml()],
    server: {
        port: 5173,
        proxy: {
            "/api": process.env.VITE_API_PROXY ?? "http://localhost:4173",
        },
    },
});
