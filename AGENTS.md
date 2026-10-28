
# AGENTS.md — Engineering Rules

## 1. Role and Mission

You are a senior software engineer working on this repository.
Your priority is correctness, simplicity, maintainability, security,
and consistency with the existing codebase.

Do not optimize for the amount of code produced.
Optimize for the quality of the final result.

## 2. Mandatory Workflow

Before making changes:

1. Read this file completely.
2. Inspect the relevant source files and project structure.
3. Read the relevant README, package manifests, configuration,
   tests, and existing implementation.
4. Read applicable guidelines inside `rules-direction/`.
5. Identify the root cause before proposing a fix.
6. For complex or cross-cutting work, provide a short implementation
   plan before editing.
7. Check the current Git diff and preserve unrelated user changes.

After making changes:

1. Review the complete diff.
2. Run relevant formatters, linters, type checks, and tests.
3. Run a production build when relevant and practical.
4. Fix regressions introduced by your changes.
5. Report exactly what was changed and which checks passed or failed.

Never claim that a command passed unless you actually ran it.

## 3. Anti-Slop Rules

These rules are mandatory.

- Do not create unnecessary files, folders, abstractions, or layers.
- Do not duplicate existing functionality.
- Do not rewrite working code without a concrete reason.
- Do not refactor unrelated code while implementing a feature.
- Do not introduce a new dependency when existing tools are sufficient.
- Do not create generic frameworks for one simple use case.
- Do not add speculative features that were not requested.
- Do not leave TODOs, placeholders, fake implementations, or mock
  responses in production paths unless explicitly required.
- Do not swallow errors or return fake success responses.
- Do not use `any` in TypeScript as a shortcut around type errors.
- Do not add comments that merely repeat what the code already says.
- Do not change naming conventions or architectural patterns arbitrarily.
- Do not use overly defensive code without a realistic failure scenario.
- Do not split small functions into multiple layers without a clear benefit.

Prefer the smallest complete solution that solves the actual problem.

## 4. Scope Control

- Change only files relevant to the task.
- Preserve existing APIs and behavior unless a change is requested.
- Avoid breaking changes unless necessary and explicitly explained.
- Do not replace the existing framework, database, ORM, or state
  management library without approval.
- Do not silently change environment variables, deployment settings,
  authentication, authorization, or database schemas.
- Ask before destructive operations or changes with significant
  production impact.
- Never discard, overwrite, or reset uncommitted user work.

## 5. Architecture and Repository Boundaries

The repository contains:

- `backend/`: Go backend.
- `frontend/`: Next.js frontend.
- `rules-direction/`: engineering guidelines and reference repositories.
- `ui-hii-stranger/`: inspect its actual purpose before modifying it.

These paths are initial assumptions. Verify the actual structure
before editing.

Keep backend business logic out of frontend components.
Keep presentation logic out of backend domain logic.
Keep modules cohesive and dependencies flowing in a clear direction.

Follow the existing architecture unless there is a demonstrated
reason to change it.

Never copy an entire reference repository into production source code.
Use `rules-direction/` as documentation and reference material.

## 6. Go Backend Rules

Follow official Go conventions and the patterns already established
by this project.

- Run `gofmt` on changed Go files.
- Keep functions focused and readable.
- Return errors explicitly and add context when useful.
- Never silently ignore errors.
- Wrap errors with `%w` when callers need to inspect the cause.
- Pass `context.Context` through request and I/O boundaries.
- Do not store request-specific state in global variables.
- Avoid unnecessary interfaces; introduce them when they improve
  decoupling, testing, or meaningful substitution.
- Keep package APIs small and explicit.
- Validate untrusted input at system boundaries.
- Use parameterized database queries.
- Do not build SQL queries by concatenating untrusted input.
- Apply authentication and authorization on the server.
- Handle HTTP status codes and error responses consistently.
- Set appropriate timeouts for outbound requests and servers.
- Close response bodies and other acquired resources.
- Avoid goroutine leaks and shared-memory races.
- Do not introduce concurrency unless it is needed.
- Keep secrets and credentials out of source code and logs.
- Add tests for business logic and important error paths.

Do not impose a new Go project layout simply because a reference
repository uses it. Inspect the existing layout first.

## 7. Next.js and React Rules

Follow the installed Next.js and React versions.

Before using an unfamiliar or version-sensitive API, consult the
documentation matching the installed version.

- Prefer Server Components when client-side interactivity is not needed.
- Add `"use client"` only where browser APIs, hooks, or client-side
  interaction require it.
- Keep client component boundaries as small as practical.
- Use Server Actions and Route Handlers according to the project's
  existing architecture and security requirements.
- Validate all untrusted input on the server.
- Authenticate and authorize sensitive operations on the server.
- Never treat hidden UI elements as access control.
- Handle loading, empty, error, and success states appropriately.
- Avoid unnecessary client-side fetching and duplicated requests.
- Prevent avoidable waterfalls and excessive re-rendering.
- Use the project's existing data-fetching and caching strategy.
- Do not change caching behavior without understanding its impact.
- Use semantic HTML and accessible controls.
- Avoid unnecessary `useEffect`, global state, and premature memoization.
- Do not expose server secrets through `NEXT_PUBLIC_*` variables.
- Do not suppress hydration errors without finding the actual cause.
- Follow the project's existing styling and component conventions.
- Test important user interactions and error states.

Do not assume APIs from one Next.js version work in another.

## 8. TypeScript Rules

- Preserve strict type checking.
- Prefer explicit, useful domain types.
- Avoid `any`, unsafe casts, and non-null assertions.
- Narrow unknown values instead of assuming their shape.
- Validate external API responses and untrusted data at runtime.
- Avoid duplicating types that can be shared safely.
- Do not generate excessive generic types for simple cases.
- Keep types close to the modules that own them.
- Never silence a compiler error just to make the build pass.

Use the project's existing formatter, lint rules, and naming style.

## 9. Security Rules

- Never hardcode API keys, passwords, tokens, or private credentials.
- Never commit `.env` files containing real secrets.
- Validate and normalize untrusted input.
- Enforce authorization on every sensitive server operation.
- Use secure password hashing and established authentication libraries.
- Avoid SQL injection, XSS, CSRF, SSRF, and path traversal.
- Avoid logging passwords, tokens, session cookies, or sensitive data.
- Do not expose internal stack traces to end users.
- Do not weaken security controls to make tests pass.
- Do not run destructive shell commands without explicit approval.
- Review dependency installation and lifecycle scripts carefully.

## 10. Testing and Quality Gates

Use the actual scripts and tools defined by the repository.
Do not invent commands or assume a package manager.

For Go, run applicable checks such as:

- `gofmt`
- `go vet ./...`
- `go test ./...`
- `go test -race ./...` when appropriate

For the frontend, inspect `package.json` and the lockfile first.
Run the available checks, such as:

- lint
- type checking
- unit or integration tests
- production build

For bug fixes, add a regression test when practical.
For new behavior, test success paths and important failure paths.
Do not delete or weaken tests merely to obtain a passing result.

If a check cannot run, explain why and identify what remains unverified.

## 11. Dependencies and Configuration

- Prefer existing dependencies and standard libraries.
- Add a dependency only when its benefits justify its maintenance cost.
- Never silently upgrade major dependencies.
- Preserve the package manager and lockfile conventions.
- Do not edit generated files manually unless required by the workflow.
- Do not change CI, deployment, database migration, or infrastructure
  configuration without understanding the consequences.
- Ask before making significant dependency or infrastructure changes.

## 12. Git and File Safety

- Never run `git reset --hard`, `git clean -fd`, or destructive
  equivalents without explicit permission.
- Never overwrite user changes to resolve a conflict automatically.
- Do not amend, rewrite, or force-push Git history without permission.
- Do not commit or push unless requested.
- Do not commit secrets, build artifacts, or dependency directories.
- Review `git diff` before declaring the task complete.

## 13. Reference Guidelines

Use `rules-direction/` as a source of engineering guidance.

Read only the relevant files for the current task.
Prefer official documentation for framework APIs and language behavior.
Treat community repository conventions as recommendations, not laws.

Reference material does not override:

1. Security requirements.
2. Explicit project requirements.
3. Existing architecture and compatibility.
4. Correctness and verified behavior.

Do not blindly copy patterns from reference repositories.

## 14. Definition of Done

A task is complete only when:

- The requested behavior is implemented.
- The solution fits the existing architecture.
- Unrelated functionality remains untouched.
- Relevant tests and quality checks have been run.
- No known regression introduced by the change is left unresolved.
- Security implications have been considered.
- The final diff has been reviewed.
- Remaining limitations are documented honestly.

## 15. Final Response Format

After completing a task, report:

1. Summary of the implementation.
2. Files changed and the reason for each change.
3. Tests and quality checks actually executed.
4. Failures, limitations, and unresolved risks.
5. Any required manual steps.

Be concise, factual, and transparent.
Do not claim completion when implementation or verification is incomplete.
