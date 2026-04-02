package adhd

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// CreateThread inserts a new CodingThread into the database for the given session.
func (s *Service) CreateThread(ctx context.Context, sessionID, rawGoal string) (*CodingThread, error) {
	id := uuid.New().String()
	narrowed := narrowGoal(rawGoal)
	threadType := guessThreadType(rawGoal)
	now := time.Now()

	_, err := s.conn.ExecContext(ctx,
		`INSERT INTO coding_threads (id, session_id, raw_goal, narrowed_goal, thread_type, status, confidence, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		id, sessionID, rawGoal, narrowed, threadType, ThreadStatusActive, 0.5, now.UnixMilli(), now.UnixMilli(),
	)
	if err != nil {
		return nil, err
	}

	return &CodingThread{
		ID:           id,
		SessionID:    sessionID,
		RawGoal:      rawGoal,
		NarrowedGoal: narrowed,
		ThreadType:   threadType,
		Status:       ThreadStatusActive,
		Confidence:   0.5,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

// GetActiveThread returns the most recently updated active thread for a session,
// or nil if none exists.
func (s *Service) GetActiveThread(ctx context.Context, sessionID string) (*CodingThread, error) {
	row := s.conn.QueryRowContext(ctx,
		`SELECT id, session_id, raw_goal, narrowed_goal, thread_type, status, next_step, next_step_reason, confidence, created_at, updated_at
		 FROM coding_threads WHERE session_id = ? AND status = ? ORDER BY updated_at DESC LIMIT 1`,
		sessionID, ThreadStatusActive,
	)

	var t CodingThread
	var nextStep, nextStepReason sql.NullString
	var createdAt, updatedAt int64
	err := row.Scan(&t.ID, &t.SessionID, &t.RawGoal, &t.NarrowedGoal, &t.ThreadType, &t.Status,
		&nextStep, &nextStepReason, &t.Confidence, &createdAt, &updatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	t.NextStep = nextStep.String
	t.NextStepReason = nextStepReason.String
	t.CreatedAt = time.UnixMilli(createdAt)
	t.UpdatedAt = time.UnixMilli(updatedAt)
	return &t, nil
}

// AddCheckpointDB persists a checkpoint to the database for the given thread.
func (s *Service) AddCheckpointDB(ctx context.Context, threadID, summary string) error {
	id := uuid.New().String()
	_, err := s.conn.ExecContext(ctx,
		`INSERT INTO checkpoints (id, thread_id, summary, created_at) VALUES (?, ?, ?, ?)`,
		id, threadID, summary, time.Now().UnixMilli(),
	)
	return err
}

// AddSideQuestDB persists a side quest to the database for the given thread.
func (s *Service) AddSideQuestDB(ctx context.Context, threadID, description string) error {
	id := uuid.New().String()
	_, err := s.conn.ExecContext(ctx,
		`INSERT INTO side_quests (id, thread_id, description, resumed, created_at) VALUES (?, ?, ?, ?, ?)`,
		id, threadID, description, false, time.Now().UnixMilli(),
	)
	return err
}

// AddDriftEventDB persists a drift event to the database for the given thread.
func (s *Service) AddDriftEventDB(ctx context.Context, threadID, signal, description string) error {
	id := uuid.New().String()
	_, err := s.conn.ExecContext(ctx,
		`INSERT INTO drift_events (id, thread_id, signal, description, created_at) VALUES (?, ?, ?, ?, ?)`,
		id, threadID, signal, description, time.Now().UnixMilli(),
	)
	return err
}

// AddNote persists a free-form note for the given thread.
func (s *Service) AddNote(ctx context.Context, threadID, text string) error {
	id := uuid.New().String()
	_, err := s.conn.ExecContext(ctx,
		`INSERT INTO notes (id, thread_id, text, created_at) VALUES (?, ?, ?, ?)`,
		id, threadID, text, time.Now().UnixMilli(),
	)
	return err
}

// UpdateConfidenceDB updates the persisted confidence score for a thread.
func (s *Service) UpdateConfidenceDB(ctx context.Context, threadID string, confidence float64) error {
	_, err := s.conn.ExecContext(ctx,
		`UPDATE coding_threads SET confidence = ?, updated_at = ? WHERE id = ?`,
		confidence, time.Now().UnixMilli(), threadID,
	)
	return err
}

// UpdateNextStep updates the next_step and next_step_reason fields for a thread.
func (s *Service) UpdateNextStep(ctx context.Context, threadID, nextStep, reason string) error {
	_, err := s.conn.ExecContext(ctx,
		`UPDATE coding_threads SET next_step = ?, next_step_reason = ?, updated_at = ? WHERE id = ?`,
		nextStep, reason, time.Now().UnixMilli(), threadID,
	)
	return err
}

// GetThreadSummary returns a human-readable summary of the given thread.
func (s *Service) GetThreadSummary(ctx context.Context, threadID string) (string, error) {
	thread, err := s.getCodingThread(ctx, threadID)
	if err != nil || thread == nil {
		return "No active thread", nil
	}

	var parts []string
	parts = append(parts, fmt.Sprintf("Thread: [%s] %s", thread.ThreadType, thread.NarrowedGoal))
	if thread.NextStep != "" {
		parts = append(parts, fmt.Sprintf("Next step: %s", thread.NextStep))
	}
	parts = append(parts, fmt.Sprintf("Confidence: %.0f%%", thread.Confidence*100))

	// Count checkpoints
	var cpCount int
	s.conn.QueryRowContext(ctx, `SELECT COUNT(*) FROM checkpoints WHERE thread_id = ?`, threadID).Scan(&cpCount) //nolint:errcheck
	parts = append(parts, fmt.Sprintf("Checkpoints: %d", cpCount))

	// Count unresolved side quests
	var sqCount int
	s.conn.QueryRowContext(ctx, `SELECT COUNT(*) FROM side_quests WHERE thread_id = ? AND resumed = 0`, threadID).Scan(&sqCount) //nolint:errcheck
	if sqCount > 0 {
		parts = append(parts, fmt.Sprintf("Parked side quests: %d", sqCount))
	}

	// Count drift events
	var driftCount int
	s.conn.QueryRowContext(ctx, `SELECT COUNT(*) FROM drift_events WHERE thread_id = ?`, threadID).Scan(&driftCount) //nolint:errcheck
	if driftCount > 0 {
		parts = append(parts, fmt.Sprintf("Drift events: %d", driftCount))
	}

	return strings.Join(parts, "\n"), nil
}

// getCodingThread fetches a single CodingThread by its primary key.
func (s *Service) getCodingThread(ctx context.Context, threadID string) (*CodingThread, error) {
	row := s.conn.QueryRowContext(ctx,
		`SELECT id, session_id, raw_goal, narrowed_goal, thread_type, status, next_step, next_step_reason, confidence, created_at, updated_at
		 FROM coding_threads WHERE id = ?`, threadID)

	var t CodingThread
	var nextStep, nextStepReason sql.NullString
	var createdAt, updatedAt int64
	err := row.Scan(&t.ID, &t.SessionID, &t.RawGoal, &t.NarrowedGoal, &t.ThreadType, &t.Status,
		&nextStep, &nextStepReason, &t.Confidence, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}
	t.NextStep = nextStep.String
	t.NextStepReason = nextStepReason.String
	t.CreatedAt = time.UnixMilli(createdAt)
	t.UpdatedAt = time.UnixMilli(updatedAt)
	return &t, nil
}

func narrowGoal(raw string) string {
	if idx := strings.Index(raw, "."); idx > 0 && idx < 120 {
		return strings.TrimSpace(raw[:idx])
	}
	if len(raw) > 120 {
		return raw[:120] + "..."
	}
	return raw
}

func guessThreadType(text string) ThreadType {
	lower := strings.ToLower(text)
	switch {
	case strings.Contains(lower, "bug") || strings.Contains(lower, "fix") || strings.Contains(lower, "broken"):
		return ThreadTypeBug
	case strings.Contains(lower, "debug") || strings.Contains(lower, "trace"):
		return ThreadTypeDebug
	case strings.Contains(lower, "refactor") || strings.Contains(lower, "clean"):
		return ThreadTypeRefactor
	case strings.Contains(lower, "spike") || strings.Contains(lower, "explore"):
		return ThreadTypeSpike
	case strings.Contains(lower, "audit") || strings.Contains(lower, "review"):
		return ThreadTypeAudit
	case strings.Contains(lower, "chore"):
		return ThreadTypeChore
	default:
		return ThreadTypeFeature
	}
}
