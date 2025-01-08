package info

import (
	"gdcl/v3/fsm"
	"gdcl/v3/nsof"
	"gdcl/v3/protocol"
	"log"
)

const (
	idle = iota
	gettingStoreNames
	selectingStore
	gettingSoupNames
	gettingAppList
)

const (
	noAction int = iota
	getStoreNames
	selectStore
	getSoupNames
	showSoupNames
	showAppList
	cancel
)

var transitions = []fsm.Transition[int, protocol.Command, int]{
	{State: idle, Event: protocol.APP_CONNECTED, Action: getStoreNames, NewState: gettingStoreNames},
	{State: idle, Fallback: true, NewState: idle},
	{State: gettingStoreNames, Event: protocol.STORE_NAMES, Action: selectStore, NewState: selectingStore},
	{State: gettingStoreNames, Event: protocol.OPERATION_CANCELED, Action: cancel, NewState: idle},
	{State: gettingStoreNames, Fallback: true, NewState: gettingStoreNames},
	{State: selectingStore, Event: protocol.RESULT, Action: getSoupNames, NewState: gettingSoupNames},
	{State: selectingStore, Event: protocol.OPERATION_CANCELED, Action: cancel, NewState: idle},
	{State: selectingStore, Fallback: true, NewState: selectingStore},
	{State: gettingSoupNames, Event: protocol.SOUP_NAMES, Action: showSoupNames, NewState: gettingSoupNames},
	{State: gettingSoupNames, Event: protocol.OPERATION_CANCELED, Action: cancel, NewState: idle},
	{State: gettingSoupNames, Fallback: true, NewState: gettingSoupNames},
	{State: gettingAppList, Event: protocol.APP_NAMES, Action: showAppList, NewState: idle},
	{State: gettingAppList, Event: protocol.OPERATION_CANCELED, Action: cancel, NewState: idle},
	{State: gettingAppList, Fallback: true, NewState: gettingAppList},
}

var (
	state = idle
)

func processIn(event *protocol.DockEvent) {
	var action int
	action, state = fsm.Input(event.Command, state, transitions)
	switch action {
	case getStoreNames:
		protocol.Events <- protocol.NewDockEvent(
			protocol.GET_STORE_NAMES,
			protocol.Out,
			[]byte{},
		)
	case selectStore:
		var eventData nsof.Data = event.Data
		stores := eventData.Factory()
		var data nsof.Data = []byte{2}
		stores[0].WriteNSOF(&data)
		protocol.Events <- protocol.NewDockEvent(
			protocol.SET_CURRENT_STORE,
			protocol.Out,
			data,
		)
	case getSoupNames:
		protocol.Events <- protocol.NewDockEvent(
			protocol.GET_SOUP_NAMES,
			protocol.Out,
			[]byte{},
		)
	case showSoupNames:
		var eventData nsof.Data = event.Data
		soups := eventData.Factory()
		log.Println(soups)
		protocol.Events <- protocol.NewDockEvent(
			protocol.GET_APP_NAMES,
			protocol.Out,
			[]byte{0, 0, 0, 0},
		)
		state = gettingAppList
	case showAppList:
		var eventData nsof.Data = event.Data
		apps := eventData.Factory()
		log.Println(apps)
		protocol.Events <- protocol.NewDockEvent(
			protocol.OPERATION_DONE,
			protocol.Out,
			[]byte{},
		)
	case cancel:
		protocol.Events <- protocol.NewDockEvent(
			protocol.OP_CANCELED_ACK,
			protocol.Out,
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
