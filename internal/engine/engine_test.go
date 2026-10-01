package engine

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/rishikdurvasula/aegis/internal/executor"
	"github.com/rishikdurvasula/aegis/internal/persistence"
	"github.com/rishikdurvasula/aegis/internal/workflow"
)

func setupEngineTest(t *testing.T) (
	*persistence.DB,
	*Engine,
	context.Context,
) {
	t.Helper()

	ctx := context.Background()

	db, err := persistence.NewDB(
		ctx,
		"postgres://localhost/aegis?sslmode=disable",
	)
	if err != nil {
		t.Fatalf("connect to database: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	exec := executor.New()

	e := New(db, exec)

	return db, e, ctx
}

func engineWorkflowID(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}

func TestWorkflowExecution(t *testing.T) {
	db, e, ctx := setupEngineTest(t)

	workflowID := engineWorkflowID("engine-success")

	w := workflow.Workflow{
		ID:     workflowID,
		Name:   "engine-test",
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
			{
				ID:           "C",
				Name:         "Task C",
				Type:         "echo",
				Status:       workflow.TaskPending,
				Dependencies: []string{"A"},
			},
			{
				ID:           "D",
				Name:         "Task D",
				Type:         "echo",
				Status:       workflow.TaskPending,
				Dependencies: []string{"B", "C"},
			},
		},
	}

	if err := db.CreateWorkflow(ctx, w); err != nil {
		t.Fatalf("CreateWorkflow failed: %v", err)
	}

	if err := e.RunWorkflow(ctx, workflowID); err != nil {
		t.Fatalf("RunWorkflow failed: %v", err)
	}

	saved, err := db.GetWorkflow(ctx, workflowID)
	if err != nil {
		t.Fatalf("GetWorkflow failed: %v", err)
	}

	if saved.Status != workflow.WorkflowCompleted {
		t.Fatalf(
			"expected workflow COMPLETED, got %s",
			saved.Status,
		)
	}

	for _, task := range saved.Tasks {
		if task.Status != workflow.TaskCompleted {
			t.Fatalf(
				"expected task %s COMPLETED, got %s",
				task.ID,
				task.Status,
			)
		}
	}
}

func TestWorkflowFailure(t *testing.T) {
	db, e, ctx := setupEngineTest(t)

	workflowID := engineWorkflowID("engine-failure")

	w := workflow.Workflow{
		ID:     workflowID,
		Name:   "failure-test",
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
				Type:         "fail",
				Status:       workflow.TaskPending,
				Dependencies: []string{"A"},
			},
			{
				ID:           "C",
				Name:         "Task C",
				Type:         "echo",
				Status:       workflow.TaskPending,
				Dependencies: []string{"B"},
			},
		},
	}

	if err := db.CreateWorkflow(ctx, w); err != nil {
		t.Fatalf("CreateWorkflow failed: %v", err)
	}

	err := e.RunWorkflow(ctx, workflowID)

	if err == nil {
		t.Fatal("expected workflow execution to fail")
	}

	saved, err := db.GetWorkflow(ctx, workflowID)
	if err != nil {
		t.Fatalf("GetWorkflow failed: %v", err)
	}

	if saved.Status != workflow.WorkflowFailed {
		t.Fatalf(
			"expected workflow FAILED, got %s",
			saved.Status,
		)
	}

	statuses := make(map[string]workflow.TaskStatus)

	for _, task := range saved.Tasks {
		statuses[task.ID] = task.Status
	}

	if statuses["A"] != workflow.TaskCompleted {
		t.Fatalf(
			"expected A COMPLETED, got %s",
			statuses["A"],
		)
	}

	if statuses["B"] != workflow.TaskFailed {
		t.Fatalf(
			"expected B FAILED, got %s",
			statuses["B"],
		)
	}

	if statuses["C"] != workflow.TaskPending {
		t.Fatalf(
			"expected C PENDING, got %s",
			statuses["C"],
		)
	}
}
