# Proposal

## Why

`untrack_repo` asks the person to confirm before it deletes a repository's
stored runs, but the question names only the repository id, which means nothing
to them. A review of the WebMCP parity work raised it (#696).

## What changes

- The confirmation looks the repository up and names its identifier, such as
  `alrayyes/payments`. A failed, slow (three seconds) or empty lookup, or a repository past the first
  500, names the id, so the question is never blocked by it. A get-by-id
  endpoint would replace the list read; it is filed separately.


## Impact

- Added requirement to `webmcp-support`.
