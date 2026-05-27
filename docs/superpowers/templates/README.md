# Six-Layer Workflow — Quick Reference

## One-Liner

> Requirement → brainstorm → spec → TaskMaster → tasks → clarify → plan → TDD(red→green) → QA gate → done

## Files You'll Create Per Feature

```
docs/
├── superpowers/specs/YYYY-MM-DD-<topic>.spec.md   # Phase 1 output
├── superpowers/plans/YYYY-MM-DD-<topic>.plan.md    # Phase 4 output
└── speckit/<feature>/
    └── tasks.md                                     # Phase 2-3 output

internal/
├── domain/<bounded-context>/        # Phase 5: entity, value object, aggregate, repo interface
├── service/<bounded-context>/       # Phase 5: command handlers, queries
├── infrastructure/                  # Phase 5: repo implementations, adapters
└── interfaces/                      # Phase 5: HTTP handlers, worker handlers
```

## Phase Checklist

- [ ] **Phase 1**: Superpower brainstorm → user approves spec → commit `.spec.md`
- [ ] **Phase 2**: TaskMaster breakdown → commit `tasks.md`
- [ ] **Phase 3**: Speckit clarify → append acceptance criteria → no TBDs remain
- [ ] **Phase 4**: Speckit plan → file paths + layer assignments → commit `.plan.md`
- [ ] **Phase 5**: Quinn writes tests (RED) → Pirlo implements (GREEN) → commit per BC
- [ ] **Phase 6**: `bash scripts/qa.sh` → all gates pass → DONE

## Agent Roles

| Agent | Role | Does | Does NOT |
|-------|------|------|----------|
| Bob (sm) | Scrum Master | TaskMaster CLI, track tasks.md state | Write code or tests |
| Quinn (qa) | QA Engineer | Write tests BEFORE implementation | Write implementation code |
| Pirlo (dev) | Developer | Implement against QA's tests, DDD layering | Write tests or manage tasks |

## Constitution §1 Cheat Sheet

```
interfaces → service → domain ← infrastructure
```

- `domain/`: zero external imports. Interfaces only.
- `service/`: depends only on `domain/`
- `infrastructure/`: implements `domain/` interfaces
- `interfaces/`: HTTP/worker adapters, depends on `service/`
- DI in `cmd/server/main.go` only
