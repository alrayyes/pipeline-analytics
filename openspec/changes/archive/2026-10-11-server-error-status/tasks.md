## 1. Route list

- [x] 1.1 Test then `web/scripts/generate-routes.ts`: patterns from `src/routes`, `[id]` kept as a pattern
- [x] 1.2 `postbuild` writes `routes.json` into the embedded `dist/`

## 2. Server

- [x] 2.1 Tests first: unknown path 404, route 200, dynamic route 200, missing file 404, `/api/` JSON 404, unreadable asset 500, no routes.json keeps the shell
- [x] 2.2 Route matcher and the 404/500 responses with their `Cache-Control`
- [x] 2.3 Log the 500's detail, send none of it

## 3. Docs and wrap-up

- [x] 3.1 README true again; the comment on `staticHandler` too
- [x] 3.2 Run the built binary with `curl -i` for each scenario
- [ ] 3.3 Archive this change once merged
