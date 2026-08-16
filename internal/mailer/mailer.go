package mailer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Provider string

const (
	ProviderMailgun  Provider = "mailgun"
	ProviderSendGrid Provider = "sendgrid"
)

type Config struct {
	Provider Provider
	APIKey   string
	Domain   string // Mailgun domain (e.g., mg.example.com)
	From     string
	FromName string
	BaseURL  string // App base URL for email links (e.g., https://app.example.com)
}

type Mailer struct {
	cfg Config
	hc  *http.Client
}

func New(cfg Config) *Mailer {
	return &Mailer{
		cfg: cfg,
		hc:  &http.Client{Timeout: 15 * time.Second},
	}
}

func (m *Mailer) Send(to, subject, body string) error {
	switch m.cfg.Provider {
	case ProviderMailgun:
		return m.sendMailgun(to, subject, body)
	case ProviderSendGrid:
		return m.sendSendGrid(to, subject, body)
	default:
		return fmt.Errorf("unknown mail provider: %s", m.cfg.Provider)
	}
}

// sendMailgun sends via Mailgun API v3
func (m *Mailer) sendMailgun(to, subject, body string) error {
	from := m.cfg.From
	if m.cfg.FromName != "" {
		from = fmt.Sprintf("%s <%s>", m.cfg.FromName, m.cfg.From)
	}

	payload := bytes.NewBufferString(fmt.Sprintf(
		"from=%s&to=%s&subject=%s&html=%s",
		urlQueryEscape(from),
		urlQueryEscape(to),
		urlQueryEscape(subject),
		urlQueryEscape(body),
	))

	req, err := http.NewRequest("POST",
		fmt.Sprintf("https://api.mailgun.net/v3/%s/messages", m.cfg.Domain),
		payload)
	if err != nil {
		return err
	}
	req.SetBasicAuth("api", m.cfg.APIKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return m.do(req)
}

// sendSendGrid sends via SendGrid v3 API
func (m *Mailer) sendSendGrid(to, subject, body string) error {
	from := map[string]string{"email": m.cfg.From}
	if m.cfg.FromName != "" {
		from["name"] = m.cfg.FromName
	}

	sgPayload := map[string]any{
		"personalizations": []map[string]any{
			{"to": []map[string]string{{"email": to}}},
		},
		"from": from,
		"subject": subject,
		"content": []map[string]string{
			{"type": "text/html", "value": body},
		},
	}

	b, _ := json.Marshal(sgPayload)
	req, err := http.NewRequest("POST", "https://api.sendgrid.com/v3/mail/send", bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+m.cfg.APIKey)
	req.Header.Set("Content-Type", "application/json")

	return m.do(req)
}

func (m *Mailer) do(req *http.Request) error {
	resp, err := m.hc.Do(req)
	if err != nil {
		return fmt.Errorf("mail request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	body, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("mail API returned %d: %s", resp.StatusCode, string(body))
}

func urlQueryEscape(s string) string {
	// Simple form encoding — RFC-compliant for Latin/ASCII
	var buf bytes.Buffer
	for _, c := range []byte(s) {
		switch {
		case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c >= '0' && c <= '9',
			c == '_', c == '-', c == '.', c == '~':
			buf.WriteByte(c)
		case c == ' ':
			buf.WriteByte('+')
		default:
			fmt.Fprintf(&buf, "%%%02X", c)
		}
	}
	return buf.String()
}
