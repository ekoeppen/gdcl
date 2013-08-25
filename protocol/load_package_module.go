package protocol

import (
	"gdcl/fsm"
)

const (
	LOAD_PACKAGE_IDLE = iota
	LOAD_PACKAGE_UP
)

type LoadPackageModule struct {
	state       int
	stateTable  map[int][]fsm.State
	packageData []byte
	ToDockLink  chan DantePacket
}

func loadPackage(state int, input interface{}, output interface{}, data interface{}) {
	module := data.(*LoadPackageModule)
	packet := DantePacketNew(LOAD_PACKAGE, module.packageData)
	module.ToDockLink <- *packet
}

func disconnect(state int, input interface{}, output interface{}, data interface{}) {
	module := data.(*LoadPackageModule)
	packet := DantePacketNew(DISCONNECT, []byte{})
	module.ToDockLink <- *packet
}

func (module *LoadPackageModule) handlePacket(packet *DantePacket) {
	module.state = fsm.Transition(module.stateTable, module.state, packet, nil, module)
}

func LoadPackageModuleNew(toDockLink chan DantePacket, packageData []byte) *LoadPackageModule {
	var module LoadPackageModule
	module.ToDockLink = toDockLink
	module.packageData = packageData
	module.stateTable = map[int][]fsm.State{
		LOAD_PACKAGE_IDLE: {{Input: DantePacketCommand{RESULT}, NewState: LOAD_PACKAGE_UP, Action: loadPackage}},
		LOAD_PACKAGE_UP:   {{Input: DantePacketCommand{RESULT}, NewState: LOAD_PACKAGE_IDLE, Action: disconnect}},
	}
	return &module
}
