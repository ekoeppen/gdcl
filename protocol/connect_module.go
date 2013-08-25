package protocol

import (
	"bytes"
	"crypto/des"
	"encoding/binary"
	"gdcl/fsm"
	"log"
)

const (
	CONN_IDLE = iota
	CONN_INITIATE
	CONN_DESKTOP_INFO
	CONN_WHICH_ICONS
	CONN_SET_TIMEOUT
	CONN_PASSWORD
	CONN_UP
)

const (
	SESSION_NONE            byte = 0
	SESSION_SETTING_UP      byte = 1
	SESSION_SYNCHRONIZE     byte = 2
	SESSION_RESTORE         byte = 3
	SESSION_LOAD_PACKAGE    byte = 4
	SESSION_TEST_COMM       byte = 5
	SESSION_LOAD_PATCH      byte = 6
	SESSION_UPDATING_STORES byte = 7
)

const (
	DESKTOP_MAC      byte = 0
	DESKTOP_WIN      byte = 1
	PROTOCOL_VERSION byte = 10
	DOCK_TIMEOUT     byte = 5
)

const (
	BACKUP_ICON   byte = 1
	RESTORE_ICON  byte = 2
	INSTALL_ICON  byte = 4
	IMPORT_ICON   byte = 8
	SYNC_ICON     byte = 16
	KEYBOARD_ICON byte = 32
	ALL_ICONS     byte = 63
)

type ConnectModule struct {
	state           int
	stateTable      map[int][]fsm.State
	sessionType     byte
	newtonChallenge uint64
	newtonPassword  uint64
	ToDockLink      chan DantePacket
}

func requestToDock(state int, input interface{}, output interface{}, data interface{}) {
	module := data.(*ConnectModule)
	packet := DantePacketNew(INITIATE_DOCKING, []byte{0, 0, 0, module.sessionType})
	module.ToDockLink <- *packet
}

func desktopInfo(state int, input interface{}, output interface{}, data interface{}) {
	module := data.(*ConnectModule)
	packet := DantePacketNew(DESKTOP_INFO, []byte{
		0, 0, 0, PROTOCOL_VERSION,
		0, 0, 0, DESKTOP_MAC,
		0x64, 0x23, 0xef, 0x02,
		0xfb, 0xcd, 0xc5, 0xa5,
		0, 0, 0, SESSION_SETTING_UP,
		0, 0, 0, 1,
		0x02, 0x05, 0x01, 0x06, 0x03, 0x07, 0x02, 0x69, 0x64, 0x07,
		0x04, 0x6e, 0x61, 0x6d, 0x65, 0x07, 0x07, 0x76, 0x65, 0x72,
		0x73, 0x69, 0x6f, 0x6e, 0x00, 0x08, 0x08, 0x38, 0x00, 0x4e,
		0x00, 0x65, 0x00, 0x77, 0x00, 0x74, 0x00, 0x6f, 0x00, 0x6e,
		0x00, 0x20, 0x00, 0x43, 0x00, 0x6f, 0x00, 0x6e, 0x00, 0x6e,
		0x00, 0x65, 0x00, 0x63, 0x00, 0x74, 0x00, 0x69, 0x00, 0x6f,
		0x00, 0x6e, 0x00, 0x20, 0x00, 0x55, 0x00, 0x74, 0x00, 0x69,
		0x00, 0x6c, 0x00, 0x69, 0x00, 0x74, 0x00, 0x69, 0x00, 0x65,
		0x00, 0x73, 0x00, 0x00, 0x00, 0x04})
	module.ToDockLink <- *packet
}

func whichIcons(state int, input interface{}, output interface{}, data interface{}) {
	module := data.(*ConnectModule)
	newton_info := input.(*DantePacket)
	buf := bytes.NewBuffer(newton_info.data[4:])
	binary.Read(buf, binary.BigEndian, &module.newtonChallenge)
	log.Printf("%08x", module.newtonChallenge)
	packet := DantePacketNew(WHICH_ICONS, []byte{0, 0, 0, ALL_ICONS})
	module.ToDockLink <- *packet
}

func setTimeout(state int, input interface{}, output interface{}, data interface{}) {
	module := data.(*ConnectModule)
	packet := DantePacketNew(SET_TIMEOUT, []byte{0, 0, 0, DOCK_TIMEOUT})
	module.ToDockLink <- *packet
}

func password(state int, input interface{}, output interface{}, data interface{}) {
	module := data.(*ConnectModule)
	var buf bytes.Buffer
	d, _ := des.NewCipher([]byte{0xe4, 0x0f, 0x7e, 0x9f, 0x0a, 0x36, 0x2c, 0xfa})
	binary.Write(&buf, binary.BigEndian, &module.newtonChallenge)
	d.Encrypt(buf.Bytes(), buf.Bytes())
	packet := DantePacketNew(PASSWORD, buf.Bytes())
	module.ToDockLink <- *packet
}

func (module *ConnectModule) handlePacket(packet *DantePacket) {
	log.Printf("Connection: %x\n", packet.command);
	module.state = fsm.Transition(module.stateTable, module.state, packet, nil, module)
}

func ConnectModuleNew(toDockLink chan DantePacket, sessionType byte) *ConnectModule {
	var module ConnectModule
	module.ToDockLink = toDockLink
	module.sessionType = sessionType
	if module.sessionType == SESSION_NONE {
		module.stateTable = map[int][]fsm.State{
			CONN_IDLE:         {{Input: DantePacketCommand{REQUEST_TO_DOCK}, NewState: CONN_INITIATE, Action: requestToDock}},
			CONN_INITIATE:     {{Input: DantePacketCommand{NEWTON_NAME}, NewState: CONN_DESKTOP_INFO, Action: desktopInfo}},
			CONN_DESKTOP_INFO: {{Input: DantePacketCommand{NEWTON_INFO}, NewState: CONN_WHICH_ICONS, Action: whichIcons}},
			CONN_WHICH_ICONS:  {{Input: DantePacketCommand{RESULT}, NewState: CONN_SET_TIMEOUT, Action: setTimeout}},
			CONN_SET_TIMEOUT:  {{Input: DantePacketCommand{PASSWORD}, NewState: CONN_UP, Action: password}},
			CONN_PASSWORD:     {{Input: DantePacketCommand{RESULT}, NewState: CONN_IDLE}},
			CONN_UP:           {{Input: DantePacketCommand{DISCONNECT}, NewState: CONN_IDLE}},
		}
	} else {
		module.stateTable = map[int][]fsm.State{
			CONN_IDLE:        {{Input: DantePacketCommand{REQUEST_TO_DOCK}, NewState: CONN_INITIATE, Action: requestToDock}},
			CONN_INITIATE:    {{Input: DantePacketCommand{NEWTON_NAME}, NewState: CONN_SET_TIMEOUT, Action: setTimeout}},
			CONN_SET_TIMEOUT: {{Input: DantePacketCommand{RESULT}, NewState: CONN_UP}},
			CONN_UP:          {{Input: DantePacketCommand{DISCONNECT}, NewState: CONN_IDLE}},
		}
	}
	return &module
}
