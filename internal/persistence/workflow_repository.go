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
