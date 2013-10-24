package protocol

import (
	"gdcl/fsm"
	"log"
)

const (
	STORAGE_IDLE = iota
	STORAGE_GET_DEFAULT_STORE
	STORAGE_GET_STORE_NAMES
)

type StorageModule struct {
	DockModule
}

func getDefaultStore(state int, input interface{}, output interface{}, data interface{}) {
	module := data.(*StorageModule)
	packet := DantePacketNew(GET_DEFAULT_STORE, []byte{})
	module.ToDockLink <- *packet
}

func gotDefaultStore(state int, input interface{}, output interface{}, data interface{}) {
	// module := data.(*StorageModule)
	log.Printf("Got default store\n")
}

func getStoreNames(state int, input interface{}, output interface{}, data interface{}) {
	module := data.(*StorageModule)
	packet := DantePacketNew(GET_STORE_NAMES, []byte{})
	module.ToDockLink <- *packet
}

func gotStoreNames(state int, input interface{}, output interface{}, data interface{}) {
	// module := data.(*StorageModule)
	log.Printf("Got store names\n")
}

func StorageModuleNew(toDockLink chan DantePacket) *StorageModule {
	var module StorageModule
	module.DockModule.DockModuleInit(toDockLink, &module)
	module.stateTable = map[int][]fsm.State{
		STORAGE_IDLE:          		{
			{Input: DantePacketCommand{APP_GET_DEFAULT_STORE}, NewState: STORAGE_GET_DEFAULT_STORE, Action: getDefaultStore},
			{Input: DantePacketCommand{APP_GET_STORE_NAMES}, NewState: STORAGE_GET_STORE_NAMES, Action: getStoreNames},
		},
		STORAGE_GET_DEFAULT_STORE:	{{Input: DantePacketCommand{DEFAULT_STORE}, NewState: STORAGE_IDLE, Action: gotDefaultStore}},
		STORAGE_GET_STORE_NAMES:	{{Input: DantePacketCommand{STORE_NAMES}, NewState: STORAGE_IDLE, Action: gotStoreNames}},
	}
	module.reader()
	return &module
}

