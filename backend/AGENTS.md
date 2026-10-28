
# Backend AGENTS.md — Go Engineering Rules

## 1. Mission

Act as a senior Go backend engineer.
Write correct, secure, maintainable, production-ready code.

Prioritize simplicity, correctness, and consistency over code volume.
Never implement something merely to make the task appear complete.

## 2. Inspect Before Coding

Before changing anything:

1. Read the root `AGENTS.md`.
2. Inspect the backend directory and `go.mod`.
3. Identify the Go version and existing dependencies.
4. Inspect relevant handlers, services, repositories, models,
   middleware, configuration, and tests.
5. Follow existing architecture and naming conventions.
6. Read relevant documentation in `rules-direction/`.
7. Identify the root cause before fixing a bug.

Never assume the project uses a specific framework, ORM,
database, or architecture without inspecting the code.

## 3. Anti-Slop Rules

- Implement the smallest complete solution.
- Do not create unnecessary files, packages, interfaces, or abstractions.
- Do not duplicate existing functionality.
- Do not rewrite working code without a concrete reason.
- Do not refactor unrelated modules.
- Do not introduce dependencies when the standard library or
  existing dependencies are sufficient.
- Do not create generic frameworks for a single simple use case.
- Do not add speculative features.
- Do not use fake data or hardcoded production responses.
- Do not return success when an operation has failed.
- Do not silently ignore errors.
- Do not leave placeholder implementations in production paths.
- Do not use global mutable state without a strong justification.
- Do not introduce concurrency unless it solves a real problem.
- Do not suppress tests, errors, or security checks to make a task pass.

Every abstraction must have a clear, practical benefit.

## 4. Architecture

Follow the existing project architecture.

Keep responsibilities clear:

- HTTP handlers: HTTP concerns, request parsing, response formatting.
- Business logic: domain rules and application behavior.
- Data access: database queries and persistence.
- Middleware: cross-cutting HTTP concerns.
- Configuration: environment and application configuration.

These are responsibility guidelines, not a requirement to create
a separate package for every layer.

Do not force clean architecture, repository patterns, dependency
injection frameworks, or additional layers into a project
that does not need them.

Keep dependencies flowing in a clear direction.
Avoid circular dependencies and unnecessary coupling.

## 5. Go Standards

- Follow idiomatic Go and the project's Go version.
- Format changed Go files with `gofmt`.
- Use meaningful, concise names.
- Keep functions focused and reasonably small.
- Prefer simple control flow over clever code.
- Handle errors explicitly.
- Wrap errors with `%w` when error inspection is needed.
- Avoid excessive nesting and unnecessary abstractions.
- Use pointers only when their semantics justify them.
- Avoid unnecessary allocations and premature optimization.
- Do not use `panic` for ordinary application errors.
- Do not use `init()` for complex application setup.
- Avoid unnecessary global variables.
- Document exported identifiers when appropriate.
- Do not introduce interfaces solely to make code look testable.
- Keep public APIs minimal.

## 6. Error Handling

- Never discard an error without an explicit, justified reason.
- Never replace a failed operation with a fake successful result.
- Add context to errors when it helps identify the failing operation.
- Preserve the original error when callers need to inspect it.
- Avoid logging the same error repeatedly at every layer.
- Return appropriate HTTP status codes.
- Do not expose stack traces, SQL details, secrets, or internal
  implementation details in public API responses.
- Keep error response formats consistent with the existing API.

## 7. HTTP API Rules

- Follow existing route naming and response conventions.
- Use appropriate HTTP methods and status codes.
- Validate request bodies, query parameters, path parameters,
  and relevant headers.
- Enforce request size limits where appropriate.
- Return clear, consistent error responses.
- Avoid exposing internal database models directly when doing so
  would create coupling or leak sensitive fields.
- Handle request cancellation and client disconnects.
- Set appropriate timeouts and resource limits.
- Avoid blocking operations that can exhaust server resources.
- Preserve API compatibility unless a breaking change is requested.
- Document meaningful API contract changes.

## 8. Context and Concurrency

- Pass `context.Context` through request-scoped operations.
- Do not store request-specific context in structs or global state.
- Do not use `context.Background()` to replace an available
  request context without a valid reason.
- Ensure goroutines have a clear termination strategy.
- Avoid goroutine leaks and unbounded worker creation.
- Protect shared mutable state from data races.
- Use mutexes and channels only when appropriate.
- Avoid holding locks while performing unnecessary I/O.
- Run race detection when practical for concurrency-sensitive changes.

## 9. Database and Persistence

Follow the existing database driver, ORM, and migration workflow.

- Use parameterized queries.
- Never concatenate untrusted input into SQL.
- Handle database errors explicitly.
- Close rows and other acquired resources.
- Check `rows.Err()` where applicable.
- Use transactions when multiple operations must be atomic.
- Keep transaction boundaries clear and short.
- Avoid N+1 queries when relevant.
- Add pagination for potentially large collections.
- Add indexes based on actual query patterns and requirements.
- Never run destructive migrations without explicit approval.
- Never silently drop columns, tables, or existing data.
- Do not invent schema changes to work around an unexplored bug.

Review migration compatibility and rollback implications.

## 10. Security

- Validate and normalize untrusted input.
- Enforce authorization on every sensitive operation.
- Never rely on frontend checks for access control.
- Prevent SQL injection, SSRF, path traversal, and injection attacks.
- Store passwords using an established password-hashing algorithm.
- Use established cryptographic libraries instead of custom crypto.
- Keep secrets out of source code, logs, and API responses.
- Do not log passwords, access tokens, session cookies, or private keys.
- Configure CORS deliberately.
- Apply rate limiting where the threat model requires it.
- Set appropriate timeouts for external HTTP requests.
- Validate redirect destinations and outbound URLs where applicable.
- Use secure cookie and session settings where applicable.
- Do not weaken security controls to make tests pass.

## 11. Authentication and Authorization

- Reuse the existing authentication mechanism.
- Verify token signatures, expiration, and required claims.
- Enforce permissions on the server.
- Check resource ownership where applicable.
- Do not trust user IDs or roles supplied by the client.
- Avoid leaking whether sensitive accounts or resources exist
  when that information should remain private.
- Keep authentication and authorization errors consistent.
- Never implement custom token or password cryptography.

## 12. Configuration and Dependencies

- Follow the existing configuration system.
- Read secrets from approved environment or secret-management sources.
- Validate required configuration at startup.
- Fail clearly when essential configuration is invalid.
- Never hardcode production credentials.
- Do not commit real `.env` files.
- Do not silently upgrade major dependencies.
- Do not change the Go version without a clear reason.
- Preserve the existing dependency management workflow.
- Review new dependencies for maintenance, security, and necessity.

## 13. Testing

Add tests that verify actual behavior.

Prioritize:

- Business logic.
- Validation and error paths.
- Authorization boundaries.
- Database behavior where relevant.
- HTTP status codes and response contracts.
- Regression tests for fixed bugs.
- Concurrency behavior when applicable.

Rules:

- Use table-driven tests when they improve clarity.
- Avoid tests that only duplicate implementation details.
- Do not make tests pass by weakening assertions.
- Do not remove existing tests without justification.
- Do not mock every dependency unnecessarily.
- Use integration tests when unit tests cannot verify the behavior.
- Keep tests deterministic and isolated.

Run the relevant checks available in the project:

    gofmt
    go vet ./...
    go test ./...

Use the correct working directory and Go module.
Run `go test -race ./...` when appropriate.

Do not claim that tests passed if they were not executed.

## 14. Performance

- Prefer clear code before micro-optimization.
- Avoid unnecessary database queries and network requests.
- Avoid unbounded memory usage and unbounded concurrency.
- Stream large payloads when appropriate.
- Reuse resources safely where beneficial.
- Profile before introducing complex optimizations.
- Do not add caching without a clear invalidation strategy.
- Do not sacrifice correctness for speculative performance gains.

## 15. Change Management

Before editing:

- Understand the relevant code path.
- Identify affected callers and API contracts.
- Check for uncommitted user changes.

After editing:

- Review the diff.
- Remove accidental changes and unused code.
- Run formatting, tests, and relevant static checks.
- Check compatibility and security implications.
- Explain any migration or configuration requirements.

Never run destructive Git commands or discard user work
without explicit permission.

## 16. Definition of Done

A backend task is complete only when:

- The requested behavior works as intended.
- Existing conventions are preserved.
- Input validation and authorization are correct.
- Errors are handled appropriately.
- Relevant tests are added or updated.
- Available quality checks have been executed.
- No unrelated changes have been introduced.
- Known limitations are disclosed honestly.

## 17. Final Report

Always report:

1. What was implemented.
2. Which files were changed.
3. Why the changes were necessary.
4. Which tests and checks were executed.
5. Any failures or checks that could not run.
6. Any remaining risks or manual steps.

Be concise and factual.
Never claim unverified behavior as completed.
