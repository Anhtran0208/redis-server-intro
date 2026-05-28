package server

import (
	"context"
	"hash/fnv"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"syscall"

	"github.com/Anhtran0208/redis-server-intro/internal/config"
	"github.com/Anhtran0208/redis-server-intro/internal/core"
	"github.com/Anhtran0208/redis-server-intro/internal/data_structure"
	"golang.org/x/sys/unix"
)

type MultiThreadServer struct {
	cfg           *config.Config
	workers       []*core.Worker
	ioHandlers    []*IOHandler
	numWorkers    int
	numIOHandlers int
	nextIOHandler uint64 // round robin
}

func NewMultiThreadServer(cfg *config.Config) (*MultiThreadServer, error) {
	multiThreadServer := &MultiThreadServer{
		cfg:           cfg,
		numWorkers:    cfg.NumWorkers,
		numIOHandlers: cfg.NumIOHandlers,
	}
	dictCfg := data_structure.DictConfig{
		MaxKeyNumber:       cfg.MaxKeyNumber,
		EvictionRatio:      cfg.EvictionRatio,
		EvictionPolicy:     data_structure.EvictionPolicy(cfg.EvictionPolicy),
		EpoolMaxSize:       cfg.EpoolMaxSize,
		EpoolLruSampleSize: cfg.EpoolLruSampleSize,
	}

	for i := 0; i < cfg.NumWorkers; i++ {
		store := core.NewStore(dictCfg)
		worker := core.NewWorker(i, store, 1024)
		multiThreadServer.workers = append(multiThreadServer.workers, worker)
	}

	for i := 0; i < cfg.NumIOHandlers; i++ {
		handler, err := NewIOHandler(i, multiThreadServer, cfg.MaxConnection)
		if err != nil {
			return nil, err
		}

		multiThreadServer.ioHandlers = append(multiThreadServer.ioHandlers, handler)
	}
	return multiThreadServer, nil
}

func (s *MultiThreadServer) getPartitionID(key string) int {
	hasher := fnv.New32()
	hasher.Write([]byte(key))
	return int(hasher.Sum32()) % s.numWorkers
}

func (s *MultiThreadServer) dispatch(task *core.Task) {
	var key string
	if len(task.Command.Args) > 0 {
		key = task.Command.Args[0]
	}
	workerID := s.getPartitionID(key)
	s.workers[workerID].TaskChannel <- task
}

func (s *MultiThreadServer) runSingleListener() {
	// Set up listener socket
	listener, err := net.Listen(s.cfg.Protocol, s.cfg.Port)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	log.Printf("Multi thread server listening on %s", s.cfg.Port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}
		s.handleAcceptedConn(conn)
	}
}

func (s *MultiThreadServer) runMultiListener() {
	var listenerWg sync.WaitGroup

	log.Printf(
		"Multi-thread server listening on %s with %d listeners, %d workers, %d I/O handlers",
		s.cfg.Port,
		s.cfg.NumListeners,
		s.numWorkers,
		s.numIOHandlers,
	)

	for i := 0; i < s.cfg.NumListeners; i++ {
		listenerWg.Add(1)

		go func(listenerID int) {
			defer listenerWg.Done()

			listener, err := createReusablePortListener(s.cfg.Protocol, s.cfg.Port)
			if err != nil {
				log.Fatalf("Listener %d failed to listen on %s: %v", listenerID, s.cfg.Port, err)
			}
			defer listener.Close()

			log.Printf("Listener %d started listening on %s", listenerID, s.cfg.Port)

			for {
				conn, err := listener.Accept()
				if err != nil {
					log.Printf("Listener %d failed to accept connection: %v", listenerID, err)
					continue
				}

				s.handleAcceptedConn(conn)
			}
		}(i)
	}
	listenerWg.Wait()
}

func (s *MultiThreadServer) RunMultiThreadServer(wg *sync.WaitGroup) {
	defer wg.Done()

	// Start all I/O handler event loops
	for _, handler := range s.ioHandlers {
		go handler.Run()
	}

	if s.cfg.NumListeners <= 1 {
		s.runSingleListener()
		return
	}

	s.runMultiListener()
}

func (s *MultiThreadServer) nextHandler() *IOHandler {
	idx := atomic.AddUint64(&s.nextIOHandler, 1)
	return s.ioHandlers[int(idx)%s.numIOHandlers]
}

// forward the new connection to an I/O handler in a round-robin manner
func (s *MultiThreadServer) handleAcceptedConn(conn net.Conn) {
	handler := s.nextHandler()

	if err := handler.AddConn(conn); err != nil {
		log.Printf("Failed to add connection to I/O handler %d: %v", handler.id, err)
		_ = conn.Close()
	}
}

func createReusablePortListener(network, addr string) (net.Listener, error) {
	lc := net.ListenConfig{
		Control: func(network, address string, c syscall.RawConn) error {
			var err error
			c.Control(func(fd uintptr) {
				err = unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEPORT, 1)
			})
			return err
		}}
	return lc.Listen(context.Background(), network, addr)
}
