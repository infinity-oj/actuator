package taskManager

import (
	"errors"
	"fmt"
)

type localTaskManager struct {
	path  string
	queue []*Task
}

func (l *localTaskManager) Fetch(tp string) (*Task, error) {
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

func (l localTaskManager) Push(_ *Task) error {
	return nil
}

func NewLocalTaskManager(queue []*Task) TaskManager {
	return &localTaskManager{
		queue: queue,
	}
}
