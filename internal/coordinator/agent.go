package coordinator

// AgentDef describes an agent definition loaded from a YAML file.
type AgentDef struct {
	Name        string   `yaml:"name" json:"name"`
	Description string   `yaml:"description" json:"description"`
	Prompt      string   `yaml:"prompt" json:"prompt"`
	Model       string   `yaml:"model,omitempty" json:"model,omitempty"`
	Tools       []string `yaml:"tools,omitempty" json:"tools,omitempty"`
	MaxTurns    int      `yaml:"max_turns,omitempty" json:"max_turns,omitempty"`
}
