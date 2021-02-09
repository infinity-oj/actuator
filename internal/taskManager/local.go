package taskManager

import (
	"errors"
	"fmt"
	"github.com/infinity-oj/server-v2/pkg/models"
)

type localTaskManager struct {
	path  string
	queue []*models.Task
}

func (l *localTaskManager) Fetch(tp string) (*models.Task, error) {
	for _, task := range l.queue {
		if task.Type == tp {
			return task, nil
		}
	}
	return nil, errors.New("no task to fetch in local task manager")
}

func (l localTaskManager) Reserve(task *Task) error {
	task.token = fmt.Sprintf("TEST_%s_%s_%s",
		task.JudgementId,
		task.TaskId,
		task.Type,
	)
	return nil
}

func (l localTaskManager) Push(_ *models.Task, a, b string) error {
	return nil
}

func NewLocalTaskManager(queue []*models.Task) TaskManager {
	return &localTaskManager{
		queue: queue,
	}
}
