package router

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"interview-chatbot/internal/contracts"
)

func TestMockClientRoutesLockedFixtures(t *testing.T) {
	client := MockClient{Threshold: 0.65}
	tests := []struct {
		name       string
		message    string
		wantMode   string
		wantTarget string
		wantSite   string
	}{
		{name: "general", message: "你好", wantMode: contracts.ModeGeneralChat},
		{name: "public price", message: "請問義美小泡芙多少錢", wantMode: contracts.ModePublicLookup, wantTarget: "義美小泡芙"},
		{name: "browser site", message: "請問 momo 義美小泡芙多少錢", wantMode: contracts.ModeBrowserLookup, wantTarget: "義美小泡芙", wantSite: "momo"},
		{name: "browser url", message: "請幫我查 https://example.com", wantMode: contracts.ModeBrowserLookup},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := client.Route(context.Background(), tt.message)
			if err != nil {
				t.Fatalf("Route returned error: %v", err)
			}
			if got.Mode != tt.wantMode {
				t.Fatalf("Mode = %s, want %s", got.Mode, tt.wantMode)
			}
			if tt.wantTarget != "" && (got.CacheHint == nil || got.CacheHint.Target != tt.wantTarget) {
				t.Fatalf("CacheHint.Target = %#v", got.CacheHint)
			}
			if tt.wantSite != "" && (got.CacheHint == nil || got.CacheHint.Site != tt.wantSite) {
				t.Fatalf("CacheHint.Site = %#v", got.CacheHint)
			}
		})
	}
}

func TestNormalizeQuestionPreservesLanguageSiteAndURL(t *testing.T) {
	got := NormalizeQuestion("  /lookup 請問 momo 義美小泡芙多少錢 https://example.com/a  ")
	want := "請問 momo 義美小泡芙多少錢 https://example.com/a"
	if got != want {
		t.Fatalf("NormalizeQuestion = %q, want %q", got, want)
	}
}

func TestValidateResultLowConfidenceClarifies(t *testing.T) {
	result := contracts.RouterResult{
		Mode:               contracts.ModePublicLookup,
		Confidence:         0.4,
		Reason:             "maybe lookup",
		NormalizedQuestion: "泡芙多少錢",
	}

	_, decision := ValidateResult(result, 0.65)
	if decision.Kind != DecisionClarify {
		t.Fatalf("decision = %s, want %s", decision.Kind, DecisionClarify)
	}
}

func TestValidateResultUnknownModeFallsBack(t *testing.T) {
	result := contracts.RouterResult{
		Mode:               "bad",
		Confidence:         0.9,
		Reason:             "bad",
		NormalizedQuestion: "請問義美小泡芙多少錢",
	}

	_, decision := ValidateResult(result, 0.65)
	if decision.Kind != DecisionFallback {
		t.Fatalf("decision = %s, want %s", decision.Kind, DecisionFallback)
	}
}

func TestValidateResultRejectsInvalidConfidenceAndMissingQuestion(t *testing.T) {
	_, decision := ValidateResult(contracts.RouterResult{
		Mode:               contracts.ModePublicLookup,
		Confidence:         1.5,
		NormalizedQuestion: "請問義美小泡芙多少錢",
	}, 0.65)
	if decision.Kind != DecisionFallback {
		t.Fatalf("invalid confidence decision = %s", decision.Kind)
	}

	_, decision = ValidateResult(contracts.RouterResult{
		Mode:       contracts.ModePublicLookup,
		Confidence: 0.9,
		Reason:     "price lookup",
	}, 0.65)
	if decision.Kind != DecisionFallback {
		t.Fatalf("missing question decision = %s", decision.Kind)
	}

	_, decision = ValidateResult(contracts.RouterResult{
		Mode:               contracts.ModePublicLookup,
		Confidence:         0.9,
		NormalizedQuestion: "   ",
	}, 0.65)
	if decision.Kind != DecisionFallback {
		t.Fatalf("blank question decision = %s", decision.Kind)
	}
}

func TestLiveClientParsesRouterJSON(t *testing.T) {
	script := writeRouterScript(t, "router.sh", "#!/bin/sh\ncat >/dev/null\nprintf '{\"mode\":\"public_lookup\",\"confidence\":0.91,\"reason\":\"price lookup\",\"normalized_question\":\"請問義美小泡芙多少錢\",\"cache_hint\":{\"lookup_type\":\"price\",\"target\":\"義美小泡芙\"}}'\n")
	client := NewLiveClient(script, 5*time.Second, 0.65)
	got, err := client.Route(context.Background(), "請問義美小泡芙多少錢")
	if err != nil {
		t.Fatalf("Route returned error: %v", err)
	}
	if got.Mode != contracts.ModePublicLookup {
		t.Fatalf("Mode = %s", got.Mode)
	}
}

func TestLiveClientInvalidJSON(t *testing.T) {
	script := writeRouterScript(t, "invalid.sh", "#!/bin/sh\ncat >/dev/null\nprintf 'not json'\n")
	client := NewLiveClient(script, 5*time.Second, 0.65)
	if _, err := client.Route(context.Background(), "請問義美小泡芙多少錢"); err == nil {
		t.Fatal("expected invalid JSON error")
	}
}

func TestLiveClientReturnsLowConfidenceForAppClarification(t *testing.T) {
	script := writeRouterScript(t, "low-confidence.sh", "#!/bin/sh\ncat >/dev/null\nprintf '{\"mode\":\"public_lookup\",\"confidence\":0.2,\"reason\":\"uncertain\",\"normalized_question\":\"泡芙\"}'\n")
	client := NewLiveClient(script, 5*time.Second, 0.65)
	got, err := client.Route(context.Background(), "泡芙")
	if err != nil {
		t.Fatalf("Route returned error: %v", err)
	}
	if got.Confidence != 0.2 {
		t.Fatalf("Confidence = %f", got.Confidence)
	}
}

func TestLiveClientDoesNotInventMissingNormalizedQuestion(t *testing.T) {
	script := writeRouterScript(t, "missing-question.sh", "#!/bin/sh\ncat >/dev/null\nprintf '{\"mode\":\"public_lookup\",\"confidence\":0.9,\"reason\":\"price lookup\"}'\n")
	client := NewLiveClient(script, 5*time.Second, 0.65)
	got, err := client.Route(context.Background(), "請問義美小泡芙多少錢")
	if err != nil {
		t.Fatalf("Route returned error: %v", err)
	}
	if got.NormalizedQuestion != "" {
		t.Fatalf("NormalizedQuestion = %q", got.NormalizedQuestion)
	}
	_, decision := ValidateResult(got, 0.65)
	if decision.Kind != DecisionFallback {
		t.Fatalf("decision = %s", decision.Kind)
	}
}

func TestLiveClientRequiresCommand(t *testing.T) {
	client := NewLiveClient("", 5*time.Second, 0.65)
	if _, err := client.Route(context.Background(), "你好"); err == nil {
		t.Fatal("expected missing command error")
	}
}

func writeRouterScript(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	return path
}
