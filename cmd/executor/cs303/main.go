package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"

	"github.com/infinity-oj/actuator/internal/taskManager"
)

func isExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func work(env Runtime, task *taskManager.Task) (warning, error string) {
	trainDataPath := "88266789-2bec-4a57-8028-be5a89350102.json"
	testDataPath := "5b8ad1e3-abbc-43b1-af6c-f542fded261e.json"

	workingDir := env.GetWorkingDir()

	hostConfig := &container.HostConfig{
		Resources: container.Resources{
			Memory:    34359720776 / 2,
			CPUPeriod: 100000,
			CPUQuota:  1000000,
			CPUCount:  10,
		},
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: workingDir,
				Target: "/work",
			},
		},
		MaskedPaths:   nil,
		ReadonlyPaths: nil,
		Init:          nil,
	}

	// cal cmd

	trainTestPath := "train_test.py"
	trainPath := "/work/train.py"
	testPath := "/work/test.py"
	pythonPath := "/root/anaconda3/envs/cs303/bin/python"
	var cmd = ""

	if isExists(path.Join(workingDir, "train_test.py")) {
		cmd += fmt.Sprintf("%s %s -i %s -t %s", pythonPath, trainTestPath, testDataPath, trainDataPath)
	} else {
		useModel := false
		if isExists(path.Join(workingDir, "train.py")) {
			cmd += fmt.Sprintf("%s %s -t %s\n", pythonPath, trainPath, trainDataPath)
			useModel = true
		}
		if isExists(path.Join(workingDir, "test.py")) {
			if useModel {
				cmd += fmt.Sprintf("%s %s -i %s -m model", pythonPath, testPath, testDataPath)
			} else {
				cmd += fmt.Sprintf("%s %s -i %s", pythonPath, testPath, testDataPath)
			}
		} else {
			return "", "test file not found"
		}
	}

	log.Println("cmd:", cmd)

	log.Println("start.sh: ", path.Join(workingDir, "start.sh"))
	err := ioutil.WriteFile(path.Join(workingDir, "start.sh"), []byte(cmd), 0644)
	if err != nil {
		log.Fatal(err)
	}

	config := &container.Config{
		Cmd:             []string{"bash", "start.sh"},
		Healthcheck:     nil,
		ArgsEscaped:     false,
		Image:           "cs303-proj3",
		Volumes:         nil,
		WorkingDir:      "/work",
		Entrypoint:      nil,
		NetworkDisabled: true,
		MacAddress:      "",
		OnBuild:         nil,
		Labels:          nil,
		StopSignal:      "",
		StopTimeout:     nil,
		Shell:           nil,
	}

	cli, err := client.NewClient("unix:///var/run/docker.sock", "v1.24", nil, nil)
	if err != nil {
		log.Fatal(err)
	}
	body, err := cli.ContainerCreate(
		context.Background(),
		config,
		hostConfig,
		nil, nil,
		fmt.Sprintf("cs303-proj3-%s", task.JudgementId),
	)
	if err != nil {
		log.Fatal(err)
	}
	containerId := body.ID
	env.SetContainer(containerId)

	log.Printf("ID: %s\n", containerId)

	err = cli.ContainerStart(
		context.TODO(),
		containerId,
		types.ContainerStartOptions{})
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	bodyCh, errCh := cli.ContainerWait(ctx, containerId, container.WaitConditionNotRunning)

	select {
	case err = <-errCh:
		return "", "Timeout after" + timeout.String()
	case body := <-bodyCh:
		log.Println(body)
	}

	data, err := cli.ContainerLogs(ctx, containerId, types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		log.Fatal(err)
	}
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	_, err = stdcopy.StdCopy(stdout, stderr, data)
	if err != nil {
		log.Fatal(err)
	}

	vol := task.Properties["volume"]
	log.Printf("Download volume: %s", vol)

	outputPath := path.Join(workingDir, "output.txt")
	log.Println("output.txt from", outputPath)

	output, err := ioutil.ReadFile(outputPath)
	if err != nil {
		if os.IsNotExist(err) {
			output = []byte{}
		} else {
			log.Fatal(err)
		}
	}

	task.Outputs = [][]byte{
		stdout.Bytes(),
		stderr.Bytes(),
		output,
	}

	return "", ""
}

var worker = "executor/cs303"
var timeout = time.Hour

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
