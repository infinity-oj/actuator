package main

import (
	"fmt"
	"github.com/infinity-oj/server-v2/pkg/models"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"time"

	"github.com/infinity-oj/actuator/internal/volume"

	"github.com/infinity-oj/actuator/internal/taskManager"
)

func work(taskManager taskManager.TaskManager) {

	task, err := taskManager.Fetch("executor/py")

	if task == nil {
		return
	}
	fmt.Println(task.TaskId)
	fmt.Println(err)

	err = taskManager.Reserve(task)
	if err != nil {
		return
	}
	//
	log.Printf("Get task, task id: %s", task.TaskId)

	vol := task.Properties["volume"]

	err = volume.DownloadVolume(vol, "./gg")

	if err != nil {
		log.Fatal(err)
	}
	mainPath := path.Join(".", "gg", vol, "main.py")
	fmt.Println(mainPath)
	cmd := exec.Command("python", mainPath)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}

	stdout, err := ioutil.TempFile("", "stdout-*")
	//stdout, err := os.OpenFile("stdout", os.O_CREATE|os.O_WRONLY, 0777)

	if err != nil {
		log.Fatalln(err)
	}
	//stderr, err := os.OpenFile("stderr", os.O_CREATE|os.O_WRONLY, 0777)
	stderr, err := ioutil.TempFile("", "stderr-*")
	if err != nil {
		log.Fatalln(err)
	}

	cmd.Stdout = stdout
	cmd.Stderr = stderr

	err = cmd.Start()
	if err != nil {
		log.Fatal(err)
	}

	cmd.Wait()
	stdin.Close()
	stdout.Close()
	stderr.Close()

	stdOut, err := ioutil.ReadFile("stdout")
	stdErr, err := ioutil.ReadFile("stderr")
	fmt.Println(string(stdOut))
	fmt.Println(string(stdErr))

	task.Outputs = models.Slots{
		&models.Slot{
			Type:  "",
			Value: string(stdOut),
		},
		&models.Slot{
			Type:  "",
			Value: string(stdErr),
		},
	}

	_ = os.Remove(stdout.Name())
	_ = os.Remove(stderr.Name())

	err = taskManager.Push(task, "", "")
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	tm := taskManager.NewRemoteTaskManager("http://127.0.0.1:8888")
	for {
		work(tm)
		time.Sleep(time.Second)
	}
}
