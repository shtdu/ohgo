# REQ-AT-003: Cron Scheduling

**Pattern:** Optional Feature
**Capability:** Task Automation

## Requirement

Where cron jobs are configured, the system shall execute specified commands or agent prompts on the defined cron schedule.

## Acceptance Criteria

- [ ] Cron expressions define the schedule
- [ ] Jobs can be enabled and disabled individually
- [ ] Jobs can be invoked outside their schedule via an immediate trigger
- [ ] Each scheduled execution records its outcome (success or failure)
- [ ] When a cron expression is invalid, the system rejects it with a parse error message identifying the malformed portion
- [ ] When cron job execution fails, the system logs the error and does not remove or disable the job

## Source Evidence

- `OpenHarness/src/openharness/tools/cron_create_tool.py`
- `OpenHarness/src/openharness/tools/cron_delete_tool.py`
- `OpenHarness/src/openharness/tools/cron_list_tool.py`
- `OpenHarness/src/openharness/tools/cron_toggle_tool.py`
- `OpenHarness/src/openharness/tools/remote_trigger_tool.py`
- `OpenHarness/src/openharness/cli.py` — `cron` subcommand
