package workflow

import (
	"fmt"
)

func Validate(w Workflow) error {
	if len(w.Tasks) == 0 { //empty workflow
		return fmt.Errorf("workflow must contain at least one task")

	}

	taskIDs := make(map[string]bool)

	for _, task := range w.Tasks {
		if task.ID == "" {
			return fmt.Errorf("task ID cannot be empty")

		}

		if taskIDs[task.ID] {
			return fmt.Errorf("duplicate task ID: %s", task.ID) //duplicate IDs

		}

		taskIDs[task.ID] = true
	}

	for _, task := range w.Tasks {
		for _, dependency := range task.Dependencies {
			if dependency == task.ID {
				return fmt.Errorf(
					"task %s cannot depend on itself",
					task.ID,
				)
			}

			if !taskIDs[dependency] {
				return fmt.Errorf(
					"task %s depends on unknown task %s",
					task.ID,
					dependency,
				)
			}
		}
	}

	if hasCycle(w.Tasks) {
		return fmt.Errorf("workflow contains a cycle")

	}

	return nil

}

func hasCycle(tasks []Task) bool {
	inDegree := make(map[string]int)
	graph := make(map[string][]string)

	for _, task := range tasks {
		inDegree[task.ID] = 0
	}

	for _, task := range tasks {
		for _, dependency := range task.Dependencies {
			graph[dependency] = append(graph[dependency], task.ID)
			inDegree[task.ID]++
		}
	}

	var queue []string

	for taskID, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, taskID)
		}
	}

	visited := 0

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		visited++

		for _, next := range graph[current] {
			inDegree[next]--

			if inDegree[next] == 0 {
				queue = append(queue, next)
			}
		}
	}

	return visited != len(tasks)
}
