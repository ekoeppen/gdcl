package protocol

import (
	"encoding/hex"
	"github.com/ekoeppen/gdcl/v2/crc16"
	"github.com/ekoeppen/gdcl/v2/fsm"
	"github.com/tarm/goserial"
	"io"
	"log"
)

type Packet struct {
	packetType    byte
	headerLength  uint16
	data          []byte
	receivedCRC   uint16
	calculatedCRC uint16
}

type Range struct {
	from byte
	to   byte
}

type Value struct {
	value byte
}

const (
	OUTSIDE_PACKET = iota
	START_SYN
	START_DLE
	INSIDE_PACKET
	DLE_IN_PACKET
	END_ETX
	END_CRC1
	PACKET_END
)

const (
	SYN byte = 22
	DLE byte = 16
	STX byte = 2
	ETX byte = 3
)

type MNPPacketLayer struct {
	state          int
	packet         Packet
	serial         io.ReadWriteCloser
	stateTable     map[int][]fsm.State
	FromConnection chan []byte
	ToConnection   chan Packet
	logging        bool
}

var packetLayer MNPPacketLayer

func (value Value) Matches(input interface{}) bool {
	return input.(byte) == value.value
}

func (inputRange Range) Matches(input interface{}) bool {
	return inputRange.from <= input.(byte) && input.(byte) <= inputRange.to
}

func startPacket(state int, input interface{}, output interface{}, data interface{}) {
	layer := data.(*MNPPacketLayer)
	layer.packet = Packet{data: make([]byte, 0, 128), receivedCRC: 0, calculatedCRC: 0}
}

func addDLE(state int, input interface{}, output interface{}, data interface{}) {
	layer := data.(*MNPPacketLayer)
	layer.packet.data = append(layer.packet.data, input.(byte))
}

func resetReceivedCRC(state int, input interface{}, output interface{}, data interface{}) {
	layer := data.(*MNPPacketLayer)
	layer.packet.receivedCRC = uint16(input.(byte))
}

func updateCalculatedCRC(state int, input interface{}, output interface{}, data interface{}) {
	layer := data.(*MNPPacketLayer)
	layer.packet.calculatedCRC = crc16.Crc16(input.(byte), layer.packet.calculatedCRC)
}

func addChar(state int, input interface{}, output interface{}, data interface{}) {
	layer := data.(*MNPPacketLayer)
	layer.packet.data = append(layer.packet.data, input.(byte))
	layer.packet.calculatedCRC = crc16.Crc16(input.(byte), layer.packet.calculatedCRC)
}

func packetReceived(state int, input interface{}, output interface{}, data interface{}) {
	layer := data.(*MNPPacketLayer)
	layer.packet.receivedCRC = layer.packet.receivedCRC + uint16(input.(byte))<<8
	if layer.packet.data[0] == 255 {
		layer.packet.headerLength = uint16(layer.packet.data[1])*256 + uint16(layer.packet.data[2])
		layer.packet.packetType = layer.packet.data[3]
		layer.packet.data = layer.packet.data[4:]
	} else {
		layer.packet.headerLength = uint16(layer.packet.data[0])
		layer.packet.packetType = layer.packet.data[1]
		layer.packet.data = layer.packet.data[2:]
	}
	layer.ToConnection <- layer.packet
}

func (layer *MNPPacketLayer) transition(input byte) {
	layer.state = fsm.Transition(layer.stateTable, layer.state, input, nil, layer)
}

func (layer *MNPPacketLayer) reader() {
	go func() {
		logBuf := make([]byte, 0, 65536)
		buf := make([]byte, 256)
		for {
			n, err := layer.serial.Read(buf)
			if err != nil {
				log.Fatal(err)
				break
			}
			for i := 0; i < n; i++ {
				logBuf = append(logBuf, buf[i])
				layer.transition(buf[i])
				if layer.state == PACKET_END {
					if layer.logging {
						log.Printf("\033[31m>>>\n%s\033[0m", hex.Dump(logBuf))
					}
					logBuf = logBuf[:0]
				}
			}
			buf = buf[:1]
		}
	}()
}

func (layer *MNPPacketLayer) writer() {
	go func() {
		for {
			buf := <-layer.FromConnection
			outBuf := make([]byte, 0, len(buf)*2+7)
			crc := uint16(0)
			outBuf = append(outBuf, SYN, DLE, STX)
			for i := 0; i < len(buf); i++ {
				outBuf = append(outBuf, buf[i])
				crc = crc16.Crc16(buf[i], crc)
				if buf[i] == DLE {
					outBuf = append(outBuf, DLE)
				}
			}
			outBuf = append(outBuf, DLE, ETX)
			crc = crc16.Crc16(ETX, crc)
			outBuf = append(outBuf, byte(crc&0xff), byte(crc>>8))
			if layer.logging {
				log.Printf("\033[32m<<<\n%s\033[0m", hex.Dump(outBuf))
			}
			layer.serial.Write(outBuf)
		}
	}()
}

func MNPPacketLayerNew(name string, speed int) *MNPPacketLayer {
	var err error
	c := &serial.Config{Name: name, Baud: speed}
	packetLayer.serial, err = serial.OpenPort(c)
	if err != nil {
		log.Fatal(err)
	}
	packetLayer.ToConnection = make(chan Packet)
	packetLayer.FromConnection = make(chan []byte)
	packetLayer.stateTable = map[int][]fsm.State{
		OUTSIDE_PACKET: {{Input: Value{SYN}, NewState: START_SYN}},
		START_SYN: {
			{Input: Value{DLE}, NewState: START_DLE},
			{Input: Range{0, 255}, NewState: OUTSIDE_PACKET},
		},
		START_DLE: {
			{Input: Value{STX}, NewState: INSIDE_PACKET, Action: startPacket},
			{Input: Range{0, 255}, NewState: OUTSIDE_PACKET},
		},
		INSIDE_PACKET: {
			{Input: Value{DLE}, NewState: DLE_IN_PACKET},
			{Input: Range{0, 255}, NewState: INSIDE_PACKET, Action: addChar},
		},
		DLE_IN_PACKET: {
			{Input: Value{DLE}, NewState: INSIDE_PACKET, Action: addDLE},
			{Input: Value{ETX}, NewState: END_ETX, Action: updateCalculatedCRC},
			{Input: Range{0, 255}, NewState: OUTSIDE_PACKET},
		},
		END_ETX:  {{Input: Range{0, 255}, NewState: END_CRC1, Action: resetReceivedCRC}},
		END_CRC1: {{Input: Range{0, 255}, NewState: PACKET_END, Action: packetReceived}},
		PACKET_END: {
			{Input: Value{SYN}, NewState: START_SYN},
			{Input: Range{0, 255}, NewState: OUTSIDE_PACKET},
		},
	}
	packetLayer.reader()
	packetLayer.writer()
	packetLayer.logging = false
	return &packetLayer
}
