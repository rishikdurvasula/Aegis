package main

import (
	"context"
	"fmt"
	"log"

	"github.com/rishikdurvasula/aegis/internal/engine"
	"github.com/rishikdurvasula/aegis/internal/executor"
	"github.com/rishikdurvasula/aegis/internal/persistence"
	"github.com/rishikdurvasula/aegis/internal/workflow"
)

func main() {
	ctx := context.Background()

	databaseURL := "postgres://localhost/aegis?sslmode=disable"

	db, err := persistence.NewDB(ctx, databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	w := workflow.Workflow{
		ID:     "demo-failure",
		Name:   "phase-1-demo",
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
				Type:         "fail",
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
		log.Fatal(err)
	}

	exec := executor.New()

	workflowEngine := engine.New(
		db,
		exec,
	)

	if err := workflowEngine.RunWorkflow(
		ctx,
		w.ID,
	); err != nil {
		log.Fatal(err)
	}

	saved, err := db.GetWorkflow(ctx, w.ID)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println()
	fmt.Println("Final workflow status:", saved.Status)

	for _, task := range saved.Tasks {
		fmt.Printf(
			"%s: %s\n",
			task.ID,
			task.Status,
		)
	}
}
