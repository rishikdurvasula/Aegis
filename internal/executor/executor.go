package executor

import (
	"context"
	"fmt"
	"time"

	"github.com/rishikdurvasula/aegis/internal/workflow"
)

type Executor struct{}

func New() *Executor {
	return &Executor{}
}

func (e *Executor) Execute(ctx context.Context, task workflow.Task) error {
	switch task.Type {

	case "echo":
		fmt.Printf("Executing task %s: %s\n", task.ID, task.Name)
		return nil

	case "sleep":
		fmt.Printf("Task %s sleeping...\n", task.ID)

		select {
		case <-time.After(1 * time.Second):
			return nil

		case <-ctx.Done():
			return ctx.Err()
		}

	case "fail":
		return fmt.Errorf("task %s intentionally failed", task.ID)

	default:
		return fmt.Errorf("unknown task type: %s", task.Type)
	}
}
