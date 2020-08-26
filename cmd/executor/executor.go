package main

import (
	"fmt"
	"github.com/infinity-oj/actuator/internal/taskManager"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"time"
)

func work() {

	task, err := taskManager.Fetch("builder/Clang")
	if task == nil {
		return
	}
	fmt.Println(task.TaskId)
	fmt.Println(err)

	err = task.Reserve()
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
	fmt.Println(stdOut)
	fmt.Println(stdErr)

	task.Outputs = [][]byte{stdOut}

	err = task.Push()
	if err != nil {
		log.Fatal(err)
	}

	os.Remove("stdin")
	os.Remove("stdout")
	os.Remove("stderr")

}

func main() {
	for {
		work()
		time.Sleep(time.Second)
	}
}
