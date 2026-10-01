package persistence

import (
	"context"
	"fmt"

	"github.com/rishikdurvasula/aegis/internal/workflow"
)

func (db *DB) CreateWorkflow(
	ctx context.Context,
	w workflow.Workflow,
) error {

	// Make sure the DAG is valid before storing anything.
	if err := workflow.Validate(w); err != nil {
		return fmt.Errorf("invalid workflow: %w", err)
	}

	// Start a database transaction.
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	// If anything fails before Commit(), undo the transaction.
	defer tx.Rollback(ctx)

	// 1. Insert the workflow.
	_, err = tx.Exec(
		ctx,
		`
		INSERT INTO workflows (id, name, status)
		VALUES ($1, $2, $3)
		`,
		w.ID,
		w.Name,
		w.Status,
	)

	if err != nil {
		return fmt.Errorf("insert workflow: %w", err)
	}

	// 2. Insert every task.
	for _, task := range w.Tasks {
		_, err = tx.Exec(
			ctx,
			`
			INSERT INTO tasks (
				id,
				workflow_id,
				name,
				task_type,
				payload,
				status
			)
			VALUES ($1, $2, $3, $4, $5, $6)
			`,
			task.ID,
			w.ID,
			task.Name,
			task.Type,
			task.Payload,
			task.Status,
		)

		if err != nil {
			return fmt.Errorf(
				"insert task %s: %w",
				task.ID,
				err,
			)
		}
	}

	// 3. Insert task dependencies.
	for _, task := range w.Tasks {
		for _, dependency := range task.Dependencies {
			_, err = tx.Exec(
				ctx,
				`
				INSERT INTO task_dependencies (
					workflow_id,
					task_id,
					depends_on_task_id
				)
				VALUES ($1, $2, $3)
				`,
				w.ID,
				task.ID,
				dependency,
			)

			if err != nil {
				return fmt.Errorf(
					"insert dependency %s -> %s: %w",
					dependency,
					task.ID,
					err,
				)
			}
		}
	}

	// 4. Everything succeeded, so make the transaction permanent.
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit workflow: %w", err)
	}

	return nil
}
func (db *DB) MarkWorkflowRunning(
	ctx context.Context,
	workflowID string,
) error {

	_, err := db.Pool.Exec(
		ctx,
		`
		UPDATE workflows
		SET status = $1,
		    started_at = NOW()
		WHERE id = $2
		`,
		workflow.WorkflowRunning,
		workflowID,
	)

	if err != nil {
		return fmt.Errorf("mark workflow running: %w", err)
	}

	return nil
}

func (db *DB) MarkWorkflowCompleted(
	ctx context.Context,
	workflowID string,
) error {

	_, err := db.Pool.Exec(
		ctx,
		`
		UPDATE workflows
		SET status = $1,
		    completed_at = NOW()
		WHERE id = $2
		`,
		workflow.WorkflowCompleted,
		workflowID,
	)

	if err != nil {
		return fmt.Errorf("mark workflow completed: %w", err)
	}

	return nil
}

func (db *DB) MarkWorkflowFailed(
	ctx context.Context,
	workflowID string,
) error {

	_, err := db.Pool.Exec(
		ctx,
		`
		UPDATE workflows
		SET status = $1,
		    completed_at = NOW()
		WHERE id = $2
		`,
		workflow.WorkflowFailed,
		workflowID,
	)

	if err != nil {
		return fmt.Errorf("mark workflow failed: %w", err)
	}

	return nil
}

func (db *DB) GetWorkflow(
	ctx context.Context,
	workflowID string,
) (*workflow.Workflow, error) {

	var w workflow.Workflow

	err := db.Pool.QueryRow(
		ctx,
		`
		SELECT
			id,
			name,
			status,
			created_at,
			started_at,
			completed_at
		FROM workflows
		WHERE id = $1
		`,
		workflowID,
	).Scan(
		&w.ID,
		&w.Name,
		&w.Status,
		&w.CreatedAt,
		&w.StartedAt,
		&w.CompletedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("get workflow: %w", err)
	}

	tasks, err := db.GetTasks(ctx, workflowID)
	if err != nil {
		return nil, err
	}

	w.Tasks = tasks

	return &w, nil
}
