package io_multiplexing

import (
	"syscall"
)

// convert from Event to EVent in KQueue
func (e Event) convertToKqEvent(flags uint16) syscall.Kevent_t {
	var filter int16 = syscall.EVFILT_WRITE
	if e.Op == OpRead {
		filter = syscall.EVFILT_READ
	}
	return syscall.Kevent_t{
		Ident:  uint64(e.FileDescriptor),
		Filter: filter,
		Flags:  flags,
	}
}

// convert KQ event to event
func convertKqToEvent(kq syscall.Kevent_t) Event {
	var op Operation = OpWrite
	if kq.Filter == syscall.EVFILT_READ {
		op = OpRead
	}
	return Event{
		FileDescriptor: int(kq.Ident),
		Op:             op,
	}
}
