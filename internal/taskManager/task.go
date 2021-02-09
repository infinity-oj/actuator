package taskManager

import (
	"github.com/infinity-oj/server-v2/pkg/models"
)

type TaskManager interface {
	Fetch(tp string) (*models.Task, error)
	Push(task *models.Task, warning, error string) error
	Reserve(task *models.Task) error
}
