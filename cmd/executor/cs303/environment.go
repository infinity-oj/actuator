package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/docker/docker/pkg/stdcopy"

	"github.com/docker/docker/api/types"

	"github.com/docker/docker/api/types/mount"

	"github.com/infinity-oj/actuator/internal/volume"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/infinity-oj/actuator/internal/taskManager"
)

type Runtime interface {
	Setup(task *taskManager.Task) error
	TearDown()

	ReadFile(path string)
	WriteFile()

	GetVolume(volumeName string) (string, error)

	NewCommand()
}

type dockerRuntime struct {
	client *client.Client

	containerId string

	WorkingDir string

	volumeMap map[string]string
}

func (e *dockerRuntime) GetVolume(volumeName string) (string, error) {
	if path, ok := e.volumeMap[volumeName]; ok {
		return path, nil
	}
	p, err := ioutil.TempDir("", volumeName)
	if err != nil {
		return "", err
	}
	err = volume.DownloadVolume(volumeName, p)
	if err != nil {
		return "", err
	}
	p = path.Join(p, volumeName)
	e.volumeMap[volumeName] = p
	return p, nil
}

func copy(src string, dst string) {
	// Read all content of src to data
	data, err := ioutil.ReadFile(src)
	checkErr(err)
	// Write data to dst
	err = ioutil.WriteFile(dst, data, 0755)
	checkErr(err)
}

func (e *dockerRuntime) Setup(task *taskManager.Task) (err error) {

	vol := task.Properties["volume"]
	log.Printf("Download volume: %s", vol)

	workingDir, err := e.GetVolume(vol)
	log.Printf("WorkingDir: %s", workingDir)
	if err != nil {
		log.Fatal(err)
	}
	trainDataPath, _ := filepath.Abs("88266789-2bec-4a57-8028-be5a89350102.json")
	testDataPath, _ := filepath.Abs("5b8ad1e3-abbc-43b1-af6c-f542fded261e.json")
	copy(trainDataPath, path.Join(workingDir, "88266789-2bec-4a57-8028-be5a89350102.json"))
	copy(testDataPath, path.Join(workingDir, "5b8ad1e3-abbc-43b1-af6c-f542fded261e.json"))

	trainDataPath = "88266789-2bec-4a57-8028-be5a89350102.json"
	testDataPath = "5b8ad1e3-abbc-43b1-af6c-f542fded261e.json"

	hostConfig := &container.HostConfig{
		Resources: container.Resources{
			Memory:    34359720776 / 2,
			CPUPeriod: 100000,
			CPUQuota:  500000,
			CPUCount:  8,
		},
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: workingDir,
				Target: "/work",
			},
		},
		MaskedPaths:   nil,
		ReadonlyPaths: nil,
		Init:          nil,
	}

	// cal cmd

	trainTestPath := "train_test.py"
	trainPath := "/work/train.py"
	testPath := "/work/test.py"
	pythonPath := "/root/anaconda3/envs/cs303/bin/python"
	var cmd = ""

	if isExists(path.Join(workingDir, "train_test.py")) {
		cmd += fmt.Sprintf("%s %s -i %s -t %s", pythonPath, trainTestPath, testDataPath, trainDataPath)
	} else {
		useModel := false
		if isExists(path.Join(workingDir, "train.py")) {
			cmd += fmt.Sprintf("%s %s -t %s\n", pythonPath, trainPath, trainDataPath)
			useModel = true
		}
		if isExists(path.Join(workingDir, "test.py")) {
			if useModel {
				cmd += fmt.Sprintf("%s %s -i %s -m model", pythonPath, testPath, testDataPath)
			} else {
				cmd += fmt.Sprintf("%s %s -i %s", pythonPath, testPath, testDataPath)
			}
		} else {
			log.Println("test file not found")
		}
	}

	log.Println("cmd:", cmd)

	log.Println("start.sh: ", path.Join(workingDir, "start.sh"))
	err = ioutil.WriteFile(path.Join(workingDir, "start.sh"), []byte(cmd), 0644)
	if err != nil {
		log.Fatal(err)
	}

	config := &container.Config{
		Cmd:             []string{"bash", "start.sh"},
		Healthcheck:     nil,
		ArgsEscaped:     false,
		Image:           "cs303-proj3",
		Volumes:         nil,
		WorkingDir:      "/work",
		Entrypoint:      nil,
		NetworkDisabled: true,
		MacAddress:      "",
		OnBuild:         nil,
		Labels:          nil,
		StopSignal:      "",
		StopTimeout:     nil,
		Shell:           nil,
	}
	body, err := e.client.ContainerCreate(
		context.Background(),
		config,
		hostConfig,
		nil, nil,
		fmt.Sprintf("cs303-proj3-%s", task.JudgementId),
	)
	if err != nil {
		log.Fatal(err)
	}
	e.containerId = body.ID
	log.Printf("ID: %s\n", e.containerId)

	err = e.client.ContainerStart(
		context.TODO(),
		e.containerId,
		types.ContainerStartOptions{})
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*20)
	defer cancel()

	bodyCh, errCh := e.client.ContainerWait(ctx, e.containerId, container.WaitConditionNotRunning)

	select {
	case err = <-errCh:
		log.Fatal(err)
	case body := <-bodyCh:
		log.Println(body)
	}

	data, err := e.client.ContainerLogs(ctx, e.containerId, types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		log.Fatal(err)
	}
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	_, err = stdcopy.StdCopy(stdout, stderr, data)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(stdout)
	log.Println(stderr)

	e.WorkingDir, err = ioutil.TempDir("", "")
	if err != nil {
		e.TearDown()
		return
	}

	return nil
}

func (e dockerRuntime) TearDown() {
	for _, v := range e.volumeMap {
		os.Remove(v)
	}
	e.volumeMap = make(map[string]string)

	timeout := time.Second * 10
	err := e.client.ContainerStop(context.Background(), e.containerId, &timeout)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Container %s stopped", e.containerId)

	err = e.client.ContainerRemove(context.Background(), e.containerId, types.ContainerRemoveOptions{})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Container %s removed", e.containerId)
}

func (e dockerRuntime) ReadFile(path string) {
	panic("implement me")
}

func (e dockerRuntime) WriteFile() {
	panic("implement me")
}

func (e dockerRuntime) NewCommand() {
	panic("implement me")
}

func NewRuntime() Runtime {
	cli, err := client.NewClient("unix:///var/run/docker.sock", "v1.24", nil, nil)
	if err != nil {
		log.Fatal(err)
	}
	return &dockerRuntime{
		client:     cli,
		WorkingDir: "",
		volumeMap:  make(map[string]string),
	}
}
