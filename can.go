//+build linux

package can

// #include <sys/socket.h>
//
// #include <linux/can.h>
// #include <linux/can/raw.h>
//
// int setupCAN(int fd, int ifindex) {
//   struct sockaddr_can addr;
//   addr.can_family = AF_CAN;
//   addr.can_ifindex = ifindex;
//   bind(fd, (struct sockaddr*) &addr, sizeof(addr));
//
//   return 0;
// }
import "C"

import (
	"errors"
	"fmt"
	"net"
	"syscall"
	"unsafe"
)

const (
	domain         = 29
	typ            = 3
	proto          = 1
	ADDR_FIELD     = uint32(0x1FFFFFFF)
	REQUEST_FIELD  = uint32(0x40000000)
	ERROR_FIELD    = uint32(0x20000000)
	EXTENDED_FIELD = uint32(0x80000000)
)

type CANFrame struct {
	ID   uint32
	Len  uint32
	Data [8]byte
}

func (fr *CANFrame) Addr() uint32 {
	return fr.ID & ADDR_FIELD
}

func (fr *CANFrame) SetAddr(addr uint32) {
	fr.ID &= ^ADDR_FIELD
	fr.ID |= (addr & ADDR_FIELD)
}

func (fr *CANFrame) IsRequest() bool {
	return (fr.ID & REQUEST_FIELD) > 0
}

func (fr *CANFrame) Request(set bool) {
	if set {
		fr.ID |= REQUEST_FIELD
	} else {
		fr.ID &= ^REQUEST_FIELD
	}
	fr.ID &= ^REQUEST_FIELD
}

func (fr *CANFrame) IsError() bool {
	return (fr.ID & ERROR_FIELD) > 0
}

func (fr *CANFrame) Error(set bool) {
	if set {
		fr.ID |= ERROR_FIELD
	} else {
		fr.ID &= ^ERROR_FIELD
	}
}

func (fr *CANFrame) IsExtended() bool {
	return (fr.ID & EXTENDED_FIELD) > 0
}

func (fr *CANFrame) String() string {
	return fmt.Sprintf("CANFrame{Addr: %#x, IsRequest: %v, Len: %d, Data: %#x}", fr.Addr(), fr.IsRequest(), fr.Len, fr.Data)
}

type CANBus struct {
	name string
	sock int
}

func NewCANBus(name string) (cb *CANBus, err error) {
	cb = &CANBus{name: name}
	// Create socket
	if cb.sock, err = syscall.Socket(domain, typ, proto); err != nil {
		return cb, err
	}
	// Find interface with name
	iface, err := net.InterfaceByName(name)
	if err != nil {
		return cb, err
	}
	if C.setupCAN(C.int(cb.sock), C.int(iface.Index)) != 0 {
		return cb, errors.New("CGo Errror")
	}
	return cb, nil
}

func (cb *CANBus) Write(fr *CANFrame) error {
	buf := (*[16]byte)(unsafe.Pointer(fr))
	if _, err := syscall.Write(cb.sock, buf[:]); err != nil {
		return err
	}
	return nil
}

func (cb *CANBus) Read(fr *CANFrame) error {
	buf := (*[16]byte)(unsafe.Pointer(fr))
	if _, err := syscall.Read(cb.sock, buf[:]); err != nil {
		return err
	}
	return nil
}
