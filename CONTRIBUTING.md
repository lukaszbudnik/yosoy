# Contributing to yosoy

Thanks for your interest in yosoy! Contributions are welcome: bug fixes, roadmap items, documentation and well-argued new ideas.

## Scope

yosoy aims to stay a **single small binary with zero required configuration**. A change fits if it helps answer:

- *"What does the network path between the caller and this pod look like?"*, or
- *"How does my infrastructure (ingress, load balancer, mesh, cache) behave when the backend does X?"*

Read the guiding principle and the **Non-goals** section in [ROADMAP.md](ROADMAP.md) before proposing a new feature.

## Before you start

- **Small fixes** (typos, obvious bugs with a test): just open a pull request.
- **Anything bigger** (new endpoints, env variables, dependencies, response format changes): open an issue first, or comment on an existing issue or roadmap item, and agree on the approach. Unsolicited large PRs are likely to be closed.
- **Security vulnerabilities:** do **not** open a public issue. Report them privately via GitHub's *Report a vulnerability* button (Security tab) or by e-mail to lukasz.budnik@gmail.com.

## Development

The required Go version is defined in [`go.mod`](go.mod).

```bash
gofmt -l .                     # must print nothing
go vet ./...
go test -race ./...
```

Build and smoke-test the Docker image (this is what CI does):

```bash
docker build -t yosoy-local:latest .
docker run --rm -p 8080:80 yosoy-local:latest
curl -s -X DELETE 'http://localhost:8080/sample/path?with=params'
```

## Pull request guidelines

- **One concern per PR.** Don't mix refactoring, formatting and features.
- **Link the issue or roadmap item** (e.g. `Closes #123`, `ROADMAP 1.2`).
- **Add or update tests.** Bug fixes need a test that fails without the fix.
- **Keep backward compatibility.** Existing endpoints, JSON field names, env variables, the default port and the default behaviour must not change without prior agreement. New behaviour should be opt-in.
- **Update the docs.** If behaviour changes, update `README.md` (and `ROADMAP.md` status / work log for roadmap items).
- **No new dependencies** without discussion. The standard library is preferred.
- **Don't change `.github/workflows/`, `Dockerfile` or release tooling** unless that is the agreed purpose of the PR.
- Write the PR description yourself: what changed, why, and how you tested it.

## AI-assisted contributions

AI-assisted contributions are welcome, on the same terms as any other contribution. You, not the tool, are the author and are accountable for the change. In practice:

1. **Disclose it.** State in the PR description whether and how AI tools were used (e.g. "code generated with X, reviewed and tested by me", "AI used for docs only", "no AI used").
2. **Understand every line.** You must be able to explain and defend the change during review without pasting the question back into a tool.
3. **Run it.** Run the commands from [Development](#development) locally before opening the PR, and say so.
4. **Be concise.** Keep issues and PR descriptions short and specific. Don't paste long generated summaries, speculative "improvements" or unrelated clean-ups.
5. **Respect the scope.** Generated features outside the project scope or roadmap will be closed.

PRs that ignore these guidelines may be closed without a detailed review. That isn't personal: maintainer time is the scarcest resource in this project.

## License

yosoy is licensed under the [Apache License 2.0](LICENSE). By submitting a contribution you agree that it is licensed under the same terms.
