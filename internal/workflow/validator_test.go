package workflow

import "testing"

func TestValidWorkflow(t *testing.T) {
	w := Workflow{
		ID:   "workflow-1",
		Name: "test",
		Tasks: []Task{
			{
				ID:           "A",
				Dependencies: []string{},
			},
			{
				ID:           "B",
				Dependencies: []string{"A"},
			},
			{
				ID:           "C",
				Dependencies: []string{"A"},
			},

			{
				ID:           "D",
				Dependencies: []string{"B", "C"},
			},
		},
	}

	err := Validate(w)

	if err != nil {
		t.Fatalf("expected valid workflow, got error: %v", err)
	}
}

func TestCycleRejected(t *testing.T) {
	w := Workflow{
		ID:   "workflow-1",
		Name: "bad-workflow",
		Tasks: []Task{
			{
				ID:           "A",
				Dependencies: []string{"C"},
			},
			{
				ID:           "B",
				Dependencies: []string{"A"},
			},
			{
				ID:           "C",
				Dependencies: []string{"B"},
			},
		},
	}

	err := Validate(w)

	if err == nil {
		t.Fatal("expected cycle to be rejected")
	}
}

func TestUnknownDependencyRejected(t *testing.T) {
	w := Workflow{
		ID:   "workflow-1",
		Name: "bad-workflow",
		Tasks: []Task{
			{
				ID:           "A",
				Dependencies: []string{"DOES_NOT_EXIST"},
			},
		},
	}

	err := Validate(w)

	if err == nil {
		t.Fatal("expected unknown dependency to be rejected")
	}
}

func TestDuplicateTaskRejected(t *testing.T) {
	w := Workflow{
		ID: "workflow-1",
		Tasks: []Task{
			{ID: "A"},
			{ID: "A"},
		},
	}

	err := Validate(w)

	if err == nil {
		t.Fatal("expected duplicate task ID to be rejected")
	}
}
