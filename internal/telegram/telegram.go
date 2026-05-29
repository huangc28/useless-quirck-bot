package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const defaultAPIBase = "https://api.telegram.org"

type Update struct {
	Message *Message `json:"message"`
}

type Message struct {
	Chat Chat   `json:"chat"`
	Text string `json:"text"`
}

type Chat struct {
	ID int64 `json:"id"`
}

func ParseUpdate(r io.Reader) (chatID int64, text string, ok bool, err error) {
	var update Update
	if err := json.NewDecoder(r).Decode(&update); err != nil {
		return 0, "", false, err
	}
	if update.Message == nil || update.Message.Chat.ID == 0 || strings.TrimSpace(update.Message.Text) == "" {
		return 0, "", false, nil
	}
	return update.Message.Chat.ID, update.Message.Text, true, nil
}

type Sender interface {
	SendMessage(ctx context.Context, chatID int64, text string) error
}

type HTTPSender struct {
	token      string
	apiBase    string
	httpClient *http.Client
}

func NewHTTPSender(token string, httpClient *http.Client) *HTTPSender {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &HTTPSender{token: token, apiBase: defaultAPIBase, httpClient: httpClient}
}

func (s *HTTPSender) SetAPIBase(base string) {
	s.apiBase = strings.TrimRight(base, "/")
}

func (s *HTTPSender) SendMessage(ctx context.Context, chatID int64, text string) error {
	if strings.TrimSpace(s.token) == "" {
		return errors.New("telegram token is empty")
	}
	payload, err := json.Marshal(map[string]any{
		"chat_id": chatID,
		"text":    text,
	})
	if err != nil {
		return err
	}
	endpoint := fmt.Sprintf("%s/bot%s/sendMessage", s.apiBase, s.token)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("telegram sendMessage failed: %s", resp.Status)
	}
	return nil
}
