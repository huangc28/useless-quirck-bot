package liveworker

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"interview-chatbot/internal/contracts"
)

const testCommandTimeout = 5 * time.Second

func TestRunDispatchesPublicLookupPrompt(t *testing.T) {
	dir := t.TempDir()
	publicPrompt := writeFile(t, dir, "public.md", "PUBLIC LOOKUP PROMPT")
	browserPrompt := writeFile(t, dir, "browser.md", "BROWSER LOOKUP PROMPT")
	capturePath := filepath.Join(dir, "codex-input.txt")
	codex := writeScript(t, dir, "codex.sh", `#!/bin/sh
cat > "$1"
printf '{"status":"ok","answer":"source-grounded answer","observed_at":"2026-05-29T12:00:00Z","evidence":[{"source":"example","url":"https://example.test/item","snippet":"product page context","observed_at":"2026-05-29T12:00:00Z"}]}'`)

	request := `{"mode":"public_lookup","question":"請問義美小泡芙多少錢","normalized_question":"請問義美小泡芙多少錢"}`
	var output bytes.Buffer

	err := Run(context.Background(), strings.NewReader(request), &output, Config{
		CodexCommand:      codex + " " + capturePath,
		PublicPromptPath:  publicPrompt,
		BrowserPromptPath: browserPrompt,
		Timeout:           testCommandTimeout,
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	var response contracts.LookupResponse
	if err := json.Unmarshal(output.Bytes(), &response); err != nil {
		t.Fatalf("output was not one lookup response JSON object: %v: %s", err, output.String())
	}
	if response.Status != contracts.StatusOK {
		t.Fatalf("Status = %s", response.Status)
	}

	captured, err := os.ReadFile(capturePath)
	if err != nil {
		t.Fatalf("read capture: %v", err)
	}
	if !strings.Contains(string(captured), "PUBLIC LOOKUP PROMPT") {
		t.Fatalf("codex input did not include public prompt: %s", captured)
	}
	if strings.Contains(string(captured), "BROWSER LOOKUP PROMPT") {
		t.Fatalf("codex input included browser prompt: %s", captured)
	}
	if !strings.Contains(string(captured), `"mode":"public_lookup"`) {
		t.Fatalf("codex input did not include lookup request JSON: %s", captured)
	}
}

func TestRunDispatchesBrowserLookupPrompt(t *testing.T) {
	dir := t.TempDir()
	publicPrompt := writeFile(t, dir, "public.md", "PUBLIC LOOKUP PROMPT")
	browserPrompt := writeFile(t, dir, "browser.md", "BROWSER LOOKUP PROMPT")
	capturePath := filepath.Join(dir, "codex-input.txt")
	codex := writeScript(t, dir, "codex.sh", `#!/bin/sh
cat > "$1"
printf '{"status":"auth_required","observed_at":"2026-05-29T12:00:00Z","error":"browser session required"}'`)

	var output bytes.Buffer
	err := Run(context.Background(), strings.NewReader(`{"mode":"browser_lookup","question":"請問 momo 義美小泡芙多少錢","normalized_question":"請問 momo 義美小泡芙多少錢","cache_hint":{"lookup_type":"price","target":"義美小泡芙","site":"momo"}}`), &output, Config{
		CodexCommand:      codex + " " + capturePath,
		PublicPromptPath:  publicPrompt,
		BrowserPromptPath: browserPrompt,
		Timeout:           testCommandTimeout,
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	response := decodeLookupResponse(t, output.Bytes())
	if response.Status != contracts.StatusAuthRequired {
		t.Fatalf("Status = %s", response.Status)
	}

	captured, err := os.ReadFile(capturePath)
	if err != nil {
		t.Fatalf("read capture: %v", err)
	}
	if !strings.Contains(string(captured), "BROWSER LOOKUP PROMPT") {
		t.Fatalf("codex input did not include browser prompt: %s", captured)
	}
	if strings.Contains(string(captured), "PUBLIC LOOKUP PROMPT") {
		t.Fatalf("codex input included public prompt: %s", captured)
	}
	if !strings.Contains(string(captured), `"site":"momo"`) {
		t.Fatalf("codex input did not preserve named site: %s", captured)
	}
}

func TestRunMalformedInputWritesErrorResponse(t *testing.T) {
	var output bytes.Buffer

	if err := Run(context.Background(), strings.NewReader(`not-json`), &output, Config{}); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	response := decodeLookupResponse(t, output.Bytes())
	if response.Status != contracts.StatusError {
		t.Fatalf("Status = %s", response.Status)
	}
	if !strings.Contains(response.Error, "malformed lookup request") {
		t.Fatalf("Error = %q", response.Error)
	}
}

func TestRunUnsupportedModeWritesErrorResponse(t *testing.T) {
	var output bytes.Buffer

	if err := Run(context.Background(), strings.NewReader(`{"mode":"unknown","question":"q","normalized_question":"q"}`), &output, Config{}); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	response := decodeLookupResponse(t, output.Bytes())
	if response.Status != contracts.StatusError {
		t.Fatalf("Status = %s", response.Status)
	}
	if !strings.Contains(response.Error, "unsupported mode") {
		t.Fatalf("Error = %q", response.Error)
	}
}

func TestRunCommandFailureWritesErrorResponse(t *testing.T) {
	dir := t.TempDir()
	publicPrompt := writeFile(t, dir, "public.md", "PUBLIC LOOKUP PROMPT")
	browserPrompt := writeFile(t, dir, "browser.md", "BROWSER LOOKUP PROMPT")
	codex := writeScript(t, dir, "codex.sh", "#!/bin/sh\necho boom >&2\nexit 7\n")
	var output bytes.Buffer

	if err := Run(context.Background(), strings.NewReader(`{"mode":"public_lookup","question":"q","normalized_question":"q"}`), &output, Config{
		CodexCommand:      codex,
		PublicPromptPath:  publicPrompt,
		BrowserPromptPath: browserPrompt,
		Timeout:           testCommandTimeout,
	}); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	response := decodeLookupResponse(t, output.Bytes())
	if response.Status != contracts.StatusError {
		t.Fatalf("Status = %s", response.Status)
	}
	if !strings.Contains(response.Error, "codex lookup failed") {
		t.Fatalf("Error = %q", response.Error)
	}
}

func TestRunInvalidLookupOutputWritesErrorResponse(t *testing.T) {
	dir := t.TempDir()
	publicPrompt := writeFile(t, dir, "public.md", "PUBLIC LOOKUP PROMPT")
	browserPrompt := writeFile(t, dir, "browser.md", "BROWSER LOOKUP PROMPT")
	codex := writeScript(t, dir, "codex.sh", "#!/bin/sh\ncat >/dev/null\nprintf 'not-json'\n")
	var output bytes.Buffer

	if err := Run(context.Background(), strings.NewReader(`{"mode":"public_lookup","question":"q","normalized_question":"q"}`), &output, Config{
		CodexCommand:      codex,
		PublicPromptPath:  publicPrompt,
		BrowserPromptPath: browserPrompt,
		Timeout:           testCommandTimeout,
	}); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	response := decodeLookupResponse(t, output.Bytes())
	if response.Status != contracts.StatusError {
		t.Fatalf("Status = %s", response.Status)
	}
	if !strings.Contains(response.Error, "invalid lookup response JSON") {
		t.Fatalf("Error = %q", response.Error)
	}
}

func TestRunTimeoutWritesErrorResponse(t *testing.T) {
	dir := t.TempDir()
	publicPrompt := writeFile(t, dir, "public.md", "PUBLIC LOOKUP PROMPT")
	browserPrompt := writeFile(t, dir, "browser.md", "BROWSER LOOKUP PROMPT")
	codex := writeScript(t, dir, "codex.sh", "#!/bin/sh\nsleep 2\n")
	var output bytes.Buffer

	if err := Run(context.Background(), strings.NewReader(`{"mode":"public_lookup","question":"q","normalized_question":"q"}`), &output, Config{
		CodexCommand:      codex,
		PublicPromptPath:  publicPrompt,
		BrowserPromptPath: browserPrompt,
		Timeout:           10 * time.Millisecond,
	}); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	response := decodeLookupResponse(t, output.Bytes())
	if response.Status != contracts.StatusError {
		t.Fatalf("Status = %s", response.Status)
	}
	if !strings.Contains(response.Error, "timeout") {
		t.Fatalf("Error = %q", response.Error)
	}
}

func TestRunPublicLookupAddsPriceCautionToGroundedAnswer(t *testing.T) {
	dir := t.TempDir()
	publicPrompt := writeFile(t, dir, "public.md", "PUBLIC LOOKUP PROMPT")
	browserPrompt := writeFile(t, dir, "browser.md", "BROWSER LOOKUP PROMPT")
	codex := writeScript(t, dir, "codex.sh", `#!/bin/sh
cat >/dev/null
printf '{"status":"ok","answer":"義美小泡芙約 NT$59","observed_at":"2026-05-29T12:00:00Z","evidence":[{"source":"example","url":"https://example.test/puff","snippet":"product page context","observed_at":"2026-05-29T12:00:00Z"}]}'`)
	var output bytes.Buffer

	if err := Run(context.Background(), strings.NewReader(`{"mode":"public_lookup","question":"請問義美小泡芙多少錢","normalized_question":"請問義美小泡芙多少錢","cache_hint":{"lookup_type":"price","target":"義美小泡芙"}}`), &output, Config{
		CodexCommand:      codex,
		PublicPromptPath:  publicPrompt,
		BrowserPromptPath: browserPrompt,
		Timeout:           testCommandTimeout,
	}); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	response := decodeLookupResponse(t, output.Bytes())
	if response.Status != contracts.StatusOK {
		t.Fatalf("Status = %s", response.Status)
	}
	if !strings.Contains(response.Answer, "實際價格請以來源頁面為準") {
		t.Fatalf("price answer missing caution: %q", response.Answer)
	}
}

func TestRunPublicLookupDoesNotDuplicateExistingPriceCaution(t *testing.T) {
	dir := t.TempDir()
	publicPrompt := writeFile(t, dir, "public.md", "PUBLIC LOOKUP PROMPT")
	browserPrompt := writeFile(t, dir, "browser.md", "BROWSER LOOKUP PROMPT")
	codex := writeScript(t, dir, "codex.sh", `#!/bin/sh
cat >/dev/null
printf '{"status":"ok","answer":"義美小泡芙 57g 每包 NT$32；實際價格請以來源商品頁或結帳頁為準。","observed_at":"2026-05-30T02:14:08+08:00","evidence":[{"source":"義美食品線上購","url":"https://imec.imeifoods.com.tw/products/milkpuff-chocolate","snippet":"頁面顯示售價 NT$32、原價 NT$36。","observed_at":"2026-05-30T02:14:08+08:00"}]}'`)
	var output bytes.Buffer

	if err := Run(context.Background(), strings.NewReader(`{"mode":"public_lookup","question":"義美小泡芙多少錢?","normalized_question":"義美小泡芙多少錢?","cache_hint":{"lookup_type":"price","target":"義美小泡芙"}}`), &output, Config{
		CodexCommand:      codex,
		PublicPromptPath:  publicPrompt,
		BrowserPromptPath: browserPrompt,
		Timeout:           testCommandTimeout,
	}); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	response := decodeLookupResponse(t, output.Bytes())
	if response.Status != contracts.StatusOK {
		t.Fatalf("Status = %s", response.Status)
	}
	if strings.Count(response.Answer, "實際價格") != 1 {
		t.Fatalf("price caution duplicated: %q", response.Answer)
	}
}

func TestRunBrowserLookupGroundedOKPasses(t *testing.T) {
	dir := t.TempDir()
	publicPrompt := writeFile(t, dir, "public.md", "PUBLIC LOOKUP PROMPT")
	browserPrompt := writeFile(t, dir, "browser.md", "BROWSER LOOKUP PROMPT")
	codex := writeScript(t, dir, "codex.sh", `#!/bin/sh
cat >/dev/null
printf '{"status":"ok","answer":"momo 頁面顯示義美小泡芙價格，實際價格請以頁面為準。","observed_at":"2026-05-29T12:00:00Z","evidence":[{"source":"momo","url":"https://www.momoshop.com.tw/","snippet":"商品頁價格區塊","observed_at":"2026-05-29T12:00:00Z"}]}'`)
	var output bytes.Buffer

	if err := Run(context.Background(), strings.NewReader(`{"mode":"browser_lookup","question":"請問 momo 義美小泡芙多少錢","normalized_question":"請問 momo 義美小泡芙多少錢","cache_hint":{"lookup_type":"price","target":"義美小泡芙","site":"momo"}}`), &output, Config{
		CodexCommand:      codex,
		PublicPromptPath:  publicPrompt,
		BrowserPromptPath: browserPrompt,
		Timeout:           testCommandTimeout,
	}); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	response := decodeLookupResponse(t, output.Bytes())
	if response.Status != contracts.StatusOK {
		t.Fatalf("Status = %s; response=%s", response.Status, output.String())
	}
	if len(response.Evidence) == 0 || response.Evidence[0].Source != "momo" {
		t.Fatalf("Evidence = %#v", response.Evidence)
	}
}

func TestRunPublicLookupSourceFreeOKBecomesNoResult(t *testing.T) {
	dir := t.TempDir()
	publicPrompt := writeFile(t, dir, "public.md", "PUBLIC LOOKUP PROMPT")
	browserPrompt := writeFile(t, dir, "browser.md", "BROWSER LOOKUP PROMPT")
	codex := writeScript(t, dir, "codex.sh", `#!/bin/sh
cat >/dev/null
printf '{"status":"ok","answer":"義美小泡芙約 NT$59","observed_at":"2026-05-29T12:00:00Z","evidence":[]}'`)
	var output bytes.Buffer

	if err := Run(context.Background(), strings.NewReader(`{"mode":"public_lookup","question":"請問義美小泡芙多少錢","normalized_question":"請問義美小泡芙多少錢","cache_hint":{"lookup_type":"price","target":"義美小泡芙"}}`), &output, Config{
		CodexCommand:      codex,
		PublicPromptPath:  publicPrompt,
		BrowserPromptPath: browserPrompt,
		Timeout:           testCommandTimeout,
	}); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	response := decodeLookupResponse(t, output.Bytes())
	if response.Status != contracts.StatusNoResult {
		t.Fatalf("Status = %s, want no_result; response=%s", response.Status, output.String())
	}
	if strings.Contains(response.Answer, "NT$59") {
		t.Fatalf("No-Result response repeated unsupported price: %q", response.Answer)
	}
}

func TestRunBrowserLookupSourceFreeOKWritesErrorResponse(t *testing.T) {
	dir := t.TempDir()
	publicPrompt := writeFile(t, dir, "public.md", "PUBLIC LOOKUP PROMPT")
	browserPrompt := writeFile(t, dir, "browser.md", "BROWSER LOOKUP PROMPT")
	codex := writeScript(t, dir, "codex.sh", `#!/bin/sh
cat >/dev/null
printf '{"status":"ok","answer":"momo price is NT$59","observed_at":"2026-05-29T12:00:00Z","evidence":[]}'`)
	var output bytes.Buffer

	if err := Run(context.Background(), strings.NewReader(`{"mode":"browser_lookup","question":"請問 momo 義美小泡芙多少錢","normalized_question":"請問 momo 義美小泡芙多少錢","cache_hint":{"lookup_type":"price","target":"義美小泡芙","site":"momo"}}`), &output, Config{
		CodexCommand:      codex,
		PublicPromptPath:  publicPrompt,
		BrowserPromptPath: browserPrompt,
		Timeout:           testCommandTimeout,
	}); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	response := decodeLookupResponse(t, output.Bytes())
	if response.Status != contracts.StatusError {
		t.Fatalf("Status = %s, want error; response=%s", response.Status, output.String())
	}
	if !strings.Contains(response.Error, "source evidence") {
		t.Fatalf("Error = %q", response.Error)
	}
}

func TestRunBrowserLookupEvidenceRequiresObservedTimeAndContext(t *testing.T) {
	dir := t.TempDir()
	publicPrompt := writeFile(t, dir, "public.md", "PUBLIC LOOKUP PROMPT")
	browserPrompt := writeFile(t, dir, "browser.md", "BROWSER LOOKUP PROMPT")
	codex := writeScript(t, dir, "codex.sh", `#!/bin/sh
cat >/dev/null
printf '{"status":"ok","answer":"momo price is NT$59","observed_at":"2026-05-29T12:00:00Z","evidence":[{"source":"momo","url":"https://www.momoshop.com.tw/"}]}'`)
	var output bytes.Buffer

	if err := Run(context.Background(), strings.NewReader(`{"mode":"browser_lookup","question":"請問 momo 義美小泡芙多少錢","normalized_question":"請問 momo 義美小泡芙多少錢","cache_hint":{"lookup_type":"price","target":"義美小泡芙","site":"momo"}}`), &output, Config{
		CodexCommand:      codex,
		PublicPromptPath:  publicPrompt,
		BrowserPromptPath: browserPrompt,
		Timeout:           testCommandTimeout,
	}); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	response := decodeLookupResponse(t, output.Bytes())
	if response.Status != contracts.StatusError {
		t.Fatalf("Status = %s, want error; response=%s", response.Status, output.String())
	}
	if !strings.Contains(response.Error, "source evidence") {
		t.Fatalf("Error = %q", response.Error)
	}
}

func decodeLookupResponse(t *testing.T, data []byte) contracts.LookupResponse {
	t.Helper()
	decoder := json.NewDecoder(bytes.NewReader(data))
	var response contracts.LookupResponse
	if err := decoder.Decode(&response); err != nil {
		t.Fatalf("decode response: %v: %s", err, string(data))
	}
	var extra json.RawMessage
	if err := decoder.Decode(&extra); err != io.EOF {
		t.Fatalf("output contained more than one JSON value: %s", string(data))
	}
	return response
}

func writeFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
	return path
}

func writeScript(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
	return path
}
