package main

import (
	"testing"

	"github.com/infinity-oj/actuator/internal/taskManager"
	"github.com/stretchr/testify/assert"
)

func TestWork(t *testing.T) {

	queue := []*taskManager.Task{
		{
			JudgementId: "jid1",
			TaskId:      "tid1",
			Type:        "executor/elf",
			Properties:  nil,
			Inputs:      [][]byte{[]byte("")},
			Outputs:     nil,
		},
		{
			JudgementId: "jid2",
			TaskId:      "tid2",
			Type:        "executor/elf",
			Properties:  nil,
			Inputs:      [][]byte{[]byte("")},
			Outputs:     nil,
		},
	}

	result := [][][]byte{
		{
			[]byte("XD"),
		},
	}

	tm := taskManager.NewLocalTaskManager(queue)
	work(tm)

	for k := range queue {
		assert.Equal(t, result[k], queue[k], "they should be equal")
	}

}
