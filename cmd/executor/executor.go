package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/infinity-oj/actuator/internal/taskManager"
)

func work(taskManager taskManager.TaskManager) {

	task, err := taskManager.Fetch("executor/elf")
	if task == nil {
		return
	}
	fmt.Println(task.TaskId)
	fmt.Println(err)

	err = taskManager.Reserve(task)
	if err != nil {
		return
	}

	if err := ioutil.WriteFile("elf", task.Inputs[0], 0755); err != nil {
		log.Fatal(err)
	}

	cmd := exec.Command("./elf")

	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}

	stdout, err := os.OpenFile("stdout", os.O_CREATE|os.O_RDWR, 0777)
	if err != nil {
		log.Fatalln(err)
	}
	stderr, err := os.OpenFile("stderr", os.O_CREATE|os.O_WRONLY, 0777)
	if err != nil {
		log.Fatalln(err)
	}

	cmd.Stdout = stdout
	cmd.Stderr = stderr

	err = cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
	_, err = stdin.Write(task.Inputs[1])
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

	task.Outputs = [][]byte{stdOut}

	err = taskManager.Push(task)
	if err != nil {
		log.Fatal(err)
	}

	os.Remove("stdin")
	os.Remove("stdout")
	os.Remove("stderr")

}

func main() {
	tm := taskManager.NewRemoteTaskManager("http://127.0.0.1:8888")
	for {
		work(tm)
		time.Sleep(time.Second)
	}
}
