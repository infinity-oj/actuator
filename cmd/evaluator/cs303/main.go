package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"

	"log"

	"github.com/infinity-oj/actuator/internal/environment"

	"github.com/infinity-oj/actuator/internal/taskManager"
)

//
//func work(taskManager taskManager.TaskManager) {
//
//	task, err := taskManager.Fetch("evaluator/cs303")
//
//	if task == nil {
//		return
//	}
//	fmt.Println(task.TaskId)
//	fmt.Println(err)
//
//	err = taskManager.Reserve(task)
//	if err != nil {
//		return
//	}
//	//
//	log.Printf("Get task, task id: %s", task.TaskId)
//
//	vol := task.Inputs[0]
//
//	err = volume.DownloadVolume(vol, "./nb")
//
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	mainPath := path.Join(".", "nb", vol)
//
//	fmt.Println(mainPath)
//
//	task.Outputs = [][]byte{
//		[]byte("66"),
//	}
//
//	err = taskManager.Push(task)
//	if err != nil {
//		log.Fatal(err)
//	}
//}

func work(env environment.Runtime, task *taskManager.Task) (warning, error string) {
	vol := task.Inputs[0]

	p, err := env.GetVolume(vol)
	if err != nil {
		log.Fatal(err)
	}

	outputPath := path.Join(p, "2")

	log.Printf("output.txt should at %s", outputPath)

	output, err := ioutil.ReadFile(outputPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", "output.txt not found"
		}
		log.Fatal(err)
	}
	str := string(output)
	outs := strings.Split(str, "\n")
	log.Printf("get %d lines", len(outs))

	if len(outs) < 25000 || len(outs) > 25001 {
		return "", fmt.Sprintf("output %d lines", len(outs))
	}

	std, err := ioutil.ReadFile("ans.json")
	var arr []int
	_ = json.Unmarshal(std, &arr)

	correct := 0
	all := 25000

	for i := 0; i < all; i += 1 {
		if fmt.Sprintf("%d", arr[i]) == strings.TrimSpace(outs[i]) {
			correct += 1
		}
	}
	fmt.Println(correct)

	score := float64(correct) / float64(all)
	log.Printf("score: %f", score)

	task.Outputs = [][]byte{
		[]byte(fmt.Sprintf("%f", score)),
	}
	return "", ""
}

var worker = "evaluator/cs303"
var timeout = time.Second * 5

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

		err = tm.Reserve(task)
		if err != nil {
			log.Printf("lock task error %s", err)
			continue
		}
		log.Printf("task locked")

		env := environment.New()
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

		env.TearDown()
	}
}
