package publisher

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"ai-content-engine-starter/internal/domain"
)

const telegramAPIBase = "https://api.telegram.org"

// Client publishes approved drafts to Telegram channels.
type Client struct {
	httpClient *http.Client
	botToken   string
	baseURL    string
}

// New creates a Telegram publisher client.
func New(httpClient *http.Client, botToken string) (*Client, error) {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	botToken = strings.TrimSpace(botToken)
	if botToken == "" {
		return nil, fmt.Errorf("bot token is empty")
	}
	return &Client{httpClient: httpClient, botToken: botToken, baseURL: telegramAPIBase}, nil
}

// PublishDraft sends draft to the target Telegram chat and returns sent message id.
func (c *Client) PublishDraft(ctx context.Context, draft domain.Draft, chatID string) (int64, error) {
	if c == nil {
		return 0, fmt.Errorf("publisher client is nil")
	}
	if ctx == nil {
		return 0, fmt.Errorf("context is nil")
	}
	if c.httpClient == nil {
		c.httpClient = http.DefaultClient
	}
	if strings.TrimSpace(c.baseURL) == "" {
		c.baseURL = telegramAPIBase
	}
	token := strings.TrimSpace(c.botToken)
	if token == "" {
		return 0, fmt.Errorf("bot token is empty")
	}
	c.botToken = token
	if draft.ID <= 0 {
		return 0, fmt.Errorf("draft id is invalid")
	}
	if draft.Status != domain.DraftStatusApproved {
		return 0, fmt.Errorf("draft status must be approved")
	}
	if strings.TrimSpace(draft.Body) == "" {
		return 0, fmt.Errorf("draft body is empty")
	}
	chatID = strings.TrimSpace(chatID)
	if chatID == "" {
		return 0, fmt.Errorf("chat id is empty")
	}

	payload := sendMessageRequest{ChatID: chatID, Text: strings.TrimSpace(draft.Body), DisableWebPagePreview: true}
	body, err := json.Marshal(payload)
	if err != nil {
		return 0, fmt.Errorf("marshal request: %w", err)
	}

	endpoint := fmt.Sprintf("%s/bot%s/sendMessage", strings.TrimRight(strings.TrimSpace(c.baseURL), "/"), c.botToken)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return 0, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		snippet, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return 0, fmt.Errorf("unexpected status code: %d: %s", resp.StatusCode, strings.TrimSpace(string(snippet)))
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("read response: %w", err)
	}
	messageID, err := parseSendMessageResponse(respBody)
	if err != nil {
		return 0, fmt.Errorf("parse response: %w", err)
	}
	if messageID <= 0 {
		return 0, fmt.Errorf("message id is invalid")
	}

	return messageID, nil
}

type sendMessageRequest struct {
	ChatID                string `json:"chat_id"`
	Text                  string `json:"text"`
	DisableWebPagePreview bool   `json:"disable_web_page_preview"`
}

type sendMessageResponse struct {
	OK          bool   `json:"ok"`
	Description string `json:"description"`
	Result      struct {
		MessageID int64 `json:"message_id"`
	} `json:"result"`
}

func parseSendMessageResponse(payload []byte) (int64, error) {
	var resp sendMessageResponse
	if err := json.Unmarshal(payload, &resp); err != nil {
		return 0, err
	}
	if !resp.OK {
		if strings.TrimSpace(resp.Description) == "" {
			return 0, fmt.Errorf("telegram api returned not ok")
		}
		return 0, fmt.Errorf("telegram api error: %s", strings.TrimSpace(resp.Description))
	}
	return resp.Result.MessageID, nil
}
