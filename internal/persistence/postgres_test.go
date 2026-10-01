package persistence

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/rishikdurvasula/aegis/internal/workflow"
)

func setupTestDB(t *testing.T) (*DB, context.Context) {
	t.Helper()

	ctx := context.Background()

	db, err := NewDB(
		ctx,
		"postgres://localhost/aegis?sslmode=disable",
	)
	if err != nil {
		t.Fatalf("connect to database: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	return db, ctx
}

func uniqueWorkflowID(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}

func TestCreateWorkflow(t *testing.T) {
	db, ctx := setupTestDB(t)

	workflowID := uniqueWorkflowID("test-create")

	w := workflow.Workflow{
		ID:     workflowID,
		Name:   "test-workflow",
		Status: workflow.WorkflowPending,

		Tasks: []workflow.Task{
			{
				ID:           "A",
				Name:         "Task A",
				Type:         "echo",
				Status:       workflow.TaskPending,
				Dependencies: []string{},
			},
			{
				ID:           "B",
				Name:         "Task B",
				Type:         "echo",
				Status:       workflow.TaskPending,
				Dependencies: []string{"A"},
			},
		},
	}

	err := db.CreateWorkflow(ctx, w)
	if err != nil {
		t.Fatalf("CreateWorkflow failed: %v", err)
	}

	saved, err := db.GetWorkflow(ctx, workflowID)
	if err != nil {
		t.Fatalf("GetWorkflow failed: %v", err)
	}

	if saved.ID != workflowID {
		t.Fatalf(
			"expected workflow ID %s, got %s",
			workflowID,
			saved.ID,
		)
	}

	if saved.Status != workflow.WorkflowPending {
		t.Fatalf(
			"expected PENDING, got %s",
			saved.Status,
		)
	}

	if len(saved.Tasks) != 2 {
		t.Fatalf(
			"expected 2 tasks, got %d",
			len(saved.Tasks),
		)
	}
}

func TestDependencyResolution(t *testing.T) {
	db, ctx := setupTestDB(t)

	workflowID := uniqueWorkflowID("test-dependencies")

	w := workflow.Workflow{
		ID:     workflowID,
		Name:   "dependency-test",
		Status: workflow.WorkflowPending,

		Tasks: []workflow.Task{
			{
				ID:           "A",
				Name:         "Task A",
				Type:         "echo",
				Status:       workflow.TaskPending,
				Dependencies: []string{},
			},
			{
				ID:           "B",
				Name:         "Task B",
				Type:         "echo",
				Status:       workflow.TaskPending,
				Dependencies: []string{"A"},
			},
		},
	}

	if err := db.CreateWorkflow(ctx, w); err != nil {
		t.Fatalf("CreateWorkflow failed: %v", err)
	}

	// Initially only A should become READY.
	if err := db.MarkReadyTasks(ctx, workflowID); err != nil {
		t.Fatalf("MarkReadyTasks failed: %v", err)
	}

	ready, err := db.GetReadyTasks(ctx, workflowID)
	if err != nil {
		t.Fatalf("GetReadyTasks failed: %v", err)
	}

	if len(ready) != 1 {
		t.Fatalf(
			"expected 1 READY task, got %d",
			len(ready),
		)
	}

	if ready[0].ID != "A" {
		t.Fatalf(
			"expected A to be READY, got %s",
			ready[0].ID,
		)
	}

	// Complete A.
	if err := db.MarkTaskRunning(
		ctx,
		workflowID,
		"A",
	); err != nil {
		t.Fatalf("MarkTaskRunning failed: %v", err)
	}

	if err := db.MarkTaskCompleted(
		ctx,
		workflowID,
		"A",
	); err != nil {
		t.Fatalf("MarkTaskCompleted failed: %v", err)
	}

	// Now B should become READY.
	if err := db.MarkReadyTasks(ctx, workflowID); err != nil {
		t.Fatalf("MarkReadyTasks failed: %v", err)
	}

	ready, err = db.GetReadyTasks(ctx, workflowID)
	if err != nil {
		t.Fatalf("GetReadyTasks failed: %v", err)
	}

	if len(ready) != 1 {
		t.Fatalf(
			"expected 1 READY task after A completed, got %d",
			len(ready),
		)
	}

	if ready[0].ID != "B" {
		t.Fatalf(
			"expected B to be READY, got %s",
			ready[0].ID,
		)
	}
}

func TestInvalidStateTransition(t *testing.T) {
	db, ctx := setupTestDB(t)

	workflowID := uniqueWorkflowID("test-transition")

	w := workflow.Workflow{
		ID:     workflowID,
		Name:   "transition-test",
		Status: workflow.WorkflowPending,

		Tasks: []workflow.Task{
			{
				ID:           "A",
				Name:         "Task A",
				Type:         "echo",
				Status:       workflow.TaskPending,
				Dependencies: []string{},
			},
		},
	}

	if err := db.CreateWorkflow(ctx, w); err != nil {
		t.Fatalf("CreateWorkflow failed: %v", err)
	}

	// A is still PENDING, so PENDING -> RUNNING should be rejected.
	err := db.MarkTaskRunning(
		ctx,
		workflowID,
		"A",
	)

	if err == nil {
		t.Fatal("expected PENDING -> RUNNING transition to fail")
	}
}
