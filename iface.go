package main

import (
	"sync"

	"github.com/songgao/water"
	_ "github.com/songgao/water/waterutil"
)

type PacketHandler interface {
	ReceivePacket(tap *Interface, packet []byte)
	ReceivePacketError(tap *Interface, err error)
	InterfaceStart(tap *Interface)
	InterfaceStop(tap *Interface)
}

type Interface struct {
	running    bool
	tap        *water.Interface
	BufferSize int
	Name       string
	handler    PacketHandler
	wg         sync.WaitGroup
}

func NewInterface(name string, bufsize int, handler PacketHandler) (*Interface, error) {
	var err error

	r := &Interface{
		BufferSize: bufsize,
		Name:       name,
		handler:    handler,
	}

	r.tap, err = water.NewTAP(name)
	return r, err
}

func (iface *Interface) ReadLoop() {
	if !iface.running {
		iface.running = true
		go iface.readLoop()
	}
}

func (iface *Interface) Stop() {
	if iface.running {
		iface.running = false
		// need at least 1 more packet to fully quit
		iface.wg.Wait()
	}
}

func (iface *Interface) readLoop() {
	iface.wg.Add(1)
	defer iface.wg.Done()

	iface.handler.InterfaceStart(iface)

	buf := make([]byte, iface.BufferSize)
	for iface.running {
		n, err := iface.tap.Read(buf)
		if err != nil {
			iface.handler.ReceivePacketError(iface, err)
			continue
		}

		if n < 1 {
			// eagain?
			continue
		}

		pkt := buf[0:n]
		iface.handler.ReceivePacket(iface, pkt)
	}

	iface.handler.InterfaceStop(iface)
}

func (iface *Interface) Write(pkt []byte) (int, error) {
	return iface.tap.Write(pkt)
}
