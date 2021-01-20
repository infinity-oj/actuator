package main

import (
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/infinity-oj/actuator/internal/volume"

	"github.com/infinity-oj/actuator/internal/taskManager"
)

type Runtime interface {
	Setup(task *taskManager.Task) error
	TearDown()

	ReadFile(path string)
	WriteFile()

	GetVolume(volumeName string) (string, error)

	NewCommand()

	GetWorkingDir() string

	SetContainer(id string)
	GetContainer() string
}

type dockerRuntime struct {
	containerId string

	WorkingDir string

	volumeMap map[string]string
}

func (e *dockerRuntime) SetContainer(id string) {
	e.containerId = id
}

func (e *dockerRuntime) GetContainer() string {
	return e.containerId
}

func (e *dockerRuntime) GetWorkingDir() string {
	return e.WorkingDir
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
	if err != nil {
		log.Fatal(err)
	}
	// Write data to dst
	err = ioutil.WriteFile(dst, data, 0755)
	if err != nil {
		log.Fatal(err)
	}
}

func (e *dockerRuntime) Setup(task *taskManager.Task) (err error) {

	vol := task.Properties["volume"]
	log.Printf("Download volume: %s", vol)

	workingDir, err := e.GetVolume(vol)
	log.Printf("WorkingDir: %s", workingDir)
	if err != nil {
		return err
	}
	trainDataPath, _ := filepath.Abs("88266789-2bec-4a57-8028-be5a89350102.json")
	testDataPath, _ := filepath.Abs("5b8ad1e3-abbc-43b1-af6c-f542fded261e.json")
	copy(trainDataPath, path.Join(workingDir, "88266789-2bec-4a57-8028-be5a89350102.json"))
	copy(testDataPath, path.Join(workingDir, "5b8ad1e3-abbc-43b1-af6c-f542fded261e.json"))

	e.WorkingDir = workingDir

	return nil
}

func (e dockerRuntime) TearDown() {
	for _, v := range e.volumeMap {
		os.Remove(v)
	}
	e.volumeMap = make(map[string]string)
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
	return &dockerRuntime{
		WorkingDir: "",
		volumeMap:  make(map[string]string),
	}
}
