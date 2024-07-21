package protocol

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
)

const (
	In byte = iota
	Out
)

type Event interface{}

const (
	APP_DATA                     = 0x00000001
	APP_QUIT                     = 0x00000002
	APP_DISCONNECT               = 0x00000003
	APP_GET_DEFAULT_STORE        = 0x00000004
	APP_GET_STORE_NAMES          = 0x00000005
	APP_QUERY_SOUP               = 0x00000006
	APP_CURSOR_FREE              = 0x00000007
	APP_CURSOR_NEXT              = 0x00000008
	APP_CURSOR_ENTRY             = 0x00000009
	APP_SET_CURRENT_STORE        = 0x0000000a
	APP_SET_CURRENT_SOUP         = 0x0000000b
	APP_GET_SOUP_NAMES           = 0x0000000c
	APP_DELETE_ENTRIES           = 0x0000000d
	APP_ADD_ENTRY                = 0x0000000e
	APP_SEND_SOUP                = 0x0000000f
	APP_GET_INFO                 = 0x00000010
	APP_CONNECTED                = 0x00000011
	LAST_APP_COMMAND             = 0x32323232
	NEWT                         = 0x6e657774
	DOCK                         = 0x646f636b
	LONGDATA                     = 0x6c647461
	REF_RESULT                   = 0x72656620
	QUERY                        = 0x71757279
	CURSOR_GOTO_KEY              = 0x676f746f
	CURSOR_MAP                   = 0x636d6170
	CURSOR_ENTRY                 = 0x63727372
	CURSOR_MOVE                  = 0x6d6f7665
	CURSOR_NEXT                  = 0x6e657874
	CURSOR_PREV                  = 0x70726576
	CURSOR_RESET                 = 0x72736574
	CURSOR_RESET_TO_END          = 0x72656e64
	CURSOR_COUNT_ENTRIES         = 0x636e7420
	CURSOR_WHICH_END             = 0x77686368
	CURSOR_FREE                  = 0x63667265
	KEYBOARD_CHAR                = 0x6b626463
	DESKTOP_INFO                 = 0x64696e66
	KEYBOARD_STRING              = 0x6b626473
	START_KEYBOARD_PASSTHROUGH   = 0x6b796264
	DEFAULT_STORE                = 0x64667374
	APP_NAMES                    = 0x6170706e
	IMPORT_PARAMETER_SLIP_RESULT = 0x69736c72
	PACKAGE_INFO                 = 0x70696e66
	SET_BASE_ID                  = 0x62617365
	BACKUP_IDS                   = 0x62696473
	BACKUP_SOUP_DONE             = 0x6273646e
	SOUP_NOT_DIRTY               = 0x6e646972
	SYNCHRONIZE                  = 0x73796e63
	CALL_RESULT                  = 0x63726573
	REMOVE_PACKAGE               = 0x726d7670
	RESULT_STRING                = 0x72657373
	SOURCE_VERSION               = 0x73766572
	ADD_ENTRY_WITH_UNIQUE_ID     = 0x61756e69
	GET_PACKAGE_INFO             = 0x6770696e
	GET_DEFAULT_STORE            = 0x67646673
	CREATE_DEFAULT_SOUP          = 0x63647370
	GET_APP_NAMES                = 0x67617070
	REG_PROTOCOL_EXTENSION       = 0x70657874
	REMOVE_PROTOCOL_EXTENSION    = 0x72706578
	SET_STORE_SIGNATURE          = 0x73736967
	SET_SOUP_SIGNATURE           = 0x73736f73
	IMPORT_PARAMETERS_SLIP       = 0x69736c70
	GET_PASSWORD                 = 0x67707764
	SEND_SOUP                    = 0x736e6473
	BACKUP_SOUP                  = 0x626b7370
	SET_STORE_NAME               = 0x73736e61
	CALL_GLOBAL_FUNCTION         = 0x6367666e
	CALL_ROOT_METHOD             = 0x63726d64
	SET_VBO_COMPRESSION          = 0x6376626f
	RESTORE_PATCH                = 0x72706174
	OPERATION_DONE               = 0x6f70646e
	OPERATION_CANCELED           = 0x6f706361
	OP_CANCELED_ACK              = 0x6f636161
	REF_TEST                     = 0x72747374
	UNKNOWN_COMMAND              = 0x756e6b6e
	PASSWORD                     = 0x70617373
	NEWTON_NAME                  = 0x6e616d65
	NEWTON_INFO                  = 0x6e696e66
	INITIATE_DOCKING             = 0x646f636b
	WHICH_ICONS                  = 0x7769636e
	REQUEST_TO_SYNC              = 0x7373796e
	SYNC_OPTIONS                 = 0x736f7074
	GET_SYNC_OPTIONS             = 0x6773796e
	SYNC_RESULTS                 = 0x73726573
	SET_STORE_GET_NAMES          = 0x7373676e
	SET_SOUP_GET_INFO            = 0x73736769
	GET_CHANGED_INDEX            = 0x63696478
	GET_CHANGED_INFO             = 0x63696e66
	REQUEST_TO_BROWSE            = 0x72746272
	GET_DEVICES                  = 0x67646576
	GET_DEFAULT_PATH             = 0x64707468
	GET_FILES_AND_FOLDERS        = 0x6766696c
	SET_PATH                     = 0x73707468
	GET_FILE_INFO                = 0x6766696e
	INTERNAL_STORE               = 0x6973746f
	RESOLVE_ALIAS                = 0x72616c69
	GET_FILTERS                  = 0x67666c74
	SET_FILTER                   = 0x73666c74
	SET_DRIVE                    = 0x73647276
	DEVICES                      = 0x64657673
	FILTERS                      = 0x66696c74
	PATH                         = 0x70617468
	FILES_AND_FOLDERS            = 0x66696c65
	FILE_INFO                    = 0x66696e66
	GET_INTERNAL_STORE           = 0x67697374
	ALIAS_RESOLVED               = 0x616c6972
	IMPORT_FILE                  = 0x696d7074
	SET_TRANSLATOR               = 0x7472616e
	TRANSLATOR_LIST              = 0x74726e6c
	IMPORTING                    = 0x64696d70
	SOUPS_CHANGED                = 0x73636867
	SET_STORE_TO_DEFAULT         = 0x73646566
	LOAD_PACKAGE_FILE            = 0x6c70666c
	RESTORE_FILE                 = 0x7273666c
	GET_RESTORE_OPTIONS          = 0x67726f70
	RESTORE_ALL                  = 0x72616c6c
	RESTORE_OPTIONS              = 0x726f7074
	RESTORE_PACKAGE              = 0x72706b67
	REQUEST_TO_RESTORE           = 0x72727374
	REQUEST_TO_INSTALL           = 0x72696e73
	REQUEST_TO_DOCK              = 0x7274646b
	CURRENT_TIME                 = 0x74696d65
	STORE_NAMES                  = 0x73746f72
	SOUP_NAMES                   = 0x736f7570
	SOUP_IDS                     = 0x73696473
	CHANGED_IDS                  = 0x63696473
	RESULT                       = 0x64726573
	ADDED_ID                     = 0x61646964
	ENTRY                        = 0x656e7472
	PACKAGE_ID_LIST              = 0x70696473
	PACKAGE                      = 0x61706b67
	INDEX_DESCRIPTION            = 0x696e6478
	INHERITANCE                  = 0x64696e68
	PATCHES                      = 0x70617463
	LAST_SYNC_TIME               = 0x73746d65
	GET_STORE_NAMES              = 0x6773746f
	GET_SOUP_NAMES               = 0x67657473
	SET_CURRENT_STORE            = 0x7373746f
	SET_CURRENT_SOUP             = 0x73736f75
	GET_SOUP_IDS                 = 0x67696473
	DELETE_ENTRIES               = 0x64656c65
	ADD_ENTRY                    = 0x61646465
	RETURN_ENTRY                 = 0x72657465
	RETURN_CHANGED_ENTRY         = 0x7263656e
	EMPTY_SOUP                   = 0x65736f75
	DELETE_SOUP                  = 0x64736f75
	LOAD_PACKAGE                 = 0x6c706b67
	GET_PACKAGE_IDS              = 0x67706964
	BACKUP_PACKAGES              = 0x62706b67
	DISCONNECT                   = 0x64697363
	DELETE_ALL_PACKAGES          = 0x64706b67
	GET_INDEX_DESCRIPTION        = 0x67696e64
	CREATE_SOUP                  = 0x63736f70
	GET_INHERITANCE              = 0x67696e68
	SET_TIMEOUT                  = 0x7374696d
	GET_PATCHES                  = 0x67706174
	DELETE_PKG_DIR               = 0x64706b64
	GET_SOUP_INFO                = 0x6773696e
	CHANGED_ENTRY                = 0x63656e74
	TEST                         = 0x74657374
	HELLO                        = 0x68656c6f
	SOUP_INFO                    = 0x73696e66
)

type SerialEvent struct {
	Direction byte
	Data      []byte
}

type MnpEvent struct {
	Direction byte
	Data      []byte
}

type DockEvent struct {
	Direction byte
	Data      []byte
	Command   uint32
	Length    uint32
}

var Events = make(chan Event, 100)

func (event SerialEvent) String() string {
	return fmt.Sprintf("Serial (%d):\n%s", event.Direction, hex.Dump(event.Data))
}

func (event MnpEvent) String() string {
	return fmt.Sprintf("MNP (%d):\n%s", event.Direction, hex.Dump(event.Data))
}

func (event DockEvent) String() string {
	return fmt.Sprintf("Dock (%d): %08x\n%s", event.Direction, event.Command, hex.Dump(event.Data))
}

func NewDockEvent(cmd uint32, direction byte, data []byte) *DockEvent {
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, uint32(NEWT))
	binary.Write(&buf, binary.BigEndian, uint32(DOCK))
	binary.Write(&buf, binary.BigEndian, cmd)
	binary.Write(&buf, binary.BigEndian, uint32(len(data)))
	buf.Write(data)
	if len(data)%4 != 0 {
		pad := 4 - len(data)%4
		buf.Write(make([]byte, pad, pad))
	}
	return &DockEvent{
		Direction: direction,
		Data:      buf.Bytes(),
	}
}

func IsQuitEvent(event Event) bool {
	switch event.(type) {
	case *DockEvent:
		return event.(*DockEvent).Command == APP_QUIT
	}
	return false
}
