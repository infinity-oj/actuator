package main

import (
	"fmt"
	"log"
	"path"
	"time"

	"github.com/infinity-oj/actuator/internal/volume"

	"github.com/infinity-oj/actuator/internal/taskManager"
)

func work(taskManager taskManager.TaskManager) {

	task, err := taskManager.Fetch("evaluator/cs303")

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

	vol := task.Inputs[0]

	err = volume.DownloadVolume(vol, "./nb")

	if err != nil {
		log.Fatal(err)
	}

	mainPath := path.Join(".", "nb", vol)

	fmt.Println(mainPath)

	task.Outputs = [][]byte{
		[]byte("66"),
	}

	err = taskManager.Push(task)
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
