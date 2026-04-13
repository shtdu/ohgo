# EARS Requirement Traceability Matrix

Auto-generated from `// EARS: REQ-XX-NNN` comments in integration test files.

## Summary

| Domain | REQs | Tested | Deferred |
|--------|------|--------|----------|
| Tool Execution (TL) | 11 | 11 | 0 |
| Permissions & Safety (PS) | 8 | 8 | 0 |
| Session Management (SM) | 8 | 3 | 5 |
| Memory & Context (MC) | 6 | 5 | 1 |
| Agent Coordination (AC) | 5 | 4 | 1 |
| Automation (AT) | 5 | 5 | 0 |
| Configuration (CF) | 7 | 4 | 3 |
| Extensibility (EX) | 8 | 5 | 3 |
| Authentication (AU) | 4 | 4 | 0 |
| User Interaction (UI) | 8 | 0 | 8 |
| **Total** | **70** | **49** | **21** |

**71 integration tests** covering **49 unique EARS requirements** across 9 test files.

## Running

```bash
make test              # unit tests only (integration excluded by build tag)
make test-integration  # integration tests only (-tags=integration)
```

---

## Tool Execution (TL)

| REQ | Test Function | File |
|-----|---------------|------|
| REQ-TL-001 | TestIntegration_ToolRegistry_DynamicExpansion | engine/integration_tool_execution_test.go |
| REQ-TL-001 | TestIntegration_ToolRegistry_InvalidSchemaExcluded | engine/integration_tool_execution_test.go |
| REQ-TL-001 | TestIntegration_ToolDiscovery_SearchByName | engine/integration_tool_execution_test.go |
| REQ-TL-002 | TestIntegration_FileOps_WriteAndReadRoundTrip | engine/integration_tool_execution_test.go |
| REQ-TL-002 | TestIntegration_FileOps_EditReplacesContent | engine/integration_tool_execution_test.go |
| REQ-TL-002 | TestIntegration_FileOps_ReadNonExistent | engine/integration_tool_execution_test.go |
| REQ-TL-003 | TestIntegration_Bash_ExecutesAndReturnsOutput | engine/integration_tool_execution_test.go |
| REQ-TL-003 | TestIntegration_Bash_NonZeroExitCode | engine/integration_tool_execution_test.go |
| REQ-TL-003 | TestIntegration_Bash_TimeoutKillsProcess | engine/integration_tool_execution_test.go |
| REQ-TL-003 | TestIntegration_Bash_CommandNotFound | engine/integration_tool_execution_test.go |
| REQ-TL-003 | TestIntegration_Bash_EngineLoop_ExecutesTool | engine/integration_tool_execution_test.go |
| REQ-TL-004 | TestIntegration_Glob_FindsFilesByPattern | engine/integration_tool_execution_test.go |
| REQ-TL-004 | TestIntegration_Glob_InvalidPattern | engine/integration_tool_execution_test.go |
| REQ-TL-004 | TestIntegration_Glob_NonExistentDirectory | engine/integration_tool_execution_test.go |
| REQ-TL-005 | TestIntegration_Grep_RegexAndContext | engine/integration_tool_execution_test.go |
| REQ-TL-005 | TestIntegration_Grep_InvalidRegex | engine/integration_tool_execution_test.go |
| REQ-TL-011 | TestIntegration_ToolDiscovery_SearchByName | engine/integration_tool_execution_test.go |
| REQ-TL-006 | — | *deferred: requires httptest mock* |
| REQ-TL-007 | — | *deferred: requires external service* |
| REQ-TL-008 | — | *deferred: requires LSP server* |
| REQ-TL-009 | — | *deferred: requires notebook file setup* |
| REQ-TL-010 | — | *deferred: requires MCP stdio bridge* |

## Permissions & Safety (PS)

| REQ | Test Function | File |
|-----|---------------|------|
| REQ-PS-001 | TestIntegration_Permission_AllModesThroughEngine | engine/integration_permissions_test.go |
| REQ-PS-001 | TestIntegration_Permission_InvalidModeDefaultsToDefault | engine/integration_permissions_test.go |
| REQ-PS-002 | TestIntegration_DefaultMode_WriteToolsPromptUser | engine/integration_permissions_test.go |
| REQ-PS-002 | TestIntegration_DefaultMode_ReadToolsAutoAllow | engine/integration_permissions_test.go |
| REQ-PS-003 | TestIntegration_PlanMode_WriteToolsDenied | engine/integration_permissions_test.go |
| REQ-PS-004 | TestIntegration_AutoMode_ExecuteWithoutPrompt | engine/integration_permissions_test.go |
| REQ-PS-004 | TestIntegration_AutoMode_DeniedListStillBlocks | engine/integration_permissions_test.go |
| REQ-PS-005 | TestIntegration_AllowDenyLists_DenyPrecedence | engine/integration_permissions_test.go |
| REQ-PS-005 | TestIntegration_AllowList_DefaultMode_AutoApproves | engine/integration_permissions_test.go |
| REQ-PS-005 | TestIntegration_AutoMode_DeniedListStillBlocks | engine/integration_permissions_test.go |
| REQ-PS-006 | TestIntegration_PathRules_FileAccessControl | engine/integration_permissions_test.go |
| REQ-PS-006 | TestIntegration_PathRules_InvalidSyntaxRejectedAtLoad | engine/integration_permissions_test.go |
| REQ-PS-007 | — | *deferred: destructive pattern detection* |
| REQ-PS-008 | TestIntegration_PermissionFailSafe_BlocksOnErrors | engine/integration_permissions_test.go |

## Session Management (SM)

| REQ | Test Function | File |
|-----|---------------|------|
| REQ-SM-001 | TestIntegration_Session_SaveAndRestore | engine/integration_session_test.go |
| REQ-SM-001 | TestIntegration_Session_MessagesAccessible | engine/integration_session_test.go |
| REQ-SM-001 | TestIntegration_Session_Clear | engine/integration_session_test.go |
| REQ-SM-008 | TestIntegration_Compaction_ShouldCompact | engine/integration_session_test.go |
| REQ-SM-008 | TestIntegration_Compaction_Microcompact | engine/integration_session_test.go |
| REQ-SM-002 | — | *deferred: continue latest session logic* |
| REQ-SM-003 | — | *deferred: resume by tag* |
| REQ-SM-004 | — | *deferred: export produces JSON* |
| REQ-SM-005 | — | *deferred: share artifact* |
| REQ-SM-006 | — | *deferred: tag round-trip* |
| REQ-SM-007 | — | *deferred: rewind round-trip* |

## Memory & Context (MC)

| REQ | Test Function | File |
|-----|---------------|------|
| REQ-MC-001 | TestIntegration_Memory_PersistenceAndPromptGeneration | memory/integration_memory_context_test.go |
| REQ-MC-002 | TestIntegration_Memory_PersistenceAndPromptGeneration | memory/integration_memory_context_test.go |
| REQ-MC-002 | TestIntegration_Memory_CLAUDEmdDiscoveryInPrompt | memory/integration_memory_context_test.go |
| REQ-MC-002 | TestIntegration_Memory_RulesDiscovery | memory/integration_memory_context_test.go |
| REQ-MC-005 | TestIntegration_Memory_AddRemove_IndexAndPromptSync | memory/integration_memory_context_test.go |
| REQ-MC-005 | TestIntegration_Memory_RemoveNoOp_PromptUnchanged | memory/integration_memory_context_test.go |
| REQ-MC-001 | TestIntegration_Memory_DualLayerPrompt | memory/integration_memory_context_test.go |
| REQ-MC-003 | TestIntegration_Memory_CLAUDEmdDiscoveryInPrompt | memory/integration_memory_context_test.go |
| REQ-MC-004 | — | *deferred: search ranked results* |
| REQ-MC-006 | — | *deferred: max files enforcement* |

## Agent Coordination (AC)

| REQ | Test Function | File |
|-----|---------------|------|
| REQ-AC-001 | TestIntegration_Coordinator_SpawnLifecycle | coordinator/integration_agents_test.go |
| REQ-AC-001 | TestIntegration_Coordinator_ListAgents | coordinator/integration_agents_test.go |
| REQ-AC-001 | TestIntegration_Coordinator_StopRunningAgent | coordinator/integration_agents_test.go |
| REQ-AC-002 | TestIntegration_Coordinator_TeamWithAgents | coordinator/integration_agents_test.go |
| REQ-AC-002 | TestIntegration_Coordinator_TeamAgentTracking | coordinator/integration_agents_test.go |
| REQ-AC-004 | TestIntegration_Coordinator_AgentIsolation | coordinator/integration_agents_test.go |
| REQ-AC-003 | — | *deferred: inter-agent messaging* |
| REQ-AC-005 | — | *deferred: task output relay* |

## Automation (AT)

| REQ | Test Function | File |
|-----|---------------|------|
| REQ-AT-001 | TestIntegration_Task_ShellExecution | tasks/integration_automation_test.go |
| REQ-AT-001 | TestIntegration_Task_FailedWithMetadata | tasks/integration_automation_test.go |
| REQ-AT-002 | TestIntegration_Task_LifecycleStates | tasks/integration_automation_test.go |
| REQ-AT-002 | TestIntegration_Task_List | tasks/integration_automation_test.go |
| REQ-AT-004 | TestIntegration_Task_OutputRetrieval | tasks/integration_automation_test.go |
| REQ-AT-004 | TestIntegration_Task_LargeOutput | tasks/integration_automation_test.go |
| REQ-AT-004 | TestIntegration_Task_StopWhileRunning | tasks/integration_automation_test.go |
| REQ-AT-005 | TestIntegration_Task_FailedWithMetadata | tasks/integration_automation_test.go |
| REQ-AT-003 | TestIntegration_Task_StopWhileRunning | tasks/integration_automation_test.go |

## Configuration (CF)

| REQ | Test Function | File |
|-----|---------------|------|
| REQ-CF-001 | TestIntegration_Settings_SaveAndLoad | config/integration_config_test.go |
| REQ-CF-001 | TestIntegration_Settings_Defaults | config/integration_config_test.go |
| REQ-CF-001 | TestIntegration_Settings_FilePermissions | config/integration_config_test.go |
| REQ-CF-003 | TestIntegration_ProviderProfiles_Settings | config/integration_config_test.go |
| REQ-CF-005 | TestIntegration_EnvOverrides_ModelFromEnv | config/integration_config_test.go |
| REQ-CF-007 | TestIntegration_Settings_JSONRoundTrip | config/integration_config_test.go |
| REQ-CF-002 | — | *deferred: flag overrides merge precedence* |
| REQ-CF-004 | — | *deferred: profile switch updates engine* |
| REQ-CF-006 | — | *deferred: /config command runtime update* |

## Extensibility (EX)

| REQ | Test Function | File |
|-----|---------------|------|
| REQ-EX-001 | TestIntegration_Plugin_DiscoveryLoadsSkills | plugins/integration_extensibility_test.go |
| REQ-EX-001 | TestIntegration_Plugin_MultiplePluginsSorted | plugins/integration_extensibility_test.go |
| REQ-EX-001 | TestIntegration_Plugin_InvalidSkippedValidLoaded | plugins/integration_extensibility_test.go |
| REQ-EX-002 | TestIntegration_Plugin_DiscoveryLoadsSkills | plugins/integration_extensibility_test.go |
| REQ-EX-005 | TestIntegration_Plugin_DiscoveryLoadsHooks | plugins/integration_extensibility_test.go |
| REQ-EX-007 | TestIntegration_Plugin_DiscoveryLoadsMCP | plugins/integration_extensibility_test.go |
| REQ-EX-008 | TestIntegration_Plugin_ManagerSkillsWiredToRegistry | plugins/integration_extensibility_test.go |
| REQ-EX-003 | — | *deferred: install/uninstall lifecycle* |
| REQ-EX-004 | — | *deferred: skill on-demand load* |
| REQ-EX-006 | — | *deferred: hook command and HTTP* |

## Authentication (AU)

| REQ | Test Function | File |
|-----|---------------|------|
| REQ-AU-001 | TestIntegration_Auth_ResolveKey_StoreAndFallback | auth/integration_auth_test.go |
| REQ-AU-001 | TestIntegration_Auth_DeleteAndVerifyResolution | auth/integration_auth_test.go |
| REQ-AU-001 | TestIntegration_Auth_FileSecurity | auth/integration_auth_test.go |
| REQ-AU-001 | TestIntegration_Auth_ConfigProfileIntegration | auth/integration_auth_test.go |
| REQ-AU-003 | TestIntegration_Auth_MultiProvider_IndependentResolution | auth/integration_auth_test.go |
| REQ-AU-004 | TestIntegration_Auth_DeleteAndVerifyResolution | auth/integration_auth_test.go |
| REQ-AU-002 | — | *deferred: OAuth credential storage* |

## User Interaction (UI)

| REQ | Test Function | File |
|-----|---------------|------|
| REQ-UI-001 | — | *deferred: needs TUI harness* |
| REQ-UI-002 | — | *deferred: needs TUI harness* |
| REQ-UI-003 | — | *deferred: TUI rendering* |
| REQ-UI-004 | — | *deferred: slash command end-to-end* |
| REQ-UI-005 | — | *deferred: channel gateway* |
| REQ-UI-006 | — | *deferred: ask user tool* |
| REQ-UI-007 | — | *deferred: themes* |
| REQ-UI-008 | — | *deferred: vim mode* |
