package core

import (
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/Anhtran0208/redis-server-intro/internal/constant"
	"github.com/Anhtran0208/redis-server-intro/internal/data_structure"
)

type Task struct {
	Command *Command
	ReplyCh chan []byte // Channel to send the result back to the client's handler
}
type Worker struct {
	id          int
	dictStore   *data_structure.Dict // data partition
	TashChannel chan *Task
}

func NewWorker(id, bufferSize int) *Worker {
	w := &Worker{
		id:          id,
		dictStore:   data_structure.CreateDict(),
		TashChannel: make(chan *Task, bufferSize),
	}
	go w.run()
	return w
}

func (w *Worker) run() {
	for task := range w.TashChannel {
		w.ExecuteAndResponse(task)
	}
}

func (w *Worker) cmdSET(args []string) []byte {
	if len(args) < 2 || len(args) == 3 || len(args) > 4 {
		return Encode(errors.New("(error) ERR wrong number of arguments for 'SET' command"), false)
	}
	var key, value string
	var ttlMs int64 = -1

	key, value = args[0], args[1]
	if len(args) > 2 {
		ttlSec, err := strconv.ParseInt(args[3], 10, 64)
		if err != nil {
			return Encode(errors.New("(error) ERR value is not an integer or out of range"), false)
		}
		ttlMs = ttlSec * 1000
	}
	w.dictStore.Set(key, w.dictStore.NewObj(key, value, ttlMs))
	return constant.RespOk
}

func (w *Worker) cmdGET(args []string) []byte {
	if len(args) != 1 {
		return Encode(errors.New("(error) ERR wrong number of arguments for 'GET' command"), false)
	}

	key := args[0]
	obj := w.dictStore.Get(key)
	if obj == nil {
		return constant.RespNil
	}

	if w.dictStore.HasExpired(key) {
		return constant.RespNil
	}

	return Encode(obj.Value, false)
}

func (w *Worker) ExecuteAndResponse(task *Task) {
	log.Printf("worker %d executes command %s", w.id, task.Command)
	var res []byte

	switch task.Command.Cmd {
	case "SET":
		res = w.cmdSET(task.Command.Args)
	case "GET":
		res = w.cmdGET(task.Command.Args)
	default:
		res = []byte(fmt.Sprintf("-CMD NOT FOUND\r\n"))
	}
	task.ReplyCh <- res
}
