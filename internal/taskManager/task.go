package taskManager

import "time"

type Task struct {
	JudgementId string
	TaskId      string
	Token       string

	Type       string
	Properties map[string]string
	Inputs     [][]byte
	Outputs    [][]byte
}

type TaskResponse struct {
	ID        uint64     `json:"id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`

	JudgementId string `json:"judgementId"`

	TaskId string `json:"taskId"`

	Type       string `json:"type"`
	Properties string `json:"properties"`
	Inputs     string `json:"inputs"`
	Outputs    string `json:"outputs"`
}

type TaskManager interface {
	Fetch(tp string) (*Task, error)
	Push(task *Task) error
	Reserve(task *Task) error
}
