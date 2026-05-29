package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"

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
	parts := strings.Fields(c.command)
	if len(parts) == 0 {
		return contracts.LookupResponse{}, errors.New("lookup worker command is empty")
	}

	runCtx := ctx
	cancel := func() {}
	if timeout > 0 {
		runCtx, cancel = context.WithTimeout(ctx, timeout)
	}
	defer cancel()

	payload, err := json.Marshal(request)
	if err != nil {
		return contracts.LookupResponse{}, err
	}
	cmd := exec.CommandContext(runCtx, parts[0], parts[1:]...)
	cmd.Stdin = bytes.NewReader(append(payload, '\n'))
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if runCtx.Err() != nil {
			return contracts.LookupResponse{}, fmt.Errorf("lookup worker timeout: %w", runCtx.Err())
		}
		return contracts.LookupResponse{}, fmt.Errorf("lookup worker failed: %w: %s", err, strings.TrimSpace(stderr.String()))
	}

	response, err := decodeSingleResponse(stdout.Bytes())
	if err != nil {
		return contracts.LookupResponse{}, err
	}
	if err := contracts.ValidateStatus(response.Status); err != nil {
		return contracts.LookupResponse{}, err
	}
	return response, nil
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
