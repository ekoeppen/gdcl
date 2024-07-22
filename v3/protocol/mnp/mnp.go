package mnp

import (
	"bytes"
	"encoding/binary"
	"gitlab.com/40hz/newton/gdcl/v3/fsm"
	"gitlab.com/40hz/newton/gdcl/v3/protocol"
)

const (
	idle = iota
	linkRequest
	dataPhase
	disconnecting
)

const (
	lr byte = 1
	ld byte = 2
	lt byte = 4
	la byte = 5
)

const (
	none = iota
	sendLinkRequestResponse
	handleLinkAcknowledgement
	closeConnection
	handleLinkTransfer
)

type packetType struct {
	packetType byte
}

type outstandingPacket struct {
	data               []byte
	sendSequenceNumber byte
}

var (
	state                      int = idle
	maxInfoLength              int
	lastAckSequenceNumber      byte
	outstandingPackets         []outstandingPacket
	maxOutstanding             byte
	sendCreditStateVariable    byte
	receiveCreditStateVariable byte
	sendStateVariable          byte
	receiveStateVariable       byte
	dockPacket                 protocol.DockEvent
	dockPacketStarted          bool
)

var transitions = []fsm.Transition[int, byte, int]{
	{State: idle, Event: lr, NewState: linkRequest, Action: sendLinkRequestResponse},
	{State: linkRequest, Event: lr, NewState: linkRequest},
	{State: linkRequest, Event: la, NewState: dataPhase},
	{State: linkRequest, Event: ld, NewState: idle},
	{State: linkRequest, Event: lt, NewState: idle},
	{State: dataPhase, Event: lr, NewState: idle},
	{State: dataPhase, Event: la, NewState: dataPhase, Action: handleLinkAcknowledgement},
	{State: dataPhase, Event: ld, NewState: idle, Action: closeConnection},
	{State: dataPhase, Event: lt, NewState: dataPhase, Action: handleLinkTransfer},
}

func processIn(event *protocol.MnpEvent) {
	var action int
	var packetType = event.Data[1]
	action, state = fsm.Input(packetType, state, transitions)
	switch action {
	case sendLinkRequestResponse:
		framingMode := event.Data[13]
		maxOutstanding = event.Data[16]
		dataPhaseOpt := event.Data[23]
		if dataPhaseOpt&0x1 == 0x1 {
			maxInfoLength = 256
		} else {
			maxInfoLength = int(event.Data[19])*256 + int(event.Data[20])
		}
		buf := []byte{23, lr, 2,
			1, 6, 1, 0, 0, 0, 0, 255,
			2, 1, framingMode,
			3, 1, maxOutstanding,
			4, 2, event.Data[19], event.Data[20],
			8, 1, dataPhaseOpt}
		outstandingPackets = make([]outstandingPacket, 0, maxOutstanding)
		protocol.Events <- &protocol.MnpEvent{
			Direction: protocol.Out,
			Data:      buf,
		}
	case handleLinkAcknowledgement:
		receiveStateVariable := event.Data[2]
		adjustedCredit := int(event.Data[3]) - len(outstandingPackets)
		if adjustedCredit > 0 && adjustedCredit <= 8 {
			receiveCreditStateVariable = byte(adjustedCredit)
		} else {
			receiveCreditStateVariable = 0
		}
		lastAckSequenceNumber = receiveStateVariable
	case handleLinkTransfer:
		receiveStateVariable = event.Data[2]
		protocol.Events <- &protocol.MnpEvent{
			Direction: protocol.Out,
			Data:      []byte{3, la, receiveStateVariable, 8},
		}
		if !dockPacketStarted {
			buf := bytes.NewBuffer(event.Data[11:])
			binary.Read(buf, binary.BigEndian, &dockPacket.Command)
			binary.Read(buf, binary.BigEndian, &dockPacket.Length)
			dockPacket.Data = event.Data[19:]
			dockPacket.Direction = protocol.In
			dockPacketStarted = true
		} else {
			dockPacket.Data = append(dockPacket.Data, event.Data[3:]...)
		}
		if uint32(len(dockPacket.Data)) >= dockPacket.Length {
			dockPacketStarted = false
			dockPacket.Data = dockPacket.Data[:dockPacket.Length]
			protocol.Events <- &dockPacket
		}
	case closeConnection:
		protocol.Events <- protocol.NewDockEvent(protocol.APP_QUIT, protocol.In, []byte{})
	}
}

func processOut(event *protocol.DockEvent) {
	eventData := event.Encode()
	for len(eventData) > 0 {
		sendStateVariable++
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, byte(2))
		binary.Write(buf, binary.BigEndian, byte(lt))
		binary.Write(buf, binary.BigEndian, byte(sendStateVariable))
		n := len(eventData)
		if n > int(maxInfoLength) {
			n = int(maxInfoLength)
		}
		buf.Write(eventData[:n])
		eventData = eventData[n:]
		protocol.Events <- &protocol.MnpEvent{
			Direction: protocol.Out,
			Data:      buf.Bytes(),
		}
	}
}

func Process(event protocol.Event) {
	switch event.(type) {
	case *protocol.MnpEvent:
		if event.(*protocol.MnpEvent).Direction == protocol.In {
			processIn(event.(*protocol.MnpEvent))
		}
	case *protocol.DockEvent:
		if event.(*protocol.DockEvent).Direction == protocol.Out {
			processOut(event.(*protocol.DockEvent))
		}
	}
}
