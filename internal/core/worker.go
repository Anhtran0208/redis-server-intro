package core

import (
	"log"
)

type Task struct {
	Command *Command
	ReplyCh chan []byte // Channel to send the result back to the client's handler
}
type Worker struct {
	id          int
	store       *Store
	executor    *Executor
	TaskChannel chan *Task
}

func NewWorker(id int, store *Store, bufferSize int) *Worker {
	w := &Worker{
		id:          id,
		store:       store,
		executor:    NewExecutor(store),
		TaskChannel: make(chan *Task, bufferSize),
	}
	go w.run()
	return w
}

func (w *Worker) run() {
	for task := range w.TaskChannel {
		w.ExecuteAndResponse(task)
	}
}

func (w *Worker) ExecuteAndResponse(task *Task) {
	log.Printf("worker %d executes command %s", w.id, task.Command)
	res := w.executor.ExecuteCMD(task.Command)
	task.ReplyCh <- res
}
