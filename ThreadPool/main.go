package main

import (
	"log"
	"net"
	"sync"
	"time"
)

// job in queue
type Job struct {
	conn net.Conn
}

// worker - thread in pool
type Worker struct {
	id      int
	jobChan chan Job
	wg      *sync.WaitGroup
}

// thread pool
type Pool struct {
	jobQueue chan Job
	workers  []*Worker
	wg       sync.WaitGroup
}

func NewWorker(id int, jobChan chan Job, wg *sync.WaitGroup) *Worker {
	return &Worker{
		id:      id,
		jobChan: jobChan,
		wg:      wg,
	}
}

func NewPool(numOfWorker int) *Pool {
	return &Pool{
		jobQueue: make(chan Job),
		workers:  make([]*Worker, numOfWorker),
	}
}

// pull job from queue
func (w *Worker) Start() {
	go func() {
		defer w.wg.Done()
		for job := range w.jobChan {
			log.Printf("Worker %d is handling job from %s", w.id, job.conn.RemoteAddr())
			handleConnection(job.conn)
		}
	}()
}

// start pool
func (p *Pool) Start() {
	for i := 0; i < len(p.workers); i++ {
		p.wg.Add(1)
		// create new worker
		worker := NewWorker(i, p.jobQueue, &p.wg)
		//start worker
		worker.Start()
	}
}

// add jobs to queue
func (p *Pool) AddJob(conn net.Conn) {
	p.jobQueue <- Job{conn: conn}
}
func main() {
	// create server port
	listener, err := net.Listen("tcp", ":3000")
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	// 1 pool with 2 threads
	pool := NewPool(2)
	pool.Start()
	for {
		// wait for client to connect and establish socket
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}
		pool.AddJob(conn)
	}

}

func handleConnection(conn net.Conn) {
	log.Println(conn.RemoteAddr())
	// read data from client
	var buffer []byte = make([]byte, 1000)

	_, err := conn.Read(buffer)
	if err != nil {
		log.Fatal(err)
	}

	// simulating process
	time.Sleep(time.Second * 10)

	// reply
	conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\nHello, world\r\n"))

	// close connection
	conn.Close()
}
