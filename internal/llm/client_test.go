package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	synerr "github.com/calvinchengx/synapse/internal/errors"
)

func TestCompleteOpenAI(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Error("missing Content-Type header")
		}

		resp := CompletionResponse{
			Choices: []Choice{
				{Message: Message{Role: "assistant", Content: `["security.md", "react.md"]`}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := &Client{
		Provider:   ProviderOpenAI,
		BaseURL:    server.URL,
		Model:      "gpt-4o-mini",
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
		MaxRetries: 0,
	}

	result, err := client.Complete(context.Background(), []Message{
		{Role: "user", Content: "test prompt"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != `["security.md", "react.md"]` {
		t.Errorf("unexpected result: %s", result)
	}
}

func TestCompleteLiteLLM(t *testing.T) {
	// LiteLLM uses the same OpenAI-compatible endpoint
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := CompletionResponse{
			Choices: []Choice{
				{Message: Message{Role: "assistant", Content: `["testing.md"]`}},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := &Client{
		Provider:   ProviderLiteLLM,
		BaseURL:    server.URL,
		Model:      "claude-3-haiku",
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
		MaxRetries: 0,
	}

	result, err := client.Complete(context.Background(), []Message{
		{Role: "user", Content: "write tests"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != `["testing.md"]` {
		t.Errorf("unexpected result: %s", result)
	}
}

func TestCompleteAnthropic(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/messages" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("x-api-key") != "test-key" {
			t.Error("missing x-api-key header")
		}
		if r.Header.Get("anthropic-version") != "2023-06-01" {
			t.Error("missing anthropic-version header")
		}

		resp := AnthropicResponse{
			Content: []ContentBlock{
				{Type: "text", Text: `["security.md"]`},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := &Client{
		Provider:   ProviderAnthropic,
		BaseURL:    server.URL,
		APIKey:     "test-key",
		Model:      "claude-3-haiku-20240307",
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
		MaxRetries: 0,
	}

	result, err := client.Complete(context.Background(), []Message{
		{Role: "user", Content: "check security"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != `["security.md"]` {
		t.Errorf("unexpected result: %s", result)
	}
}

func TestCompleteHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	}))
	defer server.Close()

	client := &Client{
		Provider:   ProviderOpenAI,
		BaseURL:    server.URL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
		MaxRetries: 0,
	}

	_, err := client.Complete(context.Background(), []Message{
		{Role: "user", Content: "test"},
	})
	if err == nil {
		t.Fatal("expected error for 500 response")
	}

	var extErr *synerr.ExternalError
	if !synerr.AsExternalError(err, &extErr) {
		t.Fatalf("expected ExternalError, got %T", err)
	}
	if extErr.Source != "llm" {
		t.Errorf("source = %q, want llm", extErr.Source)
	}
}

func TestCompleteTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := &Client{
		Provider:   ProviderOpenAI,
		BaseURL:    server.URL,
		HTTPClient: &http.Client{Timeout: 50 * time.Millisecond},
		MaxRetries: 0,
	}

	_, err := client.Complete(context.Background(), []Message{
		{Role: "user", Content: "test"},
	})
	if err == nil {
		t.Fatal("expected timeout error")
	}
}

func TestCompleteRetry(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("fail"))
			return
		}
		resp := CompletionResponse{
			Choices: []Choice{
				{Message: Message{Role: "assistant", Content: "success"}},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := &Client{
		Provider:   ProviderOpenAI,
		BaseURL:    server.URL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
		MaxRetries: 2,
	}

	result, err := client.Complete(context.Background(), []Message{
		{Role: "user", Content: "test"},
	})
	if err != nil {
		t.Fatalf("unexpected error after retries: %v", err)
	}
	if result != "success" {
		t.Errorf("result = %q, want success", result)
	}
	if attempts != 3 {
		t.Errorf("attempts = %d, want 3", attempts)
	}
}

func TestCompleteEmptyChoices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := CompletionResponse{Choices: []Choice{}}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := &Client{
		Provider:   ProviderOpenAI,
		BaseURL:    server.URL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
		MaxRetries: 0,
	}

	_, err := client.Complete(context.Background(), []Message{
		{Role: "user", Content: "test"},
	})
	if err == nil {
		t.Fatal("expected error for empty choices")
	}
}
