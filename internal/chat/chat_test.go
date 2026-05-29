package chat

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
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
	script := writeChatScript(t, "chat.sh", "#!/bin/sh\ncat >/dev/null\nprintf '你好，我可以協助查找網站資料。'\n")
	responder := NewLiveResponder(script, 5*time.Second)
	reply, err := responder.Respond(context.Background(), "你好")
	if err != nil {
		t.Fatalf("Respond returned error: %v", err)
	}
	if reply == "" {
		t.Fatal("reply is empty")
	}
}

func TestLiveResponderRequiresCommand(t *testing.T) {
	responder := NewLiveResponder("", 5*time.Second)
	if _, err := responder.Respond(context.Background(), "你好"); err == nil {
		t.Fatal("expected missing command error")
	}
}

func TestLiveResponderRejectsEmptyOutput(t *testing.T) {
	script := writeChatScript(t, "empty.sh", "#!/bin/sh\ncat >/dev/null\n")
	responder := NewLiveResponder(script, 5*time.Second)
	if _, err := responder.Respond(context.Background(), "你好"); err == nil {
		t.Fatal("expected empty response error")
	}
}

func TestLiveResponderTimeout(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell sleep script is Unix-specific")
	}
	script := writeChatScript(t, "sleep.sh", "#!/bin/sh\nsleep 2\nprintf done\n")
	responder := NewLiveResponder(script, 10*time.Millisecond)
	_, err := responder.Respond(context.Background(), "你好")
	if err == nil || !strings.Contains(err.Error(), "timeout") {
		t.Fatalf("expected timeout error, got %v", err)
	}
}

func writeChatScript(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	return path
}
