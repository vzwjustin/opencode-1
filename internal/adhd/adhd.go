package adhd

import (
	"database/sql"
	"sync"
	"time"
)

// Checkpoint records a named progress point within the current task.
type Checkpoint struct {
	ID        int64
	Label     string
	Note      string
	CreatedAt time.Time
}

// SideQuest holds a tangential idea that was parked instead of pursued.
type SideQuest struct {
	ID        int64
	Summary   string
	CreatedAt time.Time
}

// DriftEvent records a moment the agent noticed it was off-track.
type DriftEvent struct {
	ID          int64
	Description string
	CreatedAt   time.Time
}

// ThreadState holds the current goal, next step, and confidence for the active thread.
type ThreadState struct {
	Goal       string
	NextStep   string
	Confidence int // 0–100
	UpdatedAt  time.Time
}

// Service provides ADHD executive-function helpers to the coding agent.
// All state is in-memory; the sql.DB parameter is reserved for future
// persistence without requiring a schema migration today.
type Service struct {
	mu          sync.Mutex
	conn        *sql.DB // reserved, not yet used
	checkpoints []Checkpoint
	sideQuests  []SideQuest
	driftEvents []DriftEvent
	thread      ThreadState
	nextID      int64
}

// NewService constructs a new ADHD Service backed by the provided connection.
func NewService(conn *sql.DB) *Service {
	return &Service{
		conn: conn,
		thread: ThreadState{
			Goal:       "",
			NextStep:   "",
			Confidence: 100,
			UpdatedAt:  time.Now(),
		},
	}
}

// AddCheckpointInMemory records a progress checkpoint with an optional note (in-memory only).
func (s *Service) AddCheckpointInMemory(label, note string) Checkpoint {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextID++
	cp := Checkpoint{
		ID:        s.nextID,
		Label:     label,
		Note:      note,
		CreatedAt: time.Now(),
	}
	s.checkpoints = append(s.checkpoints, cp)
	return cp
}

// ParkSideQuestInMemory saves a tangential idea in memory and returns it.
func (s *Service) ParkSideQuestInMemory(summary string) SideQuest {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextID++
	sq := SideQuest{
		ID:        s.nextID,
		Summary:   summary,
		CreatedAt: time.Now(),
	}
	s.sideQuests = append(s.sideQuests, sq)
	return sq
}

// RecordDrift logs a drift event and lowers thread confidence by 10 points.
func (s *Service) RecordDrift(description string) DriftEvent {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextID++
	ev := DriftEvent{
		ID:          s.nextID,
		Description: description,
		CreatedAt:   time.Now(),
	}
	s.driftEvents = append(s.driftEvents, ev)
	if s.thread.Confidence > 10 {
		s.thread.Confidence -= 10
	} else {
		s.thread.Confidence = 0
	}
	s.thread.UpdatedAt = time.Now()
	return ev
}

// GetThreadState returns the current thread goal, next step, and confidence.
func (s *Service) GetThreadState() ThreadState {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.thread
}

// SetThreadState updates the active thread goal and next step.
func (s *Service) SetThreadState(goal, nextStep string, confidence int) ThreadState {
	s.mu.Lock()
	defer s.mu.Unlock()
	if confidence < 0 {
		confidence = 0
	}
	if confidence > 100 {
		confidence = 100
	}
	s.thread = ThreadState{
		Goal:       goal,
		NextStep:   nextStep,
		Confidence: confidence,
		UpdatedAt:  time.Now(),
	}
	return s.thread
}

// ListCheckpoints returns all recorded checkpoints in insertion order.
func (s *Service) ListCheckpoints() []Checkpoint {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Checkpoint, len(s.checkpoints))
	copy(out, s.checkpoints)
	return out
}

// ListSideQuests returns all parked side quests in insertion order.
func (s *Service) ListSideQuests() []SideQuest {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]SideQuest, len(s.sideQuests))
	copy(out, s.sideQuests)
	return out
}
