package install

import (
	"gdcl/v3/fsm"
	"gdcl/v3/protocol"
)

const (
	idle = iota
	installing
	sent
)

const (
	noAction int = iota
	sendRequest
	sendData
	installDone
	cancel
)

var transitions = []fsm.Transition[int, protocol.Command, int]{
	{State: idle, Event: protocol.APP_CONNECTED, Action: sendRequest, NewState: installing},
	{State: idle, Fallback: true, NewState: idle},
	{State: installing, Event: protocol.RESULT, Action: sendData, NewState: sent},
	{State: installing, Event: protocol.OPERATION_CANCELED, Action: cancel, NewState: idle},
	{State: sent, Event: protocol.RESULT, Action: installDone, NewState: idle},
}

var (
	state = idle
)

var PackageData []byte

func processIn(event *protocol.DockEvent) {
	var action int
	action, state = fsm.Input(event.Command, state, transitions)
	switch action {
	case sendRequest:
		protocol.Events <- protocol.NewDockEvent(
			protocol.REQUEST_TO_INSTALL,
			protocol.Out,
			[]byte{})
	case sendData:
		protocol.Events <- protocol.NewDockEvent(
			protocol.LOAD_PACKAGE,
			protocol.Out,
			PackageData)
	case installDone:
		protocol.Events <- protocol.NewDockEvent(
			protocol.DISCONNECT,
			protocol.Out,
			[]byte{})
	case cancel:
		protocol.Events <- protocol.NewDockEvent(
			protocol.OP_CANCELED_ACK,
			protocol.Out,
			[]byte{})
	}
}

func Process(event protocol.Event) {
	switch event.(type) {
	case *protocol.DockEvent:
		if event.(*protocol.DockEvent).Direction == protocol.In {
			processIn(event.(*protocol.DockEvent))
		}
	}
}
