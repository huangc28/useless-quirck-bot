package chat

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestMockResponderGreeting(t *testing.T) {
	reply, err := (MockResponder{}).Respond(context.Background(), "你好")
	if err != nil {
		t.Fatalf("Respond returned error: %v", err)
	}
	if strings.TrimSpace(reply) == "" {
		t.Fatal("mock response is empty")
	}
	if !strings.Contains(reply, "請問義美小泡芙多少錢") {
		t.Fatalf("reply does not contain example: %q", reply)
	}
}

func TestLiveResponderParsesReply(t *testing.T) {
	httpClient := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if !strings.HasPrefix(r.Header.Get("Authorization"), "Bearer ") {
			t.Fatal("missing bearer token")
		}
		return response(200, `{"reply":"你好，我可以協助查找網站資料。"}`), nil
	})}

	responder := NewLiveResponder("key", "model", "http://chat.test", httpClient)
	reply, err := responder.Respond(context.Background(), "你好")
	if err != nil {
		t.Fatalf("Respond returned error: %v", err)
	}
	if reply == "" {
		t.Fatal("reply is empty")
	}
}

func TestLiveResponderRequiresCredentials(t *testing.T) {
	responder := NewLiveResponder("", "", "http://chat.test", nil)
	if _, err := responder.Respond(context.Background(), "你好"); err == nil {
		t.Fatal("expected missing API key error")
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
