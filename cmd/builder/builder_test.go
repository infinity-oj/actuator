package main

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/infinity-oj/actuator/internal/taskManager"
)

func TestWork(t *testing.T) {

	queue := []*taskManager.Task{
		{
			JudgementId: "jid1",
			TaskId:      "tid1",
			Type:        "builder/Clang",
			Properties:  nil,
			Inputs:      [][]byte{[]byte("int main() {}")},
			Outputs:     nil,
		},
		{
			JudgementId: "jid2",
			TaskId:      "tid2",
			Type:        "builder/Clang",
			Properties:  nil,
			Inputs:      [][]byte{[]byte("int main() {}")},
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
