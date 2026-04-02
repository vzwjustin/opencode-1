package adhd

import (
	"time"
)

type ThreadType string

const (
	ThreadTypeBug      ThreadType = "bug"
	ThreadTypeFeature  ThreadType = "feature"
	ThreadTypeRefactor ThreadType = "refactor"
	ThreadTypeDebug    ThreadType = "debug"
	ThreadTypeSpike    ThreadType = "spike"
	ThreadTypeAudit    ThreadType = "audit"
	ThreadTypeChore    ThreadType = "chore"
)

type ThreadStatus string

const (
	ThreadStatusActive    ThreadStatus = "active"
	ThreadStatusPaused    ThreadStatus = "paused"
	ThreadStatusCompleted ThreadStatus = "completed"
)

type CodingThread struct {
	ID             string       `json:"id"`
	SessionID      string       `json:"session_id"`
	RawGoal        string       `json:"raw_goal"`
	NarrowedGoal   string       `json:"narrowed_goal"`
	ThreadType     ThreadType   `json:"thread_type"`
	Status         ThreadStatus `json:"status"`
	NextStep       string       `json:"next_step"`
	NextStepReason string       `json:"next_step_reason"`
	Confidence     float64      `json:"confidence"`
	CreatedAt      time.Time    `json:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at"`
}

type Note struct {
	ID        string    `json:"id"`
	ThreadID  string    `json:"thread_id"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
}
