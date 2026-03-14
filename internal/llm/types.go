package llm

// Message represents a chat message for LLM completion.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// CompletionRequest is the request body for OpenAI-compatible /v1/chat/completions.
type CompletionRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
}

// CompletionResponse is the response from /v1/chat/completions.
type CompletionResponse struct {
	Choices []Choice `json:"choices"`
}

// Choice is a single completion choice.
type Choice struct {
	Message Message `json:"message"`
}

// AnthropicRequest is the request body for Anthropic's /v1/messages endpoint.
type AnthropicRequest struct {
	Model     string    `json:"model"`
	Messages  []Message `json:"messages"`
	MaxTokens int       `json:"max_tokens"`
}

// AnthropicResponse is the response from Anthropic's /v1/messages endpoint.
type AnthropicResponse struct {
	Content []ContentBlock `json:"content"`
}

// ContentBlock is a single content block in an Anthropic response.
type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}
