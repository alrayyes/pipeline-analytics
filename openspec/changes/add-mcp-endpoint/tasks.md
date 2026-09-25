# Tasks

## 1. Dependencies

- [x] 1.1 Add `github.com/modelcontextprotocol/go-sdk` at its latest
      tagged release with `go get`, pinned exactly in `go.mod`/`go.sum`;
      verify `go build ./...` still succeeds with the new import
      unused (a stub import) so the pin lands in its own commit before
      any behavior depends on it
- [x] 1.2 Add `github.com/getkin/kin-openapi` (test-only use) at its
      latest tagged release the same way; verify `go build ./...` and
      `go vet ./...` still pass

## 2. OpenAPI spec

- [x] 2.1 Add the `/api/mcp` `POST` path to `openapi/openapi.yaml`: bearer
      + session security (matching every other data endpoint),
      request/response bodies typed loosely as MCP's JSON-RPC envelope,
      `401` `Error` response matching the existing pattern
- [x] 2.2 Run `bunx @redocly/cli lint openapi/openapi.yaml` and verify
      it's clean

## 3. MCP server scaffolding

- [x] 3.1 Write a failing `internal/httpserver/mcp_test.go` case:
      `POST /api/mcp` with no session and no bearer token returns `401`,
      matching any other data endpoint
- [x] 3.2 Write a failing case: `POST /api/mcp` with a valid session (or a
      valid API token) reaches the MCP layer and an MCP
      `tools/list` request over it returns a non-empty tool list
- [x] 3.3 Add `internal/httpserver/mcp.go`: build an
      `mcp.NewStreamableHTTPHandler` from `go-sdk`'s `mcp` package with
      no tools registered yet, mount it at `POST /api/mcp` in
      `server.go` inside the existing `requireSession`-wrapped mux;
      verify 3.1 and 3.2 pass (3.2 passes with an empty tool list at
      this point)

## 4. Pipeline tools

- [x] 4.1 Write a failing case in `mcp_test.go`: calling the
      `list_pipelines` tool returns the same pipeline IDs, names, and
      health statuses `GET /api/pipelines` returns for the same fixture
      data, respecting its pagination parameters
- [x] 4.2 Implement the `list_pipelines` tool in `mcp.go`, calling
      `deps.Metrics` the same way `pipelinesHandler.list` does; verify
      4.1 passes
- [x] 4.3 Write a failing case: calling the `get_pipeline` tool for a
      tracked pipeline ID returns its duration trend and failure-rate
      trend, matching `GET /api/pipelines/{pipelineId}`; and for an
      unknown ID returns an MCP tool error, not a crash
- [x] 4.4 Implement `get_pipeline`; verify 4.3 passes
- [x] 4.5 Write a failing case: calling `list_pipeline_steps` for a
      tracked pipeline returns the same step list (name, duration
      contribution, failure rate, flaky flag) `GET
      /api/pipelines/{pipelineId}/steps` returns
- [x] 4.6 Implement `list_pipeline_steps`; verify 4.5 passes
- [x] 4.7 Write a failing case: calling `list_pipeline_flaky_runs`
      returns the same runs `GET
      /api/pipelines/{pipelineId}/flaky-runs` returns
- [x] 4.8 Implement `list_pipeline_flaky_runs`; verify 4.7 passes

## 5. Usage and cross-pipeline tools

- [x] 5.1 Write a failing case: calling `get_repo_usage` for a tracked
      repo ID returns the same runner-minutes figures `GET
      /api/repos/{repoId}/usage` returns
- [x] 5.2 Implement `get_repo_usage`; verify 5.1 passes
- [x] 5.3 Write a failing case: calling `list_unhealthy_steps` returns
      the same steps `GET /api/steps/unhealthy` returns
- [x] 5.4 Implement `list_unhealthy_steps`; verify 5.3 passes
- [x] 5.5 Write a failing case: the tool list from `tools/list`
      contains exactly these six tools and none of them accepts a
      write/mutating call
- [x] 5.6 Verify 5.5 passes against the implementation from sections 4
      and 5

## 6. Spec parity

- [x] 6.1 Write `internal/httpserver/mcp_spec_parity_test.go`: parse
      `openapi/openapi.yaml` with `kin-openapi`, and for each of the
      six tools assert its input schema's parameter names and
      required/optional status match the corresponding path's declared
      parameters; verify it passes against the current implementation
- [x] 6.2 Confirm the test fails on a deliberate mismatch (temporarily
      rename a tool's input field or flip a `required` flag, observe
      the test fail, then revert) — proves it actually catches drift
      before relying on it

## 7. Documentation

- [ ] 7.1 Add an "MCP" section to `README.md` next to the existing
      API/installation docs: the `/api/mcp` URL, that it needs the same
      bearer token as the REST API, and how to point an MCP-capable
      client at it
- [ ] 7.2 Run this account's README-verification pass (or manually
      confirm) that any command shown actually runs as written

## 8. Verify and ship

- [ ] 8.1 Run `go build ./...`, `go vet ./...`, `go test -race -cover
      ./...`, `go mod tidy -diff`, and `golangci-lint run ./...`;
      verify all clean
- [ ] 8.2 Open a pull request with `Closes #295` and verify CI passes
