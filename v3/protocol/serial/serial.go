package serial

import (
	"gdcl/v3/protocol"
	"go.bug.st/serial"
	"log"
)

var (
	fd serial.Port
)

func SerialLoop(port string, speed int) {
	var err error
	log.Println("Starting serial loop on port", port)
	mode := &serial.Mode{
		BaudRate: speed,
	}
	fd, err = serial.Open(port, mode)
	if err != nil {
		log.Fatalf("Error opening %s: %s", port, err)
	}

	for {
		buf := make([]byte, 65536)
		n, err := fd.Read(buf)
		if err != nil {
			log.Fatal(err)
		}
		if n == 0 {
			protocol.Events <- protocol.NewDockEvent(protocol.APP_QUIT, protocol.In, []byte{})
			break
		}
		protocol.Events <- &protocol.SerialEvent{
			Direction: protocol.In,
			Data:      buf[:n],
		}

	}
	log.Println("Serial loop done")
}

func Process(event protocol.Event) {
	if protocol.IsQuitEvent(event) {
		fd.Close()
		return
	}

	switch event.(type) {
	case *protocol.SerialEvent:
		if event.(*protocol.SerialEvent).Direction == protocol.Out {
			fd.Write(event.(*protocol.SerialEvent).Data)
			fd.Drain()
		}
	}
}
