package contracts

import "testing"

func TestValidateMode(t *testing.T) {
	for _, mode := range []string{ModeGeneralChat, ModePublicLookup, ModeBrowserLookup} {
		if err := ValidateMode(mode); err != nil {
			t.Fatalf("expected mode %s to be valid: %v", mode, err)
		}
	}

	if err := ValidateMode("unsupported"); err == nil {
		t.Fatal("expected unsupported mode to fail")
	}
}

func TestValidateStatus(t *testing.T) {
	for _, status := range []string{StatusOK, StatusNoResult, StatusAuthRequired, StatusError} {
		if err := ValidateStatus(status); err != nil {
			t.Fatalf("expected status %s to be valid: %v", status, err)
		}
	}

	if err := ValidateStatus("unsupported"); err == nil {
		t.Fatal("expected unsupported status to fail")
	}
}

func TestRouterResultValid(t *testing.T) {
	result := RouterResult{
		Mode:               ModePublicLookup,
		Confidence:         0.9,
		Reason:             "price lookup",
		NormalizedQuestion: "請問義美小泡芙多少錢",
		CacheHint:          &CacheHint{LookupType: "price", Target: "義美小泡芙"},
	}

	if err := result.Valid(0.65); err != nil {
		t.Fatalf("expected valid router result: %v", err)
	}

	result.Confidence = 0.4
	if err := result.Valid(0.65); err == nil {
		t.Fatal("expected low confidence to fail")
	}
}
