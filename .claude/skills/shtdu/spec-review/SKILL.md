---
name: spec-review
description: Use when reviewing EARS requirements produced by spec-draft. Triggers when asked to review, validate, or quality-check requirements documents for pattern correctness, ME overlap, product-language, or testability.
---

# Spec Review — EARS Requirements Quality Gate

## Overview

Systematic review of EARS requirements against four quality dimensions, with emphasis on **Mutually Exclusive** overlap detection (the value spec-draft does not provide). Produces a structured report with severity ratings and fix prescriptions.

**Prerequisite:** Requirements must follow spec-draft output structure (domain folders, REQ-PREFIX-NNN naming, README.md with CE check). CE completeness was already verified during drafting — this review focuses on overlap, pattern correctness, language quality, and testability.

## Review Dimensions

Check every requirement against all four dimensions. Do not skip any.

### D1: Pattern Correctness

For each requirement, verify the EARS pattern matches the actual behavior:

| Pattern | Correct when | Wrong when |
|---------|-------------|-----------|
| **Ubiquitous** | Always-true invariant with NO conditions | Behavior depends on config, mode, or state |
| **Event-Driven** | Triggered by a specific event | Describes a continuous state or capability |
| **State-Driven** | Conditional on a sustained state | Triggered by a one-time event |
| **Optional Feature** | Behavior only exists where a feature is configured | Feature always exists in the system |
| **Unwanted Behaviour** | Describes what system shall NOT do | Describes what system SHALL do |
| **Complex** | Multi-branch conditional (if X then Y) | Single condition (use simpler pattern) |

**Red flags:**
- Ubiquitous requirement mentions "configured", "enabled", or "selected" — likely Optional Feature
- Ubiquitous requirement describes a feature that exists — not an invariant
- Optional Feature for something that always exists (e.g., CLI flags)
- Only 1 Unwanted Behaviour requirement in the entire set — error paths are under-specified
- Complex pattern with incomplete conditional matrix (if X but not if Y)

### D2: Mutually Exclusive (ME) — Primary Focus

This is the primary value of spec-review. Spec-draft handles CE; this review detects overlap systematically.

**Method: Behavioral Claim Matrix**

Extract and compare behavioral claims across all requirements. Two requirements overlap when their behavioral claims intersect.

**Step 1 — Extract claims from each requirement:**

For every requirement, extract the behavioral claim as a 4-tuple:

| Field | Question | Example |
|-------|----------|---------|
| **Actor** | Who/what triggers it? | user, system, agent, external service |
| **Trigger** | What event or condition starts it? | CLI flag, tool invocation, session start |
| **Scope** | What does it govern? | file operations, tool execution, message delivery |
| **Outcome** | What observable result? | file written, command output, error message |

Write each requirement's tuple in a comparison table. This makes overlap mechanically detectable.

**Step 2 — Pairwise intersection test:**

For each pair of requirements (within domain AND across domains), check:

```
Do they share the same Actor AND Trigger AND Scope?
  → YES: They overlap. Which one owns the behavior?
  → Partial match (same Scope, different Actor/Trigger):
    → Do they describe different aspects of the same behavior?
    → If YES: boundary must be explicit
    → If NO: no overlap, different concerns
  → No match: no overlap
```

**Step 3 — Subset test:**

For each requirement, check: "Is this requirement's scope a proper subset of another requirement's scope?" Signals:
- All acceptance criteria of REQ-A are also covered by REQ-B
- Removing REQ-A would not leave any behavior unspecified
- REQ-A exists only because it describes a specific scenario of REQ-B

**Step 4 — Cross-domain boundary probe:**

For each pair of adjacent domains in the capability tree, apply the intersection test to their requirements. Adjacent domains sharing entry points are highest-risk for overlap. For each boundary, produce:
- Which domain owns which scope
- Whether any requirement straddles the boundary
- The explicit dividing line

**Step 5 — Negative verification:**

After Steps 1-4, produce a statement for each domain: "Domain X exclusively covers [scope]. No requirement in any other domain claims [scope]."

If you cannot write this statement for a domain, there is an overlap to resolve.

**Required output: ME Overlap Matrix**

Build a domain x domain matrix. This is mandatory output in the report — it makes ME verification auditable.

```
ME Overlap Matrix:
                    | UI  | Tool | Agent | Mem | Sess | Perm | Config | Ext | Auto | Auth |
--------------------|-----|------|-------|-----|------|------|--------|-----|------|------|
User Interaction    | —   |      |       |     |      |      |        |     |      |      |
Tool Execution      |     | —    |       |     |      |      |        |     |      |      |
Agent Coordination  |     |      | —     |     |      |      |        |     | ⚠️   |      |
Memory & Context    |     |      |       | —   |      |      |        |     |      |      |
Session Mgmt        |     |      |       |     | —    |      |        |     |      |      |
Permissions         |     |      |       |     |      | —    |        |     |      |      |
Configuration       |     |      |       |     |      |      | —      |     |      |      |
Extensibility       |     |      |       |     |      |      |        | —   |      |      |
Automation          |     |      |  ⚠️   |     |      |      |        |     | —    |      |
Authentication      |     |      |       |     |      |      |        |     |      | —    |
```

**Matrix rules:**
- Empty cell = no overlap detected
- `⚠️` = overlap found (list affected requirements below the matrix)
- Diagonal (`—`) = within-domain overlaps listed separately
- Only upper triangle needed (matrix is symmetric)
- Adjacent domains in the capability tree MUST be explicitly checked (non-adjacent may be N/A)

**Below the matrix, for each ⚠️ cell:**

```markdown
⚠️ Agent Coordination ↔ Automation:
- REQ-AC-005 vs REQ-AT-002: Same (system, task lifecycle, task management, ID + state + output + stop)
- REQ-AC-005 vs REQ-AT-001: Subset — AT-001 covers task types, AC-005 covers lifecycle
```

**Overlap severity guidance:**
- Identical acceptance criteria in two requirements → MAJOR
- Same (Actor, Trigger, Scope) tuple → MAJOR
- Subset relationship → MAJOR
- Partial scope overlap in different domains → MAJOR
- Partial scope overlap in same domain → MINOR
- Fuzzy boundary without concrete claim overlap → MINOR

**Fix prescriptions must specify:**
1. Which requirement owns the overlapping behavior
2. Whether to merge, split, or add boundary language
3. The exact scope boundary after the fix

### D3: Product Language

Every requirement must describe **what the system does for users**, not how it works internally.

**Banned terms in requirement text and acceptance criteria:**
- Package names (`engine/`, `tools/`, `internal/`)
- Class or struct names (`PermissionChecker`, `ToolRegistry`)
- Implementation mechanisms (subprocess, goroutine, HTTP, file system)
- Framework concepts (middleware, dependency injection)

**Correct substitutions:**
- "The system shall..." not "The engine shall..."
- "The system shall execute tools concurrently" not "The system shall use goroutines"

**One exception:** Source Evidence section MAY reference implementation details for traceability.

### D4: Acceptance Criteria Testability

Every acceptance criterion must be verifiable with a concrete test:

**Untestable words:** "reasonable", "may", "should", "appropriate", "sufficient", "quickly", "immediately" (without timing spec), "efficiently"

**Testability rules:**
- Each criterion describes one observable outcome
- Binary pass/fail is possible (did it happen or not?)
- No restatement of the requirement itself as a criterion
- At least 3 acceptance criteria per requirement (2 is a red flag)
- Criterion does not constrain implementation unnecessarily

**Missing criteria signals:**
- Requirement describes behavior with "including X, Y, and Z" but criteria only cover X
- Requirement mentions error handling but no criterion describes the error behavior
- Requirement mentions multiple actors/stakeholders but criteria only cover one

## Review Process

```
Start → Read README (scope, tree, CE check)
  → Phase 1: Per-requirement scan
      → For each requirement in each domain:
          → Check D1 (pattern correctness)
          → Check D3 (product language)
          → Check D4 (testability)
  → Phase 2: ME overlap detection (primary)
      → Step 1: Extract behavioral claim 4-tuples (Actor, Trigger, Scope, Outcome) for ALL requirements
      → Step 2: Pairwise intersection test across ALL requirement pairs
      → Step 3: Subset test — is any requirement's scope a subset of another?
      → Step 4: Cross-domain boundary probe for adjacent domains
      → Step 5: Negative verification — write exclusive ownership statement per domain
  → Phase 3: Report
      → Merge all findings
      → Assign severity per overlap guidance
      → Produce structured report
```

Phase 1 must cover every domain. Phase 2 is the primary value — do all five steps completely. The claim extraction table (Step 1) is required output in the report.

## Report Format

For each issue:

```markdown
### [SEVERITY] REQ-ID: Short title

**Dimension:** D1/D2/D3/D4
**Problem:** [1-2 sentences]
**Fix:** [Specific rewrite or action]
```

**Severity levels:**
- **CRITICAL** — Requirement describes wrong behavior or is fundamentally untestable
- **MAJOR** — Pattern is wrong, ME overlap found, or missing requirement for core behavior
- **MINOR** — Thin acceptance criteria, vague wording, missing Source Evidence
- **NOTE** — Style suggestion or improvement opportunity

After all issues, produce a summary table:

```markdown
## Summary

| Dimension | Critical | Major | Minor | Note |
|-----------|----------|-------|-------|------|
| D1 Pattern | | | | |
| D2 ME Overlap | | | | |
| D3 Language | | | | |
| D4 Testability | | | | |
| **Total** | | | | |
```

## Common Mistakes

| Mistake | Fix |
|---------|-----|
| Skipping Phase 2 (ME) | This is the primary value of spec-review — always do all five steps |
| Comparing titles instead of behavioral claims | Extract 4-tuples and compare those, not titles |
| Reviewing only first few domains | Phase 1 must cover every domain |
| Saying "looks good" without evidence | Every pass needs justification against the dimension rules |
| Treating all issues equally | Use severity levels to prioritize |
| Fixing issues during review | Review reports problems; fixes are a separate step |
| Flagging pattern without checking the rule table | Always cross-reference the pattern definition |
| Re-running CE checks | Spec-draft already verified CE — focus on ME overlap detection |
| Omitting the claim extraction table | Step 1 output is required in the report |
