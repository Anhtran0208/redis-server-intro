package server

import (
	"io"
	"log"
	"net"
	"sync"
	"syscall"

	"github.com/Anhtran0208/redis-server-intro/internal/core"
	"github.com/Anhtran0208/redis-server-intro/internal/core/io_multiplexing"
)

type IOHandler struct {
	id            int
	ioMultiplexer io_multiplexing.IOMultiplexer
	mu            sync.Mutex
	server        *MultiThreadServer
	conns         map[int]net.Conn // map from fd -> connection
}

func NewIOHandler(id int, server *MultiThreadServer, maxEvents int) (*IOHandler, error) {
	multiplexer, err := io_multiplexing.CreateIOMultiplexer(maxEvents)
	if err != nil {
		return nil, err
	}
	return &IOHandler{
		id:            id,
		ioMultiplexer: multiplexer,
		server:        server,
		conns:         make(map[int]net.Conn),
	}, nil
}

// add conn to monitoring list of io handler
func (h *IOHandler) AddConn(conn net.Conn) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	tcpConn := conn.(*net.TCPConn)
	rawConn, err := tcpConn.SyscallConn()
	if err != nil {
		return err
	}

	// get the fd from connection and add it to the monitoring list for read operation
	var connFd int
	err = rawConn.Control(func(fd uintptr) {
		connFd = int(fd)
		log.Printf("I/O Handler %d is monitoring fd %d", h.id, connFd)
		// Store the connection object so it's not garbage collected
		h.conns[connFd] = conn
		// Add to epoll
		h.ioMultiplexer.Monitor(io_multiplexing.Event{
			FileDescriptor: connFd,
			Op:             io_multiplexing.OpRead,
		})
	})

	return err
}

func (h *IOHandler) Run() {
	log.Printf("I/O Handler %d started", h.id)
	for {
		events, err := h.ioMultiplexer.Wait()
		if err != nil {
			continue
		}

		for _, event := range events {
			h.handleEvents(event)
		}
	}
}

func (h *IOHandler) handleEvents(event io_multiplexing.Event) {
	connFd := event.FileDescriptor
	cmd, err := readCommand(connFd)
	if err != nil {
		if err == io.EOF || err == syscall.ECONNRESET {
			log.Println("client disconnected")
			h.removeConn(connFd)
			return
		}
		log.Println("read error:", err)
		return
	}

	replyCh := make(chan []byte, 1)
	task := &core.Task{
		Command: cmd,
		ReplyCh: replyCh,
	}
	// dispatch cmd to corresponding worker
	h.server.dispatch(task)

	// wait until cmd is done
	res := <-replyCh
	_, err = syscall.Write(connFd, res)
	if err != nil {
		log.Println("write error", err)
		h.removeConn(connFd)
	}
}

func (h *IOHandler) removeConn(connFd int) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if conn, ok := h.conns[connFd]; ok {
		_ = conn.Close()
		delete(h.conns, connFd)
		return
	}

	_ = syscall.Close(connFd)
}
