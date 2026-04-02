package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/opencode-ai/opencode/internal/adhd"
)

// CheckpointTool — saves a progress checkpoint on the current ADHD thread.
type CheckpointTool struct {
	adhd *adhd.Service
}

func NewCheckpointTool(adhdService *adhd.Service) BaseTool {
	return &CheckpointTool{adhd: adhdService}
}

func (t *CheckpointTool) Info() ToolInfo {
	return ToolInfo{
		Name:        "checkpoint",
		Description: "Save a progress checkpoint. Use after meaningful progress on the current task.",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"summary": map[string]any{
					"type":        "string",
					"description": "What was accomplished in this step",
				},
			},
		},
		Required: []string{"summary"},
	}
}

func (t *CheckpointTool) Run(ctx context.Context, call ToolCall) (ToolResponse, error) {
	var input struct {
		Summary string `json:"summary"`
	}
	if err := json.Unmarshal([]byte(call.Input), &input); err != nil {
		return NewTextErrorResponse("Invalid input: " + err.Error()), nil
	}

	sessionID, _ := GetContextValues(ctx)
	thread, err := t.adhd.GetActiveThread(ctx, sessionID)
	if err != nil || thread == nil {
		return NewTextResponse("No active thread — checkpoint noted but not persisted"), nil
	}

	if err := t.adhd.AddCheckpointDB(ctx, thread.ID, input.Summary); err != nil {
		return NewTextErrorResponse("Failed to save checkpoint: " + err.Error()), nil
	}

	return NewTextResponse(fmt.Sprintf("Checkpoint saved: %s", input.Summary)), nil
}

// ParkSideQuestTool — parks a tangential observation without pursuing it.
type ParkSideQuestTool struct {
	adhd *adhd.Service
}

func NewParkSideQuestTool(adhdService *adhd.Service) BaseTool {
	return &ParkSideQuestTool{adhd: adhdService}
}

func (t *ParkSideQuestTool) Info() ToolInfo {
	return ToolInfo{
		Name:        "park_side_quest",
		Description: "Park a side quest — something noticed but NOT the current focus. Do not pursue it, just note it for later.",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"description": map[string]any{
					"type":        "string",
					"description": "What the side quest is about",
				},
			},
		},
		Required: []string{"description"},
	}
}

func (t *ParkSideQuestTool) Run(ctx context.Context, call ToolCall) (ToolResponse, error) {
	var input struct {
		Description string `json:"description"`
	}
	if err := json.Unmarshal([]byte(call.Input), &input); err != nil {
		return NewTextErrorResponse("Invalid input: " + err.Error()), nil
	}

	sessionID, _ := GetContextValues(ctx)
	thread, err := t.adhd.GetActiveThread(ctx, sessionID)
	if err != nil || thread == nil {
		return NewTextResponse("Side quest noted: " + input.Description), nil
	}

	if err := t.adhd.AddSideQuestDB(ctx, thread.ID, input.Description); err != nil {
		return NewTextErrorResponse("Failed to park side quest: " + err.Error()), nil
	}

	return NewTextResponse(fmt.Sprintf("Parked side quest: %s", input.Description)), nil
}

// FlagDriftTool — flags that work is drifting from the main thread goal.
type FlagDriftTool struct {
	adhd *adhd.Service
}

func NewFlagDriftTool(adhdService *adhd.Service) BaseTool {
	return &FlagDriftTool{adhd: adhdService}
}

func (t *FlagDriftTool) Info() ToolInfo {
	return ToolInfo{
		Name:        "flag_drift",
		Description: "Flag that the current work is drifting from the main task goal. Use when you notice scope creep or tangential work.",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"description": map[string]any{
					"type":        "string",
					"description": "How the work is drifting from the main goal",
				},
			},
		},
		Required: []string{"description"},
	}
}

func (t *FlagDriftTool) Run(ctx context.Context, call ToolCall) (ToolResponse, error) {
	var input struct {
		Description string `json:"description"`
	}
	if err := json.Unmarshal([]byte(call.Input), &input); err != nil {
		return NewTextErrorResponse("Invalid input: " + err.Error()), nil
	}

	sessionID, _ := GetContextValues(ctx)
	thread, err := t.adhd.GetActiveThread(ctx, sessionID)
	if err != nil || thread == nil {
		return NewTextResponse("Drift noted: " + input.Description), nil
	}

	if err := t.adhd.AddDriftEventDB(ctx, thread.ID, "scope_drift", input.Description); err != nil {
		return NewTextErrorResponse("Failed to flag drift: " + err.Error()), nil
	}

	return NewTextResponse(fmt.Sprintf("Drift flagged: %s\nReturn to: [%s] %s", input.Description, thread.ThreadType, thread.NarrowedGoal)), nil
}

// ThreadStatusTool — shows the current thread status.
type ThreadStatusTool struct {
	adhd *adhd.Service
}

func NewThreadStatusTool(adhdService *adhd.Service) BaseTool {
	return &ThreadStatusTool{adhd: adhdService}
}

func (t *ThreadStatusTool) Info() ToolInfo {
	return ToolInfo{
		Name:        "thread_status",
		Description: "Get the current ADHD thread status — goal, next step, confidence, checkpoints, side quests.",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{},
		},
	}
}

func (t *ThreadStatusTool) Run(ctx context.Context, call ToolCall) (ToolResponse, error) {
	sessionID, _ := GetContextValues(ctx)
	thread, err := t.adhd.GetActiveThread(ctx, sessionID)
	if err != nil || thread == nil {
		return NewTextResponse("No active ADHD thread for this session."), nil
	}

	summary, err := t.adhd.GetThreadSummary(ctx, thread.ID)
	if err != nil {
		return NewTextErrorResponse("Failed to get thread status: " + err.Error()), nil
	}

	return NewTextResponse(summary), nil
}
