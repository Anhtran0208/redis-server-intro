package server

import (
	"io"
	"log"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/Anhtran0208/redis-server-intro/internal/config"
	"github.com/Anhtran0208/redis-server-intro/internal/constant"
	"github.com/Anhtran0208/redis-server-intro/internal/core"
	"github.com/Anhtran0208/redis-server-intro/internal/core/io_multiplexing"
)

var serverStatus int32 = constant.ServerStatusIdle

func RunIOMultiplexingServer(wg *sync.WaitGroup) {
	defer wg.Done()
	log.Println("Starting an I/O multiplexing TCP server on", config.Port)
	// create a listener
	listener, err := net.Listen(config.Protocol, config.Port)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	// get file descripter from listener
	tcpListener, ok := listener.(*net.TCPListener)
	if !ok {
		log.Fatal("listener is not TCP listner")
	}
	listenerFile, err := tcpListener.File()
	if err != nil {
		log.Fatal(err)
	}
	defer listenerFile.Close()
	serverFileDescriptor := int(listenerFile.Fd())

	// create an epoll isntance to monitor listener fd
	ioMultiplexer, err := io_multiplexing.CreateIOMultiplexer()
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

	var events = make([]io_multiplexing.Event, config.MaxConnection)
	var lastActiveExpireExecTime = time.Now()
	for atomic.LoadInt32(&serverStatus) != constant.ServerStatusShuttingDown {
		if time.Now().After(lastActiveExpireExecTime.Add(constant.ActiveExpireFrequency)) {
			// actively delete expired key
			//idle
			if !atomic.CompareAndSwapInt32(&serverStatus, constant.ServerStatusIdle, constant.ServerStatusBusy) {
				if serverStatus == constant.ServerStatusShuttingDown {
					return
				}
			} //busy
			core.ActiveDeleteExpiredKeys()
			// idle
			atomic.SwapInt32(&serverStatus, constant.ServerStatusIdle)
			lastActiveExpireExecTime = time.Now()
		}
		// wait for file descriptor in the monitor list to be ready
		// idle
		events, err = ioMultiplexer.Wait()
		if err != nil {
			continue
		}
		if !atomic.CompareAndSwapInt32(&serverStatus, constant.ServerStatusIdle, constant.ServerStatusBusy) {
			if serverStatus == constant.ServerStatusShuttingDown {
				return
			}
		} //busy
		// handle event
		for i := 0; i < len(events); i++ {
			// new client trying to make connect
			if events[i].FileDescriptor == serverFileDescriptor {
				// setup connection
				connFd, _, err := syscall.Accept(serverFileDescriptor)
				if err != nil {
					log.Println("err", err)
					continue
				}
				log.Printf("set up new connection")
				// ask kqueue to monitor this conn
				if err = ioMultiplexer.Monitor(io_multiplexing.Event{
					FileDescriptor: connFd,
					Op:             io_multiplexing.OpRead,
				}); err != nil {
					log.Fatal(err)
				}
			} else {
				// existing client send a new cmd
				// read cmd
				cmd, err := readCommand(events[i].FileDescriptor)
				log.Println("command: ", cmd)
				if err != nil {
					if err == io.EOF || err == syscall.ECONNRESET {
						log.Println("client disconnected")
						_ = syscall.Close(events[i].FileDescriptor)
						continue
					}
					log.Println("read error", err)
					continue
				}
				if err = core.ExecuteAndResponse(cmd, events[i].FileDescriptor); err != nil {
					log.Println("err write:", err)
				}
			}
		}
		atomic.SwapInt32(&serverStatus, constant.ServerStatusIdle)
	}
}

func readCommand(fd int) (*core.Command, error) {
	var buf = make([]byte, 512)
	n, err := syscall.Read(fd, buf)
	if err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, io.EOF
	}
	return core.ParseCmd(buf)
}

func respond(data string, fd int) error {
	if _, err := syscall.Write(fd, []byte(data)); err != nil {
		return err
	}
	return nil
}

func WaitForSignal(wg *sync.WaitGroup, signals chan os.Signal) {
	defer wg.Done()
	<-signals

	// wait for ongoing cmd to finish
	for {
		// compare current server status: if it's idle => shutting down
		if atomic.CompareAndSwapInt32(&serverStatus, constant.ServerStatusIdle, constant.ServerStatusShuttingDown) {
			log.Println("Shutting down gracefully")
			os.Exit(0) // shut down
		}
	}
}
