package taskManager

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	cookiejar "github.com/juju/persistent-cookiejar"

	"github.com/go-resty/resty/v2"
	"github.com/infinity-oj/server-v2/pkg/models"
)

type remoteTaskManager struct {
	client  *resty.Client
	baseUrl string
}

func (tm *remoteTaskManager) Reserve(task *Task) error {
	url := fmt.Sprintf("/task/%s/reservation", task.TaskId)

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

	task.token = data.Token
	return nil
}

func (tm *remoteTaskManager) Fetch(tp string) (*Task, error) {
	url := fmt.Sprintf("/task")

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

	task := &Task{
		JudgementId: tmp.JudgementId,
		TaskId:      tmp.TaskId,
		token:       "",
		Type:        tmp.Type,

		Properties: properties,
		Inputs:     strings.Split(tmp.Inputs, ","),
		Outputs:    [][]byte{},
	}
	return task, nil
}

func (tm *remoteTaskManager) CreateFile(volume, filename string, file []byte) error {

	url := fmt.Sprintf("/volume/%s/file", volume)
	resp, err := tm.client.R().
		SetFileReader(
			"file", filename, bytes.NewReader(file)).
		Post(url)

	if err != nil {
		return err
	}

	// Explore response object
	fmt.Println("Response Info:")
	fmt.Println("  ", resp.Request.URL)
	fmt.Println("  Error      :", err)
	fmt.Println("  Status Code:", resp.StatusCode())
	fmt.Println("  Status     :", resp.Status())
	fmt.Println("  Proto      :", resp.Proto())
	fmt.Println("  Time       :", resp.Time())
	fmt.Println("  Received At:", resp.ReceivedAt())
	fmt.Println("  Body       :\n", resp)
	fmt.Println()

	return nil
}

func (tm *remoteTaskManager) CreateVolume() (*models.Volume, error) {
	volume := &models.Volume{}

	resp, err := tm.client.R().
		SetResult(volume).
		Post("/volume")
	if err != nil {
		return nil, err
	}

	// Explore response object
	fmt.Println("Response Info:")
	fmt.Println("  ", resp.Request.URL)
	fmt.Println("  Error      :", err)
	fmt.Println("  Status Code:", resp.StatusCode())
	fmt.Println("  Status     :", resp.Status())
	fmt.Println("  Proto      :", resp.Proto())
	fmt.Println("  Time       :", resp.Time())
	fmt.Println("  Received At:", resp.ReceivedAt())
	fmt.Println("  Body       :\n", resp)
	fmt.Println()

	return volume, nil
}

func (tm *remoteTaskManager) Login(username, password string) error {
	request := map[string]interface{}{
		"username": username,
		"password": password,
	}

	resp, err := tm.client.R().
		SetBody(request).
		Post("/session/principal")
	if err != nil {
		return err
	}

	// Explore response object
	fmt.Println("Response Info:")
	fmt.Println("  ", resp.Request.URL)
	fmt.Println("  Error      :", err)
	fmt.Println("  Status Code:", resp.StatusCode())
	fmt.Println("  Status     :", resp.Status())
	fmt.Println("  Proto      :", resp.Proto())
	fmt.Println("  Time       :", resp.Time())
	fmt.Println("  Received At:", resp.ReceivedAt())
	fmt.Println("  Body       :\n", resp)
	fmt.Println()

	return nil
}

func (tm *remoteTaskManager) Push(task *Task, warning, error string) (err error) {

	fmt.Println("Error:", error)

	if len(task.Outputs) == 1 {

		_, err = tm.client.R().
			EnableTrace().
			SetBody(struct {
				Token   string `json:"token"`
				Outputs string `json:"outputs"`
				Warning string `json:"warning"`
				Error   string `json:"error"`
			}{
				task.token,
				string(task.Outputs[0]),
				warning,
				error,
			}).
			Put(fmt.Sprintf("/task/%s", task.TaskId))

	} else {

		vol, err := tm.CreateVolume()
		if err != nil {
			return err
		}

		for index, output := range task.Outputs {
			if len(output) == 0 {
				continue
			}
			if err := tm.CreateFile(vol.Name, fmt.Sprintf("%d", index), output); err != nil {
				return err
			}
		}

		_, err = tm.client.R().
			EnableTrace().
			SetBody(struct {
				Token   string `json:"token"`
				Outputs string `json:"outputs"`
				Warning string `json:"warning"`
				Error   string `json:"error"`
			}{
				task.token,
				vol.Name,
				warning,
				error,
			}).
			Put(fmt.Sprintf("/task/%s", task.TaskId))

	}

	return
}

func NewRemoteTaskManager(baseUrl string) TaskManager {
	client := resty.New()
	client.SetHostURL(fmt.Sprintf("%s/api/v1", baseUrl))
	jar, _ := cookiejar.New(nil)
	client.SetCookieJar(jar)

	tm := &remoteTaskManager{
		client:  client,
		baseUrl: baseUrl,
	}

	err := tm.Login("judger1", "judger1")
	if err != nil {
		panic(err)
	}

	return tm
}
