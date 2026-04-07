# EARS Requirements Review Report

**Scope:** OpenHarness full product — 67 requirements across 10 domains
**Date:** 2026-04-07
**Reviewer:** spec-review skill (shtdu:spec-review)
**Revision:** 5 — fresh review from scratch. All 17 issues resolved.

---

## Behavioral Claim Extraction Table (D2 Step 1 — Required Output)

| REQ-ID | Actor | Trigger | Scope | Outcome |
|--------|-------|---------|-------|---------|
| UI-001 | user | CLI launch | primary interaction | prompt accepted / interactive mode |
| UI-002 | user | CLI flags at invocation | session configuration | flags override defaults for session |
| UI-003 | system | TUI running | terminal display | streaming response + tool exec shown |
| UI-004 | user | slash command entry | command dispatch | built-in/plugin command executed |
| UI-005 | external service | channel message received | message routing | response returned to originating channel |
| UI-006 | agent | user input needed | interactive prompt | user decision captured via options/free-text |
| UI-007 | user | theme selection/config | visual theme | themed terminal display |
| UI-008 | user | vim mode toggle | input keybindings | vim-style editing active |
| TL-001 | system | tool catalog query | tool availability | tools with schemas returned; lifecycle hooks fire |
| TL-002 | agent | file operation invocation | file I/O | file content read/written/edited |
| TL-003 | agent | command tool invocation | shell execution | command output captured |
| TL-004 | agent | file search invocation | file pattern matching | matching paths by glob |
| TL-005 | agent | content search invocation | content matching | matching lines with context |
| TL-006 | agent | web fetch invocation | URL content retrieval | extracted text from URL |
| TL-007 | agent | web search invocation | search engine query | ranked results with title/url/summary |
| TL-008 | agent | LSP operation invocation | code intelligence | symbol info with file/line |
| TL-009 | agent | notebook edit invocation | notebook cell editing | cell replace/insert/delete |
| TL-010 | system | external tool server configured | MCP tool bridge | external tools available/executed as built-in |
| TL-011 | agent | tool search query | tool discovery | matching tools with descriptions/schemas |
| AC-001 | agent | subagent delegation | subagent creation | new isolated agent instance with task |
| AC-002 | user/agent | team creation request | team management | named group with shared channel |
| AC-003 | agent | message send request | inter-agent messaging | message delivered to target mailbox |
| AC-004 | system | subagent running | execution isolation | isolated execution with separate permissions |
| AC-005 | system | subagent task created | task lifecycle | execution coordinated + output relayed to parent |
| MC-001 | system | memory feature enabled | memory persistence | entries survive sessions |
| MC-002 | system | session start | memory discovery | memory files loaded into context |
| MC-003 | system | session start | instruction loading | CLAUDE.md etc. loaded into system prompt |
| MC-004 | agent | memory search query | memory search | ranked matching entries |
| MC-005 | agent | memory add/remove request | memory management | memory file + index updated |
| MC-006 | system | memory entry added | limit enforcement | write rejected if over limit |
| SM-001 | system | conversation activity | session persistence | state saved and retrievable |
| SM-002 | user | continue flag (-c) | session restoration | most recent session loaded |
| SM-003 | user | resume flag with ID (-r) | session restoration | specified session loaded |
| SM-004 | user | export command | transcript generation | export file in markdown/JSON |
| SM-005 | user | share command | shareable artifact | self-contained shareable document |
| SM-006 | user | tag command | session snapshot | named checkpoint created |
| SM-007 | user | rewind command | conversation truncation | recent turns removed |
| SM-008 | system | context threshold reached | context compaction | older messages summarized |
| PS-001 | system | permission mode active | permission enforcement | tools routed through permission checker |
| PS-002 | system | default mode + sensitive tool | user confirmation | prompt for approval |
| PS-003 | system | plan mode active | tool restriction | write tools disabled |
| PS-004 | system | full auto mode active | auto-execution | tools execute within boundaries |
| PS-005 | system | tool on deny/allow list | list-based filtering | denied=block, allowed=auto-approve |
| PS-006 | system | path rules configured | file access restriction | out-of-bounds paths blocked |
| PS-007 | system | destructive operation detected | warning display | user warned before execution |
| PS-008 | system | permission check error | fail-safe | tool execution blocked + error reported |
| CF-001 | system | startup | settings loading | settings read from JSON file |
| CF-003 | admin | provider profiles defined | profile management | profiles available for use |
| CF-004 | user | profile switch command | profile activation | new provider config active without restart |
| CF-005 | system | env vars set | config override | env vars used as settings values |
| CF-006 | user | slash command config change | runtime config update | settings changed without restart |
| CF-007 | system | startup | config layer merging | merged config from all layers |
| EX-001 | system | plugin directories configured | plugin discovery | plugins loaded from manifests |
| EX-002 | system | plugin becomes active | contribution registration | commands/skills/hooks/tools registered |
| EX-003 | user | plugin install/uninstall CLI | plugin lifecycle | registry updated + reload |
| EX-004 | agent | skill invocation | skill loading | skill content injected into context |
| EX-005 | system | hook event fires | hook execution | hooks run in order with alter/prevent |
| EX-006 | system | hook of supported type configured | hook type dispatch | type-specific execution |
| EX-007 | user | MCP server add/remove | MCP config management | config persisted + runtime bridge notified |
| EX-008 | user | plugin enable/disable | plugin state change | contributions loaded/removed without restart |
| AT-001 | agent | background task creation | independent execution | task runs independently |
| AT-002 | system | background task exists | task lifecycle | task tracked through state transitions |
| AT-003 | system | cron jobs configured | scheduled execution | commands/prompts run on schedule |
| AT-004 | user/agent | task output request | output retrieval | accumulated output returned |
| AT-005 | agent | task progress update | progress tracking | progress metadata persisted |
| AU-001 | system | provider configured | API key auth | authenticated request completed |
| AU-002 | user | OAuth login command | device flow auth | token obtained and stored |
| AU-003 | system | provider profile active | multi-provider auth | provider-specific credentials used |
| AU-004 | user | auth status command | status reporting | provider + credential validity shown |

---

## ME Overlap Matrix (D2 Required Output)

```
                    | UI  | Tool | Agent | Mem | Sess | Perm | Config | Ext | Auto | Auth |
--------------------|-----|------|-------|-----|------|------|--------|-----|------|------|
User Interaction    | —   |      |       |     |      |      |        |     |      |      |
Tool Execution      |     | —    |       |     |      |      |        |     |      |      |
Agent Coordination  |     |      | —     |     |      |      |        |     | ⚠️  |      |
Memory & Context    |     |      |       | —   |      |      |        |     |      |      |
Session Mgmt        |     |      |       |     | —    |      |        |     |      |      |
Permissions         |     |      |       |     |      | —    |        |     |      |      |
Configuration       |     |      |       |     |      |      | —      |     |      |      |
Extensibility       |     |      |       |     |      |      |        | —   |      |      |
Automation          |     |      |  ⚠️   |     |      |      |        |     | —    |      |
Authentication      |     |      |       |     |      |      |        |     |      | —    |
```

### ⚠️ Agent Coordination ↔ Automation

- **REQ-AC-005 vs REQ-AT-002**: AC-005 (subagent task lifecycle) is a thin wrapper over AT-002 (background task lifecycle). AC-005 has 3 ACs: AC1 is unique (output relay to parent), AC2 and AC3 explicitly delegate to AT-002 ("per REQ-AT-002"). The delegation is clean — no duplicated behavioral claims — but AC-005 has only 1 substantive acceptance criterion. See D4 issue below.
- **REQ-AC-005 vs REQ-AT-001**: AC-005 triggers on subagent task creation. AT-001 triggers on generic background task creation. Subagent tasks are a specific type of background task. Different scopes (agent-specific vs generic), no duplicated claims. Clean boundary.

**Boundary:** AT domain owns generic task lifecycle. AC domain owns agent-specific concerns (output relay to parent, agent isolation during execution). AC-005 delegates lifecycle management to AT-002 and adds agent-specific output relay.

---

## D1: Pattern Correctness Issues

### [MAJOR] REQ-AT-001: Optional Feature pattern is wrong — should be Event-Driven

**Problem:** Marked as Optional Feature ("Where a background task is created, the system shall execute it independently of the main conversation"). "A background task is created" is an event trigger, not a configuration state. The system always has the capability to create background tasks — the feature always exists. Compare with REQ-AT-003 (Cron Scheduling), which correctly uses Optional Feature because cron jobs being *configured* is a persistent state.

**Fix:** Change pattern to **Event-Driven**. Rewrite: "When a background task is created, the system shall execute it independently of the main conversation."

### [NOTE] Only 1 Unwanted Behaviour requirement in entire set (PS-008)

**Problem:** REQ-PS-008 is the only requirement using the Unwanted Behaviour pattern (1 out of 67). Error paths for critical failure scenarios are under-specified. Areas that lack "shall not" specifications include: API credential exposure, tool execution bypassing permissions, session data corruption, and memory data loss. The existing requirements cover these implicitly through positive assertions (PS-008 fail-safe, PS-007 destructive warnings, AC-004 isolation), but explicit "shall not" statements provide stronger safety guarantees.

**Fix:** Consider adding Unwanted Behaviour requirements for critical failure modes. No urgency — the positive assertions provide reasonable coverage.

---

## D2: ME Overlap Issues

### [MAJOR] AC-005: Thin wrapper — only 1 substantive acceptance criterion

**Problem:** REQ-AC-005 has 3 acceptance criteria. AC1 ("Subagent output is relayed to the parent agent upon completion") is the only unique behavioral claim. AC2 ("Subagent tasks can be stopped...per REQ-AT-002") and AC3 ("lifecycle tracked...per REQ-AT-002") are delegation references to AT-002, not independent testable claims. While the delegation is clean (no duplicated claims), a requirement with only 1 substantive AC may not justify its existence as a standalone requirement.

**Fix:** Either (a) merge AC-005's unique behavior (output relay) into AC-001 as an additional AC, or (b) add 2 more substantive ACs that are specific to subagent task lifecycle: e.g., "A subagent task's execution state (running, completed, failed) is observable via the background task query tools" and "When a subagent task completes, its result is distinguishable from other background task results as originating from a subagent."

### [MINOR] PS-004 AC4 restates PS-007 (within Permissions domain)

**Problem:** PS-004 AC4 says "Operations classified as destructive display a warning even in full auto mode." PS-007 says "When the agent attempts a destructive operation, the system shall warn the user." PS-004 AC4 restates PS-007's claim for the full auto mode context. Since PS-007 is Event-Driven and fires regardless of mode, PS-004 AC4 is technically redundant — though it serves a useful clarifying purpose (preventing the assumption that full auto bypasses all safety checks).

**Fix:** Rewrite PS-004 AC4 as an explicit cross-reference: "Destructive operation warnings (per REQ-PS-007) remain active even in full auto mode." This preserves the clarification without restating the behavior.

---

## D3: Product Language Issues

### [MINOR] REQ-AC-003 AC3: "asynchronous and non-blocking" are implementation terms

**Problem:** AC3 says "Message delivery is asynchronous and non-blocking for the sender." These terms describe implementation mechanism rather than product behavior.

**Fix:** Rewrite: "The sending agent continues execution without waiting for the receiving agent to process the message."

### [MINOR] REQ-EX-006 AC1-AC2: Implementation mechanisms in acceptance criteria

**Problem:** AC1 says "Command-type hooks run an external process and provide its output" — "external process" is implementation. AC2 says "Prompt-type hooks generate a response from the AI model" — "AI model" is borderline.

**Fix:** Rewrite AC1: "Command-type hooks execute a configured action and return its output." Rewrite AC2: "Prompt-type hooks produce an AI-generated response."

### [MINOR] REQ-UI-006 AC3: References internal subsystem

**Problem:** AC3 says "The permission system (Permissions domain) triggers prompts for tool approval." "Permission system" is an internal subsystem reference.

**Fix:** Rewrite: "Tool execution approval requests are presented through the same prompt mechanism." (Split from the rest of AC3 which covers other prompt use cases.)

---

## D4: Acceptance Criteria Testability Issues

### [CRITICAL] REQ-PS-002 AC1: "Sensitive tools (file write, bash, etc.)" — "etc." is untestable

**Problem:** AC1 says "Sensitive tools (file write, bash, etc.) require user approval." The "etc." makes the set of sensitive tools undefined and therefore untestable. The reader cannot determine which tools are in scope.

**Fix:** Replace with a definitive classification: "Tools capable of modifying the file system (file write, file edit), executing commands (bash), or making network requests require explicit user approval." Or: "Tools classified as write-capable in the tool registry require user approval."

### [MAJOR] REQ-CF-001 AC3: "Every parameter documented as configurable" — open-ended reference

**Problem:** AC3 says "Every parameter documented as configurable has a corresponding field in the settings file." "Documented as configurable" is an evolving, open-ended reference with no enumerated list. This is not testable without a definitive source of truth.

**Fix:** Replace with: "The settings file schema covers all configuration parameters defined in the Configuration domain requirements (REQ-CF-001 through REQ-CF-007)."

### [MAJOR] REQ-MC-004 AC2: "ranked results by relevance" — untestable ranking

**Problem:** AC2 says "Returns ranked results by relevance." "Relevance" is subjective — different implementations could return wildly different orderings and all pass. There is no defined ranking algorithm or minimum quality threshold.

**Fix:** Rewrite: "Results are ordered by a relevance score with a configurable minimum threshold." Or: "Results are returned in ranked order based on text similarity to the query."

### [MAJOR] REQ-EX-005 AC4: "handled according to the hook's error policy" — undefined policy

**Problem:** AC4 says "A failing hook does not terminate the session; the tool execution proceeds or is handled according to the hook's error policy." "The hook's error policy" is undefined — no requirement specifies what this policy is or where it's defined.

**Fix:** Rewrite: "A failing hook does not terminate the session. The tool execution proceeds unless the hook explicitly requests cancellation."

### [MAJOR] REQ-SM-001: No AC for session restoration

**Problem:** SM-001 has 3 ACs covering persistence (auto-save, keyed by directory/ID, survives termination). None verify that the persisted session can be successfully loaded and restored. The requirement says sessions "can be resumed after termination" but no AC tests the resume path.

**Fix:** Add: "A persisted session can be restored to a state where the agent has access to the full conversation history."

### [MAJOR] REQ-AT-002 AC3/AC5: Cross-references without testable behavior

**Problem:** AC3 "Task output retrieval follows REQ-AT-004" and AC5 "Task progress tracking follows REQ-AT-005" are cross-references, not testable criteria for this requirement. They add no verifiable claims.

**Fix:** Replace AC3: "Task output is retrievable while the task exists." Replace AC5: "Task progress percentage and status are queryable."

### [MINOR] REQ-UI-002 AC7: Cross-reference without testable behavior

**Problem:** AC7 "Flag behavior follows the precedence rules defined in the Configuration domain" is a cross-reference rather than a testable claim.

**Fix:** Replace: "CLI flags override values from the settings file for the duration of the session."

### [MINOR] REQ-AC-002: Only 3 ACs, no error cases

**Problem:** Only 3 ACs. No criterion covers creating a team with a duplicate name, or deleting a team with active agents.

**Fix:** Add: "Creating a team with a name that already exists produces an error."

### [MINOR] REQ-CF-007 AC1/AC2: Statements of fact, not behavioral criteria

**Problem:** AC1 "Global settings from ~/.openharness/settings.json" and AC2 "Project settings from .openharness/ in the working directory" describe data sources rather than what the system does with them.

**Fix:** Rewrite AC1: "The system reads configuration from the user's global settings file at startup." Rewrite AC2: "The system reads project-level settings from the project's configuration directory when present."

### [MINOR] REQ-SM-005: Only 3 ACs, vague on format

**Problem:** AC1 "Produces a self-contained shareable document" does not specify the format or delivery mechanism. Only 3 ACs.

**Fix:** Clarify: "Produces a self-contained document including the full conversation with formatted tool results."

---

## Summary

| Dimension | Critical | Major | Minor | Note |
|-----------|----------|-------|-------|------|
| D1 Pattern | 0 | 1 | 0 | 1 |
| D2 ME Overlap | 0 | 1 | 1 | 0 |
| D3 Language | 0 | 0 | 3 | 0 |
| D4 Testability | 1 | 5 | 4 | 0 |
| **Total** | **1** | **7** | **8** | **1** |

**Assessment:** Previous revisions fixed the bulk of issues (53 in Rev 1, 14 in Rev 2). This fresh review finds 17 remaining issues — primarily testability gaps (D4) and one pattern mismatch (D1). No cross-domain ME overlaps remain unfixed; the AC-005/AT-002 boundary is well-managed through delegation. The requirements set is in good shape overall.

### Priority Fixes (all resolved)

1. **REQ-AT-001** (MAJOR, D1) — Changed Optional Feature → Event-Driven
2. **REQ-PS-002 AC1** (CRITICAL, D4) — Replaced "etc." with definitive tool classification
3. **REQ-AC-005** (MAJOR, D2+D4) — Added 2 substantive ACs (subagent state observability, result distinguishability)
4. **REQ-CF-001 AC3** (MAJOR, D4) — Replaced with bounded scope reference to CF-001..CF-007
5. **REQ-MC-004 AC2** (MAJOR, D4) — Defined ranking as "text similarity to the query"
6. **REQ-EX-005 AC4** (MAJOR, D4) — Removed undefined "error policy"; replaced with explicit cancellation model
7. **REQ-SM-001** (MAJOR, D4) — Added AC for session restoration
8. **PS-004 AC4** (MINOR, D2) — Rewrote as explicit cross-reference to PS-007
9. **AC-003 AC3** (MINOR, D3) — Replaced "asynchronous and non-blocking" with product language
10. **EX-006 AC1-2** (MINOR, D3) — Replaced "external process" and "AI model" with product terms
11. **UI-006 AC3** (MINOR, D3) — Split into two ACs; removed "permission system" reference
12. **UI-002 AC7** (MINOR, D4) — Replaced cross-reference with testable claim
13. **AC-002** (MINOR, D4) — Added AC for duplicate team name error
14. **CF-007 AC1-2** (MINOR, D4) — Rewrote as behavioral criteria
15. **SM-005 AC1** (MINOR, D4) — Clarified format in AC
16. **AT-002 AC3/AC5** (MINOR, D4) — Replaced cross-references with testable claims

### Negative Verification (D2 Step 5)

- **User Interaction** exclusively covers how users communicate with the system. No other domain claims user interaction modality. ✓
- **Tool Execution** exclusively covers tool registration, discovery, and execution. No other domain claims tool catalog behavior. ✓
- **Agent Coordination** covers multi-agent collaboration. AC-005 delegates lifecycle to AT domain. Boundary documented. ✓
- **Memory and Context** exclusively covers cross-session information persistence. No other domain claims memory management. ✓
- **Session Management** exclusively covers conversation lifecycle. No other domain claims session state. ✓
- **Permissions and Safety** exclusively covers access control. PS-004 AC4 cross-references PS-007 within domain. ✓
- **Configuration** exclusively covers system customization. No other domain claims config mechanism. ✓
- **Extensibility** exclusively covers system extension (plugins, skills, hooks, MCP). No other domain claims extension mechanism. ✓
- **Task Automation** exclusively covers task automation and scheduling. AC domain delegates lifecycle to AT. ✓
- **Authentication** exclusively covers credential management. References CF domain for configuration sources. ✓
