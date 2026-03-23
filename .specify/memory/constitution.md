<!--
Sync Impact Report:
- Version Change: Template → 1.0.0
- Principles Established:
  - I. Clean Code
  - II. Domain Driven Design (DDD)
  - III. Mandatory Unit & Integration Testing (Non-Negotiable)
  - IV. Observability
- Added Sections: Quality Standards, Development Workflow
- Templates Updated:
  - .specify/templates/tasks-template.md (✅ updated to enforce mandatory testing)
  - .specify/templates/plan-template.md (✅ compatible via dynamic gates)
  - .specify/templates/spec-template.md (✅ compatible via existing testing section)
-->

# GopherTrade Constitution

## Core Principles

### I. Clean Code
Code must be written for humans first, computers second. We prioritize readability, simplicity, and maintainability over cleverness.
- **Naming**: Variables, functions, and types must reveal intent.
- **Functions**: Should do one thing and do it well. Keep them small.
- **Comments**: Explain "why", not "what". Code should be self-documenting.
- **Refactoring**: Leave the campground cleaner than you found it. Refactoring is part of the development process, not a separate phase.

### II. Domain Driven Design (DDD)
The software structure must mirror the business domain it solves.
- **Ubiquitous Language**: Use the same terminology in code as the business experts use.
- **Bounded Contexts**: Respect boundaries. Models in one context (e.g., "Sales") may differ from another (e.g., "Shipping").
- **Layered Architecture**: Isolate domain logic from infrastructure and UI. The core domain should not depend on external frameworks.

### III. Mandatory Unit & Integration Testing (NON-NEGOTIABLE)
No feature is "done" without comprehensive tests.
- **Unit Tests**: Cover core logic and domain rules. Fast and isolated.
- **Integration Tests**: Verify flow between components (DB, API, Services).
- **TDD**: Test-Driven Development is encouraged. Write the test, make it fail, then make it pass.
- **Coverage**: Must be meaningful. High coverage numbers are good, but testing critical paths is better.
- **Zero Regressions**: Bugs must be reproduced with a test case before fixing.

### IV. Observability
Systems must be debuggable in production without connecting a debugger.
- **Structured Logging**: Use JSON or structured formats. Context (IDs, trace info) must be propagated.
- **Metrics**: Expose key health and performance metrics (latency, error rates, throughput).
- **Tracing**: Request tracing across boundaries to visualize flows.
- **Errors**: Error messages must be explicit, actionable, and logged with context.

## Quality Standards

- **Idiomatic Code**: Adhere to the standard idioms of the language (e.g., Effective Go).
- **Static Analysis**: Linting is mandatory. Zero-warning policy on the main branch.
- **Dependency Management**: Dependencies must be justified and pinned. Avoid "kitchen sink" libraries.

## Development Workflow

- **Plan First**: No code is written without a plan and spec (using `speckit`).
- **Review**: All changes require a Pull Request with a code review.
- **CI/CD**: Tests must pass in CI before merging.
- **Documentation**: Documentation (README, inline) is treated as code and updated atomically.

## Governance

This Constitution serves as the primary source of engineering truth for GopherTrade.
- **Amendments**: Changes to this document require a formal Pull Request and team consensus.
- **Compliance**: All code reviews must check against these principles. Non-compliant code will be rejected.
- **Version**: 1.0.0
- **Ratified**: 2026-03-21
- **Last Amended**: 2026-03-21
