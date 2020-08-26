package taskManager

import (
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/infinity-oj/actuator/internal/crypto"
	"time"
)

const (
	URL = "http://127.0.0.1:8888"
)

type Task struct {
	JudgementId string
	TaskId      string
	Token       string

	Type       string
	Properties map[string]string
	Inputs     [][]byte
	Outputs    [][]byte
}

func (task *Task) Reserve() error {
	client := resty.New()

	url := fmt.Sprintf("%s/api/v1/task/%s/reservation",
		URL,
		task.TaskId)

	resp, err := client.R().
		EnableTrace().
		Post(url)

	if err != nil {
		return err
	}

	var data struct {
		Token string `json:"token"`
	}

	if err := json.Unmarshal(resp.Body(), &data); err != nil {
		return err
	}

	task.Token = data.Token
	return nil
}

func Fetch(tp string) (*Task, error) {
	client := resty.New()

	url := fmt.Sprintf("%s/api/v1/task", URL)

	resp, err := client.R().
		EnableTrace().
		SetQueryParam("type", tp).
		Get(url)

	if err != nil {
		return nil, err
	}

	if len(resp.Body()) == 0 {
		return nil, nil
	}

	var data []struct {
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

	if err := json.Unmarshal(resp.Body(), &data); err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, nil
	}

	tmp := data[0]

	properties := make(map[string]string)
	err = json.Unmarshal([]byte(tmp.Properties), &properties)
	if err != nil {
		return nil, err
	}

	inputs, err := crypto.EasyDecode(tmp.Inputs)
	if err != nil {
		return nil, err
	}

	task := &Task{
		JudgementId: tmp.JudgementId,
		TaskId:      tmp.TaskId,
		Token:       "",
		Type:        tmp.Type,

		Properties:  properties,
		Inputs:      inputs,
		Outputs:     [][]byte{},
	}
	return task, nil
}

func (task *Task) Push() error {
	client := resty.New()

	url := fmt.Sprintf("%s/api/v1/task/%s",
		URL,
		task.TaskId)

	_, err := client.R().
		EnableTrace().
		SetBody(struct {
			Token   string `json:"token"`
			Outputs string `json:"outputs"`
		}{
			task.Type,
			crypto.EasyEncode(task.Outputs),
		}).
		Put(url)
	if err != nil {
		return err
	}
	return nil
}
