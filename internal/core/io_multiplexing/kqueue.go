package io_multiplexing

import (
	"log"
	"syscall"
)

type KQueue struct {
	fd            int
	kqEvents      []syscall.Kevent_t
	genericEvents []Event
}

func CreateIOMultiplexer(maxEvent int) (*KQueue, error) {
	kqFD, err := syscall.Kqueue()
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	return &KQueue{
		fd:            kqFD,
		kqEvents:      make([]syscall.Kevent_t, maxEvent),
		genericEvents: make([]Event, maxEvent),
	}, nil
}

func (kq *KQueue) Monitor(event Event) error {
	// register a file id to kq
	kqEvent := event.convertToKqEvent(syscall.EV_ADD)
	// Add event.Fd to the monitoring list of kq.fd
	_, err := syscall.Kevent(kq.fd, []syscall.Kevent_t{kqEvent}, nil, nil)
	return err
}

func (kq *KQueue) Wait() ([]Event, error) {
	n, err := syscall.Kevent(kq.fd, nil, kq.kqEvents, nil)
	if err != nil {
		return nil, err
	}
	for i := 0; i < n; i++ {
		kq.genericEvents[i] = convertKqToEvent(kq.kqEvents[i])
	}

	return kq.genericEvents[:n], nil
}

func (kq *KQueue) Close() error {
	return syscall.Close(kq.fd)
}
