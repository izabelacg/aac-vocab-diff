# Development Preferences

## Planning
- Implement in phases: perform necessary refactors first, and structure
  each phase as a self-contained commit.
- Always present the full phased plan with code snippets before writing
  any implementation. Wait for my approval before proceeding.
- In plan snippets, add inline comments to explain non-obvious lines.
  In actual code, only comment where the code is not self-evident.

## Design & Architecture
- When multiple patterns apply, present 2–3 options with pros/cons
  before choosing a path — prioritize pluggability, extensibility,
  and maintainability.

## Execution Mode
- Before writing code, ask which mode I want for this task:
  - "Implement it" mode: you make the code changes end-to-end after plan approval
  - "TDD support" mode: you write the tests first, I implement, and you wait for me to signal when to move to the next phase
- Do not assume the mode, even if the task looks similar to a previous one
- In "TDD support" mode, stop after writing or updating the tests and clearly state what I should implement next
- After I reply that implementation is done, review the result, suggest the next phase, and continue only with my approval

## Testing
- Follow TDD: write failing tests first per phase, share for review,
  then implement to make them pass.
- Write tests as guardrails (protect contracts/boundaries for safe
  refactoring). Avoid duplicate, redundant, or implementation-detail
  tests.

## Git
- Never commit changes without explicit user approval.
- Use conventional commit format: `type(scope): short description`
  — common types: `feat`, `fix`, `refactor`, `chore`, `docs`, `test`
  — example: `feat(server): add per-IP rate limiting on POST /diff`
- In plan documents: include a suggested commit line at the end of each
  phase or item.
- After completing any code change: proactively suggest the commit
  message, then wait for a "go ahead" before running git commit.

## Learning
- I want to understand the "why" — explain design decisions, tradeoffs,
  and patterns as we go.
