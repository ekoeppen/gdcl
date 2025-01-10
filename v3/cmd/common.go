package cmd

import (
	"gdcl/v3/protocol"
	"gdcl/v3/protocol/dock"
	"gdcl/v3/protocol/framing"
	"gdcl/v3/protocol/mnp"
	"gdcl/v3/protocol/serial"
	"log"
)

var (
	port      string
	speed     int
	logSerial bool
	logMnp    bool
	logDock   bool
)

func logEvent(event protocol.Event) {
	switch event.(type) {
	case *protocol.SerialEvent:
		if logSerial {
			log.Println(event)
		}
	case *protocol.MnpEvent:
		if logMnp {
			log.Println(event)
		}
	case *protocol.DockEvent:
		if logDock {
			log.Println(event)
		}
	}
}

func eventLoop(port string, speed int, eventHandler func(event protocol.Event)) {
	log.Println("Starting event loop")
	go serial.SerialLoop(port, speed)
	logDock = false
	logMnp = true
	logSerial = false
	for {
		event := <-protocol.Events
		logEvent(event)

		serial.Process(event)
		framing.Process(event)
		mnp.Process(event)
		dock.Process(event)
		eventHandler(event)

		if protocol.IsQuitEvent(event) {
			break
		}
	}
	log.Println("Event loop complete")
}
