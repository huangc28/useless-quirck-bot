package liveworker

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"interview-chatbot/internal/clirunner"
	"interview-chatbot/internal/contracts"
)

type Config struct {
	CodexCommand      string
	PublicPromptPath  string
	BrowserPromptPath string
	Timeout           time.Duration
}

func Run(ctx context.Context, input io.Reader, output io.Writer, cfg Config) error {
	response := run(ctx, input, cfg)
	encoder := json.NewEncoder(output)
	return encoder.Encode(response)
}

func run(ctx context.Context, input io.Reader, cfg Config) contracts.LookupResponse {
	request, err := decodeSingleRequest(input)
	if err != nil {
		return errorResponse(fmt.Errorf("malformed lookup request: %w", err))
	}
	if err := validateLookupMode(request.Mode); err != nil {
		return errorResponse(err)
	}

	promptPath, err := promptPathForMode(request.Mode, cfg)
	if err != nil {
		return errorResponse(err)
	}
	prompt, err := os.ReadFile(promptPath)
	if err != nil {
		return errorResponse(fmt.Errorf("read prompt: %w", err))
	}
	payload, err := promptPayload(prompt, request)
	if err != nil {
		return errorResponse(err)
	}

	stdout, err := clirunner.Run(ctx, cfg.CodexCommand, payload, cfg.Timeout)
	if err != nil {
		return errorResponse(fmt.Errorf("codex lookup failed: %w", err))
	}
	response, err := decodeSingleResponse(stdout)
	if err != nil {
		return errorResponse(err)
	}
	response, err = normalizeLookupResponse(request, response)
	if err != nil {
		return errorResponse(err)
	}
	return response
}

func decodeSingleRequest(input io.Reader) (contracts.LookupRequest, error) {
	decoder := json.NewDecoder(input)
	var request contracts.LookupRequest
	if err := decoder.Decode(&request); err != nil {
		return contracts.LookupRequest{}, err
	}
	var extra json.RawMessage
	if err := decoder.Decode(&extra); err != io.EOF {
		return contracts.LookupRequest{}, errors.New("lookup worker received more than one JSON value")
	}
	return request, nil
}

func validateLookupMode(mode string) error {
	if err := contracts.ValidateMode(mode); err != nil {
		return err
	}
	if mode == contracts.ModeGeneralChat {
		return errors.New("general_chat is not a lookup worker mode")
	}
	return nil
}

func promptPathForMode(mode string, cfg Config) (string, error) {
	switch mode {
	case contracts.ModePublicLookup:
		if strings.TrimSpace(cfg.PublicPromptPath) == "" {
			return "", errors.New("public lookup prompt path is empty")
		}
		return cfg.PublicPromptPath, nil
	case contracts.ModeBrowserLookup:
		if strings.TrimSpace(cfg.BrowserPromptPath) == "" {
			return "", errors.New("browser lookup prompt path is empty")
		}
		return cfg.BrowserPromptPath, nil
	default:
		return "", fmt.Errorf("unsupported lookup mode %q", mode)
	}
}

func promptPayload(prompt []byte, request contracts.LookupRequest) ([]byte, error) {
	requestJSON, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	var payload bytes.Buffer
	payload.Write(prompt)
	payload.WriteString("\n\nLookup request JSON:\n")
	payload.Write(requestJSON)
	payload.WriteByte('\n')
	return payload.Bytes(), nil
}

func decodeSingleResponse(data []byte) (contracts.LookupResponse, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	var response contracts.LookupResponse
	if err := decoder.Decode(&response); err != nil {
		return contracts.LookupResponse{}, fmt.Errorf("invalid lookup response JSON: %w", err)
	}
	var extra json.RawMessage
	if err := decoder.Decode(&extra); err != io.EOF {
		return contracts.LookupResponse{}, errors.New("lookup command returned more than one JSON value")
	}
	return response, nil
}

func normalizeLookupResponse(request contracts.LookupRequest, response contracts.LookupResponse) (contracts.LookupResponse, error) {
	if err := contracts.ValidateStatus(response.Status); err != nil {
		return contracts.LookupResponse{}, err
	}
	if response.Status == contracts.StatusOK {
		if request.Mode == contracts.ModePublicLookup && !hasUsableEvidence(response.Evidence) {
			return noResultResponse("找不到可靠來源證據；沒有提供猜測答案。", response.ObservedAt), nil
		}
		if strings.TrimSpace(response.Answer) == "" || strings.TrimSpace(response.ObservedAt) == "" || !hasUsableEvidence(response.Evidence) {
			return contracts.LookupResponse{}, errors.New("ok lookup response requires answer, observed_at, and source evidence")
		}
		if request.Mode == contracts.ModePublicLookup && isPriceLookup(request) && !hasPriceCaution(response.Answer) {
			response.Answer = strings.TrimSpace(response.Answer) + " 實際價格請以來源頁面為準。"
		}
	}
	return response, nil
}

func hasUsableEvidence(evidence []contracts.Evidence) bool {
	for _, ev := range evidence {
		hasIdentity := strings.TrimSpace(ev.Source) != "" && strings.TrimSpace(ev.URL) != ""
		hasObservedTime := strings.TrimSpace(ev.ObservedAt) != ""
		hasContext := strings.TrimSpace(ev.Snippet) != "" || strings.TrimSpace(ev.Title) != ""
		if hasIdentity && hasObservedTime && hasContext {
			return true
		}
	}
	return false
}

func errorResponse(err error) contracts.LookupResponse {
	return contracts.LookupResponse{
		Status:     contracts.StatusError,
		ObservedAt: time.Now().UTC().Format(time.RFC3339),
		Error:      err.Error(),
	}
}

func noResultResponse(answer string, observedAt string) contracts.LookupResponse {
	if strings.TrimSpace(observedAt) == "" {
		observedAt = time.Now().UTC().Format(time.RFC3339)
	}
	return contracts.LookupResponse{
		Status:     contracts.StatusNoResult,
		Answer:     answer,
		ObservedAt: observedAt,
	}
}

func isPriceLookup(request contracts.LookupRequest) bool {
	if request.CacheHint != nil && strings.EqualFold(request.CacheHint.LookupType, "price") {
		return true
	}
	question := strings.ToLower(request.Question + " " + request.NormalizedQuestion)
	return strings.Contains(question, "price") || strings.Contains(question, "多少錢") || strings.Contains(question, "價格")
}

func hasPriceCaution(answer string) bool {
	return strings.Contains(answer, "以來源頁面") ||
		strings.Contains(answer, "以來源商品頁") ||
		strings.Contains(answer, "以頁面") ||
		strings.Contains(strings.ToLower(answer), "check")
}
