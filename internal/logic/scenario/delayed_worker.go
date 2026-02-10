package scenario

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// StartDelayedTaskWorker starts a background goroutine that polls for
// delayed tasks ready to execute. It runs every 30 seconds.
func StartDelayedTaskWorker(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		g.Log().Infof(ctx, "scenario: delayed task worker started (interval: 30s)")

		for {
			select {
			case <-ctx.Done():
				g.Log().Infof(ctx, "scenario: delayed task worker stopped")
				return
			case <-ticker.C:
				processDelayedTasks(ctx)
			}
		}
	}()
}

// processDelayedTasks finds and executes all pending delayed tasks whose execute_at <= now.
func processDelayedTasks(ctx context.Context) {
	tasks, err := GetPendingDelayedTasks(ctx)
	if err != nil {
		g.Log().Errorf(ctx, "scenario: failed to fetch delayed tasks: %v", err)
		return
	}

	if len(tasks) == 0 {
		return
	}

	g.Log().Infof(ctx, "scenario: processing %d delayed tasks", len(tasks))

	for _, task := range tasks {
		// Mark as executed first to prevent double processing
		if err := MarkDelayedTaskExecuted(ctx, task.Id); err != nil {
			g.Log().Errorf(ctx, "scenario: failed to mark task %d as executed: %v", task.Id, err)
			continue
		}

		// Resume the execution from the saved step
		go ResumeExecution(ctx, task.ExecutionId, task.StepId)
	}
}
