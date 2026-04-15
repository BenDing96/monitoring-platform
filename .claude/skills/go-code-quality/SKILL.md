---
name: go-code-quality
description: Review Go code in this project against the standard quality bar — formatting, vet, idiomatic style, error handling, concurrency, and testing. Use when the user asks to review, audit, or check Go code quality, or after non-trivial Go changes.
---

# Go Code Quality Standard

Apply this checklist when reviewing or writing Go code in `monitoring-platform`.

## Automated checks (run these first)

```bash
gofmt -l .          # must print nothing
go vet ./...        # must pass
go build ./...      # must compile
go test ./...       # must pass
```

If `golangci-lint` is installed, also run `golangci-lint run ./...`.

## Style

- Package names: short, lowercase, no underscores, no plurals.
- Exported identifiers have doc comments starting with the identifier name.
- Prefer short, descriptive receiver names (1–2 letters), consistent per type.
- No stuttering: `http.Server` not `http.HTTPServer`; `user.New` not `user.NewUser`.
- Group related declarations with `const (...)` / `var (...)` blocks.
- Imports: stdlib first, then third-party, then local — separated by blank lines.

## Errors

- Return errors, don't panic (except `init`/truly unrecoverable).
- Wrap with context: `fmt.Errorf("doing X: %w", err)`.
- Check errors immediately; don't assign to `_` without a reason comment.
- Sentinel errors: `var ErrNotFound = errors.New("not found")`, compared with `errors.Is`.
- Custom error types: implement `Error() string`, match with `errors.As`.

## Control flow

- Happy path left-aligned; early-return on errors.
- Avoid deep nesting — invert conditions and return early.
- No naked returns in long functions.

## Concurrency

- Every goroutine has a clear owner and exit condition.
- Pass `context.Context` as the first parameter; respect cancellation.
- Protect shared state with mutexes or channels — document the choice.
- Never close a channel from the receiver side.
- Use `errgroup` for fan-out/fan-in with error propagation.

## Testing

- Table-driven tests with `t.Run(tc.name, ...)` for subtest naming.
- Use `t.Helper()` in test helpers.
- Prefer `testing.T.TempDir()` over manual temp-dir cleanup.
- Don't use `time.Sleep` for synchronization — use channels or `sync` primitives.
- Benchmarks live next to the code they measure.

## Project-specific

- Module path: `monitoring-platform` (local-only for now).
- Entry point: [main.go](../../../main.go).
- Build/test via [Makefile](../../../Makefile) targets: `make build`, `make test`, `make vet`, `make fmt`, `make tidy`.

## Review output format

When reviewing, report findings grouped by severity:

1. **Blocking** — fails `gofmt`/`vet`/`build`/`test`, or introduces bugs.
2. **Should fix** — violates the standards above.
3. **Nits** — stylistic suggestions.

For each finding: file:line, the issue, and the fix.
