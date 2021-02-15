package taskManager

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	cookiejar "github.com/juju/persistent-cookiejar"
	"net/http"
	"reflect"

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

	task.Token = data.Token
	return nil
}

func (tm *remoteTaskManager) Fetch(tp string) (*Task, error) {
	url := fmt.Sprintf("/task")

	var data []models.Task

	resp, err := tm.client.R().
		SetQueryParam("type", tp).
		SetResult(&data).
		Get(url)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, errors.New("request failed")
	}

	if len(data) == 0 {
		return nil, nil
	}

	cur := data[0]

	task := &Task{
		Task:  cur,
		Token: "",
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

	volume, err := tm.CreateVolume()
	if err != nil {
		return err
	}

	for i, _ := range task.Outputs {
		slot := task.Outputs[i]
		if slot.Value == nil {
			continue
		}
		if outputBytes, ok := slot.Value.([]byte); ok {
			filename := fmt.Sprintf("%d", i)
			if err := tm.CreateFile(volume.Name, filename, outputBytes); err != nil {
				return err
			}
			slot.Type = "file"
			slot.Value = fmt.Sprintf("%s:%s", volume.Name, filename)
		} else {
			slot.Type = reflect.TypeOf(slot.Value).String()
		}
	}

	resp, err := tm.client.R().
		EnableTrace().
		SetPathParams(map[string]string{
			"taskId": task.TaskId,
		}).
		SetBody(struct {
			Token   string `json:"token"`
			Warning string `json:"warning"`
			Error   string `json:"error"`

			Outputs models.Slots `json:"outputs"`
		}{
			task.Token,
			warning,
			error,
			task.Outputs,
		}).
		Put("/task/{taskId}")

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
