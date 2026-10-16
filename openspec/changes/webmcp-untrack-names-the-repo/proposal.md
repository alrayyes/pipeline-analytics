# Proposal

## Why

`untrack_repo` asks the person to confirm before it deletes a repository's
stored runs, but the question names only the repository id, which means nothing
to them. A review of the WebMCP parity work raised it, with a wrong version
example in the `get_version` description (#696).

## What changes

- The confirmation looks the repository up and names its identifier, such as
  `alrayyes/payments`. A failed or empty lookup names the id, so the question is
  never blocked by it.
- The `get_version` description gives a version without a `v`, as the footer
  and the release history show it.

## Impact

- Added requirement to `webmcp-support`.
