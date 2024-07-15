package dock

import (
	"bytes"
	"encoding/binary"
	"gitlab.com/40hz/newton/gdcl/v3/protocol"
)

func Process(event protocol.Event) {
	switch event.(type) {
	case *protocol.DockEvent:
		dockEvent := event.(*protocol.DockEvent)
		buf := bytes.NewBuffer(dockEvent.Data[8:])
		binary.Read(buf, binary.BigEndian, &dockEvent.Command)
		binary.Read(buf, binary.BigEndian, &dockEvent.Length)
		dockEvent.Data = dockEvent.Data[16 : 16+dockEvent.Length]
	}
}
