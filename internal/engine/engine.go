package engine

import (
	"context"
	"fmt"

	"github.com/rishikdurvasula/aegis/internal/executor"
	"github.com/rishikdurvasula/aegis/internal/persistence"
)

type Engine struct {
	db       *persistence.DB
	executor *executor.Executor
}

func New(
	db *persistence.DB,
	exec *executor.Executor,
) *Engine {

	return &Engine{
		db:       db,
		executor: exec,
	}
}

func (e *Engine) RunWorkflow(
	ctx context.Context,
	workflowID string,
) error {

	if err := e.db.MarkWorkflowRunning(ctx, workflowID); err != nil {
		return err
	}

	for {
		// PENDING tasks whose dependencies are complete become READY.
		if err := e.db.MarkReadyTasks(ctx, workflowID); err != nil {
			return err
		}

		// Get runnable tasks.
		tasks, err := e.db.GetReadyTasks(ctx, workflowID)
		if err != nil {
			return err
		}

		// Nothing is currently runnable.
		if len(tasks) == 0 {

			completed, err := e.db.AllTasksCompleted(ctx, workflowID)
			if err != nil {
				return err
			}

			if completed {
				if err := e.db.MarkWorkflowCompleted(
					ctx,
					workflowID,
				); err != nil {
					return err
				}

				fmt.Printf(
					"Workflow %s completed\n",
					workflowID,
				)

				return nil
			}

			return fmt.Errorf(
				"workflow %s has unfinished tasks but no runnable tasks",
				workflowID,
			)
		}

		// Phase 1 intentionally executes sequentially.
		for _, task := range tasks {

			if err := e.db.MarkTaskRunning(
				ctx,
				workflowID,
				task.ID,
			); err != nil {
				return err
			}

			fmt.Printf("Running task: %s\n", task.ID)

			err := e.executor.Execute(ctx, task)

			if err != nil {

				_ = e.db.MarkTaskFailed(
					ctx,
					workflowID,
					task.ID,
				)

				_ = e.db.MarkWorkflowFailed(
					ctx,
					workflowID,
				)

				return fmt.Errorf(
					"task %s failed: %w",
					task.ID,
					err,
				)
			}

			if err := e.db.MarkTaskCompleted(
				ctx,
				workflowID,
				task.ID,
			); err != nil {
				return err
			}

			fmt.Printf("Completed task: %s\n", task.ID)
		}
	}
}
