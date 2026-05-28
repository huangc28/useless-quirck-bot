package contracts

import (
	"errors"
	"fmt"
)

const ModeGeneralChat = "general_chat"
const ModePublicLookup = "public_lookup"
const ModeBrowserLookup = "browser_lookup"

const StatusOK = "ok"
const StatusNoResult = "no_result"
const StatusAuthRequired = "auth_required"
const StatusError = "error"

type CacheHint struct {
	LookupType string `json:"lookup_type,omitempty"`
	Target     string `json:"target,omitempty"`
	Site       string `json:"site,omitempty"`
	URL        string `json:"url,omitempty"`
}

type RouterResult struct {
	Mode               string     `json:"mode"`
	Confidence         float64    `json:"confidence"`
	Reason             string     `json:"reason"`
	NormalizedQuestion string     `json:"normalized_question"`
	CacheHint          *CacheHint `json:"cache_hint,omitempty"`
	ForceRefresh       bool       `json:"force_refresh,omitempty"`
}

type LookupRequest struct {
	Mode               string     `json:"mode"`
	Question           string     `json:"question"`
	NormalizedQuestion string     `json:"normalized_question"`
	CacheHint          *CacheHint `json:"cache_hint,omitempty"`
	ForceRefresh       bool       `json:"force_refresh,omitempty"`
}

type Evidence struct {
	Title      string `json:"title,omitempty"`
	URL        string `json:"url,omitempty"`
	Source     string `json:"source,omitempty"`
	Snippet    string `json:"snippet,omitempty"`
	ObservedAt string `json:"observed_at,omitempty"`
}

type LookupResponse struct {
	Status     string     `json:"status"`
	Answer     string     `json:"answer,omitempty"`
	Evidence   []Evidence `json:"evidence,omitempty"`
	ObservedAt string     `json:"observed_at,omitempty"`
	Error      string     `json:"error,omitempty"`
}

func ValidateMode(mode string) error {
	switch mode {
	case ModeGeneralChat, ModePublicLookup, ModeBrowserLookup:
		return nil
	default:
		return fmt.Errorf("unsupported mode %q", mode)
	}
}

func ValidateStatus(status string) error {
	switch status {
	case StatusOK, StatusNoResult, StatusAuthRequired, StatusError:
		return nil
	default:
		return fmt.Errorf("unsupported status %q", status)
	}
}

func (r RouterResult) Valid(threshold float64) error {
	if err := ValidateMode(r.Mode); err != nil {
		return err
	}
	if r.Confidence < 0 || r.Confidence > 1 {
		return fmt.Errorf("confidence %.2f outside 0..1", r.Confidence)
	}
	if r.Confidence < threshold {
		return fmt.Errorf("confidence %.2f below threshold %.2f", r.Confidence, threshold)
	}
	if r.NormalizedQuestion == "" {
		return errors.New("normalized_question is required")
	}
	return nil
}
