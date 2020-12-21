package main

import (
	"io/ioutil"
	"log"
	"os"
	"path"
	"time"

	"github.com/infinity-oj/actuator/internal/taskManager"
)

func isExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func work(env Runtime, task *taskManager.Task) (warning, error string) {
	vol := task.Properties["volume"]
	log.Printf("Download volume: %s", vol)

	workingDir, err := env.GetVolume(vol)
	if err != nil {
		log.Fatal(err)
	}

	outputPath := path.Join(workingDir, "output.txt")
	log.Println("output.txt from", outputPath)

	output, err := ioutil.ReadFile(outputPath)

	task.Outputs = [][]byte{
		{},
		{},
		output,
	}

	return "", ""
}

var worker = "executor/cs303"
var timeout = time.Minute * 20

func main() {
	tm := taskManager.NewRemoteTaskManager("http://10.20.107.171:2333")
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

		env := NewRuntime()
		if err := env.Setup(task); err != nil {
			log.Printf("setup env error %s", err)
			continue
		}
		log.Printf("env setup")

		ch := make(chan int, 0)
		var warningMsg, errorMsg string
		go func() {
			warningMsg, errorMsg = work(env, task)
			ch <- 1
		}()
		select {
		case <-ch:
			log.Printf("done")
		case <-time.After(timeout):
			errorMsg = "timeout after " + timeout.String()
		}

		err = tm.Push(task, warningMsg, errorMsg)
		if err != nil {
			log.Printf("return task error: %s", err.Error())
		}
		log.Printf("task returned")

		env.TearDown()
	}
}
