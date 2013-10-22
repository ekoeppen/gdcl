package protocol

import (
	"gdcl/fsm"
)

const (
	STORAGE_IDLE = iota
	STORAGE_GET_DEFAULT_STORE
)

type StorageModule struct {
	DockModule
}

func getDefaultStore(state int, input interface{}, output interface{}, data interface{}) {
	module := data.(*StorageModule)
	packet := DantePacketNew(GET_DEFAULT_STORE)
	module.ToDockLink <- *packet
}

func gotDefaultStore(state int, input interface{}, output interface{}, data interface{}) {
	module := data.(*StorageModule)
}

func StorageModuleNew(toDockLink chan DantePacket) *StorageModule {
	var module StorageModule
	module.DockModule.DockModuleInit(toDockLink, &module)
	module.stateTable = map[int][]fsm.State{
		fsm.ANY:           		{{Input: DantePacketCommand{APP_GET_DEFAULT_STORE}, NewState: STORAGE_GET_DEFAULT_STORE, Action: getDefaultStore}},
		STORAGE_GET_DEFAULT_STORE:	{{Input: DantePacketCommand{APP_GET_DEFAULT_STORE}, NewState: fsm.Any, Action: gotDefaultStore}},
	}
	module.reader()
	return &module
}

