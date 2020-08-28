package taskManager

import (
	"encoding/json"
	"fmt"

	"github.com/go-resty/resty/v2"

	"github.com/infinity-oj/actuator/internal/crypto"
)



type remoteTaskManager struct {
	client  *resty.Client
	baseUrl string
}

func (tm *remoteTaskManager) Reserve(task *Task) error {
	url := fmt.Sprintf("%s/api/v1/task/%s/reservation",
		tm.baseUrl,
		task.TaskId)

	resp, err := tm.client.R().
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

func (tm *remoteTaskManager) Fetch(tp string) (*Task, error) {
	url := fmt.Sprintf("%s/api/v1/task", tm.baseUrl)

	resp, err := tm.client.R().
		EnableTrace().
		SetQueryParam("type", tp).
		Get(url)

	if err != nil {
		return nil, err
	}

	if len(resp.Body()) == 0 {
		return nil, nil
	}

	var data []TaskResponse

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

		Properties: properties,
		Inputs:     inputs,
		Outputs:    [][]byte{},
	}
	return task, nil
}

func (tm *remoteTaskManager) Push(task *Task) error {
	url := fmt.Sprintf("%s/api/v1/task/%s",
		tm.baseUrl,
		task.TaskId)

	_, err := tm.client.R().
		EnableTrace().
		SetBody(struct {
			Token   string `json:"token"`
			Outputs string `json:"outputs"`
		}{
			task.Token,
			crypto.EasyEncode(task.Outputs),
		}).
		Put(url)
	if err != nil {
		return err
	}
	return nil
}

func NewRemoteManager(baseUrl string) TaskManager {
	return &remoteTaskManager{
		client:  resty.New(),
		baseUrl: baseUrl,
	}
}