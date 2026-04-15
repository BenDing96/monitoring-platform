# monitoring-console

React TypeScript dashboard for monitoring-platform.

## Stack

- **Vite** — build tool + dev server
- **React 18** + **TypeScript**
- **Tailwind CSS v4** — utility styling
- **TanStack Query** — server state / data fetching
- **React Router v6** — client-side routing

## Local development

```bash
cd monitoring-console
npm install
npm run dev        # http://localhost:5173
```

The dev server proxies `/v1/*` to `http://localhost:8080` (the `api` service).
Start the backend first with `make kind-up && make deploy-dev` from the repo root,
or `go run ./cmd/api` from `monitoring-app/` for a faster feedback loop.

## Production build

```bash
npm run build      # outputs to dist/
```

The Docker image is a multi-stage build: `node:22-alpine` compiles to `dist/`, then
`nginx:1.27-alpine` serves the static files. The nginx config handles SPA fallback
(`try_files $uri /index.html`) and proxies `/v1/*` to the `api` service at `api:8080`.

## Project layout

```
src/
  components/       shared UI (Layout, nav)
  features/
    runs/           RunsList, RunDetail pages
  lib/
    api.ts          typed fetch client against /v1/* endpoints
  App.tsx           router setup + QueryClient
  main.tsx          entry point
deploy/helm/        Helm chart for Kubernetes deployment
nginx.conf          nginx config for the production container
```

## Pages

| Route | Component | Data source |
|---|---|---|
| `/runs` | `RunsList` | `GET /v1/runs` |
| `/runs/:traceID` | `RunDetail` | `GET /v1/runs/:traceID` |

## Adding a new page

1. Create `src/features/<name>/<Name>.tsx`.
2. Add a `<Route>` in `src/App.tsx`.
3. Add a `<NavLink>` in `src/components/Layout.tsx`.
4. Add the fetch call in `src/lib/api.ts`.
