package protocol

import (
	"github.com/ekoeppen/gdcl/v2/fsm"
	"log"
)

const (
	STORAGE_IDLE = iota
	STORAGE_GET_DEFAULT_STORE
	STORAGE_GET_STORE_NAMES
	STORAGE_QUERY_SOUP
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

func querySoup(state int, input interface{}, output interface{}, data interface{}) {
	module := data.(*StorageModule)
	packet := DantePacketNew(QUERY, []byte{})
	module.ToDockLink <- *packet
}

func setCursorId(state int, input interface{}, output interface{}, data interface{}) {
	// module := data.(*StorageModule)
	log.Printf("Setting cursor ID\n")
}

func StorageModuleNew(toDockLink chan DantePacket) *StorageModule {
	var module StorageModule
	module.DockModule.DockModuleInit(toDockLink, &module)
	module.stateTable = map[int][]fsm.State{
		STORAGE_IDLE: {
			{Input: DantePacketCommand{APP_GET_DEFAULT_STORE}, NewState: STORAGE_GET_DEFAULT_STORE, Action: getDefaultStore},
			{Input: DantePacketCommand{APP_GET_STORE_NAMES}, NewState: STORAGE_GET_STORE_NAMES, Action: getStoreNames},
			{Input: DantePacketCommand{APP_QUERY_SOUP}, NewState: STORAGE_QUERY_SOUP, Action: querySoup},
		},
		STORAGE_GET_DEFAULT_STORE: {{Input: DantePacketCommand{DEFAULT_STORE}, NewState: STORAGE_IDLE, Action: gotDefaultStore}},
		STORAGE_GET_STORE_NAMES:   {{Input: DantePacketCommand{STORE_NAMES}, NewState: STORAGE_IDLE, Action: gotStoreNames}},
		STORAGE_QUERY_SOUP: {
			{Input: DantePacketCommand{RESULT}, NewState: STORAGE_IDLE, Action: setCursorId},
			{Input: DantePacketCommand{LONGDATA}, NewState: STORAGE_IDLE, Action: setCursorId},
		},
	}
	module.reader()
	return &module
}
