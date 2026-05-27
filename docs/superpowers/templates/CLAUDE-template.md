# [Project Name]

## Process

This project uses the **Constitution-Centric Six-Layer Workflow**:

```
Requirement → Superpower(brainstorm→spec) → TaskMaster(tasks) → Speckit(clarify→plan→implement) → DDD(code) → BMAD(dev+qa) → QA(gate)
```

Read `.specify/memory/constitution.md` for full constraints.

## Quick Reference

| Phase | Tool | Output |
|-------|------|--------|
| Design | Superpower brainstorming | `docs/superpowers/specs/*.spec.md` |
| Breakdown | TaskMaster CLI | `docs/speckit/<feature>/tasks.md` |
| Clarify | Speckit clarify | tasks.md (appended) |
| Plan | Speckit plan | `docs/superpowers/plans/*.plan.md` |
| Implement | BMAD agents + DDD | code under `internal/` + `web/src/` |
| Verify | QA gate | `bash scripts/qa.sh` |

## Tech Stack

<!-- Fill with actual stack -->

## Commands

```bash
bash scripts/qa.sh                           # Run quality gate
bash scripts/ddd-scaffold.sh <context-name>   # Scaffold new bounded context
taskmaster generate --spec <spec> --output <out> --format speckit  # Break spec into tasks
```

## Rules

- No code before spec (constitution §4)
- `domain/` layer zero external dependencies (constitution §1)
- QA writes tests first, Dev implements second (TDD red→green)
- Every phase artifact is git committed — can resume from any checkpoint
