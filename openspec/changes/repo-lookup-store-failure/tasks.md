# Tasks

- [x] 1 Test, then add, `ingestion.ErrRepoNotFound` and return it from the SQLite adapter
- [x] 2 Test, then make the repo lookup answer 500 and log for a non-sentinel error
- [x] 3 Document the 500 on `getRepo` in the OpenAPI spec
- [x] 4 Test that the MCP `get_repo` tool reports a store failure as an error, not not-found
