package environment

import (
	"io/ioutil"
	"os"
	"path"

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
}

type runtime struct {
	WorkingDir string

	volumeMap map[string]string
}

func (e *runtime) GetVolume(volumeName string) (string, error) {
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

func (e *runtime) Setup(task *taskManager.Task) (err error) {
	e.WorkingDir, err = ioutil.TempDir("", "")
	if err != nil {
		e.TearDown()
		return
	}

	return nil
}

func (e runtime) TearDown() {
	for _, v := range e.volumeMap {
		os.Remove(v)
	}
	e.volumeMap = make(map[string]string)
}

func (e runtime) ReadFile(path string) {
	panic("implement me")
}

func (e runtime) WriteFile() {
	panic("implement me")
}

func (e runtime) NewCommand() {
	panic("implement me")
}

func New() Runtime {
	return &runtime{
		WorkingDir: "",
		volumeMap:  make(map[string]string),
	}
}
