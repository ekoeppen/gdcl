package serial

import (
	"fmt"
	"gitlab.com/40hz/newton/gdcl/v3/protocol"
	"go.bug.st/serial"
	"log"
)

var fd serial.Port

func SerialLoop(port string) {
	var err error
	fmt.Println("Starting serial loop")
	mode := &serial.Mode{
		BaudRate: 115200,
	}
	fd, err = serial.Open(port, mode)
	if err != nil {
		log.Fatal(err)
	}
serialLoop:
	for {
		buf := make([]byte, 65536)
		n, err := fd.Read(buf)
		if err != nil {
			log.Fatal(err)
			break serialLoop
		}
		if n == 0 {
			protocol.Events <- &protocol.DockEvent{
				Command:   protocol.APP_QUIT,
				Direction: protocol.In,
			}
			break serialLoop
		}
		protocol.Events <- &protocol.SerialEvent{
			Direction: protocol.In,
			Data:      buf[:n],
		}

	}
	fmt.Println("Serial loop done")
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
