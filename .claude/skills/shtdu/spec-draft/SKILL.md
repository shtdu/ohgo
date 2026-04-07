---
name: spec-draft
description: Use when drafting product requirements from an existing codebase. Triggers when asked to write EARS syntax specifications, product requirements, or system requirements documents that follow MECE principles.
---

# Spec Draft — EARS Requirements from Codebase

## Overview

Draft **product-level** requirements using EARS syntax by analyzing an existing codebase. Requirements describe **what the system does** (product behavior), organized by **product domain**, not by technical modules.

## Scope Selection

**Before drafting, confirm the scope with the user:**
1. Which source directory or module to analyze?
2. Is this the full product or a subsystem?

The user selects the source. Do NOT assume the entire codebase. If the user does not specify a scope, ask before proceeding.

## Core Principle: Product Domain, Not Implementation

```
WRONG: 01-engine/, 02-streaming/, 03-tools/
       (package names = implementation detail)

RIGHT: user-interaction/, tool-management/, agent-coordination/
       (product capabilities = what users experience)
```

Requirements must be intelligible to product managers, not just engineers. If a requirement mentions a package name, class name, or framework concept, rewrite it in product language.

## EARS Patterns (All 6)

| Pattern | Keyword | Syntax | Use for |
|---------|---------|--------|---------|
| **Ubiquitous** | shall | `The [system] shall [action].` | Always-true behaviors |
| **Event-Driven** | When | `When [event], the [system] shall [action].` | Triggered responses |
| **State-Driven** | While | `While [state], the [system] shall [action].` | Conditional on state |
| **Optional Feature** | Where | `Where [feature] is configured, the [system] shall [action].` | Configurable/optional behavior |
| **Unwanted Behaviour** | shall not | `If [condition], the [system] shall not [action].` | Error/negative cases |
| **Complex** | If/then | `If [condition], then the [system] shall [action].` | Multi-condition logic |

**Pattern selection rule:** If a requirement fits two patterns, pick the one that describes the **trigger** not the **implementation**. Example: permission denial is Event-Driven ("When a tool execution is denied..."), not Unwanted Behaviour, because the trigger matters more than the negation.

## MECE Decomposition Method

Top-down, 3-step:

### Step 1: Define Product Scope

List all **user-facing capabilities** the system provides. Ask: "What can a user or operator DO with this system?" Not: "What modules exist?"

Sources for discovering capabilities:
- Entry points: CLI flags, API endpoints, UI screens
- User-facing docs: README, usage examples, tutorials
- Config options: What can users configure?
- External integrations: What does the system connect to?

### Step 2: Build Capability Tree

Organize capabilities into a hierarchy. Each level must be MECE:

```
System
├── User Interaction          # How users communicate with the system
├── Tool Management           # How tools are registered, found, executed
├── Agent Coordination        # How agents collaborate
├── Memory & Context          # How information persists across sessions
├── Security & Permissions    # How access is controlled
├── Configuration             # How the system is customized
└── Integration               # How the system connects to external services
```

**MECE check:** Does every requirement fit in exactly one category? If a requirement fits in two, split it or merge categories.

### Step 3: Draft Requirements Per Capability

For each leaf in the capability tree, write requirements using the appropriate EARS pattern. Each requirement gets its own file.

### Step 4: Collectively Exhaustive Verification (REQUIRED)

Do NOT skip this step. Run the entry point enumeration check:

1. List every user-facing entry point in the scoped source: CLI flags, API endpoints, config keys, user actions, event triggers, external integrations.
2. For each entry point, verify at least one requirement covers it.
3. Any entry point with zero requirements = gap. Draft missing requirements.

Output as a checklist in README.md:
```
## CE Check — Entry Point Coverage
- [x] CLI flag `--permission-mode` → REQ-MODE-001, REQ-MODE-002
- [x] Config key `allowedTools` → REQ-AC-001
- [ ] API endpoint `/tool/execute` → NO REQUIREMENT ← GAP
```

If gaps are found, fix requirements and re-verify before finalizing.

## Output Structure

```
requirements/
├── README.md                    # Scope, capability tree, pattern summary
├── user-interaction/
│   ├── REQ-UI-001.md
│   ├── REQ-UI-002.md
│   └── ...
├── tool-management/
│   ├── REQ-TM-001.md
│   └── ...
└── [capability-domain]/
    └── REQ-[PREFIX]-NNN.md
```

### README.md contents:
1. **Product scope** — what system and version this covers
2. **Capability tree** — the MECE decomposition
3. **Pattern summary** — count of requirements per EARS pattern
4. **Traceability** — mapping from requirements to source evidence

### Per-requirement file:

```markdown
# REQ-[PREFIX]-NNN: [Short title]

**Pattern:** [EARS pattern name]
**Capability:** [Product domain from capability tree]

## Requirement

[EARS-formatted requirement statement]

## Acceptance Criteria

- [ ] [Testable criterion 1]
- [ ] [Testable criterion 2]

## Source Evidence

- [Code path, config, doc, or user flow that justifies this requirement]
```

## Anti-Patterns

| Anti-pattern | Fix |
|-------------|-----|
| Requirement mentions a package or class name | Rewrite in product language |
| Organized by `internal/` package structure | Reorganize by product capability |
| "The engine shall..." | "The system shall..." — name the product, not the module |
| Requirements overlap across categories | Tighten MECE boundaries, split or merge |
| Missing capabilities (not exhaustive) | Re-check user-facing entry points |
| Two requirements describe the same behavior | Merge or clarify scope difference |
| Implementation details in acceptance criteria | Describe observable behavior instead |

## Common Mistakes

1. **Reading code bottom-up** — Starts from packages, misses user-facing capabilities. Fix: start from entry points and user actions.
2. **Confusing "system" with "module"** — "The permission checker shall..." is wrong. "The system shall..." is right.
3. **Skipping the MECE check** — Without explicit verification, categories will overlap. Fix: after drafting, trace each requirement to exactly one category.
4. **One pattern for everything** — Defaulting to Ubiquitous ("shall") for all requirements. Fix: ask "Is this always true, or triggered by something?"
5. **Requirements that prescribe implementation** — "The system shall use goroutines" is wrong. "The system shall execute tools concurrently" is right.
