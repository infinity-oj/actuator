package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"time"

	"github.com/infinity-oj/actuator/internal/volume"

	"github.com/docker/docker/api/types"
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

func (e *dockerRuntime) Setup(task *taskManager.Task) (err error) {

	config := &container.Config{Image: "cs303-proj3"}

	body, err := e.client.ContainerCreate(
		context.Background(),
		config,
		nil, nil, nil,
		fmt.Sprintf("cs303-proj3-%s", task.JudgementId),
		)
	if err != nil {
		log.Fatal(err)
	}

	e.containerId = body.ID
	log.Printf("ID: %s\n",e.containerId )


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
	if  err != nil {
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
	cli, err := client.NewClient("tcp://localhost:2375", "v1.24", nil, nil)
	if err != nil {
		log.Fatal(err)
	}
	return &dockerRuntime{
		client: cli,
		WorkingDir: "",
		volumeMap:  make(map[string]string),
	}
}