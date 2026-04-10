package api

// UsageSnapshot tracks token usage for a single API call.
type UsageSnapshot struct {
	InputTokens              int `json:"input_tokens"`
	OutputTokens             int `json:"output_tokens"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
}

// TotalTokens returns the sum of input and output tokens.
func (u UsageSnapshot) TotalTokens() int {
	return u.InputTokens + u.OutputTokens
}
