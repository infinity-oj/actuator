package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"time"

	"github.com/infinity-oj/actuator/internal/taskManager"
)

func work(taskManager taskManager.TaskManager) {

	var task, err = taskManager.Fetch("builder/Clang")
	if task == nil {
		return
	}
	fmt.Println(task.TaskId)
	fmt.Println(err)

	err = taskManager.Reserve(task)
	if err != nil {
		return
	}

	log.Printf("Get task, token: %s", task.Token)

	if err := ioutil.WriteFile("main.cpp", task.Inputs[0], 0644); err != nil {
		log.Fatal(err)
	}
	cmd := exec.Command("g++", "main.cpp", "-o", "main")

	// 读取io.Writer类型的cmd.Stdout，再通过bytes.Buffer(缓冲byte类型的缓冲器)将byte类型转化为string类型(out.String():这是bytes类型提供的接口)
	var out bytes.Buffer
	cmd.Stdout = &out

	// Run执行c包含的命令，并阻塞直到完成。  这里stdout被取出，cmd.Wait()无法正确获取stdin,stdout,stderr，则阻塞在那了
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}

	data, err := ioutil.ReadFile("main")

	task.Outputs = [][]byte{data}

	err = taskManager.Push(task)
	if err != nil {
		log.Fatal(err)
	}

}

func main() {
	tm := taskManager.NewRemoteManager("http://127.0.0.1:8888")
	for {
		work(tm)
		time.Sleep(time.Second)
	}
}
