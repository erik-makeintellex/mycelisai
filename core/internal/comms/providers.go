package comms

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const defaultHTTPTimeout = 20 * time.Second

func defaultHTTPClient() *http.Client {
	return &http.Client{Timeout: defaultHTTPTimeout}
}

type webhookProvider struct {
	name        string
	channel     string
	description string
	endpoint    string
	headers     map[string]string
	client      *http.Client
}

func newWebhookProvider(name, channel, description, endpoint string, headers map[string]string) Provider {
	if headers == nil {
		headers = map[string]string{}
	}
	return &webhookProvider{
		name:        name,
		channel:     channel,
		description: description,
		endpoint:    strings.TrimSpace(endpoint),
		headers:     headers,
		client:      defaultHTTPClient(),
	}
}

func (p *webhookProvider) Info() ProviderInfo {
	return ProviderInfo{
		Name:        p.name,
		Channel:     p.channel,
		Description: p.description,
		Configured:  p.endpoint != "",
	}
}

func (p *webhookProvider) Send(ctx context.Context, req SendRequest) (SendResult, error) {
	if p.endpoint == "" {
		return SendResult{}, fmt.Errorf("%s provider is not configured", p.name)
	}

	payload := map[string]any{
		"recipient": req.Recipient,
		"message":   req.Message,
		"metadata":  req.Metadata,
	}
	data, _ := json.Marshal(payload)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.endpoint, bytes.NewBuffer(data))
	if err != nil {
		return SendResult{}, fmt.Errorf("build webhook request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	for k, v := range p.headers {
		httpReq.Header.Set(k, v)
	}

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return SendResult{}, fmt.Errorf("webhook send failed: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		return SendResult{}, fmt.Errorf("webhook provider %s returned %d: %s", p.name, resp.StatusCode, strings.TrimSpace(string(body)))
	}

	return SendResult{
		Provider: p.name,
		Status:   "sent",
		Metadata: map[string]any{"http_status": resp.StatusCode},
	}, nil
}

type telegramProvider struct {
	token   string
	baseURL string
	client  *http.Client
}

func newTelegramProvider(token string) Provider {
	return &telegramProvider{
		token:   strings.TrimSpace(token),
		baseURL: "https://api.telegram.org",
		client:  defaultHTTPClient(),
	}
}

func (p *telegramProvider) Info() ProviderInfo {
	return ProviderInfo{
		Name:        "telegram",
		Channel:     "chat",
		Description: "Telegram Bot API sender",
		Configured:  p.token != "",
	}
}

func (p *telegramProvider) Send(ctx context.Context, req SendRequest) (SendResult, error) {
	if p.token == "" {
		return SendResult{}, fmt.Errorf("telegram provider is not configured")
	}
	if strings.TrimSpace(req.Recipient) == "" {
		return SendResult{}, fmt.Errorf("telegram recipient (chat_id) is required")
	}

	body := map[string]any{
		"chat_id": req.Recipient,
		"text":    req.Message,
	}
	data, _ := json.Marshal(body)
	u := fmt.Sprintf("%s/bot%s/sendMessage", strings.TrimSuffix(p.baseURL, "/"), p.token)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewBuffer(data))
	if err != nil {
		return SendResult{}, fmt.Errorf("build telegram request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return SendResult{}, fmt.Errorf("telegram send failed: %w", err)
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		return SendResult{}, fmt.Errorf("telegram returned %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var parsed struct {
		Result struct {
			MessageID int64 `json:"message_id"`
		} `json:"result"`
	}
	_ = json.Unmarshal(respBody, &parsed)

	return SendResult{
		Provider:          "telegram",
		ProviderMessageID: fmt.Sprintf("%d", parsed.Result.MessageID),
		Status:            "sent",
	}, nil
}

type twilioWhatsAppProvider struct {
	accountSID string
	authToken  string
	fromNumber string
	baseURL    string
	client     *http.Client
}

func newTwilioWhatsAppProvider(accountSID, authToken, fromNumber string) Provider {
	return &twilioWhatsAppProvider{
		accountSID: strings.TrimSpace(accountSID),
		authToken:  strings.TrimSpace(authToken),
		fromNumber: strings.TrimSpace(fromNumber),
		baseURL:    "https://api.twilio.com",
		client:     defaultHTTPClient(),
	}
}

func (p *twilioWhatsAppProvider) Info() ProviderInfo {
	configured := p.accountSID != "" && p.authToken != "" && p.fromNumber != ""
	return ProviderInfo{
		Name:        "whatsapp",
		Channel:     "chat",
		Description: "Twilio WhatsApp sender",
		Configured:  configured,
	}
}

func (p *twilioWhatsAppProvider) Send(ctx context.Context, req SendRequest) (SendResult, error) {
	if p.accountSID == "" || p.authToken == "" || p.fromNumber == "" {
		return SendResult{}, fmt.Errorf("whatsapp provider is not configured")
	}
	if strings.TrimSpace(req.Recipient) == "" {
		return SendResult{}, fmt.Errorf("whatsapp recipient is required")
	}

	form := url.Values{}
	form.Set("To", ensureWhatsAppPrefix(req.Recipient))
	form.Set("From", ensureWhatsAppPrefix(p.fromNumber))
	form.Set("Body", req.Message)

	u := fmt.Sprintf("%s/2010-04-01/Accounts/%s/Messages.json", strings.TrimSuffix(p.baseURL, "/"), p.accountSID)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, u, strings.NewReader(form.Encode()))
	if err != nil {
		return SendResult{}, fmt.Errorf("build twilio request: %w", err)
	}
	httpReq.SetBasicAuth(p.accountSID, p.authToken)
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return SendResult{}, fmt.Errorf("twilio send failed: %w", err)
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		return SendResult{}, fmt.Errorf("twilio returned %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var parsed struct {
		SID string `json:"sid"`
	}
	_ = json.Unmarshal(respBody, &parsed)

	return SendResult{
		Provider:          "whatsapp",
		ProviderMessageID: parsed.SID,
		Status:            "sent",
	}, nil
}

func ensureWhatsAppPrefix(n string) string {
	n = strings.TrimSpace(n)
	if strings.HasPrefix(strings.ToLower(n), "whatsapp:") {
		return n
	}
	return "whatsapp:" + n
}

// NewGatewayFromEnv builds a communication gateway with common providers.
// Providers are always registered; if credentials are missing they remain visible as unconfigured.
func NewGatewayFromEnv() *Gateway {
	g := NewGateway()

	slackWebhook := os.Getenv("MYCELIS_COMMS_SLACK_WEBHOOK_URL")
	g.Register(newWebhookProvider(
		"slack",
		"chat",
		"Slack incoming webhook",
		slackWebhook,
		nil,
	))

	customWebhook := os.Getenv("MYCELIS_COMMS_WEBHOOK_URL")
	headers := map[string]string{}
	if token := os.Getenv("MYCELIS_COMMS_WEBHOOK_BEARER"); token != "" {
		headers["Authorization"] = "Bearer " + token
	}
	g.Register(newWebhookProvider(
		"webhook",
		"automation",
		"Generic webhook callback",
		customWebhook,
		headers,
	))

	g.Register(newTelegramProvider(os.Getenv("MYCELIS_COMMS_TELEGRAM_BOT_TOKEN")))
	g.Register(newTwilioWhatsAppProvider(
		os.Getenv("MYCELIS_COMMS_TWILIO_ACCOUNT_SID"),
		os.Getenv("MYCELIS_COMMS_TWILIO_AUTH_TOKEN"),
		os.Getenv("MYCELIS_COMMS_WHATSAPP_FROM"),
	))

	return g
}
