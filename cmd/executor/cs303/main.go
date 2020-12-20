package main

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"time"

	"github.com/infinity-oj/actuator/internal/environment"

	"github.com/infinity-oj/actuator/internal/taskManager"
)

func isExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func work(env environment.Runtime, task *taskManager.Task) (warning, error string) {
	vol := task.Properties["volume"]
	log.Printf("Download volume: %s", vol)

	workingDir, err := env.GetVolume(vol)
	if err != nil {
		log.Fatal(err)
	}

	trainTestPath := path.Join("train_test.py")

	trainDataPath, _ := filepath.Abs("88266789-2bec-4a57-8028-be5a89350102.json")
	testDataPath, _ := filepath.Abs("5b8ad1e3-abbc-43b1-af6c-f542fded261e.json")

	var cmd *exec.Cmd = nil

	log.Printf("Working dir: %s", workingDir)

	if isExists(path.Join(workingDir, trainTestPath)) {
		cmd = exec.Command("python3", trainTestPath, "-i", testDataPath, "-t", trainDataPath)
		cmd.Dir = workingDir
	} else {

		trainPath := path.Join("train.py")
		testPath := path.Join("test.py")

		useModel := false

		ctx1, cancel1 := context.WithTimeout(context.Background(), timeout)
		defer cancel1()

		if isExists(path.Join(workingDir, trainPath)) {
			cmd := exec.CommandContext(ctx1, "python3", trainPath, "-t", trainDataPath)
			cmd.Dir = workingDir

			err = cmd.Start()
			if err != nil {
				log.Fatal(err)
			}
			_ = cmd.Wait()

			useModel = true
		}

		ctx2, cancel2 := context.WithTimeout(context.Background(), timeout)
		defer cancel2()

		if isExists(path.Join(workingDir, testPath)) {

			if useModel {
				cmd = exec.CommandContext(ctx2, "python3", testPath, "-i", testDataPath, "-m", "model")
			} else {
				cmd = exec.CommandContext(ctx2, "python3", testPath, "-i", testDataPath)
			}
			cmd.Dir = workingDir
		} else {

			return "", "test file not found"
		}

	}

	if cmd == nil {
		return "", "unknown error"
	} else {

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

		_ = cmd.Wait()
		_ = stdin.Close()
		_ = stdout.Close()
		_ = stderr.Close()

		outputPath := path.Join(workingDir, "output.txt")

		stdOut, err := ioutil.ReadFile(stdout.Name())
		stdErr, err := ioutil.ReadFile(stderr.Name())
		output, err := ioutil.ReadFile(outputPath)

		task.Outputs = [][]byte{
			stdOut,
			stdErr,
			output,
		}

		_ = os.Remove(stdout.Name())
		_ = os.Remove(stderr.Name())

		return "", ""
	}

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
		log.Printf("task returned")

		env.TearDown()
	}
}
