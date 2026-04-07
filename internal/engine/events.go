package engine

// EventType identifies the kind of engine event.
type EventType int

const (
	EventTextDelta      EventType = iota
	EventToolStarted
	EventToolCompleted
	EventTurnComplete
	EventError
	EventStatus
)

// EngineEvent is emitted by the engine during the agent loop.
type EngineEvent struct {
	Type EventType
	Data any
}

// AssistantTextDelta carries incremental assistant text.
type AssistantTextDelta struct {
	Text string
}

// ToolExecutionStarted signals that a tool is about to run.
type ToolExecutionStarted struct {
	ToolName  string
	ToolInput string
}

// ToolExecutionCompleted signals that a tool has finished.
type ToolExecutionCompleted struct {
	ToolName string
	Output   string
	IsError  bool
}

// AssistantTurnComplete signals the end of a complete assistant response.
type AssistantTurnComplete struct {
	InputTokens  int
	OutputTokens int
}

// ErrorEvent carries an error message to the UI.
type ErrorEvent struct {
	Message     string
	Recoverable bool
}

// StatusEvent carries a transient status message.
type StatusEvent struct {
	Message string
}
