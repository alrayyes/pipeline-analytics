# Design

## Decisions

**A separate `/readyz`, not a stricter `/healthz`.** Liveness and
readiness answer different questions. A liveness probe that fails when the
database does would make an orchestrator restart a process that restarting
can't help. `/healthz` stays cheap and dependency-free; `/readyz` carries
the dependency check.

**The Docker `HEALTHCHECK` asks readiness.** Docker has one health status
and no restart-on-unhealthy by default, so "can this instance serve" is the
useful thing for `docker ps` and Compose `depends_on: condition:
service_healthy` to report.

**What "ready" means: the database can be read.** A trivial schema read
(`SELECT count(*) FROM sqlite_master`) within one second, not
`sql.DB.PingContext`, which can pass on a file that opens and then fails its
first query (corruption, a bad restore, a disk error). A test pins that a
non-database file fails the probe. It's the one dependency a request can't
be served without. Forge
reachability is not checked: GitHub or Forgejo being down doesn't stop the
dashboard serving stored data, and a probe depending on third parties would
flap.

**Deadline under the checker's.** The probe gets 1s and the `healthcheck`
subcommand's own request 2s, so a hung database yields a 503 the checker can
read instead of the checker timing out first.

**No leak.** The endpoint is unauthenticated, so a failure returns
`{"code":"not_ready","message":"not ready"}` and the error, which can carry
a file path, goes to the log at `Error`.

**Not request-logged.** Polled continuously like `/healthz`; a failing probe
logs its own cause, so a failure is never silent.

## Risks / Trade-offs

- **A transient database stall marks the container unhealthy for a cycle.**
  Accepted: three consecutive failures (`retries=3`) are needed, and nothing
  restarts on that status.
- **`Deps.Ready == nil` means ready.** A `Deps` built without a probe still
  answers 200, which keeps unrelated tests simple; production wiring always
  sets it, and a test pins that a closed connection goes red.
