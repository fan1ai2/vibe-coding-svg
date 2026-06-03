package ai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// OpenAIClient implements Provider using any OpenAI-compatible API.
type OpenAIClient struct {
	baseURL string
	apiKey  string
	model   string
	builder *PromptBuilder
	client  *http.Client
}

func NewOpenAIClient(baseURL, apiKey, model string, builder *PromptBuilder) *OpenAIClient {
	return &OpenAIClient{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		apiKey:  apiKey,
		model:   model,
		builder: builder,
		client: &http.Client{Timeout: 180 * time.Second},
	}
}

type chatRequest struct {
	Model     string    `json:"model"`
	Messages  []message `json:"messages"`
	Stream    bool      `json:"stream"`
	MaxTokens int       `json:"max_tokens"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type streamChunk struct {
	Choices []struct {
		Delta   struct{ Content string `json:"content"` } `json:"delta"`
		Message struct{ Content string `json:"content"` } `json:"message"`
		Text    string `json:"text"`
	} `json:"choices"`
	Error *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

func (c *OpenAIClient) Generate(ctx context.Context, prompt, style string) ([]IconCandidate, error) {
	sysPrompt := c.builder.BuildSystemPrompt(style)

	const maxRetries = 3
	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			delay := time.Duration(attempt) * 2 * time.Second
			log.Printf("[AI] retry attempt %d/%d after %v", attempt+1, maxRetries, delay)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
		}

		candidates, err := c.call(ctx, sysPrompt, prompt)
		if err != nil {
			// Retry on upstream errors (502, upstream_error)
			if strings.Contains(err.Error(), "upstream") || strings.Contains(err.Error(), "502") {
				lastErr = err
				continue
			}
			return nil, fmt.Errorf("openai call: %w", err)
		}

		if len(candidates) == 0 {
			lastErr = fmt.Errorf("no valid candidates returned")
			continue
		}

		return candidates, nil
	}

	return nil, fmt.Errorf("all %d attempts failed: %w", maxRetries, lastErr)
}

func (c *OpenAIClient) call(ctx context.Context, sysPrompt, userPrompt string) ([]IconCandidate, error) {
	body := chatRequest{
		Model: c.model,
		Messages: []message{
			{Role: "system", Content: sysPrompt},
			{Role: "user", Content: userPrompt},
		},
		Stream:    true,
		MaxTokens: 8000,
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	url := c.baseURL + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("api returned %d: %s", resp.StatusCode, string(errBody))
	}

	allBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}
	bodyStr := string(allBytes)
	log.Printf("[AI] raw body: %d bytes, HTTP status: %d", len(allBytes), resp.StatusCode)

	if strings.HasPrefix(strings.TrimSpace(bodyStr), `{"error"`) {
		return nil, fmt.Errorf("api error: %.500s", bodyStr)
	}

	fullContent, err := parseSSE(bodyStr)
	if err != nil {
		return nil, fmt.Errorf("read sse: %w", err)
	}

	// 如果 SSE 解析为空，尝试按非流式 JSON 响应解析
	if fullContent == "" {
		fullContent = parseNonStreaming(bodyStr)
	}

	log.Printf("[AI] SSE parsed %d bytes of content, raw body first 600 chars: %.600s", len(fullContent), bodyStr)

	fullContent = stripMarkdownFences(fullContent)

	var candidates []IconCandidate
	if err := json.Unmarshal([]byte(fullContent), &candidates); err != nil {
		return nil, fmt.Errorf("parse candidates: %w\nraw: %.200s", err, fullContent)
	}

	return FilterCandidates(candidates), nil
}

func parseSSE(body string) (string, error) {
	var sb strings.Builder
	scanner := bufio.NewScanner(strings.NewReader(body))
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}
		var chunk streamChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}
		// 检测 SSE 流中的错误块（某些兼容 API 在流式响应中返回错误）
		if chunk.Error != nil && chunk.Error.Message != "" {
			return "", fmt.Errorf("upstream error: %s", chunk.Error.Message)
		}
		for _, choice := range chunk.Choices {
			if choice.Delta.Content != "" {
				sb.WriteString(choice.Delta.Content)
			} else if choice.Message.Content != "" {
				sb.WriteString(choice.Message.Content)
			} else if choice.Text != "" {
				sb.WriteString(choice.Text)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return sb.String(), nil
}

func parseNonStreaming(body string) string {
	var resp struct {
		Choices []struct {
			Message struct{ Content string `json:"content"` } `json:"message"`
			Text    string `json:"text"`
		} `json:"choices"`
	}
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		return ""
	}
	for _, choice := range resp.Choices {
		if choice.Message.Content != "" {
			return choice.Message.Content
		}
		if choice.Text != "" {
			return choice.Text
		}
	}
	return ""
}

func stripMarkdownFences(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```json")
		s = strings.TrimPrefix(s, "```")
		s = strings.TrimSuffix(s, "```")
		s = strings.TrimSpace(s)
	}
	return s
}
