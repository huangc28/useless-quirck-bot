package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"interview-chatbot/internal/clirunner"
	"interview-chatbot/internal/contracts"
)

type Client interface {
	Lookup(ctx context.Context, request contracts.LookupRequest, timeout time.Duration) (contracts.LookupResponse, error)
}

type CLIClient struct {
	command string
}

func NewCLIClient(command string) *CLIClient {
	return &CLIClient{command: strings.TrimSpace(command)}
}

func (c *CLIClient) Lookup(ctx context.Context, request contracts.LookupRequest, timeout time.Duration) (contracts.LookupResponse, error) {
	if c == nil || c.command == "" {
		return contracts.LookupResponse{}, errors.New("lookup worker command is empty")
	}
	if err := contracts.ValidateMode(request.Mode); err != nil {
		return contracts.LookupResponse{}, err
	}
	if request.Mode == contracts.ModeGeneralChat {
		return contracts.LookupResponse{}, errors.New("general_chat is not a lookup worker mode")
	}
	payload, err := json.Marshal(request)
	if err != nil {
		return contracts.LookupResponse{}, err
	}
	stdout, err := clirunner.Run(ctx, c.command, append(payload, '\n'), timeout)
	if err != nil {
		return contracts.LookupResponse{}, fmt.Errorf("lookup worker failed: %w", err)
	}

	response, err := decodeSingleResponse(stdout)
	if err != nil {
		return contracts.LookupResponse{}, err
	}
	if err := validateLookupResponse(response); err != nil {
		return contracts.LookupResponse{}, err
	}
	return response, nil
}

func validateLookupResponse(response contracts.LookupResponse) error {
	if err := contracts.ValidateStatus(response.Status); err != nil {
		return err
	}
	if response.Status == contracts.StatusOK {
		if strings.TrimSpace(response.Answer) == "" || strings.TrimSpace(response.ObservedAt) == "" || !hasUsableEvidence(response.Evidence) {
			return errors.New("ok lookup response requires answer, observed_at, and source evidence")
		}
	}
	return nil
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

func decodeSingleResponse(data []byte) (contracts.LookupResponse, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	var response contracts.LookupResponse
	if err := decoder.Decode(&response); err != nil {
		return contracts.LookupResponse{}, fmt.Errorf("invalid lookup worker JSON: %w", err)
	}
	var extra json.RawMessage
	if err := decoder.Decode(&extra); err != io.EOF {
		return contracts.LookupResponse{}, errors.New("lookup worker returned more than one JSON value")
	}
	return response, nil
}
