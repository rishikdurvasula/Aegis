package persistence

import (
	"context"
	"fmt"

	"github.com/rishikdurvasula/aegis/internal/workflow"
)

// MarkReadyTasks finds PENDING tasks whose dependencies have all completed
// and changes them to READY.
func (db *DB) MarkReadyTasks(ctx context.Context, workflowID string) error {
	_, err := db.Pool.Exec(
		ctx,
		`
		UPDATE tasks t
		SET status = $1
		WHERE t.workflow_id = $2
		  AND t.status = $3
		  AND NOT EXISTS (
			  SELECT 1
			  FROM task_dependencies d
			  JOIN tasks dependency
			    ON dependency.workflow_id = d.workflow_id
			   AND dependency.id = d.depends_on_task_id
			  WHERE d.workflow_id = t.workflow_id
			    AND d.task_id = t.id
			    AND dependency.status != $4
		  )
		`,
		workflow.TaskReady,
		workflowID,
		workflow.TaskPending,
		workflow.TaskCompleted,
	)

	if err != nil {
		return fmt.Errorf("mark ready tasks: %w", err)
	}

	return nil
}

// GetReadyTasks returns tasks that are ready to execute.
func (db *DB) GetReadyTasks(
	ctx context.Context,
	workflowID string,
) ([]workflow.Task, error) {

	rows, err := db.Pool.Query(
		ctx,
		`
		SELECT
			id,
			workflow_id,
			name,
			task_type,
			payload,
			status,
			created_at,
			started_at,
			completed_at
		FROM tasks
		WHERE workflow_id = $1
		  AND status = $2
		ORDER BY created_at, id
		`,
		workflowID,
		workflow.TaskReady,
	)

	if err != nil {
		return nil, fmt.Errorf("get ready tasks: %w", err)
	}
	defer rows.Close()

	var tasks []workflow.Task

	for rows.Next() {
		var task workflow.Task

		err := rows.Scan(
			&task.ID,
			&task.WorkflowID,
			&task.Name,
			&task.Type,
			&task.Payload,
			&task.Status,
			&task.CreatedAt,
			&task.StartedAt,
			&task.CompletedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("scan ready task: %w", err)
		}

		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate ready tasks: %w", err)
	}

	return tasks, nil
}

// MarkTaskRunning changes READY -> RUNNING.
func (db *DB) MarkTaskRunning(
	ctx context.Context,
	workflowID string,
	taskID string,
) error {

	result, err := db.Pool.Exec(
		ctx,
		`
		UPDATE tasks
		SET status = $1,
		    started_at = NOW()
		WHERE workflow_id = $2
		  AND id = $3
		  AND status = $4
		`,
		workflow.TaskRunning,
		workflowID,
		taskID,
		workflow.TaskReady,
	)

	if err != nil {
		return fmt.Errorf("mark task running: %w", err)
	}

	if result.RowsAffected() != 1 {
		return fmt.Errorf("task %s was not READY", taskID)
	}

	return nil
}

// MarkTaskCompleted changes RUNNING -> COMPLETED.
func (db *DB) MarkTaskCompleted(
	ctx context.Context,
	workflowID string,
	taskID string,
) error {

	result, err := db.Pool.Exec(
		ctx,
		`
		UPDATE tasks
		SET status = $1,
		    completed_at = NOW()
		WHERE workflow_id = $2
		  AND id = $3
		  AND status = $4
		`,
		workflow.TaskCompleted,
		workflowID,
		taskID,
		workflow.TaskRunning,
	)

	if err != nil {
		return fmt.Errorf("mark task completed: %w", err)
	}

	if result.RowsAffected() != 1 {
		return fmt.Errorf("task %s was not RUNNING", taskID)
	}

	return nil
}

// MarkTaskFailed changes RUNNING -> FAILED.
func (db *DB) MarkTaskFailed(
	ctx context.Context,
	workflowID string,
	taskID string,
) error {

	result, err := db.Pool.Exec(
		ctx,
		`
		UPDATE tasks
		SET status = $1,
		    completed_at = NOW()
		WHERE workflow_id = $2
		  AND id = $3
		  AND status = $4
		`,
		workflow.TaskFailed,
		workflowID,
		taskID,
		workflow.TaskRunning,
	)

	if err != nil {
		return fmt.Errorf("mark task failed: %w", err)
	}

	if result.RowsAffected() != 1 {
		return fmt.Errorf("task %s was not RUNNING", taskID)
	}

	return nil
}

func (db *DB) GetTasks(
	ctx context.Context,
	workflowID string,
) ([]workflow.Task, error) {

	rows, err := db.Pool.Query(
		ctx,
		`
		SELECT
			id,
			workflow_id,
			name,
			task_type,
			payload,
			status,
			created_at,
			started_at,
			completed_at
		FROM tasks
		WHERE workflow_id = $1
		ORDER BY created_at, id
		`,
		workflowID,
	)

	if err != nil {
		return nil, fmt.Errorf("get tasks: %w", err)
	}
	defer rows.Close()

	var tasks []workflow.Task

	for rows.Next() {
		var task workflow.Task

		err := rows.Scan(
			&task.ID,
			&task.WorkflowID,
			&task.Name,
			&task.Type,
			&task.Payload,
			&task.Status,
			&task.CreatedAt,
			&task.StartedAt,
			&task.CompletedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("scan task: %w", err)
		}

		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tasks: %w", err)
	}

	return tasks, nil
}

func (db *DB) AllTasksCompleted(
	ctx context.Context,
	workflowID string,
) (bool, error) {

	var incomplete int

	err := db.Pool.QueryRow(
		ctx,
		`
		SELECT COUNT(*)
		FROM tasks
		WHERE workflow_id = $1
		  AND status != $2
		`,
		workflowID,
		workflow.TaskCompleted,
	).Scan(&incomplete)

	if err != nil {
		return false, fmt.Errorf("count incomplete tasks: %w", err)
	}

	return incomplete == 0, nil
}
