# Memory

ohgo remembers things between sessions. Instead of re-explaining your project or preferences every time you start a conversation, the agent carries context forward automatically.

## Two layers

ohgo has two memory layers that work together:

| Layer | Scope | Example |
|-------|-------|---------|
| **Personal** | Follows you across all projects | "I prefer table-driven tests", "Don't mock the database" |
| **Project** | Scoped to a single project directory | "The auth rewrite is driven by compliance requirements" |

Personal memory loads first, so the agent always knows your general preferences. Project memory loads second, adding context specific to the codebase you're working in.

## How it works

Memories are plain markdown files stored on disk. When a session starts, ohgo loads both `MEMORY.md` index files — one from personal memory, one from project memory — into the agent's context. The agent can then read individual memory files in full when needed.

## Memory types

ohgo organizes memories into four categories. The `type` field in a memory file's frontmatter tells the agent how to use it.

| Type | Purpose | Example |
|------|---------|---------|
| `user` | Who you are, your expertise, your preferences | "I'm a backend engineer who prefers verbose logging" |
| `feedback` | Corrections and confirmations about agent behavior | "Always use table-driven tests, not subtests" |
| `project` | Ongoing work, goals, constraints, decisions | "The auth rewrite is driven by compliance requirements" |
| `reference` | Pointers to external systems and resources | "Pipeline bugs are tracked in Linear project INGEST" |

### user

Your profile — role, skill set, communication preferences. When the agent knows you're a senior Go developer, it skips the basics and jumps to what matters.

### feedback

How you want the agent to behave. This includes both corrections ("don't do X") and confirmations ("yes, that approach was right"). Feedback memories include a reason so the agent can judge edge cases instead of blindly following a rule.

### project

Context about the work itself — why a feature exists, what deadline is looming, who the stakeholders are. Project memories decay fast, so they include the motivation to help future sessions decide if the memory is still relevant.

### reference

Shortcuts to external resources — dashboards, issue trackers, documentation, APIs. Instead of describing the resource, reference memories tell the agent *where to look*.

## File format

Each memory is a markdown file with optional YAML frontmatter:

```markdown
---
name: Short descriptive title
description: One-line summary for the memory index
type: user
---

The memory content goes here.

For feedback and project types, include:

**Why:** the reason this memory matters
**How to apply:** when this memory should influence behavior
```

When frontmatter is omitted, ohgo infers the title from the filename and the description from the first non-heading line.

## Storage layout

```
~/.ohgo/
  settings.json                    # global settings
  credentials.json                 # stored API keys
  data/
    memory/
      _personal/                   # personal memory (all projects)
        MEMORY.md                  # personal memory index
        user_role.md               # personal memory files
        feedback_testing.md
      myproject-a1b2c3/            # project memory (per-project hash)
        MEMORY.md                  # project memory index
        project_auth_rewrite.md
        reference_linear.md
```

Each project directory gets its own memory store, keyed by a SHA1 hash of the working directory path. The personal memory store lives at a fixed location and is shared across all projects.

## Commands

### List memories

```
/memory
```

Shows all memory entries for the current project and your personal memories, clearly labeled.

### Tell the agent to remember

```
Remember that I prefer table-driven tests in Go
```

The agent will save a memory file with the appropriate type and layer, then update the index.

### Tell the agent to forget

```
Forget the memory about the auth rewrite
```

The agent removes the file and updates the index.

### Search memories

```
What do you remember about testing?
```

The agent searches both personal and project memory, merging and ranking results by relevance. Metadata matches (title and description) are weighted more heavily than body text. Searches support both English words and CJK characters.

## Configuration

Memory behavior is controlled in your settings file:

| Setting | Default | What it does |
|---------|---------|--------------|
| `memory.enabled` | `true` | Turn the memory system on or off (applies to both layers) |
| `memory.max_files` | `5` | Maximum number of memory files per layer |
| `memory.max_entrypoint_lines` | `200` | Truncate each MEMORY.md index after this many lines |

When `memory.enabled` is `false`, no memories load into context and the agent won't save new ones. The files remain on disk — turning it back on restores everything.

## What gets stored (and what doesn't)

**Stored in memory:**
- User preferences and working style
- Project decisions and their rationale
- Corrections and validated approaches
- Links to external resources

**Not stored in memory:**
- Code, file contents, or project structure — the agent reads these directly from disk
- Git history — use `git log` instead
- Transient conversation context — that's part of the current session only
- Anything already documented in CLAUDE.md or README files

## Privacy

- Memories are stored locally on your machine in plain text
- Personal and project memories are isolated from each other
- You can inspect, edit, or delete any memory file directly with a text editor
- No memory data is sent anywhere beyond the API calls you already make
