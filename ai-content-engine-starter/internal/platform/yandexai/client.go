package yandexai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const defaultEndpoint = "https://llm.api.cloud.yandex.net/foundationModels/v1/completion"

// Client is a minimal Yandex AI text generation client for MVP usage.
type Client struct {
	httpClient *http.Client
	endpoint   string
	apiKey     string
	modelURI   string
}

// Config defines required runtime parameters for Yandex AI client.
type Config struct {
	Endpoint string
	APIKey   string
	ModelURI string
}

// New creates a Yandex AI client.
func New(httpClient *http.Client, cfg Config) (*Client, error) {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	endpoint := strings.TrimSpace(cfg.Endpoint)
	if endpoint == "" {
		endpoint = defaultEndpoint
	}
	if strings.TrimSpace(cfg.APIKey) == "" {
		return nil, fmt.Errorf("api key is empty")
	}
	if strings.TrimSpace(cfg.ModelURI) == "" {
		return nil, fmt.Errorf("model uri is empty")
	}

	return &Client{
		httpClient: httpClient,
		endpoint:   endpoint,
		apiKey:     strings.TrimSpace(cfg.APIKey),
		modelURI:   strings.TrimSpace(cfg.ModelURI),
	}, nil
}

// GenerateText sends prompt and returns generated response text.
func (c *Client) GenerateText(ctx context.Context, prompt string) (string, error) {
	if c == nil {
		return "", fmt.Errorf("yandex ai client is nil")
	}
	if ctx == nil {
		return "", fmt.Errorf("context is nil")
	}
	if c.httpClient == nil {
		c.httpClient = http.DefaultClient
	}
	if strings.TrimSpace(c.endpoint) == "" {
		c.endpoint = defaultEndpoint
	}
	if strings.TrimSpace(c.apiKey) == "" {
		return "", fmt.Errorf("api key is empty")
	}
	if strings.TrimSpace(c.modelURI) == "" {
		return "", fmt.Errorf("model uri is empty")
	}

	prompt = strings.TrimSpace(prompt)
	if prompt == "" {
		return "", fmt.Errorf("prompt is empty")
	}

	payload := completionRequest{
		ModelURI: c.modelURI,
		CompletionOptions: completionOptions{
			Stream:      false,
			Temperature: 0.2,
			MaxTokens:   "2000",
		},
		Messages: []message{{Role: "user", Text: prompt}},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "Api-Key "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		snippet, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return "", fmt.Errorf("unexpected status code: %d: %s", resp.StatusCode, strings.TrimSpace(string(snippet)))
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	text, err := parseGeneratedText(respBody)
	if err != nil {
		return "", fmt.Errorf("parse response: %w", err)
	}
	if text == "" {
		return "", fmt.Errorf("generated text is empty")
	}

	return text, nil
}

type completionRequest struct {
	ModelURI          string            `json:"modelUri"`
	CompletionOptions completionOptions `json:"completionOptions"`
	Messages          []message         `json:"messages"`
}

type completionOptions struct {
	Stream      bool    `json:"stream"`
	Temperature float64 `json:"temperature"`
	MaxTokens   string  `json:"maxTokens"`
}

type message struct {
	Role string `json:"role"`
	Text string `json:"text"`
}

type completionResponse struct {
	Result struct {
		Alternatives []struct {
			Message struct {
				Text string `json:"text"`
			} `json:"message"`
		} `json:"alternatives"`
	} `json:"result"`
}

func parseGeneratedText(payload []byte) (string, error) {
	var resp completionResponse
	if err := json.Unmarshal(payload, &resp); err != nil {
		return "", err
	}
	if len(resp.Result.Alternatives) == 0 {
		return "", nil
	}
	return strings.TrimSpace(resp.Result.Alternatives[0].Message.Text), nil
}
