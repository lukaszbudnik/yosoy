# yosoy roadmap & work log

Living document tracking planned improvements. Items are addressed top to bottom.
Each item has: **Problem**, **Proposal**, **Acceptance criteria**, and a **Status**.

Status legend: ⬜ todo · 🟨 in progress · ✅ done · ⏸️ parked

**Guiding principle:** yosoy stays a *single small binary with zero required configuration*. A feature belongs here only if it helps answer *"what does the network path between the caller and this pod look like?"* or *"how does my infra (ingress, LB, mesh, cache) behave when the backend does X?"*. Anything else is out of scope (see [Non-goals](#non-goals)).

---

## Phase 0 — Positioning & README

### 0.1 Reposition as a Kubernetes networking troubleshooting tool ✅

**Problem:** The README leads with "stubbing and prototyping distributed applications". The space of "echo servers" is crowded (`traefik/whoami`, `mendhak/http-https-echo`, `ealen/echo-server`, `mccutchen/go-httpbin`, `stefanprodan/podinfo`), and the README does not explain why someone should pick yosoy over them. The real differentiators — the reachability analyzer, Downward API file exposure, and CORS-friendly JSON for SPAs — are buried.

**Proposal:**
- New tagline along the lines of: *"A tiny HTTP pod that tells you who it is, who called it, and what it can reach."*
- Short "Why yosoy?" section + honest comparison table vs `whoami`, `http-https-echo`, `go-httpbin`, `podinfo`, `netshoot`.
- Lead with a Kubernetes quick start (`kubectl run` one-liner + `test/deployment.yaml`).
- Fix the **HTTP/2 claim**: the README says HTTP/2 is supported over TLS, but the server only listens on plaintext `:80`; HTTP/2 is only exercised in a unit test via `httptest.StartTLS()`. Either implement 2.3 or correct the docs.
- Add a security note: `YOSOY_SHOW_ENVS` exposes secrets injected as env vars; the ping endpoint can be used to probe internal networks.

**Acceptance criteria:**
- [x] A new visitor understands within the first screen what yosoy is for and why it differs from `whoami`.
- [x] No documented feature the binary does not actually provide.
- [x] Security note present.

### 0.2 Contribution guidelines ✅

**Problem:** There is no `CONTRIBUTING.md`, so contributors have no stated scope, quality bar or policy for AI-assisted contributions.

**Proposal:** Add `CONTRIBUTING.md` covering scope, when to open an issue first, dev commands, PR guidelines (backward compatibility, no new deps, no workflow changes), AI-assisted contribution rules (disclosure, accountability, run tests) and private security reporting.

**Acceptance criteria:**
- [x] `CONTRIBUTING.md` linked from `README.md`.
- [x] Dev commands in `CONTRIBUTING.md` pass on `main` (`gofmt`, `go vet`, `go test -race`).

---

## Phase 1 — Correctness & hygiene

### 1.1 Fix data race on request counter ⬜

**Problem:** `counter++` is executed concurrently from every request goroutine without synchronisation. `go test -race` with parallel requests will flag it, and counts are unreliable under load — exactly when someone uses yosoy to test load balancing.

**Proposal:** Use `sync/atomic` (`atomic.Int64`).

**Acceptance criteria:**
- [ ] Test issuing N parallel requests passes under `go test -race`; final counter == N.
- [ ] CI runs tests with `-race`.

### 1.2 Fix missing `Content-Type` on ping responses ⬜

**Problem:** In `ping()` `w.WriteHeader()` is called *before* `w.Header().Add("Content-Type", ...)`, so the header is silently dropped (the README sample even shows `text/plain`).

**Proposal:** Set headers before `WriteHeader`; extract a small `writeJSON(w, status, v)` helper used by all handlers.

**Acceptance criteria:**
- [ ] All JSON responses carry `Content-Type: application/json; charset=utf-8` (asserted in tests).

### 1.3 Configurable port, non-root image, graceful shutdown ⬜

**Problem:**
- Port `80` is hard-coded → the container must run as root or with `NET_BIND_SERVICE`, which fails under Pod Security Standards `restricted`.
- `http.ListenAndServe` error is ignored (bind failures exit silently with code 0).
- No graceful shutdown on `SIGTERM` → in-flight requests are dropped during rollouts, which muddies LB / rollout testing.

**Proposal:**
- `YOSOY_PORT` (default `80` for backward compatibility; document `8080` for restricted clusters).
- Image runs as non-root (`USER 65532`), optionally switch runtime to `distroless/static` or `scratch` with `CGO_ENABLED=0`.
- `http.Server` with `Shutdown(ctx)` on `SIGTERM`, configurable `YOSOY_SHUTDOWN_DELAY` (useful to demonstrate/endpoint-propagation issues).
- Log and exit non-zero on listen errors.

**Acceptance criteria:**
- [ ] `test/deployment.yaml` runs under `restricted` PSS with `runAsNonRoot: true`.
- [ ] `SIGTERM` drains in-flight requests (test).
- [ ] Default behaviour unchanged for existing users.

### 1.4 Multi-arch images & supply chain ⬜

**Problem:** `docker-publish.yml` builds only `linux/amd64`; arm64 (Apple Silicon kind/minikube, Graviton, Ampere) users get emulation or nothing. Images are unsigned, no SBOM.

**Proposal:**
- `docker/setup-qemu-action` + `docker/setup-buildx-action`, `platforms: linux/amd64,linux/arm64`.
- Enable provenance + SBOM in `build-push-action`; sign with cosign (keyless, GitHub OIDC).
- Remove `GOPROXY=direct` from the Dockerfile (bypasses the checksum-verified module proxy).

**Acceptance criteria:**
- [ ] `docker manifest inspect lukasz/yosoy:edge` lists amd64 + arm64.
- [ ] `cosign verify` succeeds for released tags.

### 1.5 CI & Dependabot maintenance load ✅

**Problem:** Dependabot runs **daily** across three ecosystems for a ~200-line project — most of the 170 commits are dependency bumps. Go version is inconsistent (`go.mod` 1.25, CI 1.25, Dockerfile 1.27).

**Proposal:** Apply the same setup as `migrator`: monthly grouped updates, native `dependabot/fetch-metadata` + `gh pr merge --auto`; single source of truth for Go version (`go-version-file: go.mod`); add `go vet` (and later `golangci-lint`, see B10).

**Acceptance criteria:**
- [x] ≤ 3 grouped Dependabot PRs per month for minor/patch updates, auto-merged when green (majors stay manual).
- [x] CI Go version read from `go.mod` (the Dockerfile builder image is kept current by Dependabot).
- [x] CI runs `go vet` and `go test -race`.

### 1.6 Replace gorilla/mux with stdlib routing ⏸️ (optional)

**Problem:** Go ≥ 1.22 `http.ServeMux` supports method + wildcard patterns; `gorilla/mux` and `gorilla/handlers` are the only runtime deps.

**Proposal:** Migrate to stdlib, keep logging / proxy-header / recovery middleware as small local functions. Only worth doing together with Phase 2 work.

---

## Phase 2 — Troubleshooting features (the reason to keep the project)

### 2.1 Richer reachability analyzer ⬜

**Problem:** `/_/yosoy/ping` returns only "succeeded" or a raw Go error string. When debugging k8s networking you usually want to know *which stage* failed and how long it took.

**Proposal:** Keep the existing endpoint backward compatible and add structured output:
- DNS stage: resolved IPs, resolver used, duration (distinguishes CoreDNS problems from network policy drops).
- Connect stage: duration, local/remote address used.
- Optional `tls=true`: handshake duration, negotiated version/ALPN, certificate chain subject/SANs/expiry — answers "is the mesh/ingress presenting the right cert?".
- Optional `http=<path>`: issue an HTTP(S) GET and return status, headers, timing.
- Error classification: `dns_error`, `timeout` (likely NetworkPolicy/firewall drop), `refused` (reachable host, nothing listening), `tls_error`.

**Acceptance criteria:**
- [ ] Existing `h/p/n/t` behaviour and response format unchanged unless new params/`Accept` used.
- [ ] Each error class covered by tests.

### 2.2 Controllable responses for infra testing ⬜

**Problem:** Testing retries, timeouts, circuit breakers, health-check behaviour and caching in ingress / service mesh needs a backend that misbehaves on demand. Today you need `go-httpbin` alongside yosoy.

**Proposal:** Query parameters or headers on the catch-all handler (e.g. `?status=503`, `?delay=2s`, `?size=1MiB`, `?cache=max-age=60`), plus an admin endpoint to flip readiness: `POST /_/yosoy/ready?state=false` and `GET /_/yosoy/healthz|readyz`. Optional failure ratio (`?fail=0.3`) to test outlier detection.

**Acceptance criteria:**
- [ ] Documented recipes: "test NGINX ingress retries", "test Istio outlier detection", "test readiness gate removal from Endpoints".
- [ ] Hard caps on delay/size to avoid self-DoS.

### 2.3 TLS and h2c ⬜

**Problem:** See 0.1 — HTTP/2 is documented but not served. TLS termination vs. passthrough is a common k8s debugging question.

**Proposal:** Optional `YOSOY_TLS_CERT`/`YOSOY_TLS_KEY` (or auto-generated self-signed cert) on a second port; optional h2c on the plaintext port. Response includes TLS info observed by the server (SNI, negotiated version, client cert subject for mTLS debugging).

**Acceptance criteria:**
- [ ] `curl --http2 -k https://...` and `curl --http2-prior-knowledge http://...` report `HTTP/2.0`.
- [ ] mTLS client cert subject visible in response when presented.

### 2.4 Pod & node identity out of the box ⬜

**Problem:** Users must wire the Downward API into files and set `YOSOY_SHOW_FILES` to see basic identity. Knowing *which node / zone* answered is the #1 question when testing topology-aware routing and `externalTrafficPolicy`.

**Proposal:** Documented env vars (`POD_NAME`, `POD_NAMESPACE`, `POD_IP`, `NODE_NAME`, zone label) surfaced in a dedicated `pod` JSON object when present, plus all local interface IPs (dual-stack debugging). Update `test/deployment.yaml`.

### 2.5 WebSocket & gRPC echo ⬜

**Problem:** Ingress/mesh WebSocket upgrades and gRPC (HTTP/2) routing are frequent failure points; yosoy cannot test them.

**Proposal:** `/_/yosoy/ws` echo endpoint; gRPC health + reflection + echo service on an optional port. Keep behind config to preserve the tiny footprint.

### 2.6 Make the ping endpoint opt-out / restrictable ⬜

**Problem:** Anyone who can reach yosoy can use `/_/yosoy/ping` as a port scanner against cluster-internal addresses (SSRF-style). Fine on a laptop, risky if yosoy is left running behind a public ingress.

**Proposal:** `YOSOY_PING_ENABLED` (default `true` to keep compatibility; consider flipping in v4) and optional `YOSOY_PING_ALLOW` CIDR / host allow-list.

---

## Phase 3 — Backlog (to be prioritised)

| # | Idea | Notes |
|---|------|-------|
| B1 | Prometheus `/metrics` | Request count per path/status, latency histogram; helps demo HPA / mesh telemetry. |
| B2 | Helm chart / Kustomize base | Deployment + Service + optional Ingress/Gateway API `HTTPRoute`. |
| B3 | Gateway API examples | `HTTPRoute` header/weight-based routing demo; replaces Ingress-only examples. |
| B4 | Request body echo | Return (truncated) request body; useful for proxies that rewrite bodies. |
| B5 | Response format negotiation | `Accept: text/plain` returns a `whoami`-style human readable output for terminal use. |
| B6 | Header redaction | Mask `Authorization` / `Cookie` values by default in response and logs (currently echoed verbatim). |
| B7 | UDP echo listener | Pair with ping `n=udp`, which today cannot really detect UDP reachability. |
| B8 | Structured JSON logging | `log/slog`, include request ID / trace headers. |
| B9 | OpenTelemetry trace propagation | Echo and continue `traceparent` to visualise mesh hops. |
| B10 | `golangci-lint` in CI | Deferred from 1.5: `errcheck` would currently fail (ignored `Encode`, `Close` and `ListenAndServe` errors); enable after 1.2/1.3. |

---

## Non-goals

- Becoming a full debugging toolbox (`tcpdump`, `dig`, shells) — use `nicolaka/netshoot` or `kubectl debug`.
- A general-purpose mock server with scripted responses (WireMock, Mockoon).
- Persisting state or requiring configuration to start.

---

## Work log

| Date | Item | Change | PR / commit |
|------|------|--------|-------------|
| 2026-11-28 | — | Roadmap created after project review | — |
| 2026-11-28 | 0.1 | README reworked: troubleshooting-focused tagline, mermaid diagram, Kubernetes quick start, use-case table, "How yosoy compares", configuration table, corrected HTTP/2 claim, security considerations, ping `Content-Type` known-bug note | `docs/positioning-and-ci` |
| 2026-11-28 | 0.2 | Added `CONTRIBUTING.md` (scope, PR rules, AI-assisted contribution policy, private security reporting); `gofmt` applied to `http_protocol_test.go` | `docs/positioning-and-ci` |
| 2026-11-28 | 1.5 | Dependabot switched from daily to monthly grouped updates (Actions, Go modules, Docker images), added native automerge workflow, CI reads Go version from `go.mod` and runs `go vet` + `go test -race` | `docs/positioning-and-ci` |
| 2026-11-28 | 1.5 | Fixed check-name clash (both Go and Docker jobs were named `build`, so a green Docker run could mask a failed Go run and let auto-merge merge a broken Dependabot PR): explicit job ids/names `Go Build and Test` / `Docker Image Build`, workflows triggered only for `main` (ported from migrator#794) | `docs/positioning-and-ci` |
