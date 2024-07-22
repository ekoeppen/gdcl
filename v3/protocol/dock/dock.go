package dock

import (
	"bytes"
	"crypto/des"
	"encoding/binary"
	"gitlab.com/40hz/newton/gdcl/v3/fsm"
	"gitlab.com/40hz/newton/gdcl/v3/protocol"
)

const (
	none           = 0
	settingUp      = 1
	synchronize    = 3
	restore        = 4
	loadPackage    = 5
	testComm       = 6
	loadPatch      = 7
	updatingStores = 8
)

const (
	idle = iota
	initiating
	sentDesktopInfo
	sentWhichIcons
	sentTimeout
	sentPassword
	up
)

const (
	noAction int = iota
	initiateDocking
	sendDesktopInfo
	sendWhichIcons
	sendPassword
	sendTimeout
	passwordError
	connected
)

const (
	desktopMac      byte = 0
	desktopWin      byte = 1
	protocolVersion byte = 10
	dockTimeout     byte = 5
)

const (
	backupIcon   byte = 1
	restoreIcon  byte = 2
	installIcon  byte = 4
	importIcon   byte = 8
	syncIcon     byte = 16
	keyboardIcon byte = 32
	allIcons     byte = 63
)

var transitions = []fsm.Transition[int, protocol.Command, int]{
	{State: idle, Event: protocol.REQUEST_TO_DOCK, Action: initiateDocking, NewState: initiating},
	{State: idle, Fallback: true, NewState: idle},
	{State: initiating, Event: protocol.NEWTON_NAME, Action: sendDesktopInfo, NewState: sentDesktopInfo},
	{State: initiating, Fallback: true, NewState: initiating},
	{State: sentDesktopInfo, Event: protocol.NEWTON_INFO, Action: sendWhichIcons, NewState: sentWhichIcons},
	{State: sentDesktopInfo, Fallback: true, NewState: sentDesktopInfo},
	{State: sentWhichIcons, Event: protocol.RESULT, Action: sendTimeout, NewState: sentTimeout},
	{State: sentWhichIcons, Fallback: true, NewState: sentWhichIcons},
	{State: sentTimeout, Event: protocol.PASSWORD, Action: sendPassword, NewState: sentPassword},
	{State: sentTimeout, Fallback: true, NewState: sentTimeout},
	{State: sentPassword, Event: protocol.HELLO, Action: connected, NewState: up},
	{State: sentPassword, Event: protocol.RESULT, Action: passwordError, NewState: idle},
	{State: up, Fallback: true, NewState: up},
}

var (
	state           = idle
	newtonChallenge uint64
)

func processIn(event *protocol.DockEvent) {
	if event.Command == protocol.NEWTON_INFO {
		buf := bytes.NewBuffer(event.Data[4:])
		binary.Read(buf, binary.BigEndian, &newtonChallenge)
	}

	var action = noAction
	action, state = fsm.Input(event.Command, state, transitions)
	switch action {
	case initiateDocking:
		protocol.Events <- protocol.NewDockEvent(
			protocol.INITIATE_DOCKING,
			protocol.Out,
			[]byte{0, 0, 0, settingUp},
		)
	case sendTimeout:
		protocol.Events <- protocol.NewDockEvent(
			protocol.SET_TIMEOUT,
			protocol.Out,
			[]byte{0, 0, 0, 10},
		)
	case sendDesktopInfo:
		protocol.Events <- protocol.NewDockEvent(
			protocol.DESKTOP_INFO,
			protocol.Out,
			[]byte{
				0, 0, 0, protocolVersion,
				0, 0, 0, desktopMac,
				0x64, 0x23, 0xef, 0x02,
				0xfb, 0xcd, 0xc5, 0xa5,
				0, 0, 0, settingUp,
				0, 0, 0, 1,
				0x02, 0x05, 0x01, 0x06, 0x03, 0x07, 0x02, 0x69, 0x64, 0x07, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x07,
				0x07, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x00, 0x08, 0x08, 0x38, 0x00, 0x4e, 0x00, 0x65,
				0x00, 0x77, 0x00, 0x74, 0x00, 0x6f, 0x00, 0x6e, 0x00, 0x20, 0x00, 0x43, 0x00, 0x6f, 0x00, 0x6e,
				0x00, 0x6e, 0x00, 0x65, 0x00, 0x63, 0x00, 0x74, 0x00, 0x69, 0x00, 0x6f, 0x00, 0x6e, 0x00, 0x20,
				0x00, 0x55, 0x00, 0x74, 0x00, 0x69, 0x00, 0x6c, 0x00, 0x69, 0x00, 0x74, 0x00, 0x69, 0x00, 0x65,
				0x00, 0x73, 0x00, 0x00, 0x00, 0x04},
		)
	case sendWhichIcons:
		protocol.Events <- protocol.NewDockEvent(
			protocol.WHICH_ICONS,
			protocol.Out,
			[]byte{0, 0, 0, allIcons},
		)
	case sendPassword:
		var buf bytes.Buffer
		d, _ := des.NewCipher([]byte{0xe4, 0x0f, 0x7e, 0x9f, 0x0a, 0x36, 0x2c, 0xfa})
		binary.Write(&buf, binary.BigEndian, newtonChallenge)
		d.Encrypt(buf.Bytes(), buf.Bytes())
		protocol.Events <- protocol.NewDockEvent(
			protocol.PASSWORD,
			protocol.Out,
			buf.Bytes(),
		)
	case connected:
		protocol.Events <- protocol.NewDockEvent(
			protocol.APP_CONNECTED,
			protocol.In,
			[]byte{},
		)
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
