# Contributing to dbt-guard

Thanks for considering a contribution. dbt-guard is a small, focused Go CLI — keep changes scoped and avoid speculative abstractions.

## Reporting bugs and proposing features

Open a [GitHub Issue](https://github.com/renatoprovi/dbt-guard/issues). For bugs, include your dbt version, the relevant `manifest.json` snippet (redact anything sensitive), and the command you ran.

## Development setup

Requirements: [Go 1.22+](https://go.dev/dl/).

```bash
git clone https://github.com/renatoprovi/dbt-guard.git
cd dbt-guard
go build ./...
go test ./...
./scripts/test-e2e.sh
```

See [docs/TESTING.md](docs/TESTING.md) for the full test matrix and [examples/README.md](examples/README.md) for the sample dbt project used in manual testing.

## Making a change

1. Fork the repo and create a branch off `main`.
2. Make your change. Add or update tests under the relevant `internal/*` package — new behavior without a test won't be merged.
3. Before opening a PR, run locally:
   ```bash
   gofmt -l .        # must print nothing
   go vet ./...
   go build ./...
   go test ./...
   ./scripts/test-e2e.sh
   ```
   This is exactly what CI checks; it's faster to catch issues locally.
4. Open a pull request against `main`. Describe *why* the change is needed, not just what it does.

## Review and merge process

Every change lands through a pull request — including the maintainer's own — and merges only after CI is green and a maintainer review/approval. Nobody pushes directly to `main`. This isn't bureaucracy for its own sake: it keeps `main` always in a state anyone can build a release from.

Only the maintainer merges into `main`; that's intentional for a young project and may change as trusted contributors emerge.

## Scope

Before adding a feature, check [docs/ROADMAP.md](docs/ROADMAP.md). If what you want to build isn't listed under "Future," open an issue to discuss it first — it saves everyone a rejected PR.
