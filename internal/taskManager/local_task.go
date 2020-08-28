package taskManager

import (
	"encoding/json"
	"errors"
	"io/ioutil"
)

type LocalTaskManager struct {
	path  string
	queue []Task
}

func (l *LocalTaskManager) init() {
	if l.queue != nil {
		return
	}
	taskFiles, err := ioutil.ReadDir(l.path)
	if err != nil {
		panic("wrong path of local task manager")
	}
	for _, p := range taskFiles {
		content, err := ioutil.ReadFile(p.Name())
		if err != nil {
			panic("file read failure in local task manager")
		}
		var task = Task{}
		err = json.Unmarshal(content, &task)
		if err != nil {
			panic("json parse failure in local task manager")
		}
		l.queue = append(l.queue, task)
	}
}

func (l *LocalTaskManager) Fetch(tp string) (*Task, error) {
	l.init()
	for _, task := range l.queue {
		if task.Type == tp {
			return &task, nil
		}
	}
	return nil, errors.New("no task to fetch in local task manager")
}

func (l LocalTaskManager) Reserve(_ *Task) error {
	return nil
}

func (l LocalTaskManager) Push(_ *Task) error {
	return nil
}
