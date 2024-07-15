package mnp

import (
	"fmt"
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
	maxInfoLength              uint16
	lastAckSequenceNumber      byte
	outstandingPackets         []outstandingPacket
	maxOutstanding             byte
	sendCreditStateVariable    byte
	receiveCreditStateVariable byte
	sendStateVariable          byte
	receiveStateVariable       byte
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
	fmt.Printf("Processing MNP packet: %d\n", len(event.Data))
	var packetType = event.Data[1]
	fmt.Printf("%02x %d -> ", packetType, state)
	action, state = fsm.Input(packetType, state, transitions)
	fmt.Printf("%d %d\n", action, state)
	switch action {
	case sendLinkRequestResponse:
		framingMode := event.Data[13]
		maxOutstanding = event.Data[16]
		dataPhaseOpt := event.Data[23]
		if dataPhaseOpt&0x1 == 0x1 {
			maxInfoLength = 256
		} else {
			maxInfoLength = 64
		}
		buf := []byte{23, lr, 2,
			1, 6, 1, 0, 0, 0, 0, 255,
			2, 1, framingMode,
			3, 1, maxOutstanding,
			4, 2, 64, 0,
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
		protocol.Events <- &protocol.DockEvent{
			Direction: protocol.In,
			Data:      event.Data[3:],
		}
	case closeConnection:
		protocol.Events <- protocol.NewDockEvent(protocol.APP_QUIT, protocol.In, []byte{})
	}
}

func processOut(event *protocol.DockEvent) {
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
