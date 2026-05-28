package server

import (
	"io"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/Anhtran0208/redis-server-intro/internal/config"
	"github.com/Anhtran0208/redis-server-intro/internal/constant"
	"github.com/Anhtran0208/redis-server-intro/internal/core"
	"github.com/Anhtran0208/redis-server-intro/internal/core/io_multiplexing"
	"github.com/Anhtran0208/redis-server-intro/internal/data_structure"
)

func RunSingleThreadServer(wg *sync.WaitGroup, cfg *config.Config) {
	defer wg.Done()
	log.Println("Starting single thread server on", cfg.Port)
	dictCfg := data_structure.DictConfig{
		MaxKeyNumber:       cfg.MaxKeyNumber,
		EvictionRatio:      cfg.EvictionRatio,
		EvictionPolicy:     data_structure.EvictionPolicy(cfg.EvictionPolicy),
		EpoolMaxSize:       cfg.EpoolMaxSize,
		EpoolLruSampleSize: cfg.EpoolLruSampleSize,
	}

	store := core.NewStore(dictCfg)
	executor := core.NewExecutor(store)

	// create listener
	listener, err := net.Listen(cfg.Protocol, cfg.Port)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	// get file descriptor from listener
	tcpListener, ok := listener.(*net.TCPListener)
	if !ok {
		log.Fatal("Listener is not TCP listener")
	}
	listenerFile, err := tcpListener.File()
	if err != nil {
		log.Fatal(err)
	}
	defer listenerFile.Close()
	serverFileDescriptor := int(listenerFile.Fd())

	// Create an ioMultiplexer instance (epoll in Linux, kqueue in MacOS)
	ioMultiplexer, err := io_multiplexing.CreateIOMultiplexer(cfg.MaxConnection)
	if err != nil {
		log.Fatal(err)
	}
	defer ioMultiplexer.Close()

	// monitor read events on server FD
	if err = ioMultiplexer.Monitor(io_multiplexing.Event{
		FileDescriptor: serverFileDescriptor,
		Op:             io_multiplexing.OpRead,
	}); err != nil {
		log.Fatal(err)
	}

	var events = make([]io_multiplexing.Event, cfg.MaxConnection)
	var lastActiveExpireExecTime = time.Now()

	for atomic.LoadInt32(&serverStatus) != constant.ServerStatusShuttingDown {
		// delete key if last execution time > 100ms
		// convert status from idle to busy
		if time.Now().After(lastActiveExpireExecTime.Add(constant.ActiveExpireFrequency)) {
			if !atomic.CompareAndSwapInt32(&serverStatus, constant.ServerStatusIdle, constant.ServerStatusBusy) {
				if atomic.LoadInt32(&serverStatus) == constant.ServerStatusShuttingDown {
					return
				}
			}
			store.ActiveDeleteExpiredKeys()
			// convert server status to idle after delete
			atomic.SwapInt32(&serverStatus, constant.ServerStatusIdle)
			lastActiveExpireExecTime = time.Now()
		}

		// wait for file descriptor in the monitor list to be ready
		events, err = ioMultiplexer.Wait()
		if err != nil {
			continue
		}

		// convert server status from idle to busy when handle events
		if !atomic.CompareAndSwapInt32(&serverStatus, constant.ServerStatusIdle, constant.ServerStatusBusy) {
			if atomic.LoadInt32(&serverStatus) == constant.ServerStatusShuttingDown {
				return
			}
		}

		// handle events
		for i := 0; i < len(events); i++ {
			// new client trying to make connect
			if events[i].FileDescriptor == serverFileDescriptor {
				// setup connection
				connFd, _, err := syscall.Accept(serverFileDescriptor)
				if err != nil {
					log.Println("Error when setting up connection", err)
					continue
				}
				log.Printf("Set up new connection")

				// ask ioMultiplexer (kqueue/epoll) to monitor this conn
				if err = ioMultiplexer.Monitor(io_multiplexing.Event{
					FileDescriptor: connFd,
					Op:             io_multiplexing.OpRead,
				}); err != nil {
					log.Fatal(err)
				}
			} else {
				// existing client send new cmd
				// read cmd
				cmd, err := readCommand(events[i].FileDescriptor)
				log.Println("Command: ", cmd)
				if err != nil {
					if err == io.EOF || err == syscall.ECONNRESET {
						log.Println("Client disconnected")
						_ = syscall.Close(events[i].FileDescriptor)
						continue
					}
					log.Println("Read error", err)
					continue
				}

				// execute cmd
				res := executor.ExecuteCMD(cmd)
				_, err = syscall.Write(events[i].FileDescriptor, res)
				if err != nil {
					log.Println("Error write", err)
				}
			}
		}
		// convert server status to idle after done
		atomic.SwapInt32(&serverStatus, constant.ServerStatusIdle)
	}
}

// read cmd from client
func readCommand(fd int) (*core.Command, error) {
	buf := make([]byte, 512)
	n, err := syscall.Read(fd, buf)
	if err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, io.EOF
	}
	return core.ParseCmd(buf[:n])
}
