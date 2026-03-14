package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	synerr "github.com/calvinchengx/synapse/internal/errors"
)

// Provider identifies the LLM backend.
type Provider string

const (
	ProviderLiteLLM   Provider = "litellm"
	ProviderAnthropic Provider = "anthropic"
	ProviderOpenAI    Provider = "openai"
)

// Client provides a unified interface for LLM completion calls.
type Client struct {
	Provider   Provider
	BaseURL    string
	APIKey     string
	Model      string
	HTTPClient *http.Client
	MaxRetries int
}

// NewClient creates an LLM client from configuration values.
func NewClient(provider, baseURL, apiKeyEnv, model string, timeout time.Duration, maxRetries int) *Client {
	apiKey := ""
	if apiKeyEnv != "" {
		apiKey = os.Getenv(apiKeyEnv)
	}

	return &Client{
		Provider: Provider(provider),
		BaseURL:  baseURL,
		APIKey:   apiKey,
		Model:    model,
		HTTPClient: &http.Client{
			Timeout: timeout,
		},
		MaxRetries: maxRetries,
	}
}

// Complete sends a chat completion request and returns the model's text response.
func (c *Client) Complete(ctx context.Context, messages []Message) (string, error) {
	var lastErr error
	for attempt := 0; attempt <= c.MaxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 1s, 2s
			select {
			case <-ctx.Done():
				return "", synerr.NewTimeoutError("llm", "complete", ctx.Err())
			case <-time.After(time.Duration(attempt) * time.Second):
			}
		}

		var result string
		var err error

		switch c.Provider {
		case ProviderAnthropic:
			result, err = c.completeAnthropic(ctx, messages)
		default:
			// LiteLLM and OpenAI both use OpenAI-compatible endpoints
			result, err = c.completeOpenAI(ctx, messages)
		}

		if err == nil {
			return result, nil
		}
		lastErr = err
	}

	return "", lastErr
}

// completeOpenAI calls an OpenAI-compatible /v1/chat/completions endpoint.
func (c *Client) completeOpenAI(ctx context.Context, messages []Message) (string, error) {
	reqBody := CompletionRequest{
		Model:       c.Model,
		Messages:    messages,
		MaxTokens:   1024,
		Temperature: 0,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", synerr.NewExternalError("llm", "marshal", err)
	}

	url := c.BaseURL + "/v1/chat/completions"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return "", synerr.NewExternalError("llm", "request", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return "", synerr.NewTimeoutError("llm", "complete", err)
		}
		return "", synerr.NewExternalError("llm", "connect", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", synerr.NewExternalError("llm", "read", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", synerr.NewExternalError("llm", "complete",
			fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody)))
	}

	var completionResp CompletionResponse
	if err := json.Unmarshal(respBody, &completionResp); err != nil {
		return "", synerr.NewExternalError("llm", "parse", err)
	}

	if len(completionResp.Choices) == 0 {
		return "", synerr.NewExternalError("llm", "complete", fmt.Errorf("no choices in response"))
	}

	return completionResp.Choices[0].Message.Content, nil
}

// completeAnthropic calls Anthropic's /v1/messages endpoint.
func (c *Client) completeAnthropic(ctx context.Context, messages []Message) (string, error) {
	reqBody := AnthropicRequest{
		Model:     c.Model,
		Messages:  messages,
		MaxTokens: 1024,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", synerr.NewExternalError("llm", "marshal", err)
	}

	url := c.BaseURL + "/v1/messages"
	if c.BaseURL == "" {
		url = "https://api.anthropic.com/v1/messages"
	}
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return "", synerr.NewExternalError("llm", "request", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return "", synerr.NewTimeoutError("llm", "complete", err)
		}
		return "", synerr.NewExternalError("llm", "connect", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", synerr.NewExternalError("llm", "read", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", synerr.NewExternalError("llm", "complete",
			fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody)))
	}

	var anthropicResp AnthropicResponse
	if err := json.Unmarshal(respBody, &anthropicResp); err != nil {
		return "", synerr.NewExternalError("llm", "parse", err)
	}

	if len(anthropicResp.Content) == 0 {
		return "", synerr.NewExternalError("llm", "complete", fmt.Errorf("no content in response"))
	}

	return anthropicResp.Content[0].Text, nil
}
