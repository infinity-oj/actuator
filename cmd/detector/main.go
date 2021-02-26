package main

import (
	"github.com/infinity-oj/actuator/internal/volume"
	"github.com/infinity-oj/server-v2/pkg/models"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/infinity-oj/actuator/internal/taskManager"
)

var worker = "basic/detector"

func isExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func main() {
	tm := taskManager.NewRemoteTaskManager("http://127.0.0.1:8888")
	for {
		time.Sleep(time.Second)

		task, err := tm.Fetch(worker)
		if err != nil {
			log.Printf("fetch task error %s", err)
			continue
		}
		if task == nil {
			continue
		}

		log.Printf("receive task: %s", task.TaskId)
		log.Printf("from judgement: %s", task.JudgementId)

		err = tm.Reserve(task)
		if err != nil {
			log.Printf("lock task error %s", err)
			continue
		}
		log.Printf("task locked")

		volumeFD := task.Inputs[0].Value.(string)
		volumeName := strings.ReplaceAll(volumeFD, ":/", "")

		p, err := ioutil.TempDir("", volumeName)
		if err != nil {
			log.Println(err)
			continue
		}
		err = volume.DownloadVolume(volumeName, p)
		if err != nil {
			log.Println(err)
			continue
		}
		p = path.Join(p, volumeName)

		log.Println("path:", p)
		for _, fun := range []nmd{tryCpp, tryJava} {
			if fun(p, task) {
				err = tm.Push(task, "", "")
				log.Println(task.Outputs)
				if err != nil {
					log.Fatal(err)
				}
				break
			}
		}
		log.Println("qwq")
	}
}

type nmd func(path string, task *taskManager.Task) bool

func readFile(path string) (string, error) {
	fileBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(fileBytes), nil
}

func tryCpp(path string, task *taskManager.Task) bool {
	path = filepath.Join(path, "main.cpp")

	str, err := readFile(path)
	if err != nil {
		return false
	}

	if isExists(path) {
		task.Outputs = models.Slots{
			{
				// code
				Type:  "string",
				Value: str,
			},
			{
				// language
				Type:  "string",
				Value: "cpp",
			},
			{
				// memory limit
				Type:  "number",
				Value: 128,
			},
			{
				// time limit
				Type:  "number",
				Value: 3,
			},
		}
		return true
	}
	return false
}

func tryJava(path string, task *taskManager.Task) bool {
	path = filepath.Join(path, "Main.java")

	str, err := readFile(path)
	if err != nil {
		return false
	}

	if isExists(path) {
		task.Outputs = models.Slots{
			{
				// code
				Type:  "string",
				Value: str,
			},
			{
				// language
				Type:  "string",
				Value: "java",
			},
			{
				// memory limit
				Type:  "number",
				Value: 128,
			},
			{
				// time limit
				Type:  "number",
				Value: 3,
			},
		}
		return true
	}
	return false
}
