package taskManager

import (
	"github.com/infinity-oj/server-v2/pkg/models"
)

type Task struct {
	models.Task
	Token string `json:"token"`
}


type TaskManager interface {
	Fetch(tp string) (*Task, error)
	Push(task *Task, warning, error string) error
	Reserve(task *Task) error
}
