package cache

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"interview-chatbot/internal/contracts"

	_ "modernc.org/sqlite"
)

const StateMiss = "miss"
const StateFresh = "fresh"
const StateStale = "stale"

type Store struct {
	db *sql.DB
}

type Entry struct {
	State              string
	Key                string
	Mode               string
	NormalizedQuestion string
	CacheHint          *contracts.CacheHint
	Response           contracts.LookupResponse
	ExpiresAt          time.Time
}

func Open(ctx context.Context, path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	store := &Store{db: db}
	if err := store.Init(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return store, nil
}

func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *Store) Init(ctx context.Context) error {
	if s == nil || s.db == nil {
		return errors.New("nil cache store")
	}
	if _, err := s.db.ExecContext(ctx, `PRAGMA journal_mode=WAL`); err != nil {
		return err
	}
	if _, err := s.db.ExecContext(ctx, `PRAGMA busy_timeout=5000`); err != nil {
		return err
	}
	_, err := s.db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS lookup_cache (
		cache_key TEXT PRIMARY KEY,
		mode TEXT NOT NULL,
		normalized_question TEXT NOT NULL,
		cache_hint_json TEXT NOT NULL,
		status TEXT NOT NULL,
		answer TEXT NOT NULL,
		evidence_json TEXT NOT NULL,
		observed_at TEXT NOT NULL,
		expires_at TEXT NOT NULL,
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL
	)`)
	return err
}

func KeyFromRouter(result contracts.RouterResult) string {
	mode := strings.TrimSpace(result.Mode)
	hint := result.CacheHint
	if hint != nil {
		lookupType := normalizePart(hint.LookupType)
		target := normalizePart(hint.Target)
		site := normalizePart(hint.Site)
		rawURL := strings.TrimSpace(hint.URL)
		switch {
		case rawURL != "":
			return mode + ":url:" + rawURL
		case lookupType != "" && site != "" && target != "":
			return mode + ":" + lookupType + ":" + site + ":" + target
		case lookupType != "" && target != "":
			return mode + ":" + lookupType + ":" + target
		case site != "" && target != "":
			return mode + ":" + site + ":" + target
		case target != "":
			return mode + ":target:" + target
		}
	}
	return mode + ":question:" + normalizePart(result.NormalizedQuestion)
}

func (s *Store) Get(ctx context.Context, key string, now time.Time) (Entry, error) {
	row := s.db.QueryRowContext(ctx, `SELECT mode, normalized_question, cache_hint_json, status, answer, evidence_json, observed_at, expires_at
		FROM lookup_cache WHERE cache_key = ?`, key)
	var mode, normalizedQuestion, cacheHintJSON, status, answer, evidenceJSON, observedAt, expiresAtRaw string
	if err := row.Scan(&mode, &normalizedQuestion, &cacheHintJSON, &status, &answer, &evidenceJSON, &observedAt, &expiresAtRaw); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Entry{State: StateMiss, Key: key}, nil
		}
		return Entry{}, err
	}

	var hint *contracts.CacheHint
	if cacheHintJSON != "" && cacheHintJSON != "null" {
		var parsed contracts.CacheHint
		if err := json.Unmarshal([]byte(cacheHintJSON), &parsed); err != nil {
			return Entry{}, err
		}
		hint = &parsed
	}
	var evidence []contracts.Evidence
	if evidenceJSON != "" {
		if err := json.Unmarshal([]byte(evidenceJSON), &evidence); err != nil {
			return Entry{}, err
		}
	}
	expiresAt, err := time.Parse(time.RFC3339Nano, expiresAtRaw)
	if err != nil {
		return Entry{}, err
	}

	state := StateFresh
	if !expiresAt.After(now) {
		state = StateStale
	}
	return Entry{
		State:              state,
		Key:                key,
		Mode:               mode,
		NormalizedQuestion: normalizedQuestion,
		CacheHint:          hint,
		Response: contracts.LookupResponse{
			Status:     status,
			Answer:     answer,
			Evidence:   evidence,
			ObservedAt: observedAt,
		},
		ExpiresAt: expiresAt,
	}, nil
}

func (s *Store) Put(ctx context.Context, key string, result contracts.RouterResult, response contracts.LookupResponse, ttl time.Duration, now time.Time) error {
	if ttl <= 0 {
		return nil
	}
	if err := contracts.ValidateStatus(response.Status); err != nil {
		return err
	}
	cacheHintJSON, err := json.Marshal(result.CacheHint)
	if err != nil {
		return err
	}
	evidenceJSON, err := json.Marshal(response.Evidence)
	if err != nil {
		return err
	}
	observedAt := response.ObservedAt
	if observedAt == "" {
		observedAt = now.Format(time.RFC3339)
	}
	expiresAt := now.Add(ttl)

	_, err = s.db.ExecContext(ctx, `INSERT INTO lookup_cache (
		cache_key, mode, normalized_question, cache_hint_json, status, answer, evidence_json, observed_at, expires_at, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(cache_key) DO UPDATE SET
		mode = excluded.mode,
		normalized_question = excluded.normalized_question,
		cache_hint_json = excluded.cache_hint_json,
		status = excluded.status,
		answer = excluded.answer,
		evidence_json = excluded.evidence_json,
		observed_at = excluded.observed_at,
		expires_at = excluded.expires_at,
		updated_at = excluded.updated_at`,
		key,
		result.Mode,
		result.NormalizedQuestion,
		string(cacheHintJSON),
		response.Status,
		response.Answer,
		string(evidenceJSON),
		observedAt,
		expiresAt.Format(time.RFC3339Nano),
		now.Format(time.RFC3339Nano),
		now.Format(time.RFC3339Nano),
	)
	return err
}

func TTLFor(result contracts.RouterResult, response contracts.LookupResponse) time.Duration {
	switch response.Status {
	case contracts.StatusError:
		return 0
	case contracts.StatusNoResult:
		return 5 * time.Minute
	case contracts.StatusAuthRequired:
		return 1 * time.Minute
	}

	if result.CacheHint != nil && strings.EqualFold(result.CacheHint.LookupType, "price") {
		return 30 * time.Minute
	}
	if result.Mode == contracts.ModePublicLookup || result.Mode == contracts.ModeBrowserLookup {
		return 60 * time.Minute
	}
	return 0
}

func ShouldForceRefresh(question string, routerResult contracts.RouterResult) bool {
	if routerResult.ForceRefresh {
		return true
	}
	lower := strings.ToLower(question)
	for _, marker := range []string{"重新查", "更新", "refresh", "不要快取"} {
		if strings.Contains(lower, strings.ToLower(marker)) {
			return true
		}
	}
	return false
}

func normalizePart(value string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
}

func DebugKey(result contracts.RouterResult) string {
	return fmt.Sprintf("%s -> %s", result.NormalizedQuestion, KeyFromRouter(result))
}
