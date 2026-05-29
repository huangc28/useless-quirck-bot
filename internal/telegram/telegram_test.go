package telegram

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestParseUpdate(t *testing.T) {
	payload := `{"update_id":1,"message":{"chat":{"id":12345},"text":"請問義美小泡芙多少錢"}}`
	chatID, text, ok, err := ParseUpdate(strings.NewReader(payload))
	if err != nil {
		t.Fatalf("ParseUpdate returned error: %v", err)
	}
	if !ok {
		t.Fatal("expected ok")
	}
	if chatID != 12345 {
		t.Fatalf("chatID = %d", chatID)
	}
	if text != "請問義美小泡芙多少錢" {
		t.Fatalf("text = %q", text)
	}
}

func TestParseUpdateIgnoresNonText(t *testing.T) {
	_, _, ok, err := ParseUpdate(strings.NewReader(`{"message":{"chat":{"id":12345}}}`))
	if err != nil {
		t.Fatalf("ParseUpdate returned error: %v", err)
	}
	if ok {
		t.Fatal("expected non-text update to be ignored")
	}
}

func TestHTTPSenderSendMessage(t *testing.T) {
	var gotPath string
	var gotBody string
	sender := NewHTTPSender("test-token", &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		gotPath = r.URL.Path
		data, _ := io.ReadAll(r.Body)
		gotBody = string(data)
		return response(200, `{"ok":true}`), nil
	})})
	sender.SetAPIBase("https://telegram.test")

	if err := sender.SendMessage(context.Background(), 12345, "hello"); err != nil {
		t.Fatalf("SendMessage returned error: %v", err)
	}
	if gotPath != "/bottest-token/sendMessage" {
		t.Fatalf("path = %q", gotPath)
	}
	if !strings.Contains(gotBody, `"chat_id":12345`) || !strings.Contains(gotBody, `"text":"hello"`) {
		t.Fatalf("unexpected body: %s", gotBody)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func response(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Status:     http.StatusText(status),
		Body:       io.NopCloser(bytes.NewBufferString(body)),
		Header:     make(http.Header),
	}
}
