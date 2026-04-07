# REQ-AT-003: Cron Scheduling

**Pattern:** Optional Feature
**Capability:** Task Automation

## Requirement

Where cron jobs are configured, the system shall execute specified commands or agent prompts on the defined cron schedule.

## Acceptance Criteria

- [ ] Cron expressions define the schedule
- [ ] Jobs can be enabled and disabled individually
- [ ] Jobs can be triggered immediately on demand
- [ ] Job execution history is maintained

## Source Evidence

- `OpenHarness/src/openharness/tools/cron_create_tool.py`
- `OpenHarness/src/openharness/tools/cron_delete_tool.py`
- `OpenHarness/src/openharness/tools/cron_list_tool.py`
- `OpenHarness/src/openharness/tools/cron_toggle_tool.py`
- `OpenHarness/src/openharness/tools/remote_trigger_tool.py`
- `OpenHarness/src/openharness/cli.py` — `cron` subcommand
